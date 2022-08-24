// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.chromium.org/luci/server/router"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetAuthState(t *testing.T) {
	t.Parallel()

	Convey("Requests which do not have the same-origin are forbidden", t, func() {
		// Set up a test router to handle auth state requests
		testRouter := router.New()
		testRouter.GET("/api/authState", nil, GetAuthState)

		// Create request
		request, err := http.NewRequest("GET", "/api/authState", nil)
		So(err, ShouldBeNil)

		// Check response status code
		response := httptest.NewRecorder()
		testRouter.ServeHTTP(response, request)
		result := response.Result()
		So(result.StatusCode, ShouldEqual, http.StatusForbidden)
	})
}
