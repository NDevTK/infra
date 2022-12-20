// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package tasks

import (
	"encoding/json"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	"infra/cmd/mallet/internal/site"
	"infra/cmdsupport/cmdlib"
	"infra/cros/recovery"
	"infra/cros/recovery/tlw"
	"infra/libs/skylab/buildbucket"
)

// RecoveryConfig subcommand: For now, print the config file content to terminal/file.
var RecoveryConfig = &subcommands.Command{
	UsageLine: "config [-deploy] [-cros] [-deep-repair] [-labstation]",
	ShortDesc: "print the JSON plan configuration file",
	LongDesc:  "print the JSON plan configuration file.",
	CommandRun: func() subcommands.CommandRun {
		c := &printConfigRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.Flags.StringVar(&c.taskName, "task-name", "recovery", "Task name of the configuration we print.")
		c.Flags.StringVar(&c.deviceType, "device", "cros", "Device type supported 'cros', 'labstation'.")
		return c
	},
}

type printConfigRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags

	taskName   string
	deviceType string
}

// Run output the content of the recovery config file.
func (c *printConfigRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

// innerRun executes internal logic of output file content.
func (c *printConfigRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	ctx := cli.GetContext(a, c, env)
	tn, err := buildbucket.NormalizeTaskName(c.taskName)
	if err != nil {
		return errors.Annotate(err, "local recovery").Err()
	}
	var dsl []tlw.DUTSetupType
	switch c.deviceType {
	case "labstation":
		dsl = append(dsl, tlw.DUTSetupTypeLabstation)
	case "android":
		dsl = append(dsl, tlw.DUTSetupTypeAndroid)
	case "cros":
		dsl = append(dsl, tlw.DUTSetupTypeCros)
	default:
		return errors.Reason("upsupported device type %s", c.deviceType).Err()
	}
	for _, ds := range dsl {
		if c, err := recovery.ParsedDefaultConfiguration(ctx, tn, ds); err != nil {
			return errors.Annotate(err, "inner run").Err()
		} else if s, err := json.MarshalIndent(c, "", "\t"); err != nil {
			return errors.Annotate(err, "inner run").Err()
		} else {
			a.GetOut().Write(s)
		}
	}
	return nil
}
