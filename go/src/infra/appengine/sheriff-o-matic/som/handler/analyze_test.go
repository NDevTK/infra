// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handler

import (
	"context"
	"fmt"
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
	"infra/appengine/sheriff-o-matic/som/analyzer/step"
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
	QueryAnalysisResponse *bisectionpb.QueryAnalysisResponse
	// Response for each call to BatchGetTestAnalyses.
	BatchGetTestAnalysesResponses []*bisectionpb.BatchGetTestAnalysesResponse
}

func (mbc *mockBisectionClient) QueryBisectionResults(c context.Context, bbid int64, stepName string) (*bisectionpb.QueryAnalysisResponse, error) {
	return mbc.QueryAnalysisResponse, nil
}

func (mbc *mockBisectionClient) BatchGetTestAnalyses(c context.Context, req *bisectionpb.BatchGetTestAnalysesRequest) (*bisectionpb.BatchGetTestAnalysesResponse, error) {
	resp := mbc.BatchGetTestAnalysesResponses[0]
	mbc.BatchGetTestAnalysesResponses = mbc.BatchGetTestAnalysesResponses[1:]
	return resp, nil
}

func TestAttachLuciBisectionResults(t *testing.T) {
	c := gaetesting.TestingContext()
	Convey("attachLUCIBisectionBuildFailureAnalyses", t, func() {
		Convey("not a compile failure", func() {
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
				QueryAnalysisResponse: &bisectionpb.QueryAnalysisResponse{
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

		Convey("compile failure, not chromium ci", func() {
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
				QueryAnalysisResponse: &bisectionpb.QueryAnalysisResponse{
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

		Convey("compile failure", func() {
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
				QueryAnalysisResponse: &bisectionpb.QueryAnalysisResponse{
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
	})

	Convey("attachLUCIBisectionTestAnalyses", t, func() {
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
						Name: "test",
					},
				},
				Reason: &messages.Reason{
					Raw: (&analyzer.BqFailure{
						Tests: []step.TestWithResult{{
							TestID:      "test1",
							VariantHash: "varianthash1",
							RefHash:     "refhash1",
							CurCounts: step.Counts{
								UnexpectedResults: 10,
								TotalResults:      10,
							},
						},
							{
								TestID:      "test2",
								VariantHash: "varianthash2",
								RefHash:     "refhash2",
								CurCounts: step.Counts{
									UnexpectedResults: 10,
									TotalResults:      10,
								},
							}},
						NumFailingTests: 2,
					}).WithKind("test"),
				},
			},
		}
		mockClient := &mockBisectionClient{
			BatchGetTestAnalysesResponses: []*bisectionpb.BatchGetTestAnalysesResponse{
				{TestAnalyses: []*bisectionpb.TestAnalysis{
					nil, // No bisection for test1.
					{
						AnalysisId: 2,
						Status:     3,
					}, // Bisection for test2.
				}},
			},
		}
		err := attachLuciBisectionResults(c, bf, mockClient)
		So(err, ShouldBeNil)
		So(bf[0].Reason.Raw.(*analyzer.BqFailure).Tests[0].LUCIBisectionResult, ShouldBeNil)
		So(bf[0].Reason.Raw.(*analyzer.BqFailure).Tests[1].LUCIBisectionResult, ShouldResemble, &step.LUCIBisectionTestAnalysis{
			AnalysisID: "2",
			Status:     bisectionpb.AnalysisStatus(3).String(),
		})

		Convey("batch, single project", func() {
			bf := []*messages.BuildFailure{}
			// Create 201 failures, each failure has one failed test.
			// This will be put into 3 batches 0..99, 100..199, 200 when calling bisection.
			for i := 0; i < 201; i++ {
				bf = append(bf, &messages.BuildFailure{
					Builders: []*messages.AlertedBuilder{
						{
							Project: "chromium",
							Bucket:  "ci",
						},
					},
					StepAtFault: &messages.BuildStep{
						Step: &messages.Step{
							Name: "test",
						},
					},
					Reason: &messages.Reason{
						Raw: (&analyzer.BqFailure{
							Tests: []step.TestWithResult{{
								TestID:      fmt.Sprintf("test%d", i),
								VariantHash: fmt.Sprintf("varianthash%d", i),
								RefHash:     fmt.Sprintf("refhash%d", i),
								CurCounts: step.Counts{
									UnexpectedResults: 10,
									TotalResults:      10,
								},
							}},
							NumFailingTests: 1,
						}).WithKind("test"),
					},
				})
			}
			// Set Test analyses in the response for each call.
			responses := make([]*bisectionpb.BatchGetTestAnalysesResponse, 3)
			numTestAnalyses := []int{100, 100, 1} // First call returns 100 test analyses, second call returns 100, third call returns 1.
			for i, num := range numTestAnalyses {
				for j := 0; j < num; j++ {
					if responses[i] == nil {
						responses[i] = &bisectionpb.BatchGetTestAnalysesResponse{TestAnalyses: []*bisectionpb.TestAnalysis{}}
					}
					responses[i].TestAnalyses = append(responses[i].TestAnalyses, &bisectionpb.TestAnalysis{
						AnalysisId: int64(j + 100*i), // AnalysisId is 0..99, 100..199, 200.
						Status:     3,
					})
				}
			}
			mockClient := &mockBisectionClient{BatchGetTestAnalysesResponses: responses}

			err := attachLuciBisectionResults(c, bf, mockClient)
			So(err, ShouldBeNil)
			for i, b := range bf {
				So(b.Reason.Raw.(*analyzer.BqFailure).Tests[0].LUCIBisectionResult, ShouldResemble, &step.LUCIBisectionTestAnalysis{
					AnalysisID: fmt.Sprint(i),
					Status:     bisectionpb.AnalysisStatus(3).String(),
				})
			}
		})
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
