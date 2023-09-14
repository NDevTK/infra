// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/system/filesystem"

	"infra/libs/cipkg/builtins"
	"infra/tools/vpython_ng/pkg/application"
	"infra/tools/vpython_ng/pkg/python"
	"infra/tools/vpython_ng/pkg/wheels"
)

func testData(filename string) string {
	return filepath.Join("..", "..", "testdata", filename)
}

var (
	testStorageDir string

	pythonRuntime = GetPythonRuntime("2.7")
	pythonEnv     = &python.Environment{
		Executable: pythonRuntime.Executable,
		CPython: &builtins.CIPDExport{
			Name: "cpython",
			Ensure: ensure.File{
				PackagesBySubdir: map[string]ensure.PackageSlice{
					"": {
						{PackageTemplate: "infra/3pp/tools/cpython/${platform}", UnresolvedVersion: "version:2@2.7.18.chromium.44"},
					},
				},
			},
		},
		Virtualenv: python.VirtualenvFromCIPD(pythonRuntime.Virtualenv),
	}
)

func setupApp(app *application.Application) {
	app.Arguments = append([]string{
		"-vpython-spec",
		testData("default.vpython3"),
		"-vpython-root",
		testStorageDir,
	}, app.Arguments...)

	app.Initialize()

	So(app.ParseEnvs(), ShouldBeNil)
	So(app.ParseArgs(), ShouldBeNil)

	So(app.LoadSpec(), ShouldBeNil)
}

func cmd(app *application.Application, env *python.Environment) *exec.Cmd {
	app.PythonExecutable = env.Executable

	setupApp(app)

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
	defer func() {
		if err = filesystem.RemoveAll(testStorageDir); err != nil {
			panic(err)
		}
	}()

	wheels.Init(filepath.Join(testStorageDir, "cipd"))

	rc := m.Run()

	os.Exit(rc)
}

func TestPythonFromPath(t *testing.T) {
	Convey("Test python from path", t, func() {
		c := cmd(&application.Application{
			Arguments:        []string{"-c", "import sys; print(sys.version)"},
			PythonExecutable: pythonEnv.Executable,
		}, pythonEnv)
		So(output(c), ShouldStartWith, "2.7")
	})
}
