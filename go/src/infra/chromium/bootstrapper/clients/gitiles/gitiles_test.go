// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gitiles

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.chromium.org/luci/common/proto"
	gitpb "go.chromium.org/luci/common/proto/git"
	gitilespb "go.chromium.org/luci/common/proto/gitiles"
	"go.chromium.org/luci/common/proto/gitiles/mock_gitiles"
	. "go.chromium.org/luci/common/testing/assertions"

	"infra/chromium/bootstrapper/clients/gob"
)

func TestClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctx = gob.UseTestClock(ctx)

	Convey("Client", t, func() {

		Convey("gitilesClientForHost", func() {

			Convey("fails if factory fails", func() {
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return nil, errors.New("fake client factory failure")
				})

				client := NewClient(ctx)
				gitilesClient, err := client.gitilesClientForHost(ctx, "fake-host")

				So(err, ShouldNotBeNil)
				So(gitilesClient, ShouldBeNil)
			})

			Convey("returns gitiles client from factory", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				mockGitilesClient := mock_gitiles.NewMockGitilesClient(ctl)
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mockGitilesClient, nil
				})

				client := NewClient(ctx)
				gitilesClient, err := client.gitilesClientForHost(ctx, "fake-host")

				So(err, ShouldBeNil)
				So(gitilesClient, ShouldEqual, mockGitilesClient)
			})

			Convey("re-uses gitiles client for host", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mock_gitiles.NewMockGitilesClient(ctl), nil
				})

				client := NewClient(ctx)
				gitilesClientFoo1, _ := client.gitilesClientForHost(ctx, "fake-host-foo")
				gitilesClientFoo2, _ := client.gitilesClientForHost(ctx, "fake-host-foo")
				gitilesClientBar, _ := client.gitilesClientForHost(ctx, "fake-host-bar")

				So(gitilesClientFoo1, ShouldNotBeNil)
				So(gitilesClientFoo2, ShouldPointTo, gitilesClientFoo1)
				So(gitilesClientBar, ShouldNotPointTo, gitilesClientFoo1)
			})

		})

		Convey("FetchLatestRevision", func() {

			Convey("fails if getting gitiles client fails", func() {
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return nil, errors.New("test gitiles client factory failure")
				})

				client := NewClient(ctx)
				revision, err := client.FetchLatestRevision(ctx, "fake-host", "fake/project", "refs/heads/fake-branch")

				So(err, ShouldNotBeNil)
				So(revision, ShouldBeEmpty)
			})

			Convey("fails if API call fails", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				mockGitilesClient := mock_gitiles.NewMockGitilesClient(ctl)
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mockGitilesClient, nil
				})
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fake Log failure"))

				client := NewClient(ctx)
				revision, err := client.FetchLatestRevision(ctx, "fake-host", "fake/project", "refs/heads/fake-branch")

				So(err, ShouldNotBeNil)
				So(revision, ShouldBeEmpty)
			})

			Convey("returns latest revision for ref", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				mockGitilesClient := mock_gitiles.NewMockGitilesClient(ctl)
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mockGitilesClient, nil
				})
				matcher := proto.MatcherEqual(&gitilespb.LogRequest{
					Project:    "fake/project",
					Committish: "refs/heads/fake-branch",
					PageSize:   1,
				})
				// Check that potentially transient errors are retried
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), matcher).
					Return(nil, status.Error(codes.NotFound, "fake transient Log failure"))
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), matcher).
					Return(nil, status.Error(codes.Unavailable, "fake transient Log failure"))
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), matcher).
					Return(&gitilespb.LogResponse{
						Log: []*gitpb.Commit{
							{Id: "fake-revision"},
						},
					}, nil)

				client := NewClient(ctx)
				revision, err := client.FetchLatestRevision(ctx, "fake-host", "fake/project", "refs/heads/fake-branch")

				So(err, ShouldBeNil)
				So(revision, ShouldEqual, "fake-revision")
			})

		})

		Convey("FetchLatestRevisionForPath", func() {

			Convey("fails if getting gitiles client fails", func() {
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return nil, errors.New("test gitiles client factory failure")
				})

				client := NewClient(ctx)
				revision, err := client.FetchLatestRevisionForPath(ctx, "fake-host", "fake/project", "refs/heads/fake-branch", "fake-path")

				So(err, ShouldNotBeNil)
				So(revision, ShouldBeEmpty)
			})

			Convey("fails if API call fails", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				mockGitilesClient := mock_gitiles.NewMockGitilesClient(ctl)
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mockGitilesClient, nil
				})
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fake Log failure"))

				client := NewClient(ctx)
				revision, err := client.FetchLatestRevisionForPath(ctx, "fake-host", "fake/project", "refs/heads/fake-branch", "fake-path")

				So(err, ShouldNotBeNil)
				So(revision, ShouldBeEmpty)
			})

			Convey("returns latest revision for path on ref", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				mockGitilesClient := mock_gitiles.NewMockGitilesClient(ctl)
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mockGitilesClient, nil
				})
				matcher := proto.MatcherEqual(&gitilespb.LogRequest{
					Project:    "fake/project",
					Committish: "refs/heads/fake-branch",
					PageSize:   1,
					Path:       "fake-path",
				})
				// Check that potentially transient errors are retried
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), matcher).
					Return(nil, status.Error(codes.NotFound, "fake transient Log failure"))
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), matcher).
					Return(nil, status.Error(codes.Unavailable, "fake transient Log failure"))
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), matcher).
					Return(&gitilespb.LogResponse{
						Log: []*gitpb.Commit{
							{Id: "fake-revision"},
						},
					}, nil)

				client := NewClient(ctx)
				revision, err := client.FetchLatestRevisionForPath(ctx, "fake-host", "fake/project", "refs/heads/fake-branch", "fake-path")

				So(err, ShouldBeNil)
				So(revision, ShouldEqual, "fake-revision")
			})

		})

		Convey("GetParentRevision", func() {

			Convey("fails if getting gitiles client fails", func() {
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return nil, errors.New("test gitiles client factory failure")
				})

				client := NewClient(ctx)
				revision, err := client.GetParentRevision(ctx, "fake-host", "fake/project", "fake-revision")

				So(err, ShouldErrLike, "test gitiles client factory failure")
				So(revision, ShouldBeEmpty)
			})

			Convey("fails if API call fails", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				mockGitilesClient := mock_gitiles.NewMockGitilesClient(ctl)
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mockGitilesClient, nil
				})
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fake Log failure"))

				client := NewClient(ctx)
				revision, err := client.GetParentRevision(ctx, "fake-host", "fake/project", "fake-revision")

				So(err, ShouldErrLike, "fake Log failure")
				So(revision, ShouldBeEmpty)
			})

			Convey("returns parent revision for revision", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				mockGitilesClient := mock_gitiles.NewMockGitilesClient(ctl)
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mockGitilesClient, nil
				})
				matcher := proto.MatcherEqual(&gitilespb.LogRequest{
					Project:    "fake/project",
					Committish: "fake-revision",
					PageSize:   2,
				})
				// Check that potentially transient errors are retried
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), matcher).
					Return(nil, status.Error(codes.NotFound, "fake transient Log failure"))
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), matcher).
					Return(nil, status.Error(codes.Unavailable, "fake transient Log failure"))
				mockGitilesClient.EXPECT().
					Log(gomock.Any(), matcher).
					Return(&gitilespb.LogResponse{
						Log: []*gitpb.Commit{
							{Id: "fake-revision"},
							{Id: "fake-parent-revision"},
						},
					}, nil)

				client := NewClient(ctx)
				revision, err := client.GetParentRevision(ctx, "fake-host", "fake/project", "fake-revision")

				So(err, ShouldBeNil)
				So(revision, ShouldEqual, "fake-parent-revision")
			})

		})

		Convey("DownloadFile", func() {

			Convey("fails if getting gitiles client fails", func() {
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return nil, errors.New("test gitiles client factory failure")
				})

				client := NewClient(ctx)
				contents, err := client.DownloadFile(ctx, "fake-host", "fake/project", "fake-revision", "fake-file")

				So(err, ShouldNotBeNil)
				So(contents, ShouldBeEmpty)
			})

			Convey("fails if API call fails", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				mockGitilesClient := mock_gitiles.NewMockGitilesClient(ctl)
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mockGitilesClient, nil
				})
				mockGitilesClient.EXPECT().
					DownloadFile(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("fake DownloadFile failure"))

				client := NewClient(ctx)
				contents, err := client.DownloadFile(ctx, "fake-host", "fake/project", "fake-revision", "fake-file")

				So(err, ShouldNotBeNil)
				So(contents, ShouldBeEmpty)
			})

			Convey("returns file contents", func() {
				ctl := gomock.NewController(t)
				defer ctl.Finish()

				mockGitilesClient := mock_gitiles.NewMockGitilesClient(ctl)
				ctx := UseGitilesClientFactory(ctx, func(ctx context.Context, host string) (GitilesClient, error) {
					return mockGitilesClient, nil
				})
				matcher := proto.MatcherEqual(&gitilespb.DownloadFileRequest{
					Project:    "fake/project",
					Committish: "fake-revision",
					Path:       "fake-file",
				})
				// Check that potentially transient errors are retried
				mockGitilesClient.EXPECT().
					DownloadFile(gomock.Any(), matcher).
					Return(nil, status.Error(codes.NotFound, "fake transient DownloadFile failure"))
				mockGitilesClient.EXPECT().
					DownloadFile(gomock.Any(), matcher).
					Return(nil, status.Error(codes.NotFound, "fake transient DownloadFile failure"))
				mockGitilesClient.EXPECT().
					DownloadFile(gomock.Any(), matcher).
					Return(&gitilespb.DownloadFileResponse{
						Contents: "fake-contents",
					}, nil)

				client := NewClient(ctx)
				contents, err := client.DownloadFile(ctx, "fake-host", "fake/project", "fake-revision", "fake-file")

				So(err, ShouldBeNil)
				So(contents, ShouldEqual, "fake-contents")
			})

		})

	})
}
