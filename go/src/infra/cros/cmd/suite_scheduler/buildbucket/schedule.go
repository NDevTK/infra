// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package buildbucket implements the interface required to schedule builder
// requests on the LUCI BuildBucket architecture.
package buildbucket

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	requestpb "go.chromium.org/chromiumos/infra/proto/go/test_platform"
	ctppb "go.chromium.org/chromiumos/infra/proto/go/test_platform/cros_test_platform"
	"go.chromium.org/luci/auth/client/authcli"
	bb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmdsupport/cmdlib"
	"infra/cros/cmd/suite_scheduler/metrics"
)

// buildBucketHost is the URL host for the BuildBucket API.
const buildBucketHost = "cr-buildbucket.appspot.com"

var (
	ctpBuilderIDProd = bb.BuilderID{
		Project: "chromeos",
		Bucket:  "testplatform",
		Builder: "cros_test_platform",
	}
	ctpBuilderIDStaging = bb.BuilderID{
		Project: "chromeos",
		Bucket:  "testplatform",
		Builder: "cros_test_platform-staging",
	}
)

// Scheduler interface type describes the BB API functionality connection.
type Scheduler interface {
	Schedule(ctpRequest *requestpb.Request) (*bb.Build, error)
}

// client implements the Scheduler interface.
type client struct {
	ctx               context.Context
	buildBucketClient bb.BuildsClient
	isProd            bool
	dryRun            bool
}

// InitScheduler returns an operable Scheduler interface.
func InitScheduler(ctx context.Context, authOpts *authcli.Flags, isProd, dryRun bool) (Scheduler, error) {
	// Build the underlying HTTP client with the proper Auth Scoping.
	httpClient, err := cmdlib.NewHTTPClient(ctx, authOpts)
	if err != nil {
		return nil, err
	}

	// Form the custom prpc client that LUCI requires.
	prpcClient := &prpc.Client{
		C:    httpClient,
		Host: buildBucketHost,
	}

	return &client{
		ctx:               ctx,
		buildBucketClient: bb.NewBuildsClient(prpcClient),
		isProd:            isProd,
		dryRun:            dryRun,
	}, nil
}

// ctpToBBRequest transforms the CTP request into a BuildBucket serviceable
// scheduling request.
func ctpToBBRequest(ctpRequest *requestpb.Request, isProd, dryRun bool) (*bb.ScheduleBuildRequest, error) {
	// Tag the entry in the requests map with the name of the suite.
	// NOTE: requests is used as opposed to request because I have seen no
	// evidence to indicate that the singular field is actively supported.
	suiteNameKey := ctpRequest.TestPlan.Suite[0].Name

	// Create a struct of the CTP recipes properties with the single request
	// entry being passed in.
	recipeProto := &ctppb.CrosTestPlatformProperties{
		Requests: map[string]*requestpb.Request{
			suiteNameKey: ctpRequest,
		},
	}

	// Transform the properties proto into a json string.
	msgJson, err := protojson.Marshal(recipeProto)
	if err != nil {
		return nil, err
	}

	// Now that we have the raw json unmarshal, transform the text into the
	// "generic" proto struct. This "generic" proto struct type is required by
	// the BuildBucket API.
	// NOTE: The default unmarshall-er from the protojson package does not throw
	// errors on unknown JSON fields. This is required because the generic
	// struct and CTP struct do not share any field names. If a different
	// unmarshall-er is chosen down the line, ensure that this functionality is
	// maintained.
	properties := &structpb.Struct{}
	err = protojson.Unmarshal(msgJson, properties)
	if err != nil {
		return nil, err
	}

	// Based on the isProd flag choose the corresponding builder identification.
	builder := &ctpBuilderIDStaging
	if isProd == true {
		builder = &ctpBuilderIDProd
	}

	// Get the chromeOS image for display in the BB tags
	image := ""
	for _, dep := range ctpRequest.Params.SoftwareDependencies {
		if dep.GetChromeosBuild() != "" {
			image = dep.GetChromeosBuild()
			break
		}
	}
	if image == "" {
		return nil, fmt.Errorf("no ChromeOS build found")
	}

	schedulerRequest := &bb.ScheduleBuildRequest{
		Builder:    builder,
		Properties: properties,
		DryRun:     dryRun,
		Tags: []*bb.StringPair{
			{
				Key:   "susch-run",
				Value: metrics.GetRunID().Id,
			},
			{
				Key:   "suite",
				Value: suiteNameKey,
			},
			{
				Key:   "user_agent",
				Value: "suite_schedulerV1.5",
			},
			{
				Key:   "label-image",
				Value: image,
			},
		},
	}

	return schedulerRequest, nil
}

// Schedule takes in a CTP request and schedules it via the BuildBucket API.
func (c *client) Schedule(ctpRequest *requestpb.Request) (*bb.Build, error) {
	schedulerRequest, err := ctpToBBRequest(ctpRequest, c.isProd, c.dryRun)
	if err != nil {
		return nil, err
	}

	// TODO(b/317084435): Handle any case where the returned status of the build is
	// a terminal "failure" status code. E.g. Failed or unspecified.
	build, err := c.buildBucketClient.ScheduleBuild(c.ctx, schedulerRequest)
	if err != nil {
		return nil, err
	}

	// TODO(b/309683890): Create and log schedule event to metrics.
	return build, nil
}

// ctpListToBatchRequest transforms the provided CTP requests into BuildBucket
// batch request with a max size of 200.
func ctpListToBatchRequest(ctpRequests []*requestpb.Request, isProd, dryRun bool) ([]*bb.BatchRequest, error) {
	batchRequests := []*bb.BatchRequest{}

	currentRequest := &bb.BatchRequest{}
	for _, ctpRequest := range ctpRequests {
		// NOTE: The BuildBucket API allows for a maximum of 200 requests per
		// batch so to adhere to this rule we need to implement the following
		// check.
		if len(currentRequest.Requests) == 200 {
			batchRequests = append(batchRequests, currentRequest)
			currentRequest = &bb.BatchRequest{}
		}

		schedulerRequest, err := ctpToBBRequest(ctpRequest, isProd, dryRun)
		if err != nil {
			return nil, err
		}

		// NOTE: The batch client is indifferent to the type of request so to
		// allow for the semi-generic type we must nest the proto as seen.
		batchRequest := &bb.BatchRequest_Request{
			Request: &bb.BatchRequest_Request_ScheduleBuild{
				ScheduleBuild: schedulerRequest,
			},
		}

		currentRequest.Requests = append(currentRequest.Requests, batchRequest)
	}
	// The final currentRequest will never be added to the BatchRequest list
	// during the loop. Add it in manually here.
	batchRequests = append(batchRequests, currentRequest)

	return batchRequests, nil
}

// BatchSchedule takes in a list of requests and schedules them via the
// BuildBucket batch API.
func (c *client) BatchSchedule(ctpRequests []*requestpb.Request) ([]*bb.BatchResponse, error) {
	// Noop, exit early to avoid scheduling an empty batch.
	if len(ctpRequests) == 0 {
		// Probably do some sort of logging here saying that there was no
		// requests passed in so the operation was a noop.
		return nil, nil
	}

	batchRequests, err := ctpListToBatchRequest(ctpRequests, c.isProd, c.dryRun)
	if err != nil {
		return nil, err
	}

	scheduleResponses := []*bb.BatchResponse{}
	for _, batch := range batchRequests {
		// TODO(b/317084435): Handle the response from the schedule build function.
		batchResponse, err := c.buildBucketClient.Batch(c.ctx, batch)
		if err != nil {
			return nil, err
		}

		scheduleResponses = append(scheduleResponses, batchResponse)

	}
	// TODO(b/309683890): Create and log schedule events to metrics.
	return scheduleResponses, nil
}
