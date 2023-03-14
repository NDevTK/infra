// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"infra/tools/vpython_ng/pkg/application"
	"infra/tools/vpython_ng/pkg/common"
	"infra/tools/vpython_ng/pkg/python"
	"infra/tools/vpython_ng/pkg/wheels"

	"go.chromium.org/luci/common/errors"
)

type PythonRuntime struct {
	Version     string
	Executable  string
	CIPDName    string
	SpecPattern string
	Virtualenv  string
}

func GetPythonRuntime(ver string) *PythonRuntime {
	switch ver {
	case "2.7":
		return &PythonRuntime{
			Version:     "2.7",
			Executable:  "python",
			CIPDName:    "cpython",
			SpecPattern: ".vpython",
			Virtualenv:  "version:2@16.7.12.chromium.7",
		}
	case "3.8":
		return &PythonRuntime{
			Version:     "3.8",
			Executable:  "python3",
			CIPDName:    "cpython3",
			SpecPattern: ".vpython3",
			Virtualenv:  "version:2@16.7.12.chromium.7",
		}
	default:
		return &PythonRuntime{
			Version:     ver,
			Executable:  "python3",
			CIPDName:    "cpython3",
			SpecPattern: ".vpython3",
			Virtualenv:  "version:2@20.17.1.chromium.8",
		}
	}
}

func DefaultPythonVersion() string {
	switch filepath.Base(os.Args[0]) {
	case "vpython", "vpython.exe":
		return "2.7"
	default:
		return "3.8"
	}
}

func main() {
	rt := GetPythonRuntime(DefaultPythonVersion())

	app := application.Application{
		PruneThreshold:    7 * 24 * time.Hour, // One week.
		MaxPrunesPerSweep: 3,

		DefaultSpecPattern: rt.SpecPattern,

		Environments: os.Environ(),
		Arguments:    os.Args[1:],
	}
	app.Initialize()

	app.Must(app.ParseEnvs())
	app.Must(app.ParseArgs())

	if app.Bypass {
		// no-op for tool mode if we are bypassing vpython
		if app.ToolMode != "" {
			return
		}
		app.PythonExecutable = rt.Executable
		app.Must(app.ExecutePython())
		return
	}

	app.Must(app.LoadSpec())

	// Update the Python Runtime based on vpython spec, if specified.
	if v := app.VpythonSpec.PythonVersion; v != "" {
		if strings.HasPrefix(rt.Version, "3.") && strings.HasPrefix(v, "2.") {
			app.Fatal(errors.Reason("Python2 specs must be explicitly executed using 'vpython'.").Err())
		}
		rt = GetPythonRuntime(v)
	}

	bundle := common.DefaultBundleDir(rt.Version)
	if app.InterpreterPath != "" {
		bundle = app.InterpreterPath
	}
	cpython, err := python.CPythonFromPath(bundle, rt.CIPDName)
	if err != nil {
		app.Fatal(err)
	}
	app.PythonExecutable = rt.Executable

	env := python.Environment{
		Executable: rt.Executable,
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
