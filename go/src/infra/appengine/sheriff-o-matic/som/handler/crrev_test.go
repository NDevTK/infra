package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/server/auth/authtest"
	"go.chromium.org/luci/server/router"

	"infra/appengine/sheriff-o-matic/som/client"
	"infra/appengine/sheriff-o-matic/som/client/test"
)

func TestRevRangeHandler(t *testing.T) {
	fakeCrRev := test.NewFakeServer()
	defer fakeCrRev.Server.Close()

	c := gaetesting.TestingContext()
	c = gologger.StdConfig.Use(c)
	crRev := client.NewCrRev(fakeCrRev.Server.URL)

	Convey("get rev range", t, func() {
		Convey("ok with positions", func() {
			c = authtest.MockAuthConfig(c)
			w := httptest.NewRecorder()
			getRevRangeHandler(&router.Context{
				Writer: w,
				Request: makeGetRequest(c,
					"startPos", "123", "endPos", "456",
					"endRev", "1a2b3c4d"),
				Params: makeParams(
					"host", "chromium", "repo", "chromium.src"),
			}, crRev)

			So(w.Code, ShouldEqual, 301)
		})
		Convey("ok with revisions", func() {
			c = authtest.MockAuthConfig(c)
			w := httptest.NewRecorder()
			getRevRangeHandler(&router.Context{
				Writer: w,
				Request: makeGetRequest(c,
					"startRev", "2a2b3c4d", "endRev", "1a2b3c4d"),
				Params: makeParams(
					"host", "chromium", "repo", "chromium.src"),
			}, crRev)

			So(w.Code, ShouldEqual, 301)
		})
		Convey("bad oauth", func() {
			w := httptest.NewRecorder()
			getRevRangeHandler(&router.Context{
				Writer: w,
				Request: makeGetRequest(c,
					"startPos", "123", "endPos", "456",
					"endRev", "1a2b3c4d"),
				Params: makeParams(
					"host", "chromium", "repo", "chromium.src"),
			}, crRev)
			So(w.Code, ShouldEqual, http.StatusMovedPermanently)
		})
		Convey("bad request", func() {
			w := httptest.NewRecorder()

			getRevRangeHandler(&router.Context{
				Writer:  w,
				Request: makeGetRequest(c),
			}, crRev)

			So(w.Code, ShouldEqual, 400)
		})
		Convey("bad start and end params", func() {
			w := httptest.NewRecorder()

			getRevRangeHandler(&router.Context{
				Writer: w,
				Request: makeGetRequest(c,
					"startPos", "123", "endRev", "1a2b3c4d"),
				Params: makeParams(
					"host", "chromium", "repo", "chromium.src"),
			}, crRev)
			So(w.Code, ShouldEqual, 400)
		})
		Convey("bad repo and host", func() {
			w := httptest.NewRecorder()

			getRevRangeHandler(&router.Context{
				Writer: w,
				Request: makeGetRequest(c,
					"startPos", "123", "endPos", "234",
					"startRev", "2a2b3c4d", "endRev", "1a2b3c4d"),
				Params: makeParams(),
			}, crRev)
			So(w.Code, ShouldEqual, 400)
		})

	})
}
