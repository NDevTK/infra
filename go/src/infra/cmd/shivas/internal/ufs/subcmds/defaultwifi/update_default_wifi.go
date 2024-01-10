// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package defaultwifi implements the subcommands to operate on UFS DefaultWifi.
package defaultwifi

import (
	"fmt"

	"github.com/maruel/subcommands"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/cmdhelp"
	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// UpdateDefaultWifiCmd Update DefaultWifi by given name.
var UpdateDefaultWifiCmd = &subcommands.Command{
	UsageLine: "DefaultWifi [Options...]",
	ShortDesc: "Update a DefaultWifi",
	LongDesc:  cmdhelp.UpdateDefaultWifiLongDesc,
	CommandRun: func() subcommands.CommandRun {
		c := &updateDefaultWifi{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.StringVar(&c.newSpecsFile, "f", "", cmdhelp.DefaultWifiUpdateFileText)

		c.Flags.StringVar(&c.name, "name", "", "name of the DefaultWifi")
		c.Flags.StringVar(&c.projectID, "project-id", "", "project ID of the GCP Secret Manager hosting the wifi secret")
		c.Flags.StringVar(&c.secretName, "secret-name", "", "the secret name in the GCP Secret Manager")
		return c
	},
}

type updateDefaultWifi struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	newSpecsFile string

	name       string
	projectID  string
	secretName string
}

func (c *updateDefaultWifi) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *updateDefaultWifi) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if err := c.validateArgs(); err != nil {
		return err
	}
	ctx := cli.GetContext(a, c, env)
	ctx = utils.SetupContext(ctx, ufsUtil.OSNamespace)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	if c.commonFlags.Verbose() {
		fmt.Printf("Using UFS service %s\n", e.UnifiedFleetService)
	}
	ic := ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.UnifiedFleetService,
		Options: site.DefaultPRPCOptions,
	})
	var wifi ufspb.DefaultWifi
	if c.newSpecsFile != "" {
		if err = utils.ParseJSONFile(c.newSpecsFile, &wifi); err != nil {
			return err
		}
	} else {
		c.parseArgs(&wifi)
	}
	if err := utils.PrintExistingDefaultWifi(ctx, ic, wifi.Name); err != nil {
		return err
	}
	wifi.Name = ufsUtil.AddPrefix(ufsUtil.DefaultWifiCollection, wifi.Name)
	res, err := ic.UpdateDefaultWifi(ctx, &ufsAPI.UpdateDefaultWifiRequest{
		DefaultWifi: &wifi,
		UpdateMask: utils.GetUpdateMask(&c.Flags, map[string]string{
			"project-id":  "wifi_secret.project_id",
			"secret-name": "wifi_secret.secret_name",
		}),
	})
	if err != nil {
		return err
	}
	res.Name = ufsUtil.RemovePrefix(res.Name)
	fmt.Println("The DefaultWifi after update:")
	utils.PrintProtoJSON(res, !utils.NoEmitMode(false))
	fmt.Printf("Successfully updated the DefaultWifi %s\n", res.Name)
	return nil
}

func (c *updateDefaultWifi) parseArgs(wifi *ufspb.DefaultWifi) {
	wifi.Name = c.name
	wifi.WifiSecret = &ufspb.Secret{ProjectId: c.projectID, SecretName: c.secretName}
}

func (c *updateDefaultWifi) validateArgs() error {
	if c.newSpecsFile != "" {
		if c.name != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-name' cannot be specified at the same time.")
		}
		if c.projectID != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-project-id' cannot be specified at the same time.")
		}
		if c.secretName != "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nThe file mode is specified. '-secret-name' cannot be specified at the same time.")
		}
	}
	if c.newSpecsFile == "" {
		if c.name == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\n'-name' is required, no mode ('-f') is specified.")
		}
		if c.projectID == "" && c.secretName == "" {
			return cmdlib.NewQuietUsageError(c.Flags, "Wrong usage!!\nNeed to specify either '-project-id' or '-secret-name'")
		}
	}
	return nil
}
