// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package compilefailuredetection

import (
	"context"
	"infra/appengine/gofindit/internal/buildbucket"
	"infra/appengine/gofindit/model"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/types/known/timestamppb"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
)

func TestFailureDetection(t *testing.T) {
	t.Parallel()

	Convey("Has Compile Step Status", t, func() {
		c := context.Background()
		Convey("No Compile Step", func() {
			build := &buildbucketpb.Build{
				Steps: []*buildbucketpb.Step{},
			}
			So(hasCompileStepStatus(c, build, buildbucketpb.Status_FAILURE), ShouldBeFalse)
		})
		Convey("Has Compile Step", func() {
			build := &buildbucketpb.Build{
				Steps: []*buildbucketpb.Step{
					{
						Name:   "compile",
						Status: buildbucketpb.Status_FAILURE,
					},
				},
			}
			So(hasCompileStepStatus(c, build, buildbucketpb.Status_FAILURE), ShouldBeTrue)
			So(hasCompileStepStatus(c, build, buildbucketpb.Status_SUCCESS), ShouldBeFalse)
		})
	})

	Convey("GetLastPassedFirstFailedBuild", t, func() {
		c := context.Background()
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		mc := buildbucket.NewMockedClient(c, ctl)
		c = mc.Ctx

		Convey("No builds", func() {
			res := &buildbucketpb.SearchBuildsResponse{
				Builds: []*buildbucketpb.Build{},
			}
			mc.Client.EXPECT().SearchBuilds(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil).AnyTimes()
			_, _, err := getLastPassedFirstFailedBuilds(c, &buildbucketpb.Build{Id: 123})
			So(err, ShouldNotBeNil)
		})

		Convey("Got succeeded builds", func() {
			res := &buildbucketpb.SearchBuildsResponse{
				Builds: []*buildbucketpb.Build{
					{
						Id:     123,
						Status: buildbucketpb.Status_FAILURE,
						Steps: []*buildbucketpb.Step{
							{
								Name:   "compile",
								Status: buildbucketpb.Status_FAILURE,
							},
						},
					},
					{
						Id:     122,
						Status: buildbucketpb.Status_FAILURE,
						Steps: []*buildbucketpb.Step{
							{
								Name:   "compile",
								Status: buildbucketpb.Status_FAILURE,
							},
						},
					},
					{
						Id:     121,
						Status: buildbucketpb.Status_INFRA_FAILURE,
					},
					{
						Id:     120,
						Status: buildbucketpb.Status_SUCCESS,
						Steps: []*buildbucketpb.Step{
							{
								Name:   "compile",
								Status: buildbucketpb.Status_SUCCESS,
							},
						},
					},
				},
			}
			mc.Client.EXPECT().SearchBuilds(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil).AnyTimes()
			lastPassedBuild, firstFailedBuild, err := getLastPassedFirstFailedBuilds(c, &buildbucketpb.Build{Id: 123})
			So(err, ShouldBeNil)
			So(lastPassedBuild.Id, ShouldEqual, 120)
			So(firstFailedBuild.Id, ShouldEqual, 122)
		})
	})

	Convey("analysisExists", t, func() {
		c := memory.Use(context.Background())

		build := &buildbucketpb.Build{
			Id: 8002,
			Builder: &buildbucketpb.BuilderID{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ios",
			},
			Number:     123,
			Status:     buildbucketpb.Status_FAILURE,
			StartTime:  &timestamppb.Timestamp{Seconds: 100},
			EndTime:    &timestamppb.Timestamp{Seconds: 101},
			CreateTime: &timestamppb.Timestamp{Seconds: 100},
			Input: &buildbucketpb.Build_Input{
				GitilesCommit: &buildbucketpb.GitilesCommit{},
			},
		}

		firstFailedBuild := &buildbucketpb.Build{
			Id: 8001,
			Builder: &buildbucketpb.BuilderID{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ios",
			},
			Number:     122,
			Status:     buildbucketpb.Status_FAILURE,
			StartTime:  &timestamppb.Timestamp{Seconds: 100},
			EndTime:    &timestamppb.Timestamp{Seconds: 101},
			CreateTime: &timestamppb.Timestamp{Seconds: 100},
		}

		lastPassedBuild := &buildbucketpb.Build{
			Id: 8000,
			Builder: &buildbucketpb.BuilderID{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "ios",
			},
			Number:     121,
			Status:     buildbucketpb.Status_FAILURE,
			StartTime:  &timestamppb.Timestamp{Seconds: 100},
			EndTime:    &timestamppb.Timestamp{Seconds: 101},
			CreateTime: &timestamppb.Timestamp{Seconds: 100},
		}

		Convey("There is no existing analysis", func() {
			check, cf, e := analysisExists(c, build, lastPassedBuild, firstFailedBuild)
			So(check, ShouldBeTrue)
			So(cf, ShouldNotBeNil)
			So(e, ShouldBeNil)
		})

		Convey("There is existing analysis", func() {
			failed_build := &model.LuciFailedBuild{
				Id: 8001,
				LuciBuild: model.LuciBuild{
					BuildId: 8001,
				},
				FailureType: model.BuildFailureType_Compile,
			}
			So(datastore.Put(c, failed_build), ShouldBeNil)
			datastore.GetTestable(c).CatchupIndexes()

			compile_failure := &model.CompileFailure{
				Id:    8001,
				Build: datastore.KeyForObj(c, failed_build),
			}
			So(datastore.Put(c, compile_failure), ShouldBeNil)
			datastore.GetTestable(c).CatchupIndexes()

			compile_failure_analysis := &model.CompileFailureAnalysis{
				CompileFailure:     datastore.KeyForObj(c, compile_failure),
				FirstFailedBuildId: 8001,
				LastPassedBuildId:  8000,
			}
			So(datastore.Put(c, compile_failure_analysis), ShouldBeNil)
			datastore.GetTestable(c).CatchupIndexes()
			check, cf, e := analysisExists(c, build, lastPassedBuild, firstFailedBuild)
			So(check, ShouldBeFalse)
			So(e, ShouldBeNil)
			So(cf, ShouldNotBeNil)
			So(cf.Id, ShouldEqual, 8002)
			So(cf.MergedFailureKey.IntID(), ShouldEqual, 8001)
		})
	})
}
