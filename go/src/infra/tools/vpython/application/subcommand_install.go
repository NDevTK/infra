// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package application

import (
	"github.com/luci/luci-go/common/cli"
	"github.com/luci/luci-go/common/errors"
	"github.com/luci/luci-go/common/logging"

	"github.com/maruel/subcommands"
	"golang.org/x/net/context"
)

var subcommandInstall = &subcommands.Command{
	UsageLine: "install",
	ShortDesc: "installs the configured VirtualEnv",
	LongDesc:  "installs the configured VirtualEnv, but doesn't run any associated commands",
	Advanced:  false,
	CommandRun: func() subcommands.CommandRun {
		return &installCommandRun{}
	},
}

type installCommandRun struct {
	subcommands.CommandRunBase
}

func (cr *installCommandRun) Run(app subcommands.Application, args []string, env subcommands.Env) int {
	c := cli.GetContext(app, cr, env)
	a := getApplication(c)

	return run(c, func(c context.Context) error {
		env, err := a.Opts.EnvConfig.Env(c)
		if err != nil {
			return errors.Annotate(err).Reason("failed to configure environment").Err()
		}

		if err := env.Setup(c, false); err != nil {
			return errors.Annotate(err).Reason("failed to setup the environment").Err()
		}

		logging.Infof(c, "Successfully setup the enviornment.")
		return nil
	})
}
