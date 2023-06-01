// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package reviewer

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/proto"
	gerritpb "go.chromium.org/luci/common/proto/gerrit"
	. "go.chromium.org/luci/common/testing/assertions"
	"go.chromium.org/luci/gae/impl/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"infra/appengine/rubber-stamper/config"
	"infra/appengine/rubber-stamper/tasks/taskspb"
)

func TestReviewBenignFileChange(t *testing.T) {
	Convey("review benign file change", t, func() {
		ctx := memory.Use(context.Background())

		ctl := gomock.NewController(t)
		defer ctl.Finish()
		gerritMock := gerritpb.NewMockGerritClient(ctl)

		t := &taskspb.ChangeReviewTask{
			Host:       "test-host",
			Number:     12345,
			Revision:   "123abc",
			Repo:       "dummy",
			AutoSubmit: false,
		}

		Convey("BenignFilePattern in RepoConfig works", func() {
			hostCfg := &config.HostConfig{
				RepoConfigs: map[string]*config.RepoConfig{
					"dummy": {
						BenignFilePattern: &config.BenignFilePattern{
							Paths: []string{
								"test/a/*",
								"test/b/c.txt",
								"test/c/**",
								"test/override/**",
								"!test/override/**",
								"test/override/a/**",
								"!test/override/a/b/*",
							},
						},
					},
				},
			}

			Convey("valid files with gitignore style patterns", func() {
				gerritMock.EXPECT().ListFiles(gomock.Any(), proto.MatcherEqual(&gerritpb.ListFilesRequest{
					Number:     t.Number,
					RevisionId: t.Revision,
				})).Return(&gerritpb.ListFilesResponse{
					Files: map[string]*gerritpb.FileInfo{
						"/COMMIT_MSG":   nil,
						"test/a/b.xtb":  nil,
						"test/b/c.txt":  nil,
						"test/c/pp.xtb": nil,
						"test/c/i/a.md": nil,
					},
				}, nil)
				invalidFiles, err := reviewBenignFileChange(ctx, hostCfg, gerritMock, t)
				So(err, ShouldBeNil)
				So(len(invalidFiles), ShouldEqual, 0)
			})
			Convey("gitigore style patterns' order matters", func() {
				gerritMock.EXPECT().ListFiles(gomock.Any(), proto.MatcherEqual(&gerritpb.ListFilesRequest{
					Number:     t.Number,
					RevisionId: t.Revision,
				})).Return(&gerritpb.ListFilesResponse{
					Files: map[string]*gerritpb.FileInfo{
						"/COMMIT_MSG":             nil,
						"test/override/1.txt":     nil,
						"test/override/a/2.txt":   nil,
						"test/override/a/b/3.txt": nil,
						"test/override/a/c/4.txt": nil,
						"test/override/ab/5.txt":  nil,
					},
				}, nil)
				invalidFiles, err := reviewBenignFileChange(ctx, hostCfg, gerritMock, t)
				So(err, ShouldBeNil)
				So(invalidFiles, ShouldResemble, []string{"test/override/1.txt", "test/override/a/b/3.txt", "test/override/ab/5.txt"})
			})
			Convey("gerrit ListFiles API returns error", func() {
				gerritMock.EXPECT().ListFiles(gomock.Any(), proto.MatcherEqual(&gerritpb.ListFilesRequest{
					Number:     t.Number,
					RevisionId: t.Revision,
				})).Return(nil, grpc.Errorf(codes.NotFound, "not found"))
				invalidFiles, err := reviewBenignFileChange(ctx, hostCfg, gerritMock, t)
				So(err, ShouldErrLike, "gerrit ListFiles rpc call failed with error")
				So(len(invalidFiles), ShouldEqual, 0)
			})
		})

		Convey("BenignFilePattern in RepoRegexpConfig works", func() {
			hostCfg := &config.HostConfig{
				RepoRegexpConfigs: []*config.HostConfig_RepoRegexpConfigPair{
					{
						Key: ".*mmy",
						Value: &config.RepoConfig{
							BenignFilePattern: &config.BenignFilePattern{
								Paths: []string{
									"test/a/*",
									"test/b/c.txt",
									"test/c/**",
									"test/override/**",
									"!test/override/**",
									"test/override/a/**",
									"!test/override/a/b/*",
								},
							},
						},
					},
				},
			}

			Convey("valid files with gitignore style patterns", func() {
				gerritMock.EXPECT().ListFiles(gomock.Any(), proto.MatcherEqual(&gerritpb.ListFilesRequest{
					Number:     t.Number,
					RevisionId: t.Revision,
				})).Return(&gerritpb.ListFilesResponse{
					Files: map[string]*gerritpb.FileInfo{
						"/COMMIT_MSG":   nil,
						"test/a/b.xtb":  nil,
						"test/b/c.txt":  nil,
						"test/c/pp.xtb": nil,
						"test/c/i/a.md": nil,
					},
				}, nil)
				invalidFiles, err := reviewBenignFileChange(ctx, hostCfg, gerritMock, t)
				So(err, ShouldBeNil)
				So(len(invalidFiles), ShouldEqual, 0)
			})
			Convey("gitigore style patterns' order matters", func() {
				gerritMock.EXPECT().ListFiles(gomock.Any(), proto.MatcherEqual(&gerritpb.ListFilesRequest{
					Number:     t.Number,
					RevisionId: t.Revision,
				})).Return(&gerritpb.ListFilesResponse{
					Files: map[string]*gerritpb.FileInfo{
						"/COMMIT_MSG":             nil,
						"test/override/1.txt":     nil,
						"test/override/a/2.txt":   nil,
						"test/override/a/b/3.txt": nil,
						"test/override/a/c/4.txt": nil,
						"test/override/ab/5.txt":  nil,
					},
				}, nil)
				invalidFiles, err := reviewBenignFileChange(ctx, hostCfg, gerritMock, t)
				So(err, ShouldBeNil)
				So(invalidFiles, ShouldResemble, []string{"test/override/1.txt", "test/override/a/b/3.txt", "test/override/ab/5.txt"})
			})
			Convey("gerrit ListFiles API returns error", func() {
				gerritMock.EXPECT().ListFiles(gomock.Any(), proto.MatcherEqual(&gerritpb.ListFilesRequest{
					Number:     t.Number,
					RevisionId: t.Revision,
				})).Return(nil, grpc.Errorf(codes.NotFound, "not found"))
				invalidFiles, err := reviewBenignFileChange(ctx, hostCfg, gerritMock, t)
				So(err, ShouldErrLike, "gerrit ListFiles rpc call failed with error")
				So(len(invalidFiles), ShouldEqual, 0)
			})
		})
	})
}

func TestRetrieveBenignFilePattern(t *testing.T) {
	Convey("retrieveBenignFilePattern works", t, func() {
		sampleBenignFilePattern := &config.BenignFilePattern{
			Paths: []string{
				"test/a/*",
				"test/b/c.txt",
			},
		}
		Convey("returns nil when hostConfig is nil", func() {
			So(retrieveBenignFilePattern(context.Background(), nil, "dummy"), ShouldBeNil)
		})
		Convey("when repoConfig exists", func() {
			hostCfg := &config.HostConfig{
				RepoRegexpConfigs: []*config.HostConfig_RepoRegexpConfigPair{
					{
						Key:   ".*mmy",
						Value: nil,
					},
				},
			}
			Convey("BenignFilePattern exists", func() {
				hostCfg.RepoConfigs = map[string]*config.RepoConfig{
					"dummy": {
						BenignFilePattern: sampleBenignFilePattern,
					},
				}
				So(retrieveBenignFilePattern(context.Background(), hostCfg, "dummy"), ShouldEqual, sampleBenignFilePattern)
			})
			Convey("BenignFilePattern is nil", func() {
				hostCfg.RepoConfigs = map[string]*config.RepoConfig{
					"dummy": {
						BenignFilePattern: nil,
					},
				}
				So(retrieveBenignFilePattern(context.Background(), hostCfg, "dummy"), ShouldBeNil)
			})
		})
		Convey("when repoConfig doesn't exist and repoRegexpConfig exists", func() {
			Convey("BenignFilePattern exists", func() {
				hostCfg := &config.HostConfig{
					RepoRegexpConfigs: []*config.HostConfig_RepoRegexpConfigPair{
						{
							Key: ".*mmy",
							Value: &config.RepoConfig{
								BenignFilePattern: sampleBenignFilePattern,
							},
						},
					},
				}
				So(retrieveBenignFilePattern(context.Background(), hostCfg, "dummy"), ShouldEqual, sampleBenignFilePattern)
			})
			Convey("BenignFilePattern is nil", func() {
				hostCfg := &config.HostConfig{
					RepoRegexpConfigs: []*config.HostConfig_RepoRegexpConfigPair{
						{
							Key: ".*mmy",
							Value: &config.RepoConfig{
								BenignFilePattern: nil,
							},
						},
					},
				}
				So(retrieveBenignFilePattern(context.Background(), hostCfg, "dummy"), ShouldBeNil)
			})
		})
		Convey("returns nil when no repo config can be found", func() {
			hostCfg := &config.HostConfig{
				RepoConfigs: map[string]*config.RepoConfig{
					"dummy": {
						BenignFilePattern: sampleBenignFilePattern,
					},
				},
				RepoRegexpConfigs: []*config.HostConfig_RepoRegexpConfigPair{
					{
						Key: ".*mmy",
						Value: &config.RepoConfig{
							BenignFilePattern: nil,
						},
					},
				},
			}
			So(retrieveBenignFilePattern(context.Background(), hostCfg, "invalid"), ShouldBeNil)
		})
	})
}
