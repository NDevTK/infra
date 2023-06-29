// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"context"
	"embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"infra/libs/cipkg_new/core"
	"infra/libs/cipkg_new/testutils"

	"go.chromium.org/luci/common/system/environ"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/afero"
)

func TestProcessCopy(t *testing.T) {
	Convey("Test action processor for copy", t, func() {
		ap := NewActionProcessor("", testutils.NewMockPackageManage(""))

		copy := &core.ActionFilesCopy{
			Files: map[string]*core.ActionFilesCopy_Source{
				"test/file": {
					Content: &core.ActionFilesCopy_Source_Local_{
						Local: &core.ActionFilesCopy_Source_Local{Path: "something"},
					},
					Mode: uint32(fs.ModeSymlink),
				},
			},
		}

		pkg, err := ap.Process(&core.Action{
			Metadata: &core.Action_Metadata{Name: "copy"},
			Deps:     []*core.Action_Dependency{ReexecDependency()},
			Spec:     &core.Action_Copy{Copy: copy},
		})
		So(err, ShouldBeNil)

		So(pkg.Dependencies, ShouldHaveLength, 1)
		So(pkg.Derivation.Args[0], ShouldStartWith, pkg.Dependencies[0].Package.Handler.OutputDirectory())
		checkReexecArg(pkg.Derivation.Args, copy)
		So(environ.New(pkg.Derivation.Env).Get(reexecActionName), ShouldEqual, pkg.Dependencies[0].Package.Handler.OutputDirectory())
	})
}

//go:embed copy_test.go
var actionCopyTestEmbed embed.FS

func TestExecuteCopy(t *testing.T) {
	Convey("Test execute action copy", t, func() {
		ctx := context.Background()
		dst := testutils.NewAferoMemMapFs()
		e := &FilesCopyExecutor{}

		Convey("output", func() {
			ctx = environ.New([]string{"somedrv=/abc/efg"}).SetInCtx(ctx)
			a := &core.ActionFilesCopy{
				Files: map[string]*core.ActionFilesCopy_Source{
					filepath.FromSlash("test/file"): {
						Content: &core.ActionFilesCopy_Source_Output_{
							Output: &core.ActionFilesCopy_Source_Output{Name: "somedrv", Path: "something"},
						},
						Mode: uint32(fs.ModeSymlink),
					},
				},
			}

			err := e.Execute(ctx, a, dst)
			So(err, ShouldBeNil)

			{
				dst := dst.(afero.Symlinker)
				l, err := dst.ReadlinkIfPossible("test/file")
				So(err, ShouldBeNil)
				So(l, ShouldEqual, filepath.FromSlash("/abc/efg/something"))
			}
		})

		Convey("local", func() {
			Convey("symlink", func() {
				a := &core.ActionFilesCopy{
					Files: map[string]*core.ActionFilesCopy_Source{
						filepath.FromSlash("test/file"): {
							Content: &core.ActionFilesCopy_Source_Local_{
								Local: &core.ActionFilesCopy_Source_Local{Path: "something"},
							},
							Mode: uint32(fs.ModeSymlink),
						},
					},
				}

				err := e.Execute(ctx, a, dst)
				So(err, ShouldBeNil)

				{
					dst := dst.(afero.Symlinker)
					l, err := dst.ReadlinkIfPossible("test/file")
					So(err, ShouldBeNil)
					So(l, ShouldEqual, "something")
				}
			})
		})

		Convey("embed", func() {
			e.StoreEmbed("something", actionCopyTestEmbed)

			a := &core.ActionFilesCopy{
				Files: map[string]*core.ActionFilesCopy_Source{
					filepath.FromSlash("test/files"): {
						Content: &core.ActionFilesCopy_Source_Embed_{
							Embed: &core.ActionFilesCopy_Source_Embed{Ref: "something"},
						},
						Mode: uint32(fs.ModeDir),
					},
					filepath.FromSlash("test/file"): {
						Content: &core.ActionFilesCopy_Source_Embed_{
							Embed: &core.ActionFilesCopy_Source_Embed{Ref: "something", Path: "copy_test.go"},
						},
						Mode: 0o666,
					},
				},
			}

			err := e.Execute(ctx, a, dst)
			So(err, ShouldBeNil)

			{
				f, err := dst.Open(filepath.FromSlash("test/files/copy_test.go"))
				So(err, ShouldBeNil)
				b, err := io.ReadAll(f)
				So(err, ShouldBeNil)
				So(string(b), ShouldContainSubstring, "Test copy ref embed")
			}
			{
				f, err := dst.Open(filepath.FromSlash("test/file"))
				So(err, ShouldBeNil)
				b, err := io.ReadAll(f)
				So(err, ShouldBeNil)
				So(string(b), ShouldContainSubstring, "Test copy ref embed")
			}
		})

		Convey("raw", func() {
			Convey("symlink", func() {
				a := &core.ActionFilesCopy{
					Files: map[string]*core.ActionFilesCopy_Source{
						filepath.FromSlash("test/file"): {
							Content: &core.ActionFilesCopy_Source_Raw{Raw: []byte("something")},
							Mode:    uint32(os.ModeSymlink),
						},
					},
				}

				e := &FilesCopyExecutor{}
				err := e.Execute(ctx, a, dst)
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "symlink is not supported")
			})

			Convey("dir", func() {
				a := &core.ActionFilesCopy{
					Files: map[string]*core.ActionFilesCopy_Source{
						filepath.FromSlash("test/file"): {
							Content: &core.ActionFilesCopy_Source_Raw{},
							Mode:    uint32(fs.ModeDir),
						},
					},
				}

				e := &FilesCopyExecutor{}
				err := e.Execute(ctx, a, dst)
				So(err, ShouldBeNil)

				{
					ok, err := afero.IsDir(dst, filepath.FromSlash("test/file"))
					So(err, ShouldBeNil)
					So(ok, ShouldBeTrue)
				}
			})

			Convey("file", func() {
				a := &core.ActionFilesCopy{
					Files: map[string]*core.ActionFilesCopy_Source{
						filepath.FromSlash("test/file"): {
							Content: &core.ActionFilesCopy_Source_Raw{Raw: []byte("something")},
							Mode:    0o777,
						},
					},
				}

				e := &FilesCopyExecutor{}
				err := e.Execute(ctx, a, dst)
				So(err, ShouldBeNil)

				{
					f, err := dst.Open(filepath.FromSlash("test/file"))
					So(err, ShouldBeNil)
					b, err := io.ReadAll(f)
					So(err, ShouldBeNil)
					So(string(b), ShouldEqual, "something")
					info, err := f.Stat()
					So(err, ShouldBeNil)
					So(info.Mode().Perm(), ShouldEqual, 0o777)
				}
			})
		})
	})
}

func TestReexecExecuteCopy(t *testing.T) {
	Convey("Test re-execute action copy", t, func() {
		ap := NewActionProcessor("", testutils.NewMockPackageManage(""))

		pkg, err := ap.Process(&core.Action{
			Metadata: &core.Action_Metadata{Name: "copy"},
			Deps:     []*core.Action_Dependency{ReexecDependency()},
			Spec: &core.Action_Copy{Copy: &core.ActionFilesCopy{
				Files: map[string]*core.ActionFilesCopy_Source{
					filepath.FromSlash("test/file"): {
						Content: &core.ActionFilesCopy_Source_Raw{Raw: []byte("something")},
						Mode:    0o777,
					},
				},
			}},
		})
		So(err, ShouldBeNil)

		dst := testutils.NewAferoMemMapFs()
		runWithDrv(dst, pkg.Derivation)

		{
			f, err := dst.Open(filepath.FromSlash("out/test/file"))
			So(err, ShouldBeNil)
			b, err := io.ReadAll(f)
			So(err, ShouldBeNil)
			So(string(b), ShouldEqual, "something")
			info, err := f.Stat()
			So(err, ShouldBeNil)
			So(info.Mode().Perm(), ShouldEqual, 0o777)
		}
	})
}
