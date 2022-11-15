// copyright 2021 the chromium os authors. all rights reserved.
// use of this source code is governed by a bsd-style license that can be
// found in the license file.

package docker

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// Tests the function which escapes special characters(ex: $, ` etc`) for a command input in the form of a string array
func TestEscapeSpecialCharacters(t *testing.T) {
	t.Parallel()
	Convey("Excape Special Characters", t, func() {
		Convey("Escapes $, \\, ` and double quote", func() {
			err := escapeSpecialChars([]string{"\\hello$", "`testString\""})
			So(err[0], ShouldEqual, "\\\\hello\\$")
			So(err[1], ShouldEqual, "\\`testString\\\"")
		})

		Convey("Does not escape anything other than $, \\, ` and double quote", func() {
			err := escapeSpecialChars([]string{"\\hello$^", "%`testString\""})
			So(err[0], ShouldEqual, "\\\\hello\\$^")
			So(err[1], ShouldEqual, "%\\`testString\\\"")
		})

		Convey("Empty array input - return empty array", func() {
			err := escapeSpecialChars([]string{})
			So(err, ShouldHaveLength, 0)
		})
	})
}
