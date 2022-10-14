package handler

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	gfipb "go.chromium.org/luci/bisection/proto"
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

type mockGoFindit struct {
	res *gfipb.QueryAnalysisResponse
}

func (mgfi *mockGoFindit) QueryGoFinditResults(c context.Context, bbid int64, stepName string) (*gfipb.QueryAnalysisResponse, error) {
	return mgfi.res, nil
}

func TestAttachGoFinditResults(t *testing.T) {
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
		mockGfi := &mockGoFindit{
			res: &gfipb.QueryAnalysisResponse{
				Analyses: []*gfipb.Analysis{
					{
						AnalysisId: 12345,
					},
				},
			},
		}
		attachGoFinditResults(c, bf, mockGfi)
		So(len(bf[0].GoFinditResult), ShouldEqual, 0)
	})

	Convey("compile failure", t, func() {
		bf := []*messages.BuildFailure{
			{
				Builders: []*messages.AlertedBuilder{
					{
						Project: "chromium",
						Bucket:  "ci",
					},
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
		mockGfi := &mockGoFindit{
			res: &gfipb.QueryAnalysisResponse{
				Analyses: []*gfipb.Analysis{
					{
						AnalysisId: 12345,
					},
				},
			},
		}
		attachGoFinditResults(c, bf, mockGfi)
		So(len(bf[0].GoFinditResult), ShouldEqual, 2)
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
