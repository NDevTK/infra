// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"os"
	"time"

	"infra/tools/vpython_ng/pkg/application"
	"infra/tools/vpython_ng/pkg/python"
	"infra/tools/vpython_ng/pkg/wheels"
)

func main() {
	app := application.Application{
		PruneThreshold:    7 * 24 * time.Hour, // One week.
		MaxPrunesPerSweep: 3,

		Environments: os.Environ(),
		Arguments:    os.Args[1:],
	}
	app.Initialize()

	app.Must(app.ParseEnvs())
	app.Must(app.ParseArgs())

	if app.Bypass {
		app.Must(app.ExecutePython())
		return
	}

	app.Must(app.LoadSpec())
	app.Must(func() error {
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

	app.MustExecutePython()
}
