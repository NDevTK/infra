// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Represents the CLI command grouping
package cli

import (
	"flag"
	"log"

	"infra/cros/cmd/ctpv2/executions"
)

// Run as build. This is in place to support backward-compatibility with
// test_runner recipes invocation of cros_test_runner.
type BuildCommand struct {
	flagSet *flag.FlagSet
}

func NewBuildCommand() *BuildCommand {
	cc := &BuildCommand{
		flagSet: flag.NewFlagSet("build", flag.ContinueOnError),
	}

	return cc
}

func (cc *BuildCommand) Is(group string) bool {
	// Always true since this is the last option.
	return true
}

func (cc *BuildCommand) Name() string {
	return "build"
}

func (cc *BuildCommand) Init(args []string) error {
	err := cc.flagSet.Parse(args)
	if err != nil {
		return err
	}

	return nil
}

// Run runs the commands to publish test results
func (cc *BuildCommand) Run() error {
	log.Printf("Running build Mode:")

	// execute hw tests.
	executions.LuciBuildExecution()

	return nil
}
