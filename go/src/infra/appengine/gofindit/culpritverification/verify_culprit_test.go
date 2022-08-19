// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// package culpritverification verifies if a suspect is a culprit.
package culpritverification

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
	bbpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/gae/impl/memory"
	"go.chromium.org/luci/gae/service/datastore"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/util/json"

	"infra/appengine/gofindit/internal/buildbucket"
	"infra/appengine/gofindit/internal/gitiles"
	"infra/appengine/gofindit/model"
)

func TestVerifySuspect(t *testing.T) {
	t.Parallel()
	c := memory.Use(context.Background())

	// Setup mock for buildbucket
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	mc := buildbucket.NewMockedClient(c, ctl)
	c = mc.Ctx
	res1 := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Project: "chromium",
			Bucket:  "findit",
			Builder: "gofindit-single-revision",
		},
		Input: &bbpb.Build_Input{
			GitilesCommit: &bbpb.GitilesCommit{
				Host:    "host",
				Project: "proj",
				Id:      "id1",
				Ref:     "ref",
			},
		},
		Id:         123,
		Status:     bbpb.Status_STARTED,
		CreateTime: &timestamppb.Timestamp{Seconds: 100},
		StartTime:  &timestamppb.Timestamp{Seconds: 101},
	}
	mc.Client.EXPECT().ScheduleBuild(gomock.Any(), gomock.Any(), gomock.Any()).Return(res1, nil).Times(1)

	res2 := &bbpb.Build{
		Builder: &bbpb.BuilderID{
			Project: "chromium",
			Bucket:  "findit",
			Builder: "gofindit-single-revision",
		},
		Input: &bbpb.Build_Input{
			GitilesCommit: &bbpb.GitilesCommit{
				Host:    "host",
				Project: "proj",
				Id:      "id2",
				Ref:     "ref",
			},
		},
		Id:         456,
		Status:     bbpb.Status_STARTED,
		CreateTime: &timestamppb.Timestamp{Seconds: 200},
		StartTime:  &timestamppb.Timestamp{Seconds: 201},
	}
	mc.Client.EXPECT().ScheduleBuild(gomock.Any(), gomock.Any(), gomock.Any()).Return(res2, nil).Times(1)
	mc.Client.EXPECT().GetBuild(gomock.Any(), gomock.Any(), gomock.Any()).Return(&bbpb.Build{}, nil).AnyTimes()

	Convey("Verify Suspect", t, func() {
		gitilesResponse := model.ChangeLogResponse{
			Log: []*model.ChangeLog{
				{
					Commit: "3424",
				},
			},
		}
		gitilesResponseStr, _ := json.Marshal(gitilesResponse)
		c = gitiles.MockedGitilesClientContext(c, map[string]string{
			"https://chromium.googlesource.com/chromium/src/+log/3425~2..3425^": string(gitilesResponseStr),
		})
		suspect := &model.Suspect{
			Score: 10,
			GitilesCommit: bbpb.GitilesCommit{
				Host:    "chromium.googlesource.com",
				Project: "chromium/src",
				Id:      "3425",
			},
		}
		So(datastore.Put(c, suspect), ShouldBeNil)
		datastore.GetTestable(c).CatchupIndexes()

		err := VerifySuspect(c, suspect, 8000, 444)
		So(err, ShouldBeNil)
		So(suspect.VerificationStatus, ShouldEqual, model.SuspectVerificationStatus_UnderVerification)

		// Check that 2 rerun builds were created, and linked to suspect
		rerun1 := &model.CompileRerunBuild{
			Id: suspect.SuspectRerunBuild.IntID(),
		}
		err = datastore.Get(c, rerun1)
		So(err, ShouldBeNil)
		So(rerun1, ShouldResemble, &model.CompileRerunBuild{
			Id:      123,
			Type:    model.RerunBuildType_CulpritVerification,
			Suspect: datastore.KeyForObj(c, suspect),
			LuciBuild: model.LuciBuild{
				BuildId:       123,
				Project:       "chromium",
				Bucket:        "findit",
				Builder:       "gofindit-single-revision",
				Status:        bbpb.Status_STARTED,
				GitilesCommit: *res1.Input.GitilesCommit,
				CreateTime:    res1.CreateTime.AsTime(),
				StartTime:     res1.StartTime.AsTime(),
			},
		})

		rerun2 := &model.CompileRerunBuild{
			Id: suspect.ParentRerunBuild.IntID(),
		}
		err = datastore.Get(c, rerun2)
		So(err, ShouldBeNil)
		So(rerun2, ShouldResemble, &model.CompileRerunBuild{
			Id:      456,
			Type:    model.RerunBuildType_CulpritVerification,
			Suspect: datastore.KeyForObj(c, suspect),
			LuciBuild: model.LuciBuild{
				BuildId:       456,
				Project:       "chromium",
				Bucket:        "findit",
				Builder:       "gofindit-single-revision",
				Status:        bbpb.Status_STARTED,
				GitilesCommit: *res2.Input.GitilesCommit,
				CreateTime:    res2.CreateTime.AsTime(),
				StartTime:     res2.StartTime.AsTime(),
			},
		})
	})
}
