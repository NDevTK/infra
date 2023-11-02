// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package application

import (
	"context"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/cipd/client/cipd"
	"go.chromium.org/luci/common/logging"
)

func TestParseArguments(t *testing.T) {
	Convey("Test parse arguments", t, func() {
		ctx := context.Background()

		app := &Application{}
		app.Initialize(ctx)

		parseArgs := func(args ...string) error {
			app.Arguments = args
			So(app.ParseEnvs(ctx), ShouldBeNil)
			return app.ParseArgs(ctx)
		}

		Convey("Test log level", func() {
			err := parseArgs(
				"-vpython-log-level",
				"warning",
			)
			So(err, ShouldBeNil)
			ctx = app.SetLogLevel(ctx)
			So(logging.GetLevel(ctx), ShouldEqual, logging.Warning)
		})

		Convey("Test unknown argument", func() {
			const unknownErr = "failed to extract flags: unknown flag: vpython-test"

			// Care but only care arguments begin with "-" or "--".
			err := parseArgs("-vpython-test")
			So(err, ShouldBeError, unknownErr)
			err = parseArgs("--vpython-test")
			So(err, ShouldBeError, unknownErr)
			err = parseArgs("-vpython-root", "root", "vpython-test")
			So(err, ShouldBeNil)

			// All arguments after the script file should be bypassed.
			err = parseArgs("-vpython-test", "test.py")
			So(err, ShouldBeError, unknownErr)
			err = parseArgs("test.py", "-vpython-test")
			So(err, ShouldBeNil)

			// Stop parsing arguments when seen --
			err = parseArgs("--", "-vpython-test")
			So(err, ShouldBeNil)
		})

		Convey("Test cipd cache dir", func() {
			err := parseArgs("-vpython-root", "root", "vpython-test")
			So(err, ShouldBeNil)
			So(app.CIPDCacheDir, ShouldStartWith, "root")
		})

		Convey("Test cipd cache dir with env", func() {
			// Don't set cipd cache dir if env provides one
			app.Environments = append(app.Environments, fmt.Sprintf("%s=%s", cipd.EnvCacheDir, "something"))
			err := parseArgs("-vpython-root", "root", "vpython-test")
			So(err, ShouldBeNil)
			So(app.CIPDCacheDir, ShouldStartWith, "something")
		})
	})
}
