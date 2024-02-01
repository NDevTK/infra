// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package godep

import (
	"testing"

	"golang.org/x/mod/modfile"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

const testGoMod = `
module test/module

go 1.21
toolchain go1.21.6

require (
	direct/dep1 v1.4
	direct/dep2 v0.0.0-20210112150236-f10218a38794
	direct/dep1/deeper v1.5
	direct/unused1 v0.5
	direct/unused2 v0.6
	direct/replaced1 v1.0
	direct/replaced2 v0.0.0-20230103053340-8a57daa72e32
)

require (
	indirect/dep1 v0.3.0   // indirect
	indirect/unused v0.4.0 // indirect
)

exclude direct/unused1 v0.2

replace (
	direct/replaced1 => direct/replaced1 v1.5
	direct/replaced2 => ../local/dir
)
`

func TestDeps(t *testing.T) {
	t.Parallel()

	prepTestDeps := func() *Deps {
		deps := NewDeps(loadTestGoMod())

		So(deps.Add("direct/dep1/pkg1", "direct/dep1", "1.20"), ShouldBeNil)
		So(deps.Add("direct/dep1/pkg1", "direct/dep1", "1.20"), ShouldBeNil) // dup is fine
		So(deps.Add("direct/dep1/pkg2", "direct/dep1", "1.20"), ShouldBeNil)

		So(deps.Add("direct/dep2", "direct/dep2", ""), ShouldBeNil)
		So(deps.Add("direct/dep2/pkg", "direct/dep2", ""), ShouldBeNil)

		So(deps.Add("indirect/dep1/pkg1", "indirect/dep1", ""), ShouldBeNil)

		So(deps.Add("direct/replaced1/pkg1", "direct/replaced1", "1.19"), ShouldBeNil)
		So(deps.Add("direct/replaced1/pkg2", "direct/replaced1", "1.19"), ShouldBeNil)

		So(deps.Add("direct/replaced2/pkg1", "direct/replaced2", "1.19"), ShouldBeNil)
		So(deps.Add("direct/replaced2/pkg2", "direct/replaced2", "1.19"), ShouldBeNil)

		return deps
	}

	expectedSavedDeps := SerializedState{
		GoMod: []byte(`module test/module

go 1.21

toolchain go1.21.6

require (
	direct/dep1 v1.4.0
	direct/dep2 v0.0.0-20210112150236-f10218a38794
	direct/replaced1 v1.0.0
	direct/replaced2 v0.0.0-20230103053340-8a57daa72e32
	indirect/dep1 v0.3.0
)

replace direct/replaced1 => direct/replaced1 v1.5.0

replace direct/replaced2 => ../local/dir
`),
		ModulesTxt: []byte(`# direct/dep1 v1.4.0
## explicit; go 1.20
direct/dep1/pkg1
direct/dep1/pkg2
# direct/dep2 v0.0.0-20210112150236-f10218a38794
## explicit
direct/dep2
direct/dep2/pkg
# direct/replaced1 v1.0.0 => direct/replaced1 v1.5.0
## explicit; go 1.19
direct/replaced1/pkg1
direct/replaced1/pkg2
# direct/replaced2 v0.0.0-20230103053340-8a57daa72e32 => ../local/dir
## explicit; go 1.19
direct/replaced2/pkg1
direct/replaced2/pkg2
# indirect/dep1 v0.3.0
## explicit
indirect/dep1/pkg1
# direct/replaced1 => direct/replaced1 v1.5.0
# direct/replaced2 => ../local/dir
`),
	}

	Convey("Adding and saving", t, func() {
		deps := prepTestDeps()

		saved, err := deps.Save()
		So(err, ShouldBeNil)

		So(string(saved.GoMod), ShouldEqual, string(expectedSavedDeps.GoMod))
		So(string(saved.ModulesTxt), ShouldEqual, string(expectedSavedDeps.ModulesTxt))
	})

	Convey("Loading", t, func() {
		deps := NewDeps(loadTestGoMod())
		So(deps.Load(expectedSavedDeps), ShouldBeNil)

		saved, err := deps.Save()
		So(err, ShouldBeNil)

		So(string(saved.GoMod), ShouldEqual, string(expectedSavedDeps.GoMod))
		So(string(saved.ModulesTxt), ShouldEqual, string(expectedSavedDeps.ModulesTxt))
	})

	Convey("AddDep errors", t, func() {
		deps := NewDeps(loadTestGoMod())

		// Wrong module name prefix.
		So(deps.Add("direct/dep2/pkg", "direct/dep2", ""), ShouldBeNil)
		So(deps.Add("direct/dep2/pkg", "direct/dep1", ""), ShouldErrLike, "not in module")
		So(deps.Add("direct/dep11/pkg", "direct/dep1", ""), ShouldErrLike, "not in module")

		// Package "switching" modules.
		So(deps.Add("direct/dep1/deeper", "direct/dep1/deeper", ""), ShouldBeNil)
		So(deps.Add("direct/dep1/deeper", "direct/dep1", ""), ShouldErrLike, "conflicting modules")

		// Package "switching" go version.
		So(deps.Add("indirect/dep1/pkg", "indirect/dep1", "1.20"), ShouldBeNil)
		So(deps.Add("indirect/dep1/pkg", "indirect/dep1", ""), ShouldErrLike, "conflicting go version")

		// Missing go.mod reference.
		So(deps.Add("unknown/pkg", "unknown", ""), ShouldErrLike, "not present in go.mod")
	})
}

func loadTestGoMod() *modfile.File {
	f, err := modfile.Parse("go.mod", []byte(testGoMod), nil)
	if err != nil {
		panic(err)
	}
	return f
}
