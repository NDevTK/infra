// Copyright 2018 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"

	tricium "infra/tricium/api/v1"
)

// These tests read from files on the filesystem, so modifying the tests may
// require modifying the example test files.
const (
	baseDir                         = "test"
	goodBsd                         = "src/good.cpp"
	goodMit                         = "src/good_mit.cpp"
	goodBsdWithoutAllRightsReserved = "src/good_wo_all_rights_reserved.cpp"
	badBsd                          = "src/bad.cpp"
	missing                         = "src/missing.cpp"
	old                             = "src/old.cpp"
)

func TestCopyrightChecker(t *testing.T) {

	Convey("Produces no comment for file with correct BSD copyright", t, func() {
		So(checkCopyright(baseDir, goodBsd), ShouldBeNil)
		So(checkCopyright(baseDir, goodBsdWithoutAllRightsReserved), ShouldBeNil)
	})

	Convey("Produces no comment for file with correct MIT copyright", t, func() {
		So(checkCopyright(baseDir, goodMit), ShouldBeNil)
	})

	Convey("Finds an issue when copyright doesn't match expected pattern", t, func() {
		c := checkCopyright(baseDir, badBsd)
		So(c, ShouldNotBeNil)
		So(c, ShouldResembleProto, &tricium.Data_Comment{
			Category: "Copyright/Incorrect",
			Message: ("Incorrect copyright statement.\n" +
				"Use the following for BSD:\n" +
				"Copyright <year> The <group> Authors. All rights reserved.\n" +
				"Use of this source code is governed by a BSD-style license that can be\n" +
				"found in the LICENSE file.\n\n" +
				"See: https://chromium.googlesource.com/chromium/src/+/main/styleguide/c++/c++.md#file-headers\n\n" +
				"Or the following for MIT: Copyright <year> The <group> Authors\n\n" +
				"Use of this source code is governed by a MIT-style\n" +
				"license that can be found in the LICENSE file or at\n" +
				"https://opensource.org/licenses/MIT."),
			Path:      badBsd,
			StartLine: 1,
			EndLine:   1,
			StartChar: 0,
			EndChar:   1,
		})
	})

	Convey("Makes a comment when there appears to be no copyright header", t, func() {
		c := checkCopyright(baseDir, missing)
		So(c, ShouldNotBeNil)
		So(c, ShouldResembleProto, &tricium.Data_Comment{
			Category: "Copyright/Missing",
			Message: ("Missing copyright statement.\n" +
				"Use the following for BSD:\n" +
				"Copyright <year> The <group> Authors. All rights reserved.\n" +
				"Use of this source code is governed by a BSD-style license that can be\n" +
				"found in the LICENSE file.\n\n" +
				"See: https://chromium.googlesource.com/chromium/src/+/main/styleguide/c++/c++.md#file-headers\n\n" +
				"Or the following for MIT: Copyright <year> The <group> Authors\n\n" +
				"Use of this source code is governed by a MIT-style\n" +
				"license that can be found in the LICENSE file or at\n" +
				"https://opensource.org/licenses/MIT."),
			Path:      missing,
			StartLine: 1,
			EndLine:   1,
			StartChar: 0,
			EndChar:   1,
		})
	})

	Convey("Makes a comment when there is a copyright statement but the old style is used", t, func() {
		c := checkCopyright(baseDir, old)
		So(c, ShouldNotBeNil)
		So(c, ShouldResembleProto, &tricium.Data_Comment{
			Category: "Copyright/OutOfDate",
			Message: "Out of date copyright statement (omit the (c) to update).\n\n" +
				"See: https://chromium.googlesource.com/chromium/src/+/main/styleguide/c++/c++.md#file-headers",
			Path:      old,
			StartLine: 1,
			EndLine:   1,
			StartChar: 0,
			EndChar:   1,
		})
	})
}
