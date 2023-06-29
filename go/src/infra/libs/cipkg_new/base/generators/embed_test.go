// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package generators

import (
	"context"
	"embed"
	"io/fs"
	"path/filepath"
	"testing"

	"infra/libs/cipkg_new/core"
	"infra/libs/cipkg_new/testutils"

	. "github.com/smartystreets/goconvey/convey"
)

//go:embed embed.go embed_test.go testdata
var embeddedFilesTestEmbed embed.FS
var embeddedFilesTestGen = InitEmbeddedFS(
	&core.Action_Metadata{Name: "embedded_files_test"},
	embeddedFilesTestEmbed,
).WithModeOverride(func(info fs.FileInfo) (fs.FileMode, error) {
	if info.Name() == "embed.go" {
		return fs.ModePerm, nil
	}
	return 0o666, nil
})

func TestEmbeddedFiles(t *testing.T) {
	Convey("Test embedded files", t, func() {
		ctx := context.Background()
		plats := Platforms{}

		Convey("Test normal", func() {
			a, err := embeddedFilesTestGen.Generate(ctx, plats)
			So(err, ShouldBeNil)

			embedFiles := testutils.Assert[*core.Action_Copy](t, a.Spec)
			So(embedFiles.Copy.Files, ShouldResemble, map[string]*core.ActionFilesCopy_Source{
				"embed.go": {
					Content: &core.ActionFilesCopy_Source_Embed_{
						Embed: &core.ActionFilesCopy_Source_Embed{Ref: embeddedFilesTestGen.ref, Path: "embed.go"},
					},
					Mode: 0o777,
				},
				"embed_test.go": {
					Content: &core.ActionFilesCopy_Source_Embed_{
						Embed: &core.ActionFilesCopy_Source_Embed{Ref: embeddedFilesTestGen.ref, Path: "embed_test.go"},
					},
					Mode: 0o666,
				},
				filepath.FromSlash("testdata/embed"): {
					Content: &core.ActionFilesCopy_Source_Embed_{
						Embed: &core.ActionFilesCopy_Source_Embed{Ref: embeddedFilesTestGen.ref, Path: "testdata/embed"},
					},
					Mode: 0o666,
				},
			})
		})
		Convey("Test subdir", func() {
			a, err := embeddedFilesTestGen.SubDir("testdata").Generate(ctx, plats)
			So(err, ShouldBeNil)

			embedFiles := testutils.Assert[*core.Action_Copy](t, a.Spec)
			So(embedFiles.Copy.Files, ShouldResemble, map[string]*core.ActionFilesCopy_Source{
				"embed": {
					Content: &core.ActionFilesCopy_Source_Embed_{
						Embed: &core.ActionFilesCopy_Source_Embed{Ref: embeddedFilesTestGen.ref, Path: "testdata/embed"},
					},
					Mode: 0o666,
				},
			})
		})
	})
}
