// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"time"

	"infra/tools/vpython_ng/pkg/application"
	"infra/tools/vpython_ng/pkg/common"
	"infra/tools/vpython_ng/pkg/python"
	"infra/tools/vpython_ng/pkg/wheels"

	"go.chromium.org/luci/common/errors"
)

const DefaultPythonVersion = "3.8"

type PythonRuntime struct {
	Version    string
	Virtualenv string
}

func GetPythonRuntime(ver string) *PythonRuntime {
	switch ver {
	case "3.8":
		return &PythonRuntime{
			Version:    "3.8",
			Virtualenv: "version:2@16.7.12.chromium.7",
		}
	default:
		return &PythonRuntime{
			Version:    ver,
			Virtualenv: "version:2@20.17.1.chromium.8",
		}
	}
}

func main() {
	rt := GetPythonRuntime(DefaultPythonVersion)

	app := application.Application{
		PruneThreshold:    7 * 24 * time.Hour, // One week.
		MaxPrunesPerSweep: 3,

		DefaultSpecPattern: ".vpython3",

		Environments: os.Environ(),
		Arguments:    os.Args[1:],
	}
	app.Initialize()

	app.Must(app.ParseEnvs())
	app.Must(app.ParseArgs())

	app.PythonExecutable = "python3"
	if app.Bypass {
		// no-op for tool mode if we are bypassing vpython
		if app.ToolMode != "" {
			return
		}
		app.Must(app.ExecutePython())
		return
	}

	app.Must(app.LoadSpec())

	// Update the Python Runtime based on vpython spec, if specified.
	if v := app.VpythonSpec.PythonVersion; v != "" {
		rt = GetPythonRuntime(v)
	}

	bundle := common.DefaultBundleDir(rt.Version)
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
		Virtualenv: python.VirtualenvFromCIPD(rt.Virtualenv),
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

	if app.Help {
		// Continue to execute python to print its help message after vpython's.
		fmt.Println(app.Usage)
	}
	app.Must(app.BuildVENV(venv))
	app.Must(app.ExecutePython())
}
