// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package generators

import (
	"context"
	"io/fs"
	"path/filepath"
	"testing"

	"infra/libs/cipkg_new/core"
	"infra/libs/cipkg_new/testutils"

	. "github.com/smartystreets/goconvey/convey"
)

func TestImport(t *testing.T) {
	Convey("Test import", t, func() {
		ctx := context.Background()
		plats := Platforms{}

		g := ImportTargets{
			Targets: map[string]ImportTarget{
				"test1":     {Source: "//path/to/host1"},
				"dir/test2": {Source: "//path/to/host2", Version: "v2"},
			},
		}
		a, err := g.Generate(ctx, plats)
		So(err, ShouldBeNil)

		imports := testutils.Assert[*core.Action_Copy](t, a.Spec)
		So(imports.Copy.Files, ShouldResemble, map[string]*core.ActionFilesCopy_Source{
			"test1": {
				Content: &core.ActionFilesCopy_Source_Local_{
					Local: &core.ActionFilesCopy_Source_Local{Path: filepath.FromSlash("//path/to/host1")},
				},
				Mode: uint32(fs.ModeSymlink),
			},
			filepath.FromSlash("dir/test2"): {
				Content: &core.ActionFilesCopy_Source_Local_{
					Local: &core.ActionFilesCopy_Source_Local{Path: filepath.FromSlash("//path/to/host2"), Version: "v2"},
				},
				Mode: uint32(fs.ModeSymlink),
			},
			filepath.FromSlash("build-support/base_import.stamp"): {
				Content: &core.ActionFilesCopy_Source_Raw{},
				Mode:    0o666,
			},
		})
	})
}
