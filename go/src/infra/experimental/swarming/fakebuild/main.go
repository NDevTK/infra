// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Binary fakebuild is a luciexe binary that pretends to do some work.
//
// To be used for Swarming and Buildbucket load testing.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/metadata"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/buildbucket"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/buildbucket/protoutil"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/data/rand/mathrand"
	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/build"

	"infra/experimental/swarming/fakebuild/fakebuildpb"
)

func main() {
	mathrand.SeedRandomly()

	inputs := &fakebuildpb.Inputs{}

	build.Main(inputs, nil, nil, func(ctx context.Context, args []string, st *build.State) error {
		for i := 0; i < int(inputs.Steps); i++ {
			sleepStep(ctx, inputs, i)
		}
		return scheduleChildBuilds(ctx, inputs)
	})
}

func randomSecs(min, max int64) int64 {
	var secs int64
	if dt := max - min; dt > 0 {
		secs = min + rand.Int63n(dt)
	} else {
		secs = min
	}
	return secs
}

func sleepStep(ctx context.Context, inputs *fakebuildpb.Inputs, idx int) {
	secs := randomSecs(inputs.SleepMinSec, inputs.SleepMaxSec)

	step, ctx := build.StartStep(ctx, fmt.Sprintf("Step %d: sleep %d", idx+1, secs))
	defer func() { step.End(nil) }()

	clock.Sleep(ctx, time.Duration(secs)*time.Second)
}

func generateTags() []*bbpb.StringPair {
	tags := strpair.Map{}
	clNum := rand.Int63n(5000000)
	for i := 0; i < 4; i++ {
		tags.Add("buildset", fmt.Sprintf("patch/gerrit/chromium-review.googlesource.com/%d/%d", clNum, i))
	}
	return protoutil.StringPairs(tags)
}

func generateRequest(builder *bbpb.BuilderID, batchSize int) *bbpb.BatchRequest {
	req := &bbpb.BatchRequest{
		Requests: []*bbpb.BatchRequest_Request{},
	}
	for i := 0; i < batchSize; i++ {
		req.Requests = append(req.Requests, &bbpb.BatchRequest_Request{
			Request: &bbpb.BatchRequest_Request_ScheduleBuild{
				ScheduleBuild: &bbpb.ScheduleBuildRequest{
					Builder: builder,
					Tags:    generateTags(), // generate a list of tags which need to be indexed.
				},
			},
		})
	}
	return req
}

func scheduleOneBatch(ctx context.Context, bbClient bbpb.BuildsClient, idx, batchSize int, cbs *fakebuildpb.ChildBuilds) error {
	step, ctx := build.StartStep(ctx, fmt.Sprintf("schedule children (%d)", idx))
	defer func() { step.End(nil) }()

	req := generateRequest(cbs.Builder, batchSize)
	res, err := bbClient.Batch(ctx, req)
	if err != nil {
		return errors.Annotate(err, "batch %d", idx).Err()
	}
	logging.Debugf(ctx, "response: %s", proto.MarshalTextString(res))

	secs := randomSecs(cbs.SleepMinSec, cbs.SleepMaxSec)
	clock.Sleep(ctx, time.Duration(secs)*time.Second)
	return nil
}

func scheduleChildBuilds(ctx context.Context, inputs *fakebuildpb.Inputs) error {
	cbs := inputs.GetChildBuilds()
	if cbs == nil {
		return nil
	}

	numBatch := 1
	if cbs.BatchSize > 0 && cbs.BatchSize < cbs.Children {
		numBatch = int(cbs.Children / cbs.BatchSize)
		if cbs.Children%cbs.BatchSize > 0 {
			numBatch += 1
		}
	}

	bbCtx := lucictx.GetBuildbucket(ctx)
	if bbCtx != nil {
		if bbCtx.GetScheduleBuildToken() != "" {
			scheduleBuildToken := bbCtx.ScheduleBuildToken
			if scheduleBuildToken != "" && scheduleBuildToken != buildbucket.DummyBuildbucketToken {
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(buildbucket.BuildbucketTokenHeader, scheduleBuildToken))
			}
		}
	}

	httpClient, err := auth.NewAuthenticator(ctx, auth.SilentLogin, chromeinfra.DefaultAuthOptions()).Client()
	switch {
	case err == auth.ErrLoginRequired:
		return errors.New("Login required: run `bb auth-login`")
	case err != nil:
		return err
	}
	bbClient := bbpb.NewBuildsPRPCClient(&prpc.Client{
		C:       httpClient,
		Host:    "cr-buildbucket-dev.appspot.com",
		Options: prpc.DefaultOptions(),
	})

	step, ctx := build.StartStep(ctx, "schedule children")
	defer func() { step.End(nil) }()

	for i := 0; i < numBatch; i++ {
		var batchSize int64
		switch {
		case cbs.BatchSize == 0:
			batchSize = cbs.Children
		case i >= 1 && i == numBatch-1:
			// last one of multiple batches.
			batchSize = cbs.Children - cbs.BatchSize*int64(i)
		default:
			batchSize = cbs.BatchSize
		}
		if err := scheduleOneBatch(ctx, bbClient, i, int(batchSize), cbs); err != nil {
			return err
		}
	}
	return nil
}
