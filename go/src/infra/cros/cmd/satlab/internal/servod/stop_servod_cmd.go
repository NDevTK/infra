// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package servod

import (
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/common/errors"

	"infra/cros/cmd/satlab/internal/site"
)

// StartServodCmd is the command that will start a servod container
var StopServodCmd = &subcommands.Command{
	UsageLine: "stop -host <hostname> [options ...]",
	ShortDesc: "starts servod container",
	LongDesc:  "Starts servod container",
	CommandRun: func() subcommands.CommandRun {
		c := &stopServodRun{}
		c.commonFlags.Register(&c.Flags)
		c.Flags.StringVar(&c.host, "host", "", "Hostname of DUT")
		c.Flags.StringVar(&c.servodContainerName, "servod-container-name", "", "Optional: name of servod container; will be fetched from UFS if not provided")
		return c
	},
}

// stopServodRun struct contains the arguments for the servod command
type stopServodRun struct {
	subcommands.CommandRunBase
	commonFlags         site.CommonFlags
	host                string
	servodContainerName string
}

// Run is what is called when a user inputs the stopServodRun command
func (c *stopServodRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// innerRun contains the actual logic of the stopServodRun command
func (c *stopServodRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	return errors.Reason("not yet implemented").Err()
}
