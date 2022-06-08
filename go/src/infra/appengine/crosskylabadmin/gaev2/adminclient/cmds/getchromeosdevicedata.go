// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmds

import (
	"context"
	"fmt"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/common/errors"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"

	"infra/appengine/crosskylabadmin/internal/ufs"
	"infra/appengine/crosskylabadmin/site"
)

// GetChromeOSDeviceData calls the GetMachineLSE RPC of UFS the way that CrOSSkylabAdmin would.
var GetChromeOSDeviceData = &subcommands.Command{
	UsageLine: `get-cros`,
	ShortDesc: `Get the ChromeOS device data`,
	CommandRun: func() subcommands.CommandRun {
		r := &getChromeOSDeviceDataRun{}
		r.authFlags.Register(&r.Flags, site.DefaultAuthOptions)
		r.Flags.StringVar(&r.ufs, "ufs", site.ProdUFS, "the UFS server")
		r.Flags.StringVar(&r.name, "name", "", "the device name")
		return r
	},
}

type getChromeOSDeviceDataRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	ufs       string
	name      string
}

// Run runs the command and returns an exit status.
func (c *getChromeOSDeviceDataRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.innerRun(ctx, a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// innerRun runs the command and returns an error.
func (c *getChromeOSDeviceDataRun) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	if len(args) != 0 {
		return errors.Reason("get chromeos device data: positional arguments are unacceptable").Err()
	}
	authOptions, err := c.authFlags.Options()
	if err != nil {
		return errors.Annotate(err, "get chromeos device data: authenticating").Err()
	}
	hc, err := auth.NewAuthenticator(ctx, auth.InteractiveLogin, authOptions).Client()
	if err != nil {
		return errors.Annotate(err, "get chromeos device data").Err()
	}
	client, err := ufs.NewClient(ctx, hc, c.ufs)
	if err != nil {
		return errors.Annotate(err, "get chromeos device data: creating client").Err()
	}
	req := &ufsAPI.GetChromeOSDeviceDataRequest{
		Hostname: c.name,
	}
	jsonMarshaler.Marshal(a.GetErr(), req)
	res, err := client.GetChromeOSDeviceData(ctx, req)
	if err != nil {
		return errors.Annotate(err, "get machine lse: inner run").Err()
	}
	jsonMarshaler.Marshal(a.GetOut(), res)
	return nil
}
