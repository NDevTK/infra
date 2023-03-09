// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package buildbucket

import (
	"context"
	"fmt"
	"reflect"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/cmd/crosfleet/internal/common"
)

// ScheduleParams encapsulates a subset of ScheduleBuildRequest fields to
// match against in the fake client.
type ScheduleParams struct {
	BuilderName string
	Tags        map[string]string
}

func (p *ScheduleParams) matches(in *buildbucketpb.ScheduleBuildRequest) bool {
	if p.BuilderName != in.Builder.Builder {
		return false
	}
	reqTags := map[string]string{}
	for _, tag := range in.GetTags() {
		reqTags[tag.Key] = tag.Value
	}

	for key, value := range p.Tags {
		reqValue, ok := reqTags[key]
		if !ok || reqValue != value {
			return false
		}
	}
	return true
}

type FakeBuildClient struct {
	ExpectedSchedule []ScheduleParams
}

// Important that this is not a pointer receiver so that it can't be nil, see
// comment in crrev.com/c/4133287. (If it's nil the library will instantiate an
// actual client).

func (f FakeBuildClient) GetBuild(context.Context, *buildbucketpb.GetBuildRequest, ...grpc.CallOption) (*buildbucketpb.Build, error) {
	// Not yet implemented.
	return nil, nil
}

func requestSummary(in *buildbucketpb.ScheduleBuildRequest) string {
	return fmt.Sprintf("builder: %+v\ntags: %+v\n", in.Builder, in.GetTags())
}

func (f FakeBuildClient) ScheduleBuild(ctx context.Context, in *buildbucketpb.ScheduleBuildRequest, opts ...grpc.CallOption) (*buildbucketpb.Build, error) {
	matchedExpectation := false
	for i, expected := range f.ExpectedSchedule {
		if expected.matches(in) {
			matchedExpectation = true
			// Matching an expectation "consumes" it.
			f.ExpectedSchedule = append(f.ExpectedSchedule[:i], f.ExpectedSchedule[i:]...)
			break
		}
	}

	if !matchedExpectation {
		return nil, fmt.Errorf("unexpected ScheduleBuild call:\n%+v\n", requestSummary(in))
	}

	return &buildbucketpb.Build{
		Id: 123,
	}, nil
}

func (f FakeBuildClient) SearchBuilds(ctx context.Context, in *buildbucketpb.SearchBuildsRequest, opts ...grpc.CallOption) (*buildbucketpb.SearchBuildsResponse, error) {
	// Not yet implemented.
	return nil, nil
}

func (f FakeBuildClient) CancelBuild(ctx context.Context, in *buildbucketpb.CancelBuildRequest, opts ...grpc.CallOption) (*buildbucketpb.Build, error) {
	// Not yet implemented.
	return nil, nil
}

type ExpectedGetIncompleteBuildsWithTagsCall struct {
	Tags     map[string]string
	Response []*buildbucketpb.Build
}

type FakeClient struct {
	Client FakeBuildClient
	// Test data for GetIncompleteBuildsWithTags.
	ExpectedGetIncompleteBuildsWithTags []*ExpectedGetIncompleteBuildsWithTagsCall
}

func (c *FakeClient) GetBuildsClient() BuildsClient {
	return c.Client
}

func (c *FakeClient) GetBuilderID() *buildbucketpb.BuilderID {
	// Not yet implemented.
	return nil
}

func (c *FakeClient) ScheduleBuild(ctx context.Context, props map[string]interface{}, dims map[string]string, tags map[string]string, priority int32) (*buildbucketpb.Build, error) {
	// Not yet implemented.
	return nil, nil
}

func (c *FakeClient) WaitForBuildStepStart(ctx context.Context, id int64, stepName string) (*buildbucketpb.Build, error) {
	// Not yet implemented.
	return nil, nil
}

func (c *FakeClient) GetAllBuildsWithTags(ctx context.Context, tags map[string]string, searchBuildsRequest *buildbucketpb.SearchBuildsRequest) ([]*buildbucketpb.Build, error) {
	// Not yet implemented.
	return nil, nil
}

func (c *FakeClient) GetBuild(ctx context.Context, ID int64, fields ...string) (*buildbucketpb.Build, error) {
	// Not yet implemented.
	return nil, nil
}

func (c *FakeClient) GetLatestGreenBuild(ctx context.Context) (*buildbucketpb.Build, error) {
	// Not yet implemented.
	return nil, nil
}

func (c *FakeClient) AnyIncompleteBuildsWithTags(ctx context.Context, tags map[string]string) (bool, int64, error) {
	// Not yet implemented.
	return false, 0, nil
}

func (c *FakeClient) GetIncompleteBuildsWithTags(ctx context.Context, tags map[string]string) ([]*buildbucketpb.Build, error) {
	if c.ExpectedGetIncompleteBuildsWithTags == nil {
		return nil, fmt.Errorf("Unexpected call to GetIncompleteBuildsWithTags")
	}

	for i, expected := range c.ExpectedGetIncompleteBuildsWithTags {
		if reflect.DeepEqual(expected.Tags, tags) {
			c.ExpectedGetIncompleteBuildsWithTags = append(c.ExpectedGetIncompleteBuildsWithTags[:i], c.ExpectedGetIncompleteBuildsWithTags[i:]...)
			return expected.Response, nil
		}
	}

	return nil, fmt.Errorf("Unexpected call to GetIncompleteBuildsWithTags:\n%v\n", tags)
}

func (c *FakeClient) CancelBuildsByUser(ctx context.Context, printer common.CLIPrinter, earliestCreateTime *timestamppb.Timestamp, user string, ids []string, reason string) error {
	// Not yet implemented.
	return nil
}

func (c *FakeClient) GetAllBuildsByUser(ctx context.Context, user string, searchBuildsRequest *buildbucketpb.SearchBuildsRequest) ([]*buildbucketpb.Build, error) {
	// Not yet implemented.
	return nil, nil
}

func (c *FakeClient) BuildURL(ID int64) string {
	return fmt.Sprintf("test/%d", ID)
}
