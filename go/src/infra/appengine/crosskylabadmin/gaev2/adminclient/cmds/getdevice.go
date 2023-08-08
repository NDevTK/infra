// Copyright 2022 The Chromium Authors
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

	"infra/appengine/crosskylabadmin/internal/ufs"
	"infra/appengine/crosskylabadmin/site"
	shivasUtils "infra/cmd/shivas/utils"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// GetDevice calls the GetMachineLSE RPC of UFS the way that CrOSSkylabAdmin would.
var GetDevice = &subcommands.Command{
	UsageLine: `get-device`,
	ShortDesc: `Get the device`,
	CommandRun: func() subcommands.CommandRun {
		r := &getDeviceRun{}
		r.authFlags.Register(&r.Flags, site.DefaultAuthOptions)
		r.Flags.StringVar(&r.ufs, "ufs", site.ProdUFS, "the UFS server")
		r.Flags.StringVar(&r.name, "name", "", "the device name")
		return r
	},
}

// GetDevice runs the get-device command.
type getDeviceRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	name      string
	ufs       string
}

// Run runs the command and returns an exit status.
func (c *getDeviceRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	ctx := cli.GetContext(a, c, env)
	if err := c.innerRun(ctx, a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

// InnerRun runs the command and returns an error.
func (c *getDeviceRun) innerRun(ctx context.Context, a subcommands.Application, args []string, env subcommands.Env) error {
	ctx = shivasUtils.SetupContext(ctx, ufsUtil.OSNamespace)
	if len(args) != 0 {
		return errors.Reason("get machine lse: positional arguments are unacceptable").Err()
	}
	authOptions, err := c.authFlags.Options()
	if err != nil {
		return errors.Annotate(err, "get machine lse: authenticating").Err()
	}
	hc, err := auth.NewAuthenticator(ctx, auth.InteractiveLogin, authOptions).Client()
	if err != nil {
		return errors.Annotate(err, "get machine lse").Err()
	}
	client, err := ufs.NewClient(ctx, hc, c.ufs)
	if err != nil {
		return errors.Annotate(err, "get machine lse: creating client").Err()
	}
	if c.name == "" {
		return errors.Reason("name cannot be empty").Err()
	}
	req := &ufsAPI.GetDeviceDataRequest{
		Hostname: c.name,
	}
	jsonMarshaler.Marshal(a.GetErr(), req)
	res, err := client.GetDeviceData(ctx, req)
	if err != nil {
		return errors.Annotate(err, "get machine lse: inner run").Err()
	}
	jsonMarshaler.Marshal(a.GetOut(), res)
	return nil
}
