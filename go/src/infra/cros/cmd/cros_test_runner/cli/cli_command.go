// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Represents the CLI command grouping
package cli

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
)

// CLI command executed the provisioning as a CLI
type CLICommand struct {
	mode    string
	flagSet *flag.FlagSet
}

func NewCLICommand() *CLICommand {
	cc := &CLICommand{
		flagSet: flag.NewFlagSet("cli", flag.ContinueOnError),
	}

	cc.flagSet.StringVar(&cc.mode, "mode", "", fmt.Sprintf("Specify the port for the server. Default value %d.", 0))
	return cc
}

func (cc *CLICommand) Is(group string) bool {
	return strings.HasPrefix(group, "c")
}

func (cc *CLICommand) Name() string {
	return "cli"
}

func (cc *CLICommand) Init(args []string) error {
	err := cc.flagSet.Parse(args)
	if err != nil {
		return err
	}

	if err = cc.validate(); err != nil {
		return err
	}

	return nil
}

// validate checks if inputs are ok
func (sc *CLICommand) validate() error {
	if sc.mode == "" {
		return errors.New("mode not specified")
	}

	if sc.mode != "hw" && sc.mode != "vm" && sc.mode != "local" {
		return errors.New("invalid mode! Only 'hw', 'vm' and 'local' is supported!")
	}

	return nil
}

// Run runs the commands to publish test results
func (cc *CLICommand) Run() error {
	log.Printf("Running CLI Mode:")

	// TODO: Call to different flow depending on mode value

	// Call to hw flow
	// Call to vm flow
	// Call to local flow

	return nil
}
