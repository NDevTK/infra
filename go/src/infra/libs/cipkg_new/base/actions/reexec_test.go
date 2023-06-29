// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"infra/libs/cipkg_new/core"
	"infra/libs/cipkg_new/testutils"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
	"go.chromium.org/luci/common/system/environ"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestProcessReexec(t *testing.T) {
	Convey("Test action processor for reexec", t, func() {
		ap := NewActionProcessor("", testutils.NewMockPackageManage(""))

		reexec := &core.ActionReexec{}

		pkg, err := ap.Process(&core.Action{
			Metadata: &core.Action_Metadata{Name: "url"},
			Spec:     &core.Action_Reexec{Reexec: reexec},
		})
		So(err, ShouldBeNil)

		checkReexecArg(pkg.Derivation.Args, reexec)
	})
}

func TestReexecExecuteReexec(t *testing.T) {
	Convey("Test re-execute action reexec", t, func() {
		ap := NewActionProcessor("", testutils.NewMockPackageManage(""))

		pkg, err := ap.Process(ReexecDependency().Action)
		So(err, ShouldBeNil)

		dst := testutils.NewAferoMemMapFs()
		runWithDrv(dst, pkg.Derivation)

		{
			self, err := os.Executable()
			So(err, ShouldBeNil)

			srcFile, err := os.Open(self)
			So(err, ShouldBeNil)
			srcBytes, err := io.ReadAll(srcFile)
			So(err, ShouldBeNil)

			cipkgExec := filepath.FromSlash("out/cipkg_exec")
			if runtime.GOOS == "windows" {
				cipkgExec = filepath.FromSlash("out/cipkg_exec.exe")
			}
			dstFile, err := dst.Open(cipkgExec)
			So(err, ShouldBeNil)
			dstBytes, err := io.ReadAll(dstFile)
			So(err, ShouldBeNil)

			So(dstBytes, ShouldResemble, srcBytes)
		}
	})
}

func runWithDrv(dst afero.Fs, drv *core.Derivation) {
	env := environ.New(drv.Env)
	env.Set("out", "out")
	NewMain().RunWithArgs(dst, env, drv.Args, func() {
		panic("unreachable")
	})
}

func checkReexecArg(args []string, m proto.Message) {
	m, err := anypb.New(m)
	So(err, ShouldBeNil)
	b, err := protojson.Marshal(m)
	So(err, ShouldBeNil)
	So(args, ShouldContain, string(b))
}
