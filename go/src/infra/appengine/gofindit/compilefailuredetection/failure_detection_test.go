// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package compilefailuredetection

import (
	"context"
	"infra/appengine/gofindit/internal/buildbucket"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
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
}
