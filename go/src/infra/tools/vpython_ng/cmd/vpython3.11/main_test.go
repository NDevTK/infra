// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"infra/tools/vpython_ng/pkg/application"
	"infra/tools/vpython_ng/pkg/python"
	"infra/tools/vpython_ng/pkg/wheels"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/system/exitcode"
	"go.chromium.org/luci/common/system/filesystem"

	. "github.com/smartystreets/goconvey/convey"
)

func testData(filename string) string {
	return filepath.Join("..", "..", "testdata", filename)
}

var testStorageDir string

func cmd(app *application.Application) *exec.Cmd {
	app.Arguments = append([]string{
		"-vpython-root",
		testStorageDir,
	}, app.Arguments...)
	app.PythonExecutable = "python3"

	app.Initialize()

	So(app.ParseEnvs(), ShouldBeNil)
	So(app.ParseArgs(), ShouldBeNil)

	So(app.LoadSpec(), ShouldBeNil)

	env := python.Environment{
		Executable: app.PythonExecutable,
		CPython:    python.CPython3FromCIPD("version:2@3.11.1.chromium.25"),
		Virtualenv: python.VirtualenvFromCIPD("version:2@20.17.1.chromium.8"),
	}
	wheels, err := wheels.FromSpec(app.VpythonSpec, env.Pep425Tags())
	So(err, ShouldBeNil)
	venv := env.WithWheels(wheels)

	So(app.BuildVENV(venv), ShouldBeNil)

	// Release all the resources so the temporary vpython root directory can be
	// removed on Windows.
	defer app.Close()

	return app.GetExecCommand()
}

func output(c *exec.Cmd) interface{} {
	var out strings.Builder
	c.Stdout = &out
	c.Stderr = &out
	if err := c.Run(); err != nil {
		return errors.Annotate(err, out.String()).Err()
	}
	return strings.TrimSpace(out.String())
}

func TestMain(m *testing.M) {
	var err error

	if testStorageDir, err = os.MkdirTemp("", "vpython-test-"); err != nil {
		panic(err)
	}

	rc := m.Run()

	if err = filesystem.RemoveAll(testStorageDir); err != nil {
		panic(err)
	}

	os.Exit(rc)
}

func TestBadCWD(t *testing.T) {
	Convey("Test bad cwd", t, func() {
		cwd, err := os.Getwd()
		So(err, ShouldBeNil)
		err = os.Chdir(testData("test_bad_cwd"))
		So(err, ShouldBeNil)

		c := cmd(&application.Application{
			Arguments: []string{
				"bisect.py",
			},
		})
		So(output(c), ShouldEqual, "SUCCESS")

		err = os.Chdir(cwd)
		So(err, ShouldBeNil)
	})
}

func TestExitCode(t *testing.T) {
	Convey("Test exit code", t, func() {
		c := cmd(&application.Application{
			Arguments: []string{
				"-vpython-spec",
				testData("default.vpython3"),
				testData("test_exit_code.py"),
			},
		})

		err := output(c)
		rc, has := exitcode.Get(err.(error))
		So(has, ShouldBeTrue)
		So(rc, ShouldEqual, 42)
	})
}

func TestParseArguments(t *testing.T) {
	parseArgs := func(args ...string) error {
		app := &application.Application{
			Arguments: args,
		}
		app.Initialize()
		return app.ParseArgs()
	}

	Convey("Test unknown argument", t, func() {
		err := parseArgs(
			"-vpython-spec",
			testData("default.vpython3"),
			"-vpython-test",
		)
		So(err.Error(), ShouldContainSubstring, "-vpython-test")

		err = parseArgs(
			"-vpython-spec",
			testData("default.vpython3"),
			"--",
			"-vpython-test",
		)
		So(err, ShouldBeNil)
	})
}

func BenchmarkStartup(b *testing.B) {
	Convey("Benchmark startup", b, func() {
		c := func() *exec.Cmd {
			return cmd(&application.Application{
				Arguments: []string{
					"-vpython-spec",
					testData("default.vpython3"),
					"-c",
					"print(1)",
				},
			})
		}
		So(output(c()), ShouldEqual, "1")
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			_ = c()
		}
	})
}
