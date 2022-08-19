// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package rerun

import (
	"context"
	"infra/appengine/gofindit/internal/buildbucket"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/gae/impl/memory"
)

func TestRerun(t *testing.T) {
	t.Parallel()

	Convey("getRerunPropertiesAndDimensions", t, func() {
		c := memory.Use(context.Background())
		cl := testclock.New(testclock.TestTimeUTC)
		c = clock.Set(c, cl)

		// Setup mock for buildbucket
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		mc := buildbucket.NewMockedClient(c, ctl)
		c = mc.Ctx
		bootstrapProperties := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"bs_key_1": structpb.NewStringValue("bs_val_1"),
			},
		}

		targetBuilder := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"builder": structpb.NewStringValue("linux-test"),
				"group":   structpb.NewStringValue("buildergroup1"),
			},
		}

		res := &bbpb.Build{
			Builder: &bbpb.BuilderID{
				Project: "chromium",
				Bucket:  "ci",
				Builder: "linux-test",
			},
			Input: &bbpb.Build_Input{
				Properties: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"builder_group":         structpb.NewStringValue("buildergroup1"),
						"$bootstrap/properties": structpb.NewStructValue(bootstrapProperties),
						"another_prop":          structpb.NewStringValue("another_val"),
					},
				},
			},
			Infra: &bbpb.BuildInfra{
				Swarming: &bbpb.BuildInfra_Swarming{
					TaskDimensions: []*bbpb.RequestedDimension{
						{
							Key:   "dimen_key_1",
							Value: "dimen_val_1",
						},
						{
							Key:   "os",
							Value: "ubuntu",
						},
						{
							Key:   "gpu",
							Value: "Intel",
						},
					},
				},
			},
		}
		mc.Client.EXPECT().GetBuild(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil).AnyTimes()
		props, dimens, err := getRerunPropertiesAndDimensions(c, 1234, 4646418413256704)
		So(err, ShouldBeNil)
		So(props, ShouldResemble, &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"builder_group":  structpb.NewStringValue("buildergroup1"),
				"target_builder": structpb.NewStructValue(targetBuilder),
				"$bootstrap/properties": structpb.NewStructValue(&structpb.Struct{
					Fields: map[string]*structpb.Value{
						"bs_key_1": structpb.NewStringValue("bs_val_1"),
					},
				}),
				"analysis_id": structpb.NewNumberValue(4646418413256704),
			},
		})
		So(dimens, ShouldResemble, []*bbpb.RequestedDimension{
			{
				Key:   "os",
				Value: "ubuntu",
			},
			{
				Key:   "gpu",
				Value: "Intel",
			},
		})
	})
}
