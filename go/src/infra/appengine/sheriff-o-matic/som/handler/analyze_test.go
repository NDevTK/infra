package handler

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"golang.org/x/net/context"

	"infra/appengine/sheriff-o-matic/som/analyzer"
	"infra/appengine/sheriff-o-matic/som/analyzer/step"
	"infra/appengine/sheriff-o-matic/som/client"
	"infra/appengine/sheriff-o-matic/som/client/mock"
	testhelper "infra/appengine/sheriff-o-matic/som/client/test"
	"infra/appengine/sheriff-o-matic/som/model"
	"infra/monitoring/messages"

	"go.chromium.org/gae/impl/dummy"
	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/gae/service/info"
	tq "go.chromium.org/gae/service/taskqueue"
	"go.chromium.org/gae/service/urlfetch"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/server/auth/authtest"
	"go.chromium.org/luci/server/router"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/logging/gologger"

	"github.com/golang/mock/gomock"
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

func setUpGitiles(c context.Context) context.Context {
	return urlfetch.Set(c, &testhelper.MockGitilesTransport{
		Responses: map[string]string{
			gkTreesURL: `{    "chromium": {
        "build-db": "waterfall_build_db.json",
        "masters": {
            "https://build.chromium.org/p/chromium": ["*"]
        },
        "open-tree": true,
        "password-file": "/creds/gatekeeper/chromium_status_password",
        "revision-properties": "got_revision_cp",
        "set-status": true,
        "status-url": "https://chromium-status.appspot.com",
        "track-revisions": true
    }}`,
			gkTreesInternalURL: `{    "chromium": {
        "build-db": "waterfall_build_db.json",
        "masters": {
            "https://build.chromium.org/p/chromium": ["*"]
        },
        "open-tree": true,
        "password-file": "/creds/gatekeeper/chromium_status_password",
        "revision-properties": "got_revision_cp",
        "set-status": true,
        "status-url": "https://chromium-status.appspot.com",
        "track-revisions": true
    }}`,
			gkTreesCorpURL: `{    "chromium": {
        "build-db": "waterfall_build_db.json",
        "masters": {
            "https://build.chromium.org/p/chromium": ["*"]
        },
        "open-tree": true,
        "password-file": "/creds/gatekeeper/chromium_status_password",
        "revision-properties": "got_revision_cp",
        "set-status": true,
        "status-url": "https://chromium-status.appspot.com",
        "track-revisions": true
    }}`,
			gkUnkeptTreesURL: `{    "chromium": {
        "build-db": "waterfall_build_db.json",
        "masters": {
            "https://build.chromium.org/p/chromium": ["*"]
        },
        "open-tree": true,
        "password-file": "/creds/gatekeeper/chromium_status_password",
        "revision-properties": "got_revision_cp",
        "set-status": true,
        "status-url": "https://chromium-status.appspot.com",
        "track-revisions": true
    }}`,
			gkConfigInternalURL: `
{
  "comment": ["This is a configuration file for gatekeeper_ng.py",
              "Look at that for documentation on this file's format."],
  "masters": {
    "https://build.chromium.org/p/chromium": [
      {
        "categories": [
          "chromium_tree_closer"
        ],
        "builders": {
          "Win": {
            "categories": [
              "chromium_windows"
            ]
          },
          "*": {}
        }
      }
    ]
   }
}`,

			gkConfigURL: `
{
  "comment": ["This is a configuration file for gatekeeper_ng.py",
              "Look at that for documentation on this file's format."],
  "masters": {
    "https://build.chromium.org/p/chromium": [
      {
        "categories": [
          "chromium_tree_closer"
        ],
        "builders": {
          "Win": {
            "categories": [
              "chromium_windows"
            ]
          },
          "*": {}
        }
      }
    ]
   }
}`,
			gkConfigCorpURL: `
{
  "comment": ["This is a configuration file for gatekeeper_ng.py",
              "Look at that for documentation on this file's format."],
  "masters": {
    "https://build.chromium.org/p/chromium": [
      {
        "categories": [
          "chromium_tree_closer"
        ],
        "builders": {
          "Win": {
            "categories": [
              "chromium_windows"
            ]
          },
          "*": {}
        }
      }
    ]
   }
}`,
			gkUnkeptConfigURL: `
{
  "comment": ["This is a configuration file for gatekeeper_ng.py",
              "Look at that for documentation on this file's format."],
  "masters": {
    "https://build.chromium.org/p/chromium": [
      {
        "categories": [
          "chromium_tree_closer"
        ],
        "builders": {
          "Win": {
            "categories": [
              "chromium_windows"
            ]
          },
          "*": {}
        }
      }
    ]
   }
}`,
		}})
}

func TestGenerateAlerts(t *testing.T) {
	Convey("bad request", t, func() {
		c := gaetesting.TestingContext()
		c = info.SetFactory(c, func(ic context.Context) info.RawInterface {
			return giMock{dummy.Info(), "", time.Now(), nil}
		})
		c = setUpGitiles(c)
		w := httptest.NewRecorder()

		ctx := &router.Context{
			Context: c,
			Writer:  w,
			Request: makeGetRequest(),
			Params:  makeParams("tree", "unknown.tree"),
		}
		_, _ = generateAlerts(ctx)

		So(w.Code, ShouldEqual, http.StatusNotFound)
	})

	Convey("ok request", t, func() {
		c := newTestContext()
		c = info.SetFactory(c, func(ic context.Context) info.RawInterface {
			return giMock{dummy.Info(), "", time.Now(), nil}
		})
		c = setUpGitiles(c)

		c = gologger.StdConfig.Use(c)
		c = authtest.MockAuthConfig(c)

		mockCtrl := gomock.NewController(t)
		bbMock := mock.NewMockBuildbotClient(mockCtrl)
		biMock := mock.NewMockBuildInfoClient(mockCtrl)

		c = client.WithMiloBuildbot(c, bbMock)
		c = client.WithMiloBuildInfo(c, biMock)

		c = client.WithReader(c, testhelper.MockReader{
			BuildExtracts: map[string]*messages.BuildExtract{
				"chromium": {},
			},
		})

		w := httptest.NewRecorder()
		r := makeGetRequest()
		c = info.SetFactory(c, func(ic context.Context) info.RawInterface {
			return giMock{dummy.Info(), "", time.Now(), nil}
		})
		c = setUpGitiles(c)

		ctx := &router.Context{
			Context: c,
			Writer:  w,
			Request: r,
			Params:  makeParams("tree", "chromium"),
		}
		_, _ = generateAlerts(ctx)

		So(w.Code, ShouldEqual, http.StatusOK)
	})

	Convey("ok request, no gitiles", t, func() {
		c := newTestContext()
		c = info.SetFactory(c, func(ic context.Context) info.RawInterface {
			return giMock{dummy.Info(), "", time.Now(), nil}
		})
		c = urlfetch.Set(c, &testhelper.MockGitilesTransport{})
		c = client.WithReader(c, testhelper.MockReader{
			BuildExtracts: map[string]*messages.BuildExtract{
				"chromium": {},
			},
		})
		w := httptest.NewRecorder()
		r := makeGetRequest()

		ctx := &router.Context{
			Context: c,
			Writer:  w,
			Request: r,
			Params:  makeParams("tree", "chromium"),
		}
		_, err := generateAlerts(ctx)

		So(err, ShouldNotBeNil)
	})
}

func TestStoreAlertsSummary(t *testing.T) {
	Convey("success", t, func() {
		c := gaetesting.TestingContext()
		c = info.SetFactory(c, func(ic context.Context) info.RawInterface {
			return giMock{dummy.Info(), "", clock.Now(c), nil}
		})
		c = setUpGitiles(c)
		a := analyzer.New(5, 100)
		err := storeAlertsSummary(c, a, "some tree", &messages.AlertsSummary{
			Alerts: []messages.Alert{
				{
					Title: "foo",
					Extension: messages.BuildFailure{
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

func TestEnqueueLogDiffTask(t *testing.T) {
	Convey("success", t, func() {
		c := gaetesting.TestingContext()
		c = gologger.StdConfig.Use(c)
		c = info.SetFactory(c, func(ic context.Context) info.RawInterface {
			return mck{giMock{dummy.Info(), "", time.Now(), nil}}
		})
		tqt := tq.GetTestable(c)
		tqt.CreateQueue("logdiff")
		alerts := []messages.Alert{
			{
				Title: "foo",
				Extension: messages.BuildFailure{
					RegressionRanges: []*messages.RegressionRange{
						{Repo: "some repo", URL: "about:blank", Positions: []string{}, Revisions: []string{}},
					},
					Builders: []messages.AlertedBuilder{
						{Name: "chromium.test", URL: "https://uberchromegw.corp.google.com/i/chromium.webkit/builders/WebKit%20Win7%20%28dbg%29", FirstFailure: 15038, LatestFailure: 15038},
					},
					StepAtFault: &messages.BuildStep{
						Step: &messages.Step{
							Name: "test",
						},
					},
				},
			},
		}
		err := enqueueLogDiffTask(c, alerts)
		So(err, ShouldBeNil)
	})

	Convey("fail with non existing queue", t, func() {
		c := gaetesting.TestingContext()
		c = info.SetFactory(c, func(ic context.Context) info.RawInterface {
			return mck{giMock{dummy.Info(), "", time.Now(), nil}}
		})
		tqt := tq.GetTestable(c)
		tqt.CreateQueue("badqueue")
		alerts := []messages.Alert{
			{
				Title: "foo",
				Extension: messages.BuildFailure{
					RegressionRanges: []*messages.RegressionRange{
						{Repo: "some repo", URL: "about:blank", Positions: []string{}, Revisions: []string{}},
					},
					Builders: []messages.AlertedBuilder{
						{Name: "chromium.test", URL: "https://uberchromegw.corp.google.com/i/chromium.webkit/builders/WebKit%20Win7%20%28dbg%29", FirstFailure: 15038, LatestFailure: 15038},
					},
					StepAtFault: &messages.BuildStep{
						Step: &messages.Step{
							Name: "test",
						},
					},
				},
			},
		}
		err := enqueueLogDiffTask(c, alerts)
		So(err, ShouldNotBeNil)
	})
}

type fakeReasonRaw struct {
	signature string
	title     string
}

func (f *fakeReasonRaw) Signature() string {
	if f.signature != "" {
		return f.signature
	}

	return "fakeSignature"
}

func (f *fakeReasonRaw) Kind() string {
	return "fakeKind"
}

func (f *fakeReasonRaw) Title([]*messages.BuildStep) string {
	if f.title == "" {
		return "fakeTitle"
	}
	return f.title
}

func (f *fakeReasonRaw) Severity() messages.Severity {
	return messages.NewFailure
}

func TestMergeAlertsByReason(t *testing.T) {
	Convey("test MergeAlertsByReason", t, func() {
		c := newTestContext()
		w := httptest.NewRecorder()

		ctx := &router.Context{
			Context: c,
			Writer:  w,
			Request: makeGetRequest(),
			Params:  makeParams("tree", "unknown.tree"),
		}

		tests := []struct {
			name    string
			in      []messages.Alert
			want    []model.Annotation
			wantErr error
		}{
			{
				name: "empty",
				want: []model.Annotation{},
			},
			{
				name: "no merges",
				in: []messages.Alert{
					{
						Type: messages.AlertBuildFailure,
						Extension: messages.BuildFailure{
							Reason: &messages.Reason{
								Raw: &fakeReasonRaw{
									signature: "reason_a",
								},
							},
						},
						Key: "a",
					},
					{
						Type: messages.AlertBuildFailure,
						Extension: messages.BuildFailure{
							Reason: &messages.Reason{
								Raw: &fakeReasonRaw{
									signature: "reason_b",
								},
							},
						},
						Key: "b",
					},
				},
				want: []model.Annotation{},
			},
			{
				name: "multiple builders fail on bad_test",
				in: []messages.Alert{
					{
						Type: messages.AlertBuildFailure,
						Extension: messages.BuildFailure{
							Reason: &messages.Reason{
								Raw: &fakeReasonRaw{
									signature: "bad_test",
								},
							},
						},
						Key: "buildera.bad_test",
					},
					{
						Type: messages.AlertBuildFailure,
						Extension: messages.BuildFailure{
							Reason: &messages.Reason{
								Raw: &fakeReasonRaw{
									signature: "bad_test",
								},
							},
						},
						Key: "builderb.bad_test",
					},
					{
						Type: messages.AlertBuildFailure,
						Extension: messages.BuildFailure{
							Reason: &messages.Reason{
								Raw: &fakeReasonRaw{
									signature: "bad_test",
								},
							},
						},
						Key: "builderc.bad_test",
					},
				},
				want: []model.Annotation{
					{
						Tree:      datastore.MakeKey(c, "Tree", "unknown.tree"),
						KeyDigest: fmt.Sprintf("%x", sha1.Sum([]byte("buildera.bad_test"))),
						Key:       "buildera.bad_test",
						GroupID:   "fakeTitle",
					},
					{
						Tree:      datastore.MakeKey(c, "Tree", "unknown.tree"),
						KeyDigest: fmt.Sprintf("%x", sha1.Sum([]byte("builderb.bad_test"))),
						Key:       "builderb.bad_test",
						GroupID:   "fakeTitle",
					},
					{
						Tree:      datastore.MakeKey(c, "Tree", "unknown.tree"),
						KeyDigest: fmt.Sprintf("%x", sha1.Sum([]byte("builderc.bad_test"))),
						Key:       "builderc.bad_test",
						GroupID:   "fakeTitle",
					},
				},
			},
		}

		for _, test := range tests {
			test := test
			Convey(test.name, func() {
				groups, err := mergeAlertsByReason(ctx, test.in)
				So(err, ShouldResemble, test.wantErr)
				So(groups, ShouldNotBeNil)

				allAnns := []model.Annotation{}
				q := datastore.NewQuery("Annotation")
				So(datastore.GetAll(c, q, &allAnns), ShouldBeNil)

				sort.Sort(annList(allAnns))
				sort.Sort(annList(test.want))
				So(allAnns, ShouldResemble, test.want)
			})
		}
	})
}

type annList []model.Annotation

func (a annList) Len() int {
	return len(a)
}

func (a annList) Less(i, j int) bool {
	return a[i].Key < a[j].Key
}

func (a annList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func TestAttachTestResults(t *testing.T) {
	Convey("basic", t, func() {

		c := newTestContext()
		c = authtest.MockAuthConfig(c)
		fakeTRServer := testhelper.NewFakeServer()
		defer fakeTRServer.Server.Close()

		testHistory := map[string]interface{}{
			"test-builder": &client.BuilderTestHistory{
				BuildNumbers:   []int64{10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
				ChromeRevision: []string{"10", "9", "8", "7", "6", "5", "4", "3", "2", "1", "0"},
				Tests: map[string]*client.TestResultHistory{
					"test 1": {
						Results: [][]interface{}{
							{float64(5), "B"},
							{float64(5), "A"},
						},
					},
				},
				FailureMap: map[string]string{
					"A": "PASS",
					"B": "FAIL",
				},
			},
		}
		fakeTRServer.JSONResponse = testHistory

		c = client.WithTestResults(c, fakeTRServer.Server.URL)
		Convey("empty alert", func() {
			alert := &messages.Alert{}
			err := attachTestResults(c, alert)
			So(err, ShouldNotBeNil)
		})

		Convey("with test failure", func() {
			alert := &messages.Alert{
				Extension: messages.BuildFailure{
					Builders: []messages.AlertedBuilder{
						{
							Master:        "master",
							Name:          "test-builder",
							LatestFailure: 10,
							FirstFailure:  8,
							LatestPassing: 7,
						},
					},
					StepAtFault: &messages.BuildStep{
						Step: &messages.Step{
							Name: "test step",
						},
					},
					Reason: &messages.Reason{
						Raw: &step.TestFailure{
							TestNames: []string{"test 1"},
							StepName:  "test step",
						},
					},
				},
			}
			err := attachTestResults(c, alert)
			So(err, ShouldBeNil)
			alertTestResults := alert.Extension.(messages.BuildFailure).Reason.Raw.(*step.TestFailure).AlertTestResults
			So(alertTestResults, ShouldNotBeEmpty)
			So(alertTestResults, ShouldResemble, []messages.AlertTestResults{
				{
					TestName: "test 1",
					MasterResults: []messages.MasterResults{
						{
							MasterName: "master",
							BuilderResults: []messages.BuilderResults{
								{
									BuilderName: "test-builder",
									Results: []messages.Results{
										{
											BuildNumber: 10,
											Revision:    "10",
											Actual:      []string{"FAIL"},
										},

										{
											BuildNumber: 9,
											Revision:    "9",
											Actual:      []string{"FAIL"},
										},

										{
											BuildNumber: 8,
											Revision:    "8",
											Actual:      []string{"FAIL"},
										},

										{
											BuildNumber: 7,
											Revision:    "7",
											Actual:      []string{"FAIL"},
										},

										{
											BuildNumber: 6,
											Revision:    "6",
											Actual:      []string{"FAIL"},
										},
									},
								},
							},
						},
					},
				},
			})
		})
	})
}
