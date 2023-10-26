// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package application

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"infra/tools/vpython/pkg/python"
	"infra/tools/vpython/pkg/wheels"

	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/system/exitcode"
	"go.chromium.org/luci/common/system/filesystem"

	. "github.com/smartystreets/goconvey/convey"
)

const defaultPythonVersion = "3.8"

func testData(filename string) string {
	return filepath.Join("..", "..", "testdata", filename)
}

var testStorageDir string

func getPythonEnvironment(ver string) *python.Environment {
	return map[string]*python.Environment{
		"3.8": {
			Executable: "python3",
			CPython:    python.CPython3FromCIPD("version:2@3.8.10.chromium.24"),
			Virtualenv: python.VirtualenvFromCIPD("version:2@16.7.12.chromium.7"),
		},
		"3.11": {
			Executable: "python3",
			CPython:    python.CPython3FromCIPD("version:2@3.11.5.chromium.30"),
			Virtualenv: python.VirtualenvFromCIPD("version:2@20.17.1.chromium.8"),
		},
	}[ver]
}

func setupApp(ctx context.Context, app *Application) context.Context {
	app.Arguments = append([]string{
		"-vpython-root",
		testStorageDir,
	}, app.Arguments...)

	app.Initialize(ctx)

	So(app.ParseEnvs(ctx), ShouldBeNil)
	So(app.ParseArgs(ctx), ShouldBeNil)
	ctx = app.SetLogLevel(ctx)
	So(app.LoadSpec(ctx), ShouldBeNil)
	return ctx
}

func buildVENV(ctx context.Context, app *Application, venv generators.Generator) {
	ap := actions.NewActionProcessor()
	wheels.MustSetTransformer(app.CIPDCacheDir, ap)
	So(app.BuildVENV(ctx, ap, venv), ShouldBeNil)

	// Release all the resources so the temporary vpython root directory can be
	// removed on Windows.
	app.Close()
}

func cmd(tb testing.TB, app *Application, env *python.Environment) *exec.Cmd {
	tb.Helper()

	ctx := context.Background()
	if env == nil {
		env = getPythonEnvironment(defaultPythonVersion)
	}
	app.PythonExecutable = env.Executable

	ctx = setupApp(ctx, app)

	venv := env.WithWheels(wheels.FromSpec(app.VpythonSpec, env.Pep425Tags()))
	buildVENV(ctx, app, venv)

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
	reexec := actions.NewReexecRegistry()
	wheels.MustSetExecutor(reexec)
	reexec.Intercept(context.Background())

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

func TestPythonBasic(t *testing.T) {
	Convey("Test python basic", t, func() {
		var env *python.Environment
		for _, ver := range []string{"3.8", "3.11"} {
			Convey(ver, func() {
				env = getPythonEnvironment(ver)

				Convey("test bad cwd", func() {
					cwd, err := os.Getwd()
					So(err, ShouldBeNil)
					err = os.Chdir(testData("test_bad_cwd"))
					So(err, ShouldBeNil)

					c := cmd(t, &Application{
						Arguments: []string{
							"bisect.py",
						},
					}, env)
					So(output(c), ShouldEqual, "SUCCESS")

					err = os.Chdir(cwd)
					So(err, ShouldBeNil)
				})

				Convey("Test exit code", func() {
					c := cmd(t, &Application{
						Arguments: []string{
							"-vpython-spec",
							testData("default.vpython3"),
							testData("test_exit_code.py"),
						},
					}, env)

					err := output(c).(error)
					rc, has := exitcode.Get(err)
					So(has, ShouldBeTrue)
					So(rc, ShouldEqual, 42)
				})
			})
		}
	})
}

func TestPythonFromPath(t *testing.T) {
	Convey("Test python from path", t, func() {
		ctx := context.Background()
		env := getPythonEnvironment(defaultPythonVersion)

		app := &Application{
			Arguments: []string{
				"-vpython-spec",
				testData("default.vpython3"),
				testData("test_exit_code.py"),
			},
			PythonExecutable: env.Executable,
		}
		ctx = setupApp(ctx, app)

		// We are not actually building venv, but this should also work for python
		// package.
		buildVENV(ctx, app, env.CPython)

		// Python located at ${CPython}/bin/python3
		dir := filepath.Dir(filepath.Dir(app.PythonExecutable))
		py, err := python.CPythonFromPath(dir, "cpython3")
		So(err, ShouldBeNil)
		env.CPython = py

		// Run actual command
		c := cmd(t, app, env)
		err = output(c).(error)
		rc, has := exitcode.Get(err)
		So(has, ShouldBeTrue)
		So(rc, ShouldEqual, 42)
	})
}

func BenchmarkStartup(b *testing.B) {
	Convey("Benchmark startup", b, func() {
		c := func() *exec.Cmd {
			return cmd(b, &Application{
				Arguments: []string{
					"-vpython-spec",
					testData("default.vpython3"),
					"-c",
					"print(1)",
				},
			}, nil)
		}
		So(output(c()), ShouldEqual, "1")
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			_ = c()
		}
	})
}
