// Copyright 2019 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package hwid

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/chromiumos/infra/proto/go/manufacturing"
	"go.chromium.org/luci/appengine/gaetesting"
)

var dutlabelJSON string = `
{
 "labels": [
  {
   "name": "variant",
   "value": "ampton"
  },
  {
   "name": "phase",
   "value": "PVT"
  },
  {
   "name": "touchscreen"
  }
 ],
 "possible_labels": [
  "sku",
  "phase",
  "touchscreen"
 ]
}
`

var bomJSON string=`
{
 "components": [
  {
   "componentClass": "ro_main_firmware",
   "name": "Google_Coral_10068_59_0"
  },
  {
   "componentClass": "sku",
   "name": "36"
  },
  {
	  "componentClass": "storage",
	  "name": "samsung_16g_2"
  }
  ]
}
`


func TestGetHwidData(t *testing.T) {
	ctx := gaetesting.TestingContextWithAppID("go-test")

	Convey("Get hwid data", t, func() {
		Convey("Happy path", func() {
			mockHwidServerForBOM :=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, bomJSON)
			}))

			mockHwidServerForDutLabel := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, dutlabelJSON)
				hwidServerURL =mockHwidServerForBOM.URL + "/%s/%s/%s"
			}))
			defer mockHwidServerForDutLabel.Close()
			defer mockHwidServerForBOM.Close()
			hwidServerURL = mockHwidServerForDutLabel.URL + "/%s/%s/%s"

			data, err := GetHwidData(ctx, "AMPTON C3B-A2B-D2K-H9I-A2S", "secret")
			So(err, ShouldBeNil)
			So(data.Phase, ShouldEqual, manufacturing.Config_PHASE_PVT)
			So(data.Variant.GetValue(), ShouldEqual, "36")
		})

		Convey("Invaid key", func() {
			mockHwidServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintln(w, "bad key")
			}))
			defer mockHwidServer.Close()
			hwidServerURL = mockHwidServer.URL + "/%s/%s/%s"

			_, err := GetHwidData(ctx, "AMPTON C3B-A2B-D2K-H9I-A2S", "secret")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "bad key")
		})
	})
}
