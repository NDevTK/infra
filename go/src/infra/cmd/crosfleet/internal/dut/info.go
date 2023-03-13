// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"fmt"
	"infra/cmd/crosfleet/internal/common"
	dutinfopb "infra/cmd/crosfleet/internal/proto"
	"infra/cmd/crosfleet/internal/site"
	"infra/cmd/crosfleet/internal/ufs"
	"infra/cmdsupport/cmdlib"
	"infra/libs/skylab/common/heuristics"
	ufsapi "infra/unifiedfleet/api/v1/rpc"
	ufsutil "infra/unifiedfleet/app/util"
	"strings"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
)

const (
	infoCmdName = "info"
)

var info = &subcommands.Command{
	UsageLine: fmt.Sprintf("%s HOSTNAME [HOSTNAME...]", infoCmdName),
	ShortDesc: "print DUT information",
	LongDesc: `Print DUT information from UFS.

This command's behavior is subject to change without notice.
Do not build automation around this subcommand.`,
	CommandRun: func() subcommands.CommandRun {
		c := &infoRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.printer.Register(&c.Flags)
		return c
	},
}

type infoRun struct {
	subcommands.CommandRunBase
	authFlags authcli.Flags
	envFlags  common.EnvFlags
	printer   common.CLIPrinter
}

func (c *infoRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		cmdlib.PrintError(a, err)
		return 1
	}
	return 0
}

func (c *infoRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if len(args) == 0 {
		return fmt.Errorf("missing DUT hostname arg")
	}
	ctx := cli.GetContext(a, c, env)
	ufsClient, err := ufs.NewUFSClient(ctx, c.envFlags.Env().UFSService, &c.authFlags)
	if err != nil {
		return err
	}

	var infoList dutinfopb.DUTInfoList
	for _, hostname := range args {
		info, allInfoFound, err := getDutInfo(ctx, ufsClient, hostname)
		// If outputting the command as JSON, collect all DUT info in a proto
		// message first, then print together as one JSON object.
		// Otherwise, just print each separately from this loop.
		// Swallow all errors from here on, since we have partial info to print.
		infoList.DUTs = append(infoList.DUTs, info)
		c.printer.WriteTextStdout("%s\n", dutInfoAsBashVariables(info))
		if !allInfoFound {
			c.printer.WriteTextStdout("Couldn't fetch complete DUT info for %s, possibly due to transient UFS RPC errors;\nrun `crosfleet dut %s %s` to try again", hostname, infoCmdName, hostname)
		}
		if err != nil {
			c.printer.WriteTextStdout("RPC error: %s", err.Error())
		}
	}
	c.printer.WriteJSONStdout(&infoList)
	return nil
}

// getDutInfo returns information about the DUT with the given hostname, and a
// bool indicating whether all information fields were found in UFS.
func getDutInfo(ctx context.Context, ufsClient ufs.Client, hostname string) (*dutinfopb.DUTInfo, bool, error) {
	info := &dutinfopb.DUTInfo{
		Hostname: heuristics.NormalizeBotNameToDeviceName(hostname),
	}

	ctx = contextWithOSNamespace(ctx)
	var err error
	info.LabSetup, err = ufsClient.GetMachineLSE(ctx, &ufsapi.GetMachineLSERequest{
		Name: ufsutil.AddPrefix(ufsutil.MachineLSECollection, info.Hostname),
	})
	if err != nil {
		return info, false, err
	}
	if names := info.LabSetup.GetMachines(); len(names) > 0 && names[0] != "" {
		info.Machine, err = ufsClient.GetMachine(ctx, &ufsapi.GetMachineRequest{
			Name: ufsutil.AddPrefix(ufsutil.MachineCollection, names[0]),
		})
		if err != nil {
			return info, false, err
		}
	}
	allFieldsFound := info.Hostname != "" && info.LabSetup != nil && info.Machine != nil
	return info, allFieldsFound, nil
}

// dutInfoAsBashVariables returns a pretty-printed string containing info about
// the given DUT formatted as bash variables. Only the variables that are found
// in the DUT info proto message are printed.
func dutInfoAsBashVariables(info *dutinfopb.DUTInfo) string {
	var bashVars []string

	hostname := info.GetHostname()
	if hostname != "" {
		bashVars = append(bashVars,
			fmt.Sprintf("DUT_HOSTNAME=%s", hostname))
	}

	chromeOSMachine := info.GetMachine().GetChromeosMachine()
	if chromeOSMachine != nil {
		bashVars = append(bashVars,
			fmt.Sprintf("MODEL=%s\nBOARD=%s",
				chromeOSMachine.GetModel(),
				chromeOSMachine.GetBuildTarget()))
	}

	servo := info.GetLabSetup().GetChromeosMachineLse().GetDeviceLse().GetDut().GetPeripherals().GetServo()
	if servo != nil {
		bashVars = append(bashVars,
			fmt.Sprintf("SERVO_HOSTNAME=%s\nSERVO_PORT=%d\nSERVO_SERIAL=%s",
				servo.GetServoHostname(),
				servo.GetServoPort(),
				servo.GetServoSerial()))
	}

	return strings.Join(bashVars, "\n")
}
