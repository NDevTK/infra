// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package run

import (
	"github.com/maruel/subcommands"
)

// runSuiteFlags holds the flags necessary for test execution
type runSuiteFlags struct {
	subcommands.CommandRunBase

	model     string
	board     string
	milestone string
	build     string
	pool      string
	suite     string
}

// registerRunSuiteFlags registers the test execution flags.
func registerRunSuiteFlags(c *runSuite) {
	c.Flags.StringVar(&c.model, "model", "", "model specifies what model a test should run on")
	c.Flags.StringVar(&c.board, "board", "", "board is the board to run against")
	c.Flags.StringVar(&c.milestone, "milestone", "", "milestone of the ChromeOS image")
	c.Flags.StringVar(&c.build, "build", "", "build version of the ChromeOS image")
	c.Flags.StringVar(&c.pool, "pool", "", "pool specifies what `label-pool` dimension we should run a test on")
	c.Flags.StringVar(&c.suite, "suite", "", "test suite to execute")
}
