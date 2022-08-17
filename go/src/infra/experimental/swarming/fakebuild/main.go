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

	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/buildbucket"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/buildbucket/protoutil"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/data/rand/mathrand"
	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/common/errors"
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

		if err := scheduleChildBuilds(ctx, bbClient, inputs); err != nil {
			return err
		}
		return searchBuilds(ctx, bbClient, inputs)
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

func gerritChange(change, patchset int64) *bbpb.GerritChange {
	return &bbpb.GerritChange{
		Host:     "chromium-review.googlesource.com",
		Project:  "chromium/src",
		Change:   change,
		Patchset: patchset,
	}
}

func generateGerritChangesAndTags(req *bbpb.ScheduleBuildRequest) {
	change := rand.Int63n(5000000)

	var changes []*bbpb.GerritChange
	tags := strpair.Map{}
	for i := 1; i <= 4; i++ {
		changes = append(changes, gerritChange(change, int64(i)))
		tags.Add("buildset", fmt.Sprintf("patch/gerrit/chromium-review.googlesource.com/%d/%d", change, i))
	}
	req.GerritChanges = changes
	req.Tags = protoutil.StringPairs(tags)
}

func generateScheduleRequest(builder *bbpb.BuilderID, batchSize int) *bbpb.BatchRequest {
	req := &bbpb.BatchRequest{
		Requests: []*bbpb.BatchRequest_Request{},
	}
	for i := 0; i < batchSize; i++ {
		subReq := &bbpb.ScheduleBuildRequest{
			Builder: builder,
		}
		generateGerritChangesAndTags(subReq)
		req.Requests = append(req.Requests, &bbpb.BatchRequest_Request{
			Request: &bbpb.BatchRequest_Request_ScheduleBuild{
				ScheduleBuild: subReq,
			},
		})
	}
	return req
}

func scheduleOneBatch(ctx context.Context, bbClient bbpb.BuildsClient, idx, batchSize int, cbs *fakebuildpb.ChildBuilds) error {
	step, ctx := build.StartStep(ctx, fmt.Sprintf("schedule children (%d)", idx))
	defer func() { step.End(nil) }()

	req := generateScheduleRequest(cbs.Builder, batchSize)
	res, err := bbClient.Batch(ctx, req)
	if err != nil {
		return errors.Annotate(err, "batch %d", idx).Err()
	}
	log := step.Log("response")
	marsh := jsonpb.Marshaler{Indent: "  "}
	if err = marsh.Marshal(log, res); err != nil {
		return errors.Annotate(err, "failed to marshal proto").Err()
	}

	secs := randomSecs(cbs.SleepMinSec, cbs.SleepMaxSec)
	clock.Sleep(ctx, time.Duration(secs)*time.Second)
	return nil
}

func scheduleChildBuilds(ctx context.Context, bbClient bbpb.BuildsClient, inputs *fakebuildpb.Inputs) error {
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

func searchBuildStep(ctx context.Context, stepName string, bbClient bbpb.BuildsClient, req *bbpb.SearchBuildsRequest, sbs *fakebuildpb.SearchBuilds) error {
	step, ctx := build.StartStep(ctx, stepName)
	defer func() { step.End(nil) }()

	res, err := bbClient.SearchBuilds(ctx, req)
	if err != nil {
		return errors.Annotate(err, stepName).Err()
	}
	log := step.Log("response")
	marsh := jsonpb.Marshaler{Indent: "  "}
	if err = marsh.Marshal(log, res); err != nil {
		return errors.Annotate(err, "failed to marshal proto").Err()
	}
	secs := randomSecs(sbs.SleepMinSec, sbs.SleepMaxSec)
	clock.Sleep(ctx, time.Duration(secs)*time.Second)
	return nil
}

// searchBuildsByBuildsetTag simulates milo to search related builds by buildset tag.
func searchBuildsByBuildsetTag(ctx context.Context, bbClient bbpb.BuildsClient, sbs *fakebuildpb.SearchBuilds) error {
	change := rand.Int63n(5000000)
	patchset := rand.Int63n(20)
	req := &bbpb.SearchBuildsRequest{
		Predicate: &bbpb.BuildPredicate{
			Tags:                []*bbpb.StringPair{{Key: "buildset", Value: fmt.Sprintf("patch/gerrit/chromium-review.googlesource.com/%d/%d", change, patchset)}},
			IncludeExperimental: true,
		},
	}
	return searchBuildStep(ctx, fmt.Sprintf("search related builds for CL %d/%d", change, patchset), bbClient, req, sbs)
}

// SearchBuildsByGerritChange simulates CV to search builds by gerrit change.
func SearchBuildsByGerritChange(ctx context.Context, bbClient bbpb.BuildsClient, sbs *fakebuildpb.SearchBuilds, idx int) error {
	change := rand.Int63n(5000000)
	patchset := rand.Int63n(20)

	req := &bbpb.SearchBuildsRequest{
		Predicate: &bbpb.BuildPredicate{
			GerritChanges:       []*bbpb.GerritChange{gerritChange(change, patchset)},
			IncludeExperimental: true,
		},
		Mask: &bbpb.BuildMask{
			Fields: &fieldmaskpb.FieldMask{
				Paths: []string{
					"builder",
					"create_time",
					"id",
					"output.properties",
					"status",
					"status_details",
					"summary_markdown",
					"update_time",
					"input.gerrit_changes",
					"infra.buildbucket.requested_properties",
				},
			},
		},
	}
	return searchBuildStep(ctx, fmt.Sprintf("search builds (%d)", idx), bbClient, req, sbs)
}

func searchBuildsByBuilder(ctx context.Context, bbClient bbpb.BuildsClient, sbs *fakebuildpb.SearchBuilds) error {
	req := &bbpb.SearchBuildsRequest{
		Predicate: &bbpb.BuildPredicate{
			Builder: &bbpb.BuilderID{
				Project: "infra",
				Bucket:  "loadtest",
				Builder: "fake-1m-no-bn",
			},
			Status:              bbpb.Status_ENDED_MASK,
			IncludeExperimental: true,
		},
		PageSize: 200,
	}
	return searchBuildStep(ctx, "search builds by builder", bbClient, req, sbs)
}

func searchBuilds(ctx context.Context, bbClient bbpb.BuildsClient, inputs *fakebuildpb.Inputs) error {
	sbs := inputs.GetSearchBuilds()
	if sbs == nil {
		return nil
	}

	steps := int(sbs.Steps)
	if steps > 2 {
		steps = steps - 2
	}

	for i := 0; i < steps; i++ {
		if err := SearchBuildsByGerritChange(ctx, bbClient, sbs, i); err != nil {
			return errors.Annotate(err, "search build %d", i).Err()
		}
	}
	if err := searchBuildsByBuildsetTag(ctx, bbClient, sbs); err != nil {
		return err
	}
	return searchBuildsByBuilder(ctx, bbClient, sbs)
}
