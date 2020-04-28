// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package inventory

import (
	"context"
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	swarming_api "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"

	skycmdlib "infra/cmd/skylab/internal/cmd/cmdlib"
	flagx "infra/cmd/skylab/internal/flagx"
	"infra/cmd/skylab/internal/site"
	"infra/cmdsupport/cmdlib"
	"infra/libs/skylab/swarming"
)

// DutList subcommand: Get DUT information
var DutList = &subcommands.Command{
	UsageLine: "dut-list [-pool POOL] [-model MODEL] [-board BOARD]",
	ShortDesc: "List hostnames of devices matching search criteria",
	LongDesc: `List hostnames of devices matching search criteria.

	If no criteria are provided, dut-list will list all the DUT hostnames in Skylab.

	Search criteria include pool, model, board`,
	CommandRun: func() subcommands.CommandRun {
		c := &dutListRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.board, "board", "", "board name")
		c.Flags.StringVar(&c.model, "model", "", "model name")
		c.Flags.StringVar(&c.pool, "pool", "", "pool name")
		c.Flags.StringVar(&c.servoType, "servo-type", "", "the type of servo")
		c.Flags.BoolVar(&c.useInventory, "use-inventory", false, "use the inventory service if set, use swarming if unset")
		c.Flags.Var(flagx.Dims(&c.dims), "dims", "List of additional dimensions in format key1=value1,key2=value2,... .")
		return c
	},
}

type dutListRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  skycmdlib.EnvFlags

	board        string
	model        string
	pool         string
	servoType    string
	useInventory bool
	// nil is a valid value for dims.
	dims map[string]string
}

type dutListParams struct {
	board     string
	model     string
	pool      string
	servoType string
	// dims is the additional dimensions used to filter DUTs.
	// nil is a valid value for dims.
	dims map[string]string
}

func (c *dutListRun) getUseInventory() bool {
	return c.useInventory
}

func (c *dutListRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, errors.Annotate(err, "dut-list").Err())
		return 1
	}
	return 0
}

func (c *dutListRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if len(args) != 0 {
		return cmdlib.NewUsageError(c.Flags, "unexpected positional argument.")
	}

	ctx := cli.GetContext(a, c, env)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}

	params := dutListParams{c.board, c.model, c.pool, c.servoType, c.dims}
	var listed []*swarming.ListedHost

	if c.getUseInventory() {
		// TODO(gregorynisbet): remove this
		panic("QUERYING INVENTORY DIRECTLY NOT YET SUPPORTED")
	} else {
		sc, err := swarming.NewClient(hc, c.envFlags.Env().SwarmingService)
		if err != nil {
			return err
		}
		listed, err = listDutsSwarming(ctx, sc, params)
		if err != nil {
			return err
		}
	}
	for _, l := range listed {
		fmt.Printf("%s\n", l.String())
	}

	return nil
}

func dimsOfDutListParams(params dutListParams) ([]*swarming_api.SwarmingRpcsStringPair, error) {
	// TODO(gregorynisbet): support servoType
	if params.servoType != "" {
		return nil, errors.New("servoType not yet supported")
	}
	makePair := func(key string, value string) *swarming_api.SwarmingRpcsStringPair {
		out := &swarming_api.SwarmingRpcsStringPair{}
		out.Key = key
		out.Value = value
		return out
	}
	out := []*swarming_api.SwarmingRpcsStringPair{makePair("pool", "ChromeOSSkylab")}
	if params.model != "" {
		out = append(out, makePair("label-model", params.model))
	}
	if params.board != "" {
		out = append(out, makePair("label-board", params.board))
	}
	if params.pool != "" {
		out = append(out, makePair("label-pool", params.pool))
	}
	for k, v := range params.dims {
		out = append(out, makePair(k, v))
	}
	return out, nil
}

func listDutsSwarming(ctx context.Context, sc *swarming.Client, params dutListParams) ([]*swarming.ListedHost, error) {
	dims, err := dimsOfDutListParams(params)
	if err != nil {
		return nil, err
	}
	listed, err := sc.GetListedBots(ctx, dims)
	if err != nil {
		return nil, err
	}
	return listed, nil
}
