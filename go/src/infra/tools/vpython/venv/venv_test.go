// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package venv

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"infra/tools/vpython/api/env"
	"infra/tools/vpython/filesystem"
	"infra/tools/vpython/filesystem/testfs"
	"infra/tools/vpython/python"

	"github.com/luci/luci-go/common/errors"

	"golang.org/x/net/context"

	. "github.com/luci/luci-go/common/testing/assertions"
	. "github.com/smartystreets/goconvey/convey"
)

const testDataDir = "test_data"

type resolvedInterpreter struct {
	i       *python.Interpreter
	version python.Version
}

func resolveFromPath(vers python.Version) *resolvedInterpreter {
	c := context.Background()
	i, err := python.Find(c, vers)
	if err != nil {
		return nil
	}
	if err := filesystem.AbsPath(&i.Python); err != nil {
		panic(err)
	}

	ri := resolvedInterpreter{
		i: i,
	}
	if ri.version, err = ri.i.GetVersion(c); err != nil {
		panic(err)
	}
	return &ri
}

var (
	pythonGeneric = resolveFromPath(python.Version{})
	python27      = resolveFromPath(python.Version{2, 7, 0})
	python3       = resolveFromPath(python.Version{3, 0, 0})
)

func TestResolvePythonInterpreter(t *testing.T) {
	t.Parallel()

	Convey(`Resolving a Python interpreter`, t, func() {
		c := context.Background()
		cfg := Config{
			Spec: &env.Spec{},
		}

		// Tests to run if we have Python 2.7 installed.
		if python27 != nil {
			Convey(`When Python 2.7 is requested, it gets resolved.`, func() {
				cfg.Spec.PythonVersion = "2.7"
				So(cfg.resolvePythonInterpreter(c), ShouldBeNil)
				So(cfg.Python, ShouldEqual, python27.i.Python)

				vers, err := python.ParseVersion(cfg.Spec.PythonVersion)
				So(err, ShouldBeNil)
				So(vers.IsSatisfiedBy(python27.version), ShouldBeTrue)
			})

			Convey(`Fails when Python 9999 is requested, but a Python 2 interpreter is forced.`, func() {
				cfg.Python = python27.i.Python
				cfg.Spec.PythonVersion = "9999"
				So(cfg.resolvePythonInterpreter(c), ShouldErrLike, "doesn't match specification")
			})
		}

		// Tests to run if we have Python 2.7 and a generic Python installed.
		if pythonGeneric != nil && python27 != nil {
			// Our generic Python resolves to a known version, so we can proceed.
			Convey(`When no Python version is specified, spec resolves to generic.`, func() {
				So(cfg.resolvePythonInterpreter(c), ShouldBeNil)
				So(cfg.Python, ShouldEqual, pythonGeneric.i.Python)

				vers, err := python.ParseVersion(cfg.Spec.PythonVersion)
				So(err, ShouldBeNil)
				So(vers.IsSatisfiedBy(pythonGeneric.version), ShouldBeTrue)
			})
		}

		// Tests to run if we have Python 3 installed.
		if python3 != nil {
			Convey(`When Python 3 is requested, it gets resolved.`, func() {
				cfg.Spec.PythonVersion = "3"
				So(cfg.resolvePythonInterpreter(c), ShouldBeNil)
				So(cfg.Python, ShouldEqual, python3.i.Python)

				vers, err := python.ParseVersion(cfg.Spec.PythonVersion)
				So(err, ShouldBeNil)
				So(vers.IsSatisfiedBy(python3.version), ShouldBeTrue)
			})

			Convey(`Fails when Python 9999 is requested, but a Python 3 interpreter is forced.`, func() {
				cfg.Python = python3.i.Python
				cfg.Spec.PythonVersion = "9999"
				So(cfg.resolvePythonInterpreter(c), ShouldErrLike, "doesn't match specification")
			})
		}
	})
}

// testingPackageLoader is a map of a CIPD package name to the root directory
// that it should be loaded from.
type testingPackageLoader map[string]string

func (pl testingPackageLoader) Resolve(c context.Context, root string, packages []*env.Spec_Package) error {
	for _, pkg := range packages {
		pkg.Version = "resolved"
	}
	return nil
}

func (pl testingPackageLoader) Ensure(c context.Context, root string, packages []*env.Spec_Package) error {
	for _, pkg := range packages {
		if err := pl.installPackage(pkg.Path, root); err != nil {
			return err
		}
	}
	return nil
}

func (pl testingPackageLoader) installPackage(name, root string) error {
	testName := pl[name]
	if testName == "" {
		return errors.Reason("could not resolve package for %(name)q").
			D("name", name).
			Err()
	}
	sourcePath := filepath.Join(testDataDir, testName)

	switch st, err := os.Stat(sourcePath); {
	case err != nil:
		return errors.Annotate(err).Reason("could not stat source: %(source)s").
			D("source", sourcePath).
			Err()

	case st.IsDir():
		if err := recursiveCopyDir(sourcePath, root); err != nil {
			return errors.Annotate(err).Reason("failed to recursively copy").Err()
		}

	case strings.HasSuffix(sourcePath, ".zip"):
		// If it's a file, it's a ZIP file. Unpack it into destination.
		if err := unzip(sourcePath, root); err != nil {
			return errors.Annotate(err).Reason("failed to un-zip archive").Err()
		}

	default:
		return errors.Reason("don't know how to handle: %(path)s").
			D("path", sourcePath).
			Err()
	}
	return nil
}

func recursiveCopyDir(src, dst string) error {
	// Recursively copy from sourcePath to root.
	return filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil || path == src {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return errors.Annotate(err).Reason("failed to get relative path").Err()
		}

		dst := filepath.Join(dst, rel)

		opener := func() (io.ReadCloser, error) { return os.Open(path) }
		if err := copyFileOrDir(opener, dst, fi); err != nil {
			return errors.Annotate(err).Reason("failed to copy: [%(src)s] => [%(dst)s]").
				D("src", path).
				D("dst", dst).
				Err()
		}
		return nil
	})
}

func unzip(src, dst string) error {
	fd, err := zip.OpenReader(src)
	if err != nil {
		return errors.Annotate(err).Reason("failed to open ZIP reader").Err()
	}
	defer fd.Close()

	for _, f := range fd.File {
		if err := copyFileOrDir(f.Open, filepath.Join(dst, f.Name), f.FileInfo()); err != nil {
			return errors.Annotate(err).Reason("failed to extract file: %(name)s").
				D("name", f.Name).
				Err()
		}
	}
	return nil
}

// copyFile copies a source file and its mode to a destination.
func copyFileOrDir(opener func() (io.ReadCloser, error), dst string, fi os.FileInfo) error {
	if fi.IsDir() {
		if err := os.MkdirAll(dst, 0755); err != nil {
			return errors.Annotate(err).Reason("failed to mkdir").Err()
		}
		return nil
	}

	srcFD, err := opener()
	if err != nil {
		return errors.Annotate(err).Reason("failed to create source").Err()
	}
	defer srcFD.Close()

	dstFD, err := os.Create(dst)
	if err != nil {
		return errors.Annotate(err).Reason("failed to create dest").Err()
	}
	defer dstFD.Close()

	if _, err := io.Copy(dstFD, srcFD); err != nil {
		return errors.Annotate(err).Reason("failed to copy").Err()
	}
	if err := os.Chmod(dst, fi.Mode()); err != nil {
		return errors.Annotate(err).Reason("failed to chmod").Err()
	}
	return nil
}

type setupCheckManifest struct {
	Interpreter string `json:"interpreter"`
	Pants       string `json:"pants"`
	Shirt       string `json:"shirt"`
}

func TestVirtualEnv(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		ri   *resolvedInterpreter
	}{
		{"python27", python27},
		{"python3", python3},
	} {
		tc := tc
		t.Run(fmt.Sprintf(`Testing Virtualenv for: %s`, tc.name), func(t *testing.T) {
			t.Parallel()

			conveyOp := Convey
			if tc.ri == nil {
				// No interpreter found, skip this test.
				conveyOp = SkipConvey
			}
			conveyOp(`Testing Setup`, t, testfs.MustWithTempDir("TestVirtualEnv", func(tdir string) {
				c := context.Background()
				config := Config{
					BaseDir:    tdir,
					MaxHashLen: 4,
					Package: env.Spec_Package{
						Path:    "foo/bar/virtualenv",
						Version: "unresolved",
					},
					Python: tc.ri.i.Python,
					Spec: &env.Spec{
						Wheel: []*env.Spec_Package{
							&env.Spec_Package{Path: "foo/bar/shirt", Version: "unresolved"},
							&env.Spec_Package{Path: "foo/bar/pants", Version: "unresolved"},
						},
					},
					Loader: testingPackageLoader{
						"foo/bar/virtualenv": "virtualenv-15.1.0.zip",
						"foo/bar/shirt":      "shirt",
						"foo/bar/pants":      "pants",
					},
				}
				v, err := config.Env(c)
				So(err, ShouldBeNil)

				// The setup should be successful.
				So(v.Setup(c, false), ShouldBeNil)

				testScriptPath := filepath.Join(testDataDir, "setup_check.py")
				checkOut := filepath.Join(tdir, "output.json")
				i := v.Interpreter()
				So(i.Run(c, testScriptPath, "--json-output", checkOut), ShouldBeNil)

				var m setupCheckManifest
				So(loadJSON(checkOut, &m), ShouldBeNil)
				So(m.Interpreter, ShouldStartWith, v.Root)
				So(m.Pants, ShouldStartWith, v.Root)
				So(m.Shirt, ShouldStartWith, v.Root)

				// We should be able to delete it.
				So(v.Delete(c), ShouldBeNil)
			}))
		})
	}
}

func loadJSON(path string, dst interface{}) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Annotate(err).Reason("failed to open file").Err()
	}
	if err := json.Unmarshal(content, dst); err != nil {
		return errors.Annotate(err).Reason("failed to unmarshal JSON").Err()
	}
	return nil
}
