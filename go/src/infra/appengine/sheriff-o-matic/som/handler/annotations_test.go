package handler

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"infra/appengine/sheriff-o-matic/som/client"
	"infra/appengine/sheriff-o-matic/som/model"
	"infra/monorail"

	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/server/auth/authtest"
	"go.chromium.org/luci/server/auth/xsrf"
	"go.chromium.org/luci/server/router"
	"golang.org/x/net/context"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilterAnnotations(t *testing.T) {
	Convey("Test filter annotation", t, func() {
		activeKeys := map[string]interface{}{
			"alert_1": nil,
			"alert_2": nil,
			"alert_3": nil,
		}

		annotations := []*model.Annotation{
			{
				Key:     "alert_1",
				GroupID: "group_1",
			},
			{
				Key:     "alert_2",
				GroupID: "group_2",
			},
			{
				Key:     "group_2",
				GroupID: "",
			},
			{
				Key:     "group_3",
				GroupID: "",
			},
			{
				Key:     "group_1",
				GroupID: "",
			},
		}
		result := filterAnnotations(annotations, activeKeys)
		So(len(result), ShouldEqual, 4)
		So(result[0].Key, ShouldEqual, "alert_1")
		So(result[1].Key, ShouldEqual, "alert_2")
		So(result[2].Key, ShouldEqual, "group_2")
		So(result[3].Key, ShouldEqual, "group_1")
	})
}

func TestFilterDuplicateBugs(t *testing.T) {
	Convey("Test filter annotation", t, func() {
		bugs := []*model.MonorailBug{
			{
				BugID:     "bug_1",
				ProjectID: "project_1",
			},
			{
				BugID:     "bug_2",
				ProjectID: "project_2",
			},
			{
				BugID:     "bug_1",
				ProjectID: "project_1",
			},
			{
				BugID:     "bug_3",
				ProjectID: "project_3",
			},
		}

		result := filterDuplicateBugs(bugs)
		So(len(result), ShouldEqual, 3)
		So(result[0].BugID, ShouldEqual, "bug_1")
		So(result[1].BugID, ShouldEqual, "bug_2")
		So(result[2].BugID, ShouldEqual, "bug_3")
	})
}

func TestAnnotations(t *testing.T) {
	newContext := func() (context.Context, testclock.TestClock) {
		c := gaetesting.TestingContext()
		c = authtest.MockAuthConfig(c)
		c = gologger.StdConfig.Use(c)

		cl := testclock.New(testclock.TestRecentTimeUTC)
		c = clock.Set(c, cl)
		return c, cl
	}
	Convey("/annotations", t, func() {

		w := httptest.NewRecorder()
		c, cl := newContext()
		tok, err := xsrf.Token(c)
		So(err, ShouldBeNil)

		monorailMux := http.NewServeMux()
		monorailResponse := func(w http.ResponseWriter, r *http.Request) {
			logging.Errorf(c, "got monorailMux request")
			res := &monorail.IssuesListResponse{
				Items:        []*monorail.Issue{},
				TotalResults: 0,
			}
			bytes, err := json.Marshal(res)
			if err != nil {
				logging.Errorf(c, "error marshaling response: %v", err)
			}
			w.Write(bytes)
		}
		monorailMux.HandleFunc("/", monorailResponse)

		monorailServer := httptest.NewServer(monorailMux)
		defer monorailServer.Close()
		Monorail := client.NewMonorail(c, monorailServer.URL)

		Bqh := &BugQueueHandler{
			Monorail: Monorail,
		}
		ah := &AnnotationHandler{
			Bqh: Bqh,
		}

		Convey("GET", func() {
			Convey("no annotations yet", func() {
				ah.GetAnnotationsHandler(&router.Context{
					Context: c,
					Writer:  w,
					Request: makeGetRequest(),
				}, nil)

				r, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				body := string(r)
				So(w.Code, ShouldEqual, 200)
				So(body, ShouldEqual, "[]")
			})

			ann := &model.Annotation{
				KeyDigest:        fmt.Sprintf("%x", sha1.Sum([]byte("foobar"))),
				Key:              "foobar",
				Bugs:             []model.MonorailBug{{BugID: "111", ProjectID: "fuchsia"}, {BugID: "222", ProjectID: "chromium"}},
				SnoozeTime:       123123,
				ModificationTime: datastore.RoundTime(clock.Now(c).Add(4 * time.Hour)),
			}

			So(datastore.Put(c, ann), ShouldBeNil)
			datastore.GetTestable(c).CatchupIndexes()

			Convey("basic annotation", func() {
				ah.GetAnnotationsHandler(&router.Context{
					Context: c,
					Writer:  w,
					Request: makeGetRequest(),
				}, map[string]interface{}{ann.Key: nil})

				r, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				body := string(r)
				So(w.Code, ShouldEqual, 200)
				rslt := []*model.Annotation{}
				So(json.NewDecoder(strings.NewReader(body)).Decode(&rslt), ShouldBeNil)
				So(rslt, ShouldHaveLength, 1)
				So(rslt[0], ShouldResemble, ann)
			})

			Convey("basic annotation, alert no longer active", func() {
				ah.GetAnnotationsHandler(&router.Context{
					Context: c,
					Writer:  w,
					Request: makeGetRequest(),
				}, nil)

				r, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				body := string(r)
				So(w.Code, ShouldEqual, 200)
				rslt := []*model.Annotation{}
				So(json.NewDecoder(strings.NewReader(body)).Decode(&rslt), ShouldBeNil)
				So(rslt, ShouldHaveLength, 0)
			})
		})

		Convey("POST", func() {
			Convey("invalid action", func() {
				ah.PostAnnotationsHandler(&router.Context{
					Context: c,
					Writer:  w,
					Request: makePostRequest(""),
					Params:  makeParams("action", "lolwut"),
				})

				So(w.Code, ShouldEqual, 400)
			})

			Convey("invalid json", func() {
				ah.PostAnnotationsHandler(&router.Context{
					Context: c,
					Writer:  w,
					Request: makePostRequest("invalid json"),
					Params:  makeParams("annKey", "foobar", "action", "add"),
				})

				So(w.Code, ShouldEqual, http.StatusBadRequest)
			})

			ann := &model.Annotation{
				Tree:             datastore.MakeKey(c, "Tree", "tree.unknown"),
				Key:              "foobar",
				KeyDigest:        fmt.Sprintf("%x", sha1.Sum([]byte("foobar"))),
				ModificationTime: datastore.RoundTime(clock.Now(c)),
			}
			cl.Add(time.Hour)

			makeChange := func(data map[string]interface{}, tok string) string {
				change, err := json.Marshal(map[string]interface{}{
					"xsrf_token": tok,
					"data":       data,
				})
				So(err, ShouldBeNil)
				return string(change)
			}
			Convey("add, bad xsrf token", func() {
				ah.PostAnnotationsHandler(&router.Context{
					Context: c,
					Writer:  w,
					Request: makePostRequest(makeChange(map[string]interface{}{
						"snoozeTime": 123123,
					}, "no good token")),
					Params: makeParams("annKey", "foobar", "action", "add"),
				})

				So(w.Code, ShouldEqual, http.StatusForbidden)
			})

			Convey("add", func() {
				ann = &model.Annotation{
					Tree:             datastore.MakeKey(c, "Tree", "tree.unknown"),
					Key:              "foobar",
					KeyDigest:        fmt.Sprintf("%x", sha1.Sum([]byte("foobar"))),
					ModificationTime: datastore.RoundTime(clock.Now(c)),
				}
				change := map[string]interface{}{}
				Convey("snoozeTime", func() {
					ah.PostAnnotationsHandler(&router.Context{
						Context: c,
						Writer:  w,
						Request: makePostRequest(makeChange(map[string]interface{}{
							"snoozeTime": 123123,
							"key":        "foobar",
						}, tok)),
						Params: makeParams("action", "add", "tree", "tree.unknown"),
					})

					So(w.Code, ShouldEqual, 200)

					So(datastore.Get(c, ann), ShouldBeNil)
					So(ann.SnoozeTime, ShouldEqual, 123123)
				})

				Convey("bugs", func() {
					change["bugs"] = []model.MonorailBug{{BugID: "123123", ProjectID: "chromium"}}
					change["key"] = "foobar"
					ah.PostAnnotationsHandler(&router.Context{
						Context: c,
						Writer:  w,
						Request: makePostRequest(makeChange(change, tok)),
						Params:  makeParams("action", "add", "tree", "tree.unknown"),
					})

					So(w.Code, ShouldEqual, 200)

					So(datastore.Get(c, ann), ShouldBeNil)
					So(ann.Bugs, ShouldResemble, []model.MonorailBug{{BugID: "123123", ProjectID: "chromium"}})
				})
			})

			Convey("remove", func() {
				Convey("can't remove non-existent annotation", func() {
					ah.PostAnnotationsHandler(&router.Context{
						Context: c,
						Writer:  w,
						Request: makePostRequest(makeChange(map[string]interface{}{"key": "foobar"}, tok)),
						Params:  makeParams("action", "remove", "tree", "tree.unknown"),
					})

					So(w.Code, ShouldEqual, 404)
				})

				ann.SnoozeTime = 123
				So(datastore.Put(c, ann), ShouldBeNil)

				Convey("basic", func() {
					So(ann.SnoozeTime, ShouldEqual, 123)

					ah.PostAnnotationsHandler(&router.Context{
						Context: c,
						Writer:  w,
						Request: makePostRequest(makeChange(map[string]interface{}{
							"key":        "foobar",
							"snoozeTime": true,
						}, tok)),
						Params: makeParams("action", "remove", "tree", "tree.unknown"),
					})

					So(w.Code, ShouldEqual, 200)
					So(datastore.Get(c, ann), ShouldBeNil)
					So(ann.SnoozeTime, ShouldEqual, 0)
				})
			})
		})

		Convey("refreshAnnotations", func() {
			Convey("handler", func() {
				c, _ := newContext()

				ah.RefreshAnnotationsHandler(&router.Context{
					Context: c,
					Writer:  w,
					Request: makeGetRequest(),
				})

				b, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				So(w.Code, ShouldEqual, 200)
				So(string(b), ShouldEqual, "{}")
			})
		})
	})
}
