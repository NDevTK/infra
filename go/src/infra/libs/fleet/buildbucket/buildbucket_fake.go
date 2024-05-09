// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package buildbucket

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
)

type FakeBuildClient struct{}

func (f FakeBuildClient) GetBuild(context.Context, *buildbucketpb.GetBuildRequest, ...grpc.CallOption) (*buildbucketpb.Build, error) {
	// Not yet implemented.
	return nil, fmt.Errorf("GetBuild not yet implemented")
}

func (f FakeBuildClient) ScheduleBuild(ctx context.Context, in *buildbucketpb.ScheduleBuildRequest, opts ...grpc.CallOption) (*buildbucketpb.Build, error) {
	return &buildbucketpb.Build{
		Id: 123,
	}, nil
}

func (f FakeBuildClient) SearchBuilds(ctx context.Context, in *buildbucketpb.SearchBuildsRequest, opts ...grpc.CallOption) (*buildbucketpb.SearchBuildsResponse, error) {
	// Not yet implemented.
	return nil, fmt.Errorf("SearchBuilds not yet implemented")
}

type FakeClient struct {
	buildBucketClient FakeBuildClient
}

func (c *FakeClient) GetLatestGreenBuild(ctx context.Context) (*buildbucketpb.Build, error) {
	// Not yet implemented.
	return &buildbucketpb.Build{
		Id:     1234,
		Number: 1234,
	}, nil
}

func (c *FakeClient) BuildURL(ID int64) string {
	return fmt.Sprintf("test/%d", ID)
}

func (c FakeClient) ScheduleCTPBuild(ctx context.Context) (*buildbucketpb.Build, error) {
	return &buildbucketpb.Build{
		Id: 123,
	}, nil
}
