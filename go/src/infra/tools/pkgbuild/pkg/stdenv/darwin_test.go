// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package stdenv

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/common/exec/execmock"
)

func TestImportDarwin(t *testing.T) {
	Convey("import darwin", t, func() {
		ctx := execmock.Init(context.Background())
		xcodeSelectUses := execmock.Simple.WithArgs("xcode-select", "--print-path").Mock(ctx, execmock.SimpleInput{
			Stdout: "/path/to/xcode.app",
		})
		xcodeBuildUses := execmock.Simple.WithArgs(filepath.FromSlash("/path/to/xcode.app/usr/bin/xcodebuild"), "-version").Mock(ctx, execmock.SimpleInput{
			Stdout: "xcodeversion",
		})

		gs, err := importDarwin(ctx, &Config{
			FindBinary: func(bin string) (string, error) {
				return filepath.FromSlash(fmt.Sprintf("/bin/%s", bin)), nil
			},
			BuildPlatform: generators.NewPlatform("darwin", "arm64"),
		})
		So(err, ShouldBeNil)
		So(xcodeSelectUses.Snapshot(), ShouldHaveLength, 1)
		So(xcodeBuildUses.Snapshot(), ShouldHaveLength, 1)
		So(gs, ShouldContain, &generators.ImportTargets{
			Name: "xcode_import",
			Targets: map[string]generators.ImportTarget{
				"Developer": {Source: "/path/to/xcode.app", Mode: fs.ModeSymlink, Version: "xcodeversion"},
			},
		})

		// All imports on Mac should be symlink.
		for _, g := range gs {
			if targets, ok := g.(*generators.ImportTargets); ok {
				for _, t := range targets.Targets {
					_, _ = Println("checking", t)
					So(t.Mode&fs.ModeSymlink, ShouldNotBeEmpty)
				}
			}
		}
	})
}
