// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gitignore

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"infra/cmd/cloudbuildhelper/fileset"
)

func TestExcluder(t *testing.T) {
	t.Parallel()

	Convey("findRepoRoot works", t, func(c C) {
		tmp := newTempDir(c)

		Convey("No .git at all", func() {
			root, err := findRepoRoot(tmp.join("."))
			So(err, ShouldBeNil)
			So(root, ShouldEqual, tmp.join("."))
		})

		Convey("Already given the root", func() {
			tmp.mkdir("a/.git")

			root, err := findRepoRoot(tmp.join("a"))
			So(err, ShouldBeNil)
			So(root, ShouldEqual, tmp.join("a"))
		})

		Convey("Discovers it few layers up", func() {
			tmp.mkdir("a/.git")
			tmp.mkdir("a/b/c")

			root, err := findRepoRoot(tmp.join("a/b/c"))
			So(err, ShouldBeNil)
			So(root, ShouldEqual, tmp.join("a"))
		})

		Convey("Skips files", func() {
			tmp.mkdir("a/.git")
			tmp.touch("a/b/.git")
			tmp.mkdir("a/b/c")

			root, err := findRepoRoot(tmp.join("a/b/c"))
			So(err, ShouldBeNil)
			So(root, ShouldEqual, tmp.join("a"))
		})
	})

	Convey("scanUp works", t, func(c C) {
		tmp := newTempDir(c)
		So(scanUp(tmp.join("a/b/c"), tmp.join("."), ".gitignore"), ShouldResemble, []string{
			tmp.join(".gitignore"),
			tmp.join("a/.gitignore"),
			tmp.join("a/b/.gitignore"),
		})
	})

	Convey("scanDown works", t, func(c C) {
		tmp := newTempDir(c)
		tmp.touch(".gitignore")
		tmp.touch("stuff/stuff")
		tmp.touch("a/b/c/.gitignore")
		tmp.touch("a/d/.gitignore")
		tmp.touch("a/d/stuff")

		paths, err := scanDown(nil, tmp.join("."), ".gitignore")
		So(err, ShouldBeNil)
		So(paths, ShouldResemble, []string{
			tmp.join(".gitignore"),
			tmp.join("a/b/c/.gitignore"),
			tmp.join("a/d/.gitignore"),
		})
	})

	Convey("With temp dir", t, func(c C) {
		tmp := newTempDir(c)
		tmp.mkdir(".git") // pretend to be the repo root

		excluder := func(p string) fileset.Excluder {
			cb, err := NewExcluder(tmp.join(p), ".gitignore")
			So(err, ShouldBeNil)
			return func(rel string, isDir bool) bool {
				return cb(tmp.join(rel), isDir)
			}
		}

		Convey("Noop excluder", func() {
			cb := excluder(".")

			So(cb(".", true), ShouldBeFalse)
			So(cb("a/b/c", false), ShouldBeFalse)
		})

		Convey("Simple excluder", func() {
			tmp.put(".gitignore", "*.out")
			cb := excluder(".")

			So(cb("abc.go", false), ShouldBeFalse)
			So(cb("abc.out", false), ShouldBeTrue)
			So(cb("abc.out", true), ShouldBeTrue)
			So(cb("1/2/3/abc.go", false), ShouldBeFalse)
			So(cb("1/2/3/abc.out", false), ShouldBeTrue)
			So(cb("abc.out/1/2/3", false), ShouldBeTrue)
		})

		Convey("Complex excluder", func() {
			tmp.put(".gitignore", "/dir/*\n!/dir/?z")
			cb := excluder(".")

			So(cb("dir", true), ShouldBeFalse)
			So(cb("dir/az", false), ShouldBeFalse)
			So(cb("dir/bz", false), ShouldBeFalse)
			So(cb("dir/ay", false), ShouldBeTrue)
			So(cb("dir/abc", false), ShouldBeTrue)
			So(cb("another/dir/abc", false), ShouldBeFalse)
		})

		Convey("Inherited .gitignore", func() {
			tmp.put(".gitignore", "*.pyc")
			tmp.put("a/.gitignore", "*.a\n/hidden")
			tmp.put("a/z/z/b/.gitignore", "*.b")

			// No matter where we start, all .gitignore files are respected.
			for _, start := range []string{".", "a", "a/z/z", "a/z/z/b"} {
				cb := excluder(start)

				So(cb("a/z/z/b/1.pyc", false), ShouldBeTrue)
				So(cb("a/z/z/b/1.a", false), ShouldBeTrue)
				So(cb("a/z/z/b/1.b", false), ShouldBeTrue)
				So(cb("a/z/z/b/1.good", false), ShouldBeFalse)
			}

			// Entries relative to .gitignore location are respected.
			cb := excluder(".")
			So(cb("hidden", true), ShouldBeFalse)
			So(cb("a/hidden", true), ShouldBeTrue)
			So(cb("a/z/hidden", true), ShouldBeFalse)
		})

		Convey("#include support", func() {
			tmp.put(".gitignore", "root-hidden\n#comment\n#!include:included")
			tmp.put("included", "include-hidden")

			cb := excluder(".")
			So(cb("visible", false), ShouldBeFalse)
			So(cb("root-hidden", false), ShouldBeTrue)
			So(cb("include-hidden", false), ShouldBeTrue)
		})
	})
}

func TestPatternExcluder(t *testing.T) {
	t.Parallel()

	Convey("Works", t, func() {
		exc := NewPatternExcluder([]string{"*.bad", "*.worse", "hidden"})

		So(exc("something.good", false), ShouldBeFalse)
		So(exc("something.good", true), ShouldBeFalse)
		So(exc("a/b/c/something.good", false), ShouldBeFalse)
		So(exc("something.good/a/b/c", false), ShouldBeFalse)

		So(exc("something.bad", false), ShouldBeTrue)
		So(exc("something.bad", true), ShouldBeTrue)
		So(exc("a/b/c/something.bad", false), ShouldBeTrue)
		So(exc("something.bad/a/b/c", false), ShouldBeTrue)

		So(exc("something.worse", false), ShouldBeTrue)
		So(exc("something.worse", true), ShouldBeTrue)
		So(exc("a/b/c/something.worse", false), ShouldBeTrue)
		So(exc("something.worse/a/b/c", false), ShouldBeTrue)

		So(exc("something/hidden", true), ShouldBeTrue)
		So(exc("hidden", true), ShouldBeTrue)
		So(exc("hidden/something", false), ShouldBeTrue)
	})

	Convey("Empty", t, func() {
		exc := NewPatternExcluder(nil)
		So(exc("stuff", false), ShouldBeFalse)
	})
}

type tmpDir struct {
	p string
	c C
}

func newTempDir(c C) tmpDir {
	tmp, err := ioutil.TempDir("", "gitignore_test")
	c.So(err, ShouldBeNil)
	c.Reset(func() { os.RemoveAll(tmp) })
	return tmpDir{tmp, c}
}

func (t tmpDir) join(p string) string {
	return filepath.Join(t.p, filepath.FromSlash(p))
}

func (t tmpDir) mkdir(p string) {
	t.c.So(os.MkdirAll(t.join(p), 0777), ShouldBeNil)
}

func (t tmpDir) put(p, data string) {
	t.mkdir(path.Dir(p))
	f, err := os.Create(t.join(p))
	t.c.So(err, ShouldBeNil)
	_, err = f.Write([]byte(data))
	t.c.So(err, ShouldBeNil)
	t.c.So(f.Close(), ShouldBeNil)
}

func (t tmpDir) touch(p string) {
	t.put(p, "")
}
