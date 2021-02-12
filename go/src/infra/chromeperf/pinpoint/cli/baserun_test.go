// Copyright 2021 The Chromium Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFactorySettingFallbackCasese(t *testing.T) {

	Convey("When PINPOINT_CACHE_DIR is not defined", t, func() {
		ctx := context.Background()
		tc, creds, err := getFactorySettings(ctx, "pinpoint-stable.endpoints.chromeperf.cloud.goog")
		So(err, ShouldBeNil)
		So(tc, ShouldNotBeNil)
		// This is not nice, but we're checking that the directory is the default we're expecting.
		hd, err := os.UserHomeDir()
		So(err, ShouldBeNil)
		So(tc.cacheFile, ShouldEqual, filepath.Join(hd, ".cache", "pinpoint-cli", "cached-token"))
		So(creds, ShouldNotBeNil)
	})

	Convey("When provided service domain is not valid, we expect it to still work", t, func() {
		ctx := context.Background()
		tc, creds, err := getFactorySettings(ctx, "undefined")
		So(err, ShouldBeNil)
		So(tc, ShouldNotBeNil)
		// This is not nice, but we're checking that the directory is the default we're expecting.
		hd, err := os.UserHomeDir()
		So(err, ShouldBeNil)
		So(tc.cacheFile, ShouldEqual, filepath.Join(hd, ".cache", "pinpoint-cli", "cached-token"))
		So(creds, ShouldNotBeNil)
	})

}
