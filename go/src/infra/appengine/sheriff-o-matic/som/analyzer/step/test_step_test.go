// Copyright 2016 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package step

import (
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/context"

	"infra/appengine/sheriff-o-matic/som/client"
	testhelper "infra/appengine/sheriff-o-matic/som/client/test"
	te "infra/appengine/sheriff-o-matic/som/testexpectations"
	"infra/appengine/test-results/model"
	"infra/monitoring/messages"

	"go.chromium.org/gae/impl/dummy"
	"go.chromium.org/gae/service/info"
	"go.chromium.org/gae/service/urlfetch"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/server/auth/authtest"

	. "github.com/smartystreets/goconvey/convey"
)

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
	data, _ := json.Marshal(map[string]*te.BuilderConfig{})

	return urlfetch.Set(c, &testhelper.MockGitilesTransport{
		Responses: map[string]string{
			"https://chromium.googlesource.com/chromium/src/+/master/third_party/blink/tools/blinkpy/common/config/builders.json?format=TEXT": string(data),
		},
	})
}

func TestTestStepFailureAlerts(t *testing.T) {
	tr, fa := true, false
	truePtr, falsePtr := &tr, &fa

	Convey("test TestFailureAnalyzer", t, func() {
		maxFailedTests = 2
		Convey("analyze", func() {
			tests := []struct {
				name          string
				failures      []*messages.BuildStep
				testResults   *model.FullResult
				finditResults []*messages.FinditResult
				wantResult    []messages.ReasonRaw
				wantErr       error
			}{
				{
					name:       "empty",
					wantResult: []messages.ReasonRaw{},
				},
				{
					name: "non-test failure",
					failures: []*messages.BuildStep{
						{
							Master: &messages.MasterLocation{URL: url.URL{
								Scheme: "https",
								Host:   "build.chromium.org",
								Path:   "/p/fake.Master",
							}},
							Build: &messages.Build{
								BuilderName: "fake_builder",
							},
							Step: &messages.Step{
								Name: "tests_compile",
							},
						},
					},
					wantResult: []messages.ReasonRaw{
						nil,
					},
				},
				{
					name: "test step failure",
					failures: []*messages.BuildStep{
						{
							Master: &messages.MasterLocation{URL: url.URL{
								Scheme: "https",
								Host:   "build.chromium.org",
								Path:   "/p/fake.Master",
							}},
							Build: &messages.Build{
								BuilderName: "fake_builder",
							},
							Step: &messages.Step{
								Name: "something_tests",
							},
						},
					},
					testResults: &model.FullResult{
						Tests: model.FullTest{
							"test_a": &model.FullTestLeaf{
								Expected:   []string{"PASS"},
								Actual:     []string{"FAIL"},
								Unexpected: truePtr,
							},
						},
					},
					finditResults: []*messages.FinditResult{
						{
							TestName:    "test_a",
							IsFlakyTest: false,
							SuspectedCLs: []messages.SuspectCL{
								{
									RepoName:       "repo",
									Revision:       "deadbeef",
									CommitPosition: 1234,
								},
							},
						},
					},
					wantResult: []messages.ReasonRaw{
						&TestFailure{
							TestNames: []string{"test_a"},
							StepName:  "something_tests",
							Tests: []TestWithResult{
								{
									TestName: "test_a",
									IsFlaky:  false,
									SuspectedCLs: []messages.SuspectCL{
										{
											RepoName:       "repo",
											Revision:       "deadbeef",
											CommitPosition: 1234,
										},
									},
									Artifacts: []ArtifactLink{},
								},
							},
						},
					},
				},
				{
					name: "test step failure (too many failures)",
					failures: []*messages.BuildStep{
						{
							Master: &messages.MasterLocation{URL: url.URL{
								Scheme: "https",
								Host:   "build.chromium.org",
								Path:   "/p/fake.Master",
							}},
							Build: &messages.Build{
								BuilderName: "fake_builder",
							},
							Step: &messages.Step{
								Name: "something_tests",
							},
						},
					},
					testResults: &model.FullResult{
						Tests: model.FullTest{
							"test_a": &model.FullTestLeaf{
								Expected:   []string{"PASS"},
								Actual:     []string{"FAIL"},
								Unexpected: truePtr,
							},
							"test_b": &model.FullTestLeaf{
								Expected:   []string{"PASS"},
								Actual:     []string{"FAIL"},
								Unexpected: truePtr,
							},
							"test_c": &model.FullTestLeaf{
								Expected:   []string{"PASS"},
								Actual:     []string{"FAIL"},
								Unexpected: truePtr,
							},
						},
					},
					finditResults: []*messages.FinditResult{
						{
							TestName:    "test_a",
							IsFlakyTest: false,
							SuspectedCLs: []messages.SuspectCL{
								{
									RepoName:       "repo",
									Revision:       "deadbeef",
									CommitPosition: 1234,
								},
							},
						},
						{
							TestName:    "test_b",
							IsFlakyTest: false,
							SuspectedCLs: []messages.SuspectCL{
								{
									RepoName:       "repo",
									Revision:       "deadbeef",
									CommitPosition: 1234,
								},
							},
						},
					},
					wantResult: []messages.ReasonRaw{
						&TestFailure{
							TestNames: []string{tooManyFailuresText, "test_a", "test_b"},
							StepName:  "something_tests",
							Tests: []TestWithResult{
								{
									TestName: "test_a",
									IsFlaky:  false,
									SuspectedCLs: []messages.SuspectCL{
										{
											RepoName:       "repo",
											Revision:       "deadbeef",
											CommitPosition: 1234,
										},
									},
									Artifacts: []ArtifactLink{},
								},
								{
									TestName: "test_b",
									IsFlaky:  false,
									SuspectedCLs: []messages.SuspectCL{
										{
											RepoName:       "repo",
											Revision:       "deadbeef",
											CommitPosition: 1234,
										},
									},
									Artifacts: []ArtifactLink{},
								},
								{
									TestName:  tooManyFailuresText,
									IsFlaky:   false,
									Artifacts: []ArtifactLink{},
								},
							},
						},
					},
				},
				{
					name: "test step failure with weird step name",
					failures: []*messages.BuildStep{
						{
							Master: &messages.MasterLocation{URL: url.URL{
								Scheme: "https",
								Host:   "build.chromium.org",
								Path:   "/p/fake.Master",
							}},
							Build: &messages.Build{
								BuilderName: "fake_builder",
							},
							Step: &messages.Step{
								Name: "something_tests on windows_7",
							},
						},
					},
					testResults: &model.FullResult{
						Tests: model.FullTest{
							"test_a": &model.FullTestLeaf{
								Expected:   []string{"PASS"},
								Actual:     []string{"FAIL"},
								Unexpected: truePtr,
							},
						},
					},
					finditResults: []*messages.FinditResult{
						{
							TestName:    "test_a",
							IsFlakyTest: false,
							SuspectedCLs: []messages.SuspectCL{
								{
									RepoName:       "repo",
									Revision:       "deadbeef",
									CommitPosition: 1234,
								},
							},
						},
					},
					wantResult: []messages.ReasonRaw{
						&TestFailure{
							TestNames: []string{"test_a"},
							StepName:  "something_tests",
							Tests: []TestWithResult{
								{
									TestName: "test_a",
									IsFlaky:  false,
									SuspectedCLs: []messages.SuspectCL{
										{
											RepoName:       "repo",
											Revision:       "deadbeef",
											CommitPosition: 1234,
										},
									},
									Artifacts: []ArtifactLink{},
								},
							},
						},
					},
				},
				{
					name: "test step failure with weird step name on perf",
					failures: []*messages.BuildStep{
						{
							Master: &messages.MasterLocation{URL: url.URL{
								Scheme: "https",
								Host:   "build.chromium.org",
								Path:   "/p/chromium.perf",
							}},
							Build: &messages.Build{
								BuilderName: "fake_builder",
							},
							Step: &messages.Step{
								Name: "something_tests on windows_7 on Intel GPU",
								Logs: [][]interface{}{
									{
										"swarming.summary",
										"foo",
									},
								},
							},
						},
					},
					wantResult: []messages.ReasonRaw{
						&TestFailure{
							TestNames: []string{},
							StepName:  "something_tests",
						},
					},
				},
				{
					name: "flaky",
					failures: []*messages.BuildStep{
						{
							Master: &messages.MasterLocation{URL: url.URL{
								Scheme: "https",
								Host:   "build.chromium.org",
								Path:   "/p/fake.Master",
							}},
							Build: &messages.Build{
								BuilderName: "fake_builder",
							},
							Step: &messages.Step{
								Name: "something_tests",
							},
						},
					},
					testResults: &model.FullResult{
						Tests: model.FullTest{
							"test_a": &model.FullTestLeaf{
								Expected:   []string{"PASS"},
								Actual:     []string{"FAIL", "PASS"},
								Unexpected: falsePtr,
							},
						},
					},
					wantResult: []messages.ReasonRaw{
						&TestFailure{
							TestNames: []string{},
							StepName:  "something_tests",
						},
					},
				},
				{
					name: "test findit found flaky",
					failures: []*messages.BuildStep{
						{
							Master: &messages.MasterLocation{URL: url.URL{
								Scheme: "https",
								Host:   "build.chromium.org",
								Path:   "/p/fake.Master",
							}},
							Build: &messages.Build{
								BuilderName: "fake_builder",
							},
							Step: &messages.Step{
								Name: "something_tests",
							},
						},
					},
					testResults: &model.FullResult{
						Tests: model.FullTest{
							"test_a": &model.FullTestLeaf{
								Expected:   []string{"PASS"},
								Actual:     []string{"FAIL"},
								Unexpected: truePtr,
							},
						},
					},
					finditResults: []*messages.FinditResult{
						{
							TestName:     "test_a",
							IsFlakyTest:  true,
							SuspectedCLs: []messages.SuspectCL{},
						},
					},
					wantResult: []messages.ReasonRaw{
						&TestFailure{
							TestNames: []string{"test_a"},
							StepName:  "something_tests",
							Tests: []TestWithResult{
								{
									TestName:     "test_a",
									IsFlaky:      true,
									SuspectedCLs: []messages.SuspectCL{},
									Artifacts:    []ArtifactLink{},
								},
							},
						},
					},
				},
			}

			for _, test := range tests {

				test := test
				Convey(test.name, func() {
					c := gaetesting.TestingContext()
					c = authtest.MockAuthConfig(c)
					c = gologger.StdConfig.Use(c)

					testResultsFake := testhelper.NewFakeServer()
					defer testResultsFake.Server.Close()
					finditFake := testhelper.NewFakeServer()
					defer finditFake.Server.Close()
					te.LayoutTestExpectations = map[string]string{}

					c = info.SetFactory(c, func(ic context.Context) info.RawInterface {
						return giMock{dummy.Info(), "", time.Now(), nil}
					})

					c = setUpGitiles(c)
					c = client.WithFindit(c, finditFake.Server.URL)

					testResultsFake.JSONResponse = test.testResults

					// knownResults determines what tests the test results server knows about. Set
					// this up so that we don't quit early.
					knownResults := model.BuilderData{}
					if test.testResults != nil {
						for _, failure := range test.failures {
							knownResults.Masters = append(knownResults.Masters, model.Master{
								Name: failure.Master.Name(),
								Tests: map[string]*model.Test{
									GetTestSuite(failure): {
										Builders: []string{failure.Build.BuilderName},
									},
								},
							})
						}
					}

					testResultsFake.PerURLResponse = map[string]interface{}{
						"/data/builders": knownResults,
					}

					c = client.WithTestResults(c, testResultsFake.Server.URL)
					finditFake.JSONResponse = &client.FinditAPIResponse{Results: test.finditResults}
					gotResult, gotErr := testFailureAnalyzer(c, test.failures, "chromium.perf")
					So(gotErr, ShouldEqual, test.wantErr)
					So(gotResult, ShouldResemble, test.wantResult)
				})
			}
		})
	})
}

func TestUnexpected(t *testing.T) {
	Convey("unexpected", t, func() {
		r := &model.FullResult{}
		So(unexpectedFailures(r), ShouldBeEmpty)

		tptr := true
		fptr := false
		Convey("pass", func() {
			r := &model.FullResult{
				Tests: model.FullTest{
					"foo": &model.FullTestLeaf{
						Actual:     []string{"PASS"},
						Expected:   []string{"PASS"},
						Unexpected: &fptr,
					},
				},
			}
			So(unexpectedFailures(r), ShouldBeEmpty)
		})

		Convey("fail", func() {
			r := &model.FullResult{
				Tests: model.FullTest{
					"foo": &model.FullTestLeaf{
						Actual:     []string{"FAIL"},
						Expected:   []string{"PASS"},
						Unexpected: &tptr,
					},
				},
			}
			So(unexpectedFailures(r), ShouldResemble, []string{"foo"})
		})

		Convey("jbudorik's case", func() {
			r := &model.FullResult{
				Tests: model.FullTest{
					"/a/fetch-event-respond-with-readable-stream-chunk.https.html": &model.FullTestLeaf{
						Actual:     []string{"PASS"},
						Bugs:       []string{"crbug.com/807954"},
						Expected:   []string{"CRASH"},
						Unexpected: &tptr,
					},
					"/b/fetch-event-respond-with-readable-stream-chunk.https.html": &model.FullTestLeaf{
						Actual:     []string{"PASS"},
						Bugs:       []string{"crbug.com/807954"},
						Expected:   []string{"CRASH"},
						Unexpected: &tptr,
					},
				},
			}

			So(unexpectedFailures(r), ShouldResemble, []string{})
		})
	})
}

func TestBasicFailure(t *testing.T) {
	Convey("basicFailure", t, func() {
		bf := &basicFailure{Name: "basic"}
		title := bf.Title([]*messages.BuildStep{
			{
				Master: &messages.MasterLocation{},
				Build:  &messages.Build{BuilderName: "basic.builder"},
				Step:   &messages.Step{Name: "step"},
			},
		})

		So(title, ShouldEqual, "step failing on /basic.builder")

		So(bf.Signature(), ShouldEqual, bf.Name)
		So(bf.Kind(), ShouldEqual, "basic")
	})
}

func TestTruncateTestName(t *testing.T) {
	Convey("testTrunc", t, func() {
		t := &TestFailure{
			TestNames: []string{"hi"},
		}

		Convey("basic", func() {
			So(t.testTrunc(), ShouldResemble, "hi")
		})

		Convey("multiple tests", func() {
			t.TestNames = []string{"a", "b"}
			So(t.testTrunc(), ShouldResemble, "a and 1 other(s)")
		})

		Convey("chromium tree example", func() {
			t.TestNames = []string{"virtual/outofblink-cors/http/tests/xmlhttprequest/redirect-cross-origin-post.html"}
			So(t.testTrunc(), ShouldResemble, "virtual/.../redirect-cross-origin-post.html")
		})

		Convey("chromium.perf tree example", func() {
			t.TestNames = []string{"smoothness.top_25_smooth/https://plus.google.com/110031535020051778989/posts"}
			So(t.testTrunc(), ShouldResemble, "smoothness.top_25_smooth/https://plus.google.com/110031535020051778989/posts")
		})
	})
}

func TestGetTestSuite(t *testing.T) {
	Convey("GetTestSuite", t, func() {
		s := &messages.BuildStep{
			Step: &messages.Step{
				Name: "thing_tests",
			},
		}
		url, err := url.Parse("https://build.chromium.org/p/chromium.linux")
		So(err, ShouldBeNil)
		s.Master = &messages.MasterLocation{
			URL: *url,
		}
		Convey("basic", func() {
			So(GetTestSuite(s), ShouldEqual, "thing_tests")
		})
		Convey("with suffixes", func() {
			s.Step.Name = "thing_tests on Intel GPU on Linux"
			So(GetTestSuite(s), ShouldEqual, "thing_tests")
		})
		Convey("on perf", func() {
			url, err := url.Parse("https://build.chromium.org/p/chromium.perf")
			So(err, ShouldBeNil)
			s.Master = &messages.MasterLocation{
				URL: *url,
			}
			s.Step.Logs = [][]interface{}{
				{
					"swarming.summary",
					"foo",
				},
			}
			Convey("with suffixes", func() {
				s.Step.Name = "battor.power_cases on Intel GPU on Linux"
				So(GetTestSuite(s), ShouldEqual, "battor.power_cases")
			})
			Convey("C++ test with suffixes", func() {
				s.Step.Name = "performance_browser_tests on Intel GPU on Linux"
				So(GetTestSuite(s), ShouldEqual, "performance_browser_tests")
			})
			Convey("not a test", func() {
				s.Step.Logs = nil
				s.Step.Name = "something_random"
				So(GetTestSuite(s), ShouldEqual, "")
			})
		})
	})
}
