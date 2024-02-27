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
		Builder: "cros_test_platform-staging",
	}
)

// Scheduler interface type describes the BB API functionality connection.
type Scheduler interface {
	Schedule(ctpRequest *requestpb.Request, buildID, eventID, configName string) (*bb.Build, error)
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
func ctpToBBRequest(ctpRequest *requestpb.Request, isProd, dryRun bool, buildID, eventID, configName string) (*bb.ScheduleBuildRequest, error) {
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
	msgJSON, err := protojson.Marshal(recipeProto)
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
		// These tags will appear on the Milo UI and will help us search for
		// builds in plx.
		Tags: []*bb.StringPair{
			{
				Key:   "kron-run",
				Value: metrics.GetRunID().Id,
			},
			{
				Key:   "build-id",
				Value: buildID,
			},
			{
				Key:   "event-id",
				Value: eventID,
			},
			{
				Key:   "suite",
				Value: suiteNameKey,
			},
			{
				Key:   "user_agent",
				Value: "kron",
			},
			{
				Key:   "label-image",
				Value: image,
			},
			{
				Key:   "suite-scheduler-config",
				Value: configName,
			},
		},
	}

	return schedulerRequest, nil
}

// Schedule takes in a CTP request and schedules it via the BuildBucket API.
func (c *client) Schedule(ctpRequest *requestpb.Request, buildID, eventID, configName string) (*bb.Build, error) {
	schedulerRequest, err := ctpToBBRequest(ctpRequest, c.isProd, c.dryRun, buildID, eventID, configName)
	if err != nil {
		return nil, err
	}

	// TODO(b/317084435): Handle any case where the returned status of the build is
	// a terminal "failure" status code. E.g. Failed or unspecified.
	build, err := c.buildBucketClient.ScheduleBuild(c.ctx, schedulerRequest)
	if err != nil {
		return nil, err
	}

	return build, nil
}
