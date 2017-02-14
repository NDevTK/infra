// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package vpython

import (
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/system/environ"

	"golang.org/x/net/context"
)

var (
	// EnvSpecPath is the exported enviornment variable for the specification path.
	//
	// This is added to the bootstrap enviornment used by Run to allow subprocess
	// "vpython" invocations to automatically inherit the same environment.
	EnvSpecPath = "VPYTHON_VENV_SPEC_PATH"
)

// Run sets up a Python VirtualEnv and executes the supplied Options.
//
// Run returns nil if if the Python environment was successfully set-up and the
// Python interpreter was successfully run with a zero return code. If the
// Python interpreter returns a non-zero return code, a PythonError (potentially
// wrapped) will be returned.
//
// A generalized return code to return for an error value can be obtained via
// ReturnCode.
//
// Run consists of:
//
//	- Identify the target Python script to run (if there is one).
//	- Identifying the Python interpreter to use.
//	- Composing the environment specification.
//	- Constructing the virtual environment (download, install).
//	- Execute the Python process with the supplied arguments.
//
// The Python subprocess is bound to the lifetime of ctx, and will be terminated
// if ctx is cancelled.
func Run(c context.Context, opts Options) error {
	// Resolve our Options.
	if err := opts.resolve(c); err != nil {
		return errors.Annotate(err).Reason("could not resolve options").Err()
	}

	// Create our virtual enviornment root directory.
	venv, err := opts.EnvConfig.Env(c)
	if err != nil {
		return errors.Annotate(err).Reason("failed to resolve VirtualEnv").Err()
	}
	if err := venv.Setup(c, opts.WaitForEnv); err != nil {
		return errors.Annotate(err).Reason("failed to setup VirtualEnv").Err()
	}

	// Build the augmented environment variables.
	e := opts.Environ
	if e.Len() == 0 {
		// If no environment was supplied, use the system environment.
		e = environ.System()
	}
	e.Set("VIRTUAL_ENV", venv.Root) // Set by VirtualEnv script.
	if venv.SpecPath != "" {
		e.Set(EnvSpecPath, venv.SpecPath)
	}

	// Run our bootstrapped Python command.
	i := venv.Interpreter()
	i.WorkDir = opts.WorkDir
	i.Isolated = true
	i.ConnectSTDIN = true
	i.Env = e.Sorted()
	if err := i.Run(c, opts.Args...); err != nil {
		return errors.Annotate(err).Reason("failed to execute Python bootstrap").Err()
	}
	return nil
}
