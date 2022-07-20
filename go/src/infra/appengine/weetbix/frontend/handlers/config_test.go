// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.chromium.org/luci/gae/impl/memory"

	"infra/appengine/weetbix/internal/config"
	configpb "infra/appengine/weetbix/proto/config"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("With Router", t, func() {
		ctx := memory.Use(context.Background())

		router := routerForTesting()

		Convey("Get", func() {
			get := func() *http.Response {
				url := fmt.Sprintf("/api/projects/%s/config", testProject)
				request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
				So(err, ShouldBeNil)

				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				return response.Result()
			}

			err := config.SetTestProjectConfig(ctx, map[string]*configpb.ProjectConfig{
				testProject: {
					Monorail: &configpb.MonorailProject{
						DisplayPrefix: "mybug.com",
					},
				},
			})
			So(err, ShouldBeNil)

			response := get()
			So(response.StatusCode, ShouldEqual, 200)

			b, err := io.ReadAll(response.Body)
			So(err, ShouldBeNil)

			var responseBody projectConfig
			So(json.Unmarshal(b, &responseBody), ShouldBeNil)
			expected := projectConfig{
				Project: testProject,
				Monorail: &monorail{
					DisplayPrefix: "mybug.com",
				},
			}
			So(responseBody, ShouldResemble, expected)
		})
	})
}
