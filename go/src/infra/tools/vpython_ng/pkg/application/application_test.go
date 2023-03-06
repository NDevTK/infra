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
	})
}
