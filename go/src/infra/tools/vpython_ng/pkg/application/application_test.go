package application

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/common/logging"
)

func TestParseArguments(t *testing.T) {
	Convey("Test parse arguments", t, func() {
		app := &Application{}
		app.Initialize()

		parseArgs := func(args ...string) error {
			app.Arguments = args
			return app.ParseArgs()
		}

		Convey("Test log level", func() {
			err := parseArgs(
				"-vpython-log-level",
				"warning",
			)
			So(err, ShouldBeNil)
			So(logging.GetLevel(app.Context), ShouldEqual, logging.Warning)
		})

		Convey("Test unknown argument", func() {
			err := parseArgs(
				"-vpython-test",
			)
			So(err.Error(), ShouldContainSubstring, "-vpython-test")

			err = parseArgs(
				"--",
				"-vpython-test",
			)
			So(err, ShouldBeNil)
		})
	})
}
