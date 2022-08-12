// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"time"

	"infra/tools/vpython_ng/pkg/application"
	"infra/tools/vpython_ng/pkg/python"
	"infra/tools/vpython_ng/pkg/wheels"

	"go.chromium.org/luci/common/errors"
)

func main() {
	app := application.Application{
		PruneThreshold:    7 * 24 * time.Hour, // One week.
		MaxPrunesPerSweep: 3,

		Environments: os.Environ(),
		Arguments:    os.Args[1:],

		PythonExecutable: "python3",
	}
	app.Initialize()

	app.Must(app.ParseEnvs())
	app.Must(app.ParseArgs())

	if app.Bypass {
		// no-op for tool mode if we are bypassing vpython
		if app.ToolMode != "" {
			return
		}
		app.Must(app.ExecutePython())
		return
	}

	app.Must(app.LoadSpec())

	bundle := "3.8"
	if app.InterpreterPath != "" {
		bundle = app.InterpreterPath
	}
	cpython, err := python.CPythonFromPath(bundle, "cpython3")
	if err != nil {
		app.Fatal(err)
	}
	env := python.Environment{
		Executable: app.PythonExecutable,
		CPython:    cpython,
		Virtualenv: python.VirtualenvFromCIPD("version:2@16.7.10.chromium.7"),
	}
	wheel, err := wheels.FromSpec(app.VpythonSpec, env.Pep425Tags())
	if err != nil {
		app.Fatal(err)
	}
	venv := env.WithWheels(wheel)

	if !app.Help && app.ToolMode != "" {
		app.Must(func() error {
			switch app.ToolMode {
			case "install":
				app.PruneThreshold = 0
				return app.BuildVENV(venv)
			case "verify":
				return wheels.Verify(app.VpythonSpec)
			default:
				return errors.Reason("unknown -vpython-tool command: %s", app.ToolMode).Err()
			}
		}())
		return
	}

	app.Must(app.BuildVENV(venv))

	if app.Help {
		// Continue to execute python to print its help message after vpython's.
		fmt.Println(app.Usage)
	}
	app.Must(app.ExecutePython())
}
