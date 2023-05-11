// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handler

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	bisectionpb "go.chromium.org/luci/bisection/proto/v1"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/gae/impl/dummy"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/gae/service/info"

	"infra/appengine/sheriff-o-matic/som/analyzer"
	"infra/monitoring/messages"
)

func newTestContext() context.Context {
	c := gaetesting.TestingContext()
	ta := datastore.GetTestable(c)
	ta.Consistent(true)
	c = gologger.StdConfig.Use(c)
	return c
}

type giMock struct {
	info.RawInterface
	token  string
	expiry time.Time
	err    error
}

func (gi giMock) AccessToken(scopes ...string) (token string, expiry time.Time, err error) {
	return gi.token, gi.expiry, gi.err
}

type mockBisectionClient struct {
	res *bisectionpb.QueryAnalysisResponse
}

func (mbc *mockBisectionClient) QueryBisectionResults(c context.Context, bbid int64, stepName string) (*bisectionpb.QueryAnalysisResponse, error) {
	return mbc.res, nil
}

func TestAttachLuciBisectionResults(t *testing.T) {
	c := gaetesting.TestingContext()
	Convey("not a compile failure", t, func() {
		bf := []*messages.BuildFailure{
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "chromium",
						Bucket:  "ci",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "step",
					},
				},
			},
		}
		mockClient := &mockBisectionClient{
			res: &bisectionpb.QueryAnalysisResponse{
				Analyses: []*bisectionpb.Analysis{
					{
						AnalysisId: 12345,
					},
				},
			},
		}
		err := attachLuciBisectionResults(c, bf, mockClient)
		So(err, ShouldBeNil)
		So(bf[0].LuciBisectionResult, ShouldBeNil)
	})

	Convey("compile failure, not chromium ci", t, func() {
		bf := []*messages.BuildFailure{
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "chromium",
						Bucket:  "bucket",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "compile",
					},
				},
			},
		}
		mockClient := &mockBisectionClient{
			res: &bisectionpb.QueryAnalysisResponse{
				Analyses: []*bisectionpb.Analysis{
					{
						AnalysisId: 12345,
					},
				},
			},
		}
		err := attachLuciBisectionResults(c, bf, mockClient)
		So(err, ShouldBeNil)
		So(bf[0].LuciBisectionResult.IsSupported, ShouldEqual, false)
		So(bf[0].LuciBisectionResult.Analysis, ShouldBeNil)
	})

	Convey("compile failure", t, func() {
		bf := []*messages.BuildFailure{
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "chromium",
						Bucket:  "ci",
					},
				},
				StepAtFault: &messages.BuildStep{
					Step: &messages.Step{
						Name: "compile",
					},
				},
			},
		}
		mockClient := &mockBisectionClient{
			res: &bisectionpb.QueryAnalysisResponse{
				Analyses: []*bisectionpb.Analysis{
					{
						AnalysisId: 12345,
					},
				},
			},
		}
		err := attachLuciBisectionResults(c, bf, mockClient)
		So(err, ShouldBeNil)
		So(bf[0].LuciBisectionResult.IsSupported, ShouldEqual, true)
		So(bf[0].LuciBisectionResult.Analysis.AnalysisId, ShouldEqual, 12345)
	})
}

func TestStoreAlertsSummary(t *testing.T) {
	Convey("success", t, func() {
		c := gaetesting.TestingContext()
		c = info.SetFactory(c, func(ic context.Context) info.RawInterface {
			return giMock{dummy.Info(), "", clock.Now(c), nil}
		})
		a := analyzer.New(5, 100)
		err := storeAlertsSummary(c, a, "some tree", &messages.AlertsSummary{
			Alerts: []*messages.Alert{
				{
					Title: "foo",
					Extension: &messages.BuildFailure{
						RegressionRanges: []*messages.RegressionRange{
							{Repo: "some repo", URL: "about:blank", Positions: []string{}, Revisions: []string{}},
						},
					},
				},
			},
		})
		So(err, ShouldBeNil)
	})
}
