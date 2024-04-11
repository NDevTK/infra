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
	"infra/cros/cmd/kron/builds"
	"infra/cros/cmd/kron/metrics"
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
		Builder: "cros_test_platform-dev",
	}
)

// Scheduler interface type describes the BB API functionality connection.
type Scheduler interface {
	Schedule(requests []*builds.EventWrapper, configName string) (*bb.Build, error)
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

// mergeRequests merge all CTP requests into one CTP recipe input properties object.
func mergeRequests(requests []*builds.EventWrapper) *ctppb.CrosTestPlatformProperties {
	ctpRequestInputProps := &ctppb.CrosTestPlatformProperties{
		Requests: map[string]*requestpb.Request{},
	}

	// Add all CTP Test Requests to the input properties struct mapped by their
	// unique request metadata.
	for _, request := range requests {
		key := fmt.Sprintf("%s.%s.%s", request.CtpRequest.Params.SoftwareAttributes.BuildTarget.Name, request.Event.ConfigName, request.Event.SuiteName)
		if _, ok := ctpRequestInputProps.Requests[key]; ok {
			// If the key is duplicated for some reason then add the eventUuid
			// to differentiate.
			key = fmt.Sprintf("%s.%s", key, request.Event.EventUuid)

		}
		ctpRequestInputProps.Requests[key] = request.CtpRequest
	}

	return ctpRequestInputProps
}

// generateBBRequest creates a BuildBucket Request proto with proper metadata in
// the tags.
func generateBBRequest(suiteName, configName string, dryRun bool, builder *bb.BuilderID, properties *structpb.Struct, requests []*builds.EventWrapper) (*bb.ScheduleBuildRequest, error) {
	// Generate the BuildBucket request.
	schedulerRequest := &bb.ScheduleBuildRequest{
		Builder:    builder,
		Properties: properties,
		DryRun:     dryRun,
		// These tags will appear on the Milo UI and will help us search for
		// builds in plx.
		Tags: []*bb.StringPair{
			{
				Key:   "kron-run",
				Value: metrics.GetRunID(),
			},
			{
				Key:   "suite",
				Value: suiteName,
			},
			{
				Key:   "label-suite",
				Value: suiteName,
			},
			{
				Key:   "user_agent",
				Value: "kron",
			},
			{
				Key:   "suite-scheduler-config",
				Value: configName,
			},
		},
	}

	// Add all image, buildUuid, and eventUuid fields per test request.
	for _, request := range requests {
		image := ""
		for _, dep := range request.CtpRequest.Params.SoftwareDependencies {

			// The SoftwareDependencies proto type includes many types of deps,
			// so search for one which can provide the image value.
			if dep.GetChromeosBuild() != "" {
				image = dep.GetChromeosBuild()
				break
			}
		}

		// A CTP request cannot function with a nil image value so throw an
		// error here.
		if image == "" {
			return nil, fmt.Errorf("no ChromeOS build found")
		}

		schedulerRequest.Tags = append(schedulerRequest.Tags,
			&bb.StringPair{
				Key:   "build-id",
				Value: request.Event.BuildUuid,
			})
		schedulerRequest.Tags = append(schedulerRequest.Tags,
			&bb.StringPair{
				Key:   "event-id",
				Value: request.Event.EventUuid,
			})
		schedulerRequest.Tags = append(schedulerRequest.Tags,
			&bb.StringPair{
				Key:   "label-image",
				Value: image,
			})
	}

	return schedulerRequest, nil

}

// ctpToBBRequest transforms the CTP request into a BuildBucket serviceable
// scheduling request.
func ctpToBBRequest(requests []*builds.EventWrapper, isProd, dryRun bool, configName string) (*bb.ScheduleBuildRequest, error) {
	// Create a struct of the CTP recipes properties with the single request
	// entry being passed in.
	ctpRequestInputProps := mergeRequests(requests)

	// Transform the properties proto into a json string.
	msgJSON, err := protojson.Marshal(ctpRequestInputProps)
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
	err = protojson.Unmarshal(msgJSON, properties)
	if err != nil {
		return nil, err
	}

	// Based on the isProd flag choose the corresponding builder identification.
	builder := &ctpBuilderIDStaging
	if isProd {
		builder = &ctpBuilderIDProd
	}

	// Since we combine requests of the same SuSch config, they will all share
	// the same suite value. Pull the first one for simplicity.
	suiteName := requests[0].CtpRequest.TestPlan.Suite[0].Name

	// Generate the BuildBucket request.
	schedulerRequest, err := generateBBRequest(suiteName, configName, dryRun, builder, properties, requests)
	if err != nil {
		return nil, err
	}

	return schedulerRequest, nil
}

// Schedule takes in a batch of EventWrappers and schedules it via the BuildBucket API.
func (c *client) Schedule(requests []*builds.EventWrapper, configName string) (*bb.Build, error) {
	schedulerRequest, err := ctpToBBRequest(requests, c.isProd, c.dryRun, configName)
	if err != nil {
		return nil, err
	}

	build, err := c.buildBucketClient.ScheduleBuild(c.ctx, schedulerRequest)
	if err != nil {
		return nil, err
	}

	return build, nil
}
