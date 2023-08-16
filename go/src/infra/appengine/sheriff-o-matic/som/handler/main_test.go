// Copyright 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handler

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/gae/service/datastore"
	"go.chromium.org/luci/server/auth/authtest"
	"go.chromium.org/luci/server/auth/xsrf"
	"go.chromium.org/luci/server/router"

	"infra/appengine/sheriff-o-matic/som/model"
	"infra/monitoring/messages"
)

var _ = fmt.Printf

func TestMain(t *testing.T) {
	Convey("main", t, func() {
		c := gaetesting.TestingContext()
		c = authtest.MockAuthConfig(c)
		c = gologger.StdConfig.Use(c)

		cl := testclock.New(testclock.TestRecentTimeUTC)
		c = clock.Set(c, cl)

		w := httptest.NewRecorder()

		monorailMux := http.NewServeMux()
		monorailServer := httptest.NewServer(monorailMux)
		defer monorailServer.Close()
		tok, err := xsrf.Token(c)
		So(err, ShouldBeNil)
		Convey("/api/v1", func() {
			alertIdx := datastore.IndexDefinition{
				Kind:     "AlertJSONNonGrouping",
				Ancestor: true,
				SortBy: []datastore.IndexColumn{
					{
						Property: "Resolved",
					},
					{
						Property:   "Date",
						Descending: false,
					},
				},
			}
			revisionSummaryIdx := datastore.IndexDefinition{
				Kind:     "RevisionSummaryJSON",
				Ancestor: true,
				SortBy: []datastore.IndexColumn{
					{
						Property:   "Date",
						Descending: false,
					},
				},
			}
			indexes := []*datastore.IndexDefinition{&alertIdx, &revisionSummaryIdx}
			datastore.GetTestable(c).AddIndexes(indexes...)

			Convey("GetTrees", func() {
				Convey("no trees yet", func() {
					trees, err := GetTrees(c)

					So(err, ShouldBeNil)
					So(string(trees), ShouldEqual, "[]")
				})

				tree := &model.Tree{
					Name:        "oak",
					DisplayName: "Oak",
				}
				So(datastore.Put(c, tree), ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				Convey("basic tree", func() {
					trees, err := GetTrees(c)

					So(err, ShouldBeNil)
					So(string(trees), ShouldEqual, `[{"name":"oak","display_name":"Oak","bb_project_filter":""}]`)
				})
			})

			Convey("/alerts", func() {
				contents, _ := json.Marshal(&messages.Alert{
					Key: "test",
				})
				alertJSON := &model.AlertJSON{
					ID:       "test",
					Tree:     datastore.MakeKey(c, "Tree", "chromeos"),
					Resolved: false,
					Date:     time.Unix(1, 0).UTC(),
					Contents: []byte(contents),
				}
				contents2, _ := json.Marshal(&messages.Alert{
					Key: "test2",
				})
				oldResolvedJSON := &model.AlertJSON{
					ID:       "test2",
					Tree:     datastore.MakeKey(c, "Tree", "chromeos"),
					Resolved: true,
					Date:     time.Unix(1, 0).UTC(),
					Contents: []byte(contents2),
				}
				contents3, _ := json.Marshal(&messages.Alert{
					Key: "test3",
				})
				newResolvedJSON := &model.AlertJSON{
					ID:       "test3",
					Tree:     datastore.MakeKey(c, "Tree", "chromeos"),
					Resolved: true,
					Date:     clock.Now(c),
					Contents: []byte(contents3),
				}

				Convey("GET", func() {
					Convey("no alerts yet", func() {
						GetAlertsHandler(&router.Context{
							Writer:  w,
							Request: makeGetRequest(c),
							Params:  makeParams("tree", "chromeos"),
						})

						_, err := ioutil.ReadAll(w.Body)
						So(err, ShouldBeNil)
						So(w.Code, ShouldEqual, 200)
					})

					So(datastorePutAlertJSON(c, alertJSON), ShouldBeNil)
					datastore.GetTestable(c).CatchupIndexes()

					Convey("basic alerts", func() {
						GetAlertsHandler(&router.Context{
							Writer:  w,
							Request: makeGetRequest(c),
							Params:  makeParams("tree", "chromeos"),
						})

						r, err := ioutil.ReadAll(w.Body)
						So(err, ShouldBeNil)
						So(w.Code, ShouldEqual, 200)
						summary := &messages.AlertsSummary{}
						err = json.Unmarshal(r, &summary)
						So(err, ShouldBeNil)
						So(summary.Alerts, ShouldHaveLength, 1)
						So(summary.Alerts[0].Key, ShouldEqual, "test")
						So(summary.Resolved, ShouldHaveLength, 0)
						// TODO(seanmccullough): Remove all of the POST /alerts handling
						// code and tests except for whatever chromeos needs.
					})

					So(datastorePutAlertJSON(c, oldResolvedJSON), ShouldBeNil)
					So(datastorePutAlertJSON(c, newResolvedJSON), ShouldBeNil)

					Convey("resolved alerts", func() {
						GetAlertsHandler(&router.Context{
							Writer:  w,
							Request: makeGetRequest(c),
							Params:  makeParams("tree", "chromeos"),
						})

						r, err := ioutil.ReadAll(w.Body)
						So(err, ShouldBeNil)
						So(w.Code, ShouldEqual, 200)
						summary := &messages.AlertsSummary{}
						err = json.Unmarshal(r, &summary)
						So(err, ShouldBeNil)
						So(summary.Alerts, ShouldHaveLength, 1)
						So(summary.Alerts[0].Key, ShouldEqual, "test")
						So(summary.Resolved, ShouldHaveLength, 1)
						So(summary.Resolved[0].Key, ShouldEqual, "test3")
						// TODO(seanmccullough): Remove all of the POST /alerts handling
						// code and tests except for whatever chromeos needs.
					})
				})
			})

			Convey("/unresolved", func() {
				contents, _ := json.Marshal(&messages.Alert{
					Key: "test",
				})
				alertJSON := &model.AlertJSON{
					ID:       "test",
					Tree:     datastore.MakeKey(c, "Tree", "chromeos"),
					Resolved: false,
					Date:     time.Unix(1, 0).UTC(),
					Contents: []byte(contents),
				}
				contents2, _ := json.Marshal(&messages.Alert{
					Key: "test2",
				})
				oldResolvedJSON := &model.AlertJSON{
					ID:       "test2",
					Tree:     datastore.MakeKey(c, "Tree", "chromeos"),
					Resolved: true,
					Date:     time.Unix(1, 0).UTC(),
					Contents: []byte(contents2),
				}
				contents3, _ := json.Marshal(&messages.Alert{
					Key: "test3",
				})
				newResolvedJSON := &model.AlertJSON{
					ID:       "test3",
					Tree:     datastore.MakeKey(c, "Tree", "chromeos"),
					Resolved: true,
					Date:     clock.Now(c),
					Contents: []byte(contents3),
				}

				Convey("GET", func() {
					Convey("no alerts yet", func() {
						GetUnresolvedAlertsHandler(&router.Context{
							Writer:  w,
							Request: makeGetRequest(c),
							Params:  makeParams("tree", "chromeos"),
						})

						_, err := ioutil.ReadAll(w.Body)
						So(err, ShouldBeNil)
						So(w.Code, ShouldEqual, 200)
					})

					So(datastorePutAlertJSON(c, alertJSON), ShouldBeNil)
					So(datastorePutAlertJSON(c, oldResolvedJSON), ShouldBeNil)
					So(datastorePutAlertJSON(c, newResolvedJSON), ShouldBeNil)
					datastore.GetTestable(c).CatchupIndexes()

					Convey("basic alerts", func() {
						GetUnresolvedAlertsHandler(&router.Context{
							Writer:  w,
							Request: makeGetRequest(c),
							Params:  makeParams("tree", "chromeos"),
						})

						r, err := ioutil.ReadAll(w.Body)
						So(err, ShouldBeNil)
						So(w.Code, ShouldEqual, 200)
						summary := &messages.AlertsSummary{}
						err = json.Unmarshal(r, &summary)
						So(err, ShouldBeNil)
						So(summary.Alerts, ShouldHaveLength, 1)
						So(summary.Alerts[0].Key, ShouldEqual, "test")
						So(summary.Resolved, ShouldBeNil)
					})
				})
			})

			Convey("/resolved", func() {
				contents, _ := json.Marshal(&messages.Alert{
					Key: "test",
				})
				alertJSON := &model.AlertJSON{
					ID:       "test",
					Tree:     datastore.MakeKey(c, "Tree", "chromeos"),
					Resolved: false,
					Date:     time.Unix(1, 0).UTC(),
					Contents: []byte(contents),
				}
				contents2, _ := json.Marshal(&messages.Alert{
					Key: "test2",
				})
				oldResolvedJSON := &model.AlertJSON{
					ID:       "test2",
					Tree:     datastore.MakeKey(c, "Tree", "chromeos"),
					Resolved: true,
					Date:     time.Unix(1, 0).UTC(),
					Contents: []byte(contents2),
				}
				contents3, _ := json.Marshal(&messages.Alert{
					Key: "test3",
				})
				newResolvedJSON := &model.AlertJSON{
					ID:       "test3",
					Tree:     datastore.MakeKey(c, "Tree", "chromeos"),
					Resolved: true,
					Date:     clock.Now(c),
					Contents: []byte(contents3),
				}

				Convey("GET", func() {
					Convey("no alerts yet", func() {
						GetResolvedAlertsHandler(&router.Context{
							Writer:  w,
							Request: makeGetRequest(c),
							Params:  makeParams("tree", "chromeos"),
						})

						_, err := ioutil.ReadAll(w.Body)
						So(err, ShouldBeNil)
						So(w.Code, ShouldEqual, 200)
					})

					So(datastorePutAlertJSON(c, alertJSON), ShouldBeNil)
					So(datastorePutAlertJSON(c, oldResolvedJSON), ShouldBeNil)
					So(datastorePutAlertJSON(c, newResolvedJSON), ShouldBeNil)
					datastore.GetTestable(c).CatchupIndexes()

					Convey("resolved alerts", func() {
						GetResolvedAlertsHandler(&router.Context{
							Writer:  w,
							Request: makeGetRequest(c),
							Params:  makeParams("tree", "chromeos"),
						})

						r, err := ioutil.ReadAll(w.Body)
						So(err, ShouldBeNil)
						So(w.Code, ShouldEqual, 200)
						summary := &messages.AlertsSummary{}
						err = json.Unmarshal(r, &summary)
						So(err, ShouldBeNil)
						So(summary.Alerts, ShouldBeNil)
						So(summary.Resolved, ShouldHaveLength, 1)
						So(summary.Resolved[0].Key, ShouldEqual, "test3")
						// TODO(seanmccullough): Remove all of the POST /alerts handling
						// code and tests except for whatever chromeos needs.
					})
				})
			})
		})

		Convey("cron", func() {
			Convey("flushOldAnnotations", func() {
				getAllAnns := func() []*model.Annotation {
					anns := []*model.Annotation{}
					q := datastoreCreateAnnotationQuery()
					So(datastoreGetAnnotationsByQuery(c, &anns, q), ShouldBeNil)
					return anns
				}

				ann := &model.Annotation{
					KeyDigest:        fmt.Sprintf("%x", sha1.Sum([]byte("foobar"))),
					Key:              "foobar",
					ModificationTime: datastore.RoundTime(cl.Now()),
				}
				So(datastorePutAnnotation(c, ann), ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				Convey("current not deleted", func() {
					num, err := flushOldAnnotations(c)
					So(err, ShouldBeNil)
					So(num, ShouldEqual, 0)
					So(getAllAnns(), ShouldResemble, []*model.Annotation{ann})
				})

				ann.ModificationTime = cl.Now().Add(-(annotationExpiration + time.Hour))
				So(datastorePutAnnotation(c, ann), ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				Convey("old deleted", func() {
					num, err := flushOldAnnotations(c)
					So(err, ShouldBeNil)
					So(num, ShouldEqual, 1)
					So(getAllAnns(), ShouldResemble, []*model.Annotation{})
				})

				datastore.GetTestable(c).CatchupIndexes()
				q := datastoreCreateAnnotationQuery()
				anns := []*model.Annotation{}
				datastore.GetTestable(c).CatchupIndexes()
				datastoreGetAnnotationsByQuery(c, &anns, q)
				datastoreDeleteAnnotations(c, anns)
				anns = []*model.Annotation{
					{
						KeyDigest:        fmt.Sprintf("%x", sha1.Sum([]byte("foobar2"))),
						Key:              "foobar2",
						ModificationTime: datastore.RoundTime(cl.Now()),
					},
					{
						KeyDigest:        fmt.Sprintf("%x", sha1.Sum([]byte("foobar"))),
						Key:              "foobar",
						ModificationTime: datastore.RoundTime(cl.Now().Add(-(annotationExpiration + time.Hour))),
					},
				}
				So(datastorePutAnnotations(c, anns), ShouldBeNil)
				datastore.GetTestable(c).CatchupIndexes()

				Convey("only delete old", func() {
					num, err := flushOldAnnotations(c)
					So(err, ShouldBeNil)
					So(num, ShouldEqual, 1)
					So(getAllAnns(), ShouldResemble, anns[:1])
				})

				Convey("handler", func() {
					FlushOldAnnotationsHandler(c)
				})
			})

			Convey("clientmon", func() {
				body := &eCatcherReq{XSRFToken: tok}
				bodyBytes, err := json.Marshal(body)
				So(err, ShouldBeNil)
				ctx := &router.Context{
					Writer:  w,
					Request: makePostRequest(c, string(bodyBytes)),
					Params:  makeParams("xsrf_token", tok),
				}

				PostClientMonHandler(ctx)
				So(w.Code, ShouldEqual, 200)
			})

			Convey("treelogo", func() {
				ctx := &router.Context{
					Writer:  w,
					Request: makeGetRequest(c),
					Params:  makeParams("tree", "chromium"),
				}

				getTreeLogo(ctx, "", &noopSigner{})
				So(w.Code, ShouldEqual, 302)
			})

			Convey("treelogo fail", func() {
				ctx := &router.Context{
					Writer:  w,
					Request: makeGetRequest(c),
					Params:  makeParams("tree", "chromium"),
				}

				getTreeLogo(ctx, "", &noopSigner{fmt.Errorf("fail")})
				So(w.Code, ShouldEqual, 500)
			})
		})
	})
}

type noopSigner struct {
	err error
}

func (n *noopSigner) SignBytes(c context.Context, b []byte) (string, []byte, error) {
	return string(b), b, n.err
}

func makeGetRequest(ctx context.Context, queryParams ...string) *http.Request {
	if len(queryParams)%2 != 0 {
		return nil
	}
	params := make([]string, len(queryParams)/2)
	for i := range params {
		params[i] = fmt.Sprintf("%s=%s", queryParams[2*i], queryParams[2*i+1])
	}
	paramsStr := strings.Join(params, "&")
	url := fmt.Sprintf("/doesntmatter?%s", paramsStr)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	return req
}

func makePostRequest(ctx context.Context, body string) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, "POST", "/doesntmatter", strings.NewReader(body))
	return req
}

func makeParams(items ...string) httprouter.Params {
	if len(items)%2 != 0 {
		return nil
	}

	params := make([]httprouter.Param, len(items)/2)
	for i := range params {
		params[i] = httprouter.Param{
			Key:   items[2*i],
			Value: items[2*i+1],
		}
	}

	return params
}
