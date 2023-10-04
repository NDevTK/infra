// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"infra/tools/vpython_ng/pkg/application"
	"infra/tools/vpython_ng/pkg/common"
	"infra/tools/vpython_ng/pkg/python"
	"infra/tools/vpython_ng/pkg/wheels"

	"go.chromium.org/luci/cipkg/base/actions"
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

func Main(ctx context.Context) error {
	reexecRegistry := actions.NewReexecRegistry()
	wheels.MustSetExecutor(reexecRegistry)
	reexecRegistry.Intercept(ctx)

	rt := GetPythonRuntime(DefaultPythonVersion())
	app := &application.Application{
		PruneThreshold:    7 * 24 * time.Hour, // One week.
		MaxPrunesPerSweep: 3,

		DefaultSpecPattern: rt.SpecPattern,

		Environments: os.Environ(),
		Arguments:    os.Args[1:],
	}
	ctx = app.Initialize(ctx)
	if err := app.ParseEnvs(ctx); err != nil {
		return err
	}
	if err := app.ParseArgs(ctx); err != nil {
		return err
	}
	ctx = app.SetLogLevel(ctx)

	actionProcessor := actions.NewActionProcessor()
	wheels.MustSetTransformer(app.CIPDCacheDir, actionProcessor)

	if app.Bypass {
		// no-op for tool mode if we are bypassing vpython
		if app.ToolMode != "" {
			return nil
		}
		app.PythonExecutable = rt.Executable
		return app.ExecutePython(ctx)
	}

	if err := app.LoadSpec(ctx); err != nil {
		return err
	}

	// Update the Python Runtime based on vpython spec, if specified.
	if v := app.VpythonSpec.PythonVersion; v != "" {
		if strings.HasPrefix(rt.Version, "3.") && strings.HasPrefix(v, "2.") {
			return errors.Reason("Python2 specs must be explicitly executed using 'vpython'.").Err()
		}
		rt = GetPythonRuntime(v)
	}
	app.PythonExecutable = rt.Executable

	bundle := common.DefaultBundleDir(rt.Version)
	if app.InterpreterPath != "" {
		bundle = app.InterpreterPath
	}
	cpython, err := python.CPythonFromPath(bundle, rt.CIPDName)
	if err != nil {
		return err
	}

	env := python.Environment{
		Executable: app.PythonExecutable,
		CPython:    cpython,
		Virtualenv: python.VirtualenvFromCIPD(rt.Virtualenv),
	}
	venv := env.WithWheels(wheels.FromSpec(app.VpythonSpec, env.Pep425Tags()))

	if !app.Help && app.ToolMode != "" {
		switch app.ToolMode {
		case "install":
			app.PruneThreshold = 0
			return app.BuildVENV(ctx, actionProcessor, venv)
		case "verify":
			return wheels.Verify(app.VpythonSpec)
		default:
			return errors.Reason("unknown -vpython-tool command: %s", app.ToolMode).Err()
		}
	}

	if app.Help {
		// Continue to execute python to print its help message after vpython's.
		fmt.Println(app.Usage)
	}
	if err := app.BuildVENV(ctx, actionProcessor, venv); err != nil {
		return err
	}
	if err := app.ExecutePython(ctx); err != nil {
		return err
	}
	return errors.New("unreachable")
}

func main() {
	ctx := context.Background()
	if err := Main(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}
