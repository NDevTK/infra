// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	_ "embed"
	"testing"

	"infra/libs/cipkg_new/core"
	"infra/libs/cipkg_new/testutils"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessCIPD(t *testing.T) {
	Convey("Test action processor for cipd", t, func() {
		ap := NewActionProcessor("", testutils.NewMockPackageManage(""))

		cipd := &core.ActionCIPDExport{
			EnsureFile: "infra/tools/luci/vpython/linux-amd64 git_revision:98782288dfc349541691a2a5dfc0e44327f22731",
		}

		pkg, err := ap.Process(&core.Action{
			Metadata: &core.Action_Metadata{Name: "url"},
			Deps:     []*core.Action_Dependency{ReexecDependency()},
			Spec:     &core.Action_Cipd{Cipd: cipd},
		})
		So(err, ShouldBeNil)

		So(pkg.Dependencies, ShouldHaveLength, 1)
		So(pkg.Derivation.Args[0], ShouldStartWith, pkg.Dependencies[0].Package.Handler.OutputDirectory())
		checkReexecArg(pkg.Derivation.Args, cipd)
	})
}
