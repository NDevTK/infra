// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
	"testing"

	"infra/tools/vpython_ng/pkg/application"
	"infra/tools/vpython_ng/pkg/python"
	"infra/tools/vpython_ng/pkg/wheels"

	"go.chromium.org/luci/common/system/exitcode"
	"go.chromium.org/luci/common/system/filesystem"
)

var testStorageDir string

func testData(filename string) string {
	return filepath.Join("..", "..", "testdata", filename)
}

func must(tb testing.TB, err error) {
	if err != nil {
		tb.Fatal(err)
	}
}

func run(tb testing.TB, spec, script string, app *application.Application) error {
	app.ForkExec = true // Prevent using execve
	app.Arguments = append([]string{
		"-vpython-root",
		testStorageDir,
		"-vpython-spec",
		testData(spec),
		testData(script),
	}, app.Arguments...)

	app.Initialize()

	must(tb, app.ParseEnvs())
	must(tb, app.ParseArgs())

	must(tb, app.LoadSpec())
	must(tb, func() error {
		env := python.NewEnvironment(python.Versions{
			CPython:    "version:2@3.8.10.chromium.24",
			VirtualENV: "version:2@16.7.10.chromium.7",
		})
		wheels, err := wheels.FromSpec(app.VpythonSpec, env.Pep425Tags())
		if err != nil {
			return err
		}
		venv := env.WithWheels(wheels)
		return app.BuildVENV(venv)
	}())

	// Usually we don't need to RUnlock the venv package. It will be unlocked
	// when the vpython process exits. However for tests, the lock will prevent
	// the temporary vpython root directory from being removed on Windows.
	defer app.VENVPackage.RUnlock()
	return app.ExecutePython()
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

func TestExitCode(t *testing.T) {
	err := run(t, "default.vpython3", "test_exit_code.py", &application.Application{})

	if rc, has := exitcode.Get(err); !has || rc != 42 {
		t.Errorf("expect error with code 42, got %v", err)
	}
}
