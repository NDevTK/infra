// Copyright 2016 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/auth/authtest"
	"go.chromium.org/luci/server/router"
)

var _ = fmt.Printf

func TestMain(t *testing.T) {
	t.Parallel()

	Convey("main", t, func() {
		c := gaetesting.TestingContext()
		c = gologger.StdConfig.Use(c)

		cl := testclock.New(testclock.TestRecentTimeUTC)
		c = clock.Set(c, cl)

		w := httptest.NewRecorder()

		Convey("index", func() {
			Convey("pathless", func() {
				(&SOMHandlers{}).indexPage(&router.Context{
					Writer:  w,
					Request: makeGetRequest(c, "/"),
				})

				So(w.Code, ShouldEqual, 302)
			})

			Convey("anonymous", func() {
				(&SOMHandlers{}).indexPage(&router.Context{
					Writer:  w,
					Request: makeGetRequest(c, "/chromium"),
				})

				r, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				body := string(r)
				So(w.Code, ShouldEqual, 500)
				So(body, ShouldNotContainSubstring, "som-app")
				So(body, ShouldContainSubstring, "login")
			})

			authState := &authtest.FakeState{
				Identity: "user:user@example.com",
			}
			c = auth.WithState(c, authState)

			Convey("No access", func() {
				(&SOMHandlers{}).indexPage(&router.Context{
					Writer:  w,
					Request: makeGetRequest(c, "/chromium"),
				})

				So(w.Code, ShouldEqual, 200)
				r, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				body := string(r)
				So(body, ShouldNotContainSubstring, "som-app")
				So(body, ShouldContainSubstring, "Access denied")
			})
			authState.IdentityGroups = []string{authGroup}

			Convey("good path", func() {
				(&SOMHandlers{}).indexPage(&router.Context{
					Writer:  w,
					Request: makeGetRequest(c, "/chromium"),
				})
				r, err := ioutil.ReadAll(w.Body)
				So(err, ShouldBeNil)
				body := string(r)
				So(body, ShouldContainSubstring, "som-app")
				So(w.Code, ShouldEqual, 200)
			})
		})

		Convey("noop", func() {
			noopHandler(nil)
		})
	})
}

func makeGetRequest(ctx context.Context, path string) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, "GET", path, nil)
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
