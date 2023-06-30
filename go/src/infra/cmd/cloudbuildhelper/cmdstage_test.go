// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"infra/cmd/cloudbuildhelper/fileset"
	"infra/cmd/cloudbuildhelper/manifest"

	. "github.com/smartystreets/goconvey/convey"
)

const testPinsYAML = `
pins:
- image: example.com/some/image
  tag: some-tag
  digest: sha256:431d3cca5e6a1f043b48ec556a9c60876e7ae1db18e20d0b860cb9c77e4c0b1b
`

func TestStage(t *testing.T) {
	t.Parallel()

	Convey("With temp", t, func() {
		ctx := context.Background()

		write := func(root, path, body string) string {
			p := filepath.Join(root, path)
			So(os.WriteFile(p, []byte(body), 0600), ShouldBeNil)
			return p
		}
		read := func(root, path string) string {
			blob, err := os.ReadFile(filepath.Join(root, path))
			So(err, ShouldBeNil)
			return string(blob)
		}

		Convey("Bunch of files", func() {
			ctxDir := t.TempDir()
			outDir := t.TempDir()

			write(ctxDir, "a", "file a")
			write(ctxDir, "b", "file b")

			err := stage(ctx,
				&manifest.Manifest{
					ContextDir: ctxDir,
				},
				func(fs *fileset.Set) error {
					return fs.Materialize(outDir)
				},
			)

			So(err, ShouldBeNil)
			So(read(outDir, "a"), ShouldEqual, "file a")
			So(read(outDir, "b"), ShouldEqual, "file b")
		})

		Convey("Resolving explicitly set Dockerfile", func() {
			ctxDir := t.TempDir()
			outDir := t.TempDir()

			dockerfile := write(ctxDir, "Dockerfile", "FROM example.com/some/image:some-tag")

			err := stage(ctx,
				&manifest.Manifest{
					ContextDir: ctxDir,
					Dockerfile: dockerfile,
					ImagePins:  write(t.TempDir(), "pins.yaml", testPinsYAML),
				},
				func(fs *fileset.Set) error {
					return fs.Materialize(outDir)
				},
			)

			So(err, ShouldBeNil)
			So(read(outDir, "Dockerfile"), ShouldEqual,
				"FROM example.com/some/image@sha256:431d3cca5e6a1f043b48ec556a9c60876e7ae1db18e20d0b860cb9c77e4c0b1b")
		})

		Convey("Resolving ${contextdir}/Dockerfile", func() {
			ctxDir := t.TempDir()
			outDir := t.TempDir()

			write(ctxDir, "Dockerfile", "FROM example.com/some/image:some-tag")

			err := stage(ctx,
				&manifest.Manifest{
					ContextDir: ctxDir,
					ImagePins:  write(t.TempDir(), "pins.yaml", testPinsYAML),
				},
				func(fs *fileset.Set) error {
					return fs.Materialize(outDir)
				},
			)

			So(err, ShouldBeNil)
			So(read(outDir, "Dockerfile"), ShouldEqual,
				"FROM example.com/some/image@sha256:431d3cca5e6a1f043b48ec556a9c60876e7ae1db18e20d0b860cb9c77e4c0b1b")
		})

		Convey("Explicitly set Dockerfile is missing", func() {
			ctxDir := t.TempDir()
			outDir := t.TempDir()

			err := stage(ctx,
				&manifest.Manifest{
					ContextDir: ctxDir,
					Dockerfile: filepath.Join(ctxDir, "Dockerfile"),
				},
				func(fs *fileset.Set) error {
					return fs.Materialize(outDir)
				},
			)

			So(errors.Is(err, fs.ErrNotExist), ShouldBeTrue)
		})

		Convey("Using build steps", func() {
			ctxDir := t.TempDir()
			inpDir := t.TempDir()
			outDir := t.TempDir()

			write(ctxDir, "a", "file a")
			write(inpDir, "b", "file b")

			m, err := manifest.Load(write(inpDir, "manifest.yaml", `{
				"name": "ignored",
				"contextdir": "to-be-set-later",
				"inputsdir": "to-be-set-later",
				"build": [
					{
						"copy": "${inputsdir}/b",
						"dest": "${contextdir}/b"
					}
				]
			}`))
			So(err, ShouldBeNil)
			m.ContextDir = ctxDir
			m.InputsDir = inpDir
			So(m.Finalize(), ShouldBeNil)

			err = stage(ctx, m, func(fs *fileset.Set) error {
				return fs.Materialize(outDir)
			})

			So(err, ShouldBeNil)
			So(read(outDir, "a"), ShouldEqual, "file a")
			So(read(outDir, "b"), ShouldEqual, "file b")
		})
	})
}
