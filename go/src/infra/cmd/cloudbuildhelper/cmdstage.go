// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/common/logging"
)

var cmdStage = &subcommands.Command{
	UsageLine: "stage -input-manifest <path> -stage-location <path> [...]",
	ShortDesc: "prepares the context directory or tarball",
	LongDesc: `Prepares the context directory or tarball.

Evaluates input YAML manifest specified via "-input-manifest" and executes all
local build steps there. Materializes the resulting context dir in a location
specified by "-stage-location". If it ends in "*.tar.gz", then the result is
a tarball, otherwise it is a new directory (attempting to output to an existing
directory is an error).

The contents of this directory/tarball is exactly what will be sent to the
docker daemon or to a Cloud Build worker.
`,

	CommandRun: func() subcommands.CommandRun {
		c := &cmdStageRun{}
		c.init()
		return c
	},
}

type cmdStageRun struct {
	commandBase

	inputManifest string
	stageLocation string
}

func (c *cmdStageRun) init() {
	c.commandBase.init(c.exec, false) // no auth

	c.Flags.StringVar(&c.inputManifest, "input-manifest", "", "Where to read YAML with input from.")
	c.Flags.StringVar(&c.stageLocation, "stage-location", "", "Where to put the prepared context dir.")
}

func (c *cmdStageRun) exec(ctx context.Context) error {
	switch {
	case c.inputManifest == "":
		return errBadFlag("-input-manifest", "this flag is required")
	case c.stageLocation == "":
		return errBadFlag("-stage-location", "this flag is required")
	}

	// TODO(vadimsh): Validate formal correctness of CLI flags, then call the
	// actual implementation that lives in a proper go package with proper API and
	// unit tests.

	logging.Infof(ctx, "Hello, world!")
	return nil
}
