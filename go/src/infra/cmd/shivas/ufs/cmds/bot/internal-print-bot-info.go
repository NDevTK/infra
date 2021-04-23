// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmds

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/maruel/subcommands"
	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/cli"
	"go.chromium.org/luci/grpc/prpc"

	"infra/cmd/shivas/site"
	"infra/cmd/shivas/utils"
	"infra/cmdsupport/cmdlib"
	"infra/cros/dutstate"
	"infra/libs/skylab/inventory"
	"infra/libs/skylab/inventory/swarming"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// PrintBotInfo subcommand: Print Swarming dimensions for a DUT.
var PrintBotInfo = &subcommands.Command{
	UsageLine: "internal-print-bot-info DUT hostname/Asset tag",
	ShortDesc: "print Swarming bot info for a DUT",
	LongDesc: `Print Swarming bot info for a DUT.

For internal use only.`,
	CommandRun: func() subcommands.CommandRun {
		c := &printBotInfoRun{}
		c.authFlags.Register(&c.Flags, site.DefaultAuthOptions)
		c.envFlags.Register(&c.Flags)
		c.commonFlags.Register(&c.Flags)

		c.Flags.BoolVar(&c.byHostname, "by-hostname", false, "Lookup by hostname instead of ID/Asset tag.")
		return c
	},
}

type printBotInfoRun struct {
	subcommands.CommandRunBase
	authFlags   authcli.Flags
	envFlags    site.EnvFlags
	commonFlags site.CommonFlags

	byHostname bool
}

func (c *printBotInfoRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if err := c.innerRun(a, args, env); err != nil {
		fmt.Fprintf(a.GetErr(), "%s: %s\n", a.GetName(), err)
		return 1
	}
	return 0
}

func (c *printBotInfoRun) innerRun(a subcommands.Application, args []string, env subcommands.Env) error {
	if len(args) != 1 {
		return cmdlib.NewUsageError(c.Flags, "exactly one DUT hostname must be provided")
	}
	ctx := cli.GetContext(a, c, env)
	ctx = utils.SetupContext(ctx, ufsUtil.OSNamespace)
	hc, err := cmdlib.NewHTTPClient(ctx, &c.authFlags)
	if err != nil {
		return err
	}
	e := c.envFlags.Env()
	if c.commonFlags.Verbose() {
		fmt.Printf("Using UnifiedFleet service %s\n", e.UnifiedFleetService)
	}
	ufsClient := ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C:       hc,
		Host:    e.UnifiedFleetService,
		Options: site.DefaultPRPCOptions,
	})
	stderr := a.GetErr()
	r := func(e error) { fmt.Fprintf(stderr, "sanitize dimensions: %s\n", err) }
	bi, err := botInfoForSU(ctx, ufsClient, args[0], r)
	if err != nil {
		if status.Code(err) != codes.NotFound {
			return err
		}
		bi, err = botInfoForDUT(ctx, ufsClient, args[0], c.byHostname, r)
		if err != nil {
			return err
		}
	}
	enc, err := json.Marshal(bi)
	if err != nil {
		return err
	}
	a.GetOut().Write(enc)
	return nil
}

type botInfo struct {
	Dimensions swarming.Dimensions
	State      botState
}

type botState map[string][]string

func botInfoForSU(ctx context.Context, c ufsAPI.FleetClient, id string, r swarming.ReportFunc) (botInfo, error) {
	req := &ufsAPI.GetSchedulingUnitRequest{
		Name: ufsUtil.AddPrefix(ufsUtil.SchedulingUnitCollection, id),
	}
	su, err := c.GetSchedulingUnit(ctx, req)
	if err != nil {
		return botInfo{}, err
	}
	var dutsBotInfo []botInfo
	duts := su.GetMachineLSEs()
	for _, hostname := range duts {
		dbi, err := botInfoForDUT(ctx, c, hostname, true, r)
		if err != nil {
			return botInfo{}, err
		}
		dutsBotInfo = append(dutsBotInfo, dbi)
	}
	dims := map[string][]string{
		"dut_name":        {ufsUtil.RemovePrefix(su.GetName())},
		"dut_id":          {su.GetName()},
		"label-pool":      su.GetPools(),
		"label-dut_count": {fmt.Sprintf("%d", len(duts))},
		"label-multiduts": {"True"},
		"dut_state":       SchedulingUnitDutState(dutsBotInfo),
	}
	if len(duts) > 0 {
		dims["managed_duts"] = duts
	}
	suLabels := []string{"label-board", "label-model"}
	for _, v := range suLabels {
		label := JoinSingleValueLabel(v, dutsBotInfo)
		if len(label) > 0 {
			dims[v] = label
		}
	}
	botInfo := botInfo{
		Dimensions: dims,
		State:      make(botState),
	}
	return botInfo, nil
}

func botInfoForDUT(ctx context.Context, c ufsAPI.FleetClient, id string, byHostname bool, r swarming.ReportFunc) (botInfo, error) {
	req := &ufsAPI.GetChromeOSDeviceDataRequest{}
	if byHostname {
		req.Hostname = id
	} else {
		req.ChromeosDeviceId = id
	}
	data, err := c.GetChromeOSDeviceData(ctx, req)
	if err != nil {
		return botInfo{}, err
	}
	dutStateInfo := dutstate.Read(ctx, c, data.GetLabConfig().GetName())
	dut := data.GetDutV1()
	botInfo := botInfo{
		Dimensions: botDimensionsForDUT(dut, dutStateInfo, r),
		State:      botStateForDUT(dut),
	}
	return botInfo, nil
}

func botStateForDUT(d *inventory.DeviceUnderTest) botState {
	s := make(botState)
	for _, kv := range d.GetCommon().GetAttributes() {
		k, v := kv.GetKey(), kv.GetValue()
		s[k] = append(s[k], v)
	}
	s["storage_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetStorageState().String()[len("HARDWARE_"):]}
	s["servo_usb_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetServoUsbState().String()[len("HARDWARE_"):]}
	s["battery_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetBatteryState().String()[len("HARDWARE_"):]}
	s["rpm_state"] = []string{d.GetCommon().GetLabels().GetPeripherals().GetRpmState().String()}
	return s
}

func botDimensionsForDUT(d *inventory.DeviceUnderTest, ds dutstate.Info, r swarming.ReportFunc) swarming.Dimensions {
	c := d.GetCommon()
	dims := swarming.Convert(c.GetLabels())
	dims["dut_id"] = []string{c.GetId()}
	dims["dut_name"] = []string{c.GetHostname()}
	if v := c.GetHwid(); v != "" {
		dims["hwid"] = []string{v}
	}
	if v := c.GetSerialNumber(); v != "" {
		dims["serial_number"] = []string{v}
	}
	if v := c.GetLocation(); v != nil {
		dims["location"] = []string{formatLocation(v)}
	}
	dims["dut_state"] = []string{string(ds.State)}
	swarming.Sanitize(dims, r)
	return dims
}

func formatLocation(loc *inventory.Location) string {
	return fmt.Sprintf("%s-row%d-rack%d-host%d",
		loc.GetLab().GetName(),
		loc.GetRow(),
		loc.GetRack(),
		loc.GetHost(),
	)
}

func JoinSingleValueLabel(label string, botInfo []botInfo) []string {
	res := make([]string, 0)
	d := make(map[string]int)
	for _, bi := range botInfo {
		if l, ok := bi.Dimensions[label]; ok {
			d[l[0]] += 1
			// If we found a same label appears multiple times, we give
			// them a numeric prefix(e.g. coral, coral2, coral3).
			suffix := ""
			if d[l[0]] > 1 {
				suffix = fmt.Sprintf("%d", d[l[0]])
			}
			res = append(res, l[0]+suffix)
		}
	}
	return res
}

func SchedulingUnitDutState(botInfo []botInfo) []string {
	dutStateMap := map[string]int{
		"ready":               1,
		"needs_repair":        2,
		"repair_failed":       3,
		"needs_manual_repair": 4,
		"needs_replacement":   4,
		"needs_deploy":        4,
	}
	record := 0
	for _, bi := range botInfo {
		if s, ok := bi.Dimensions["dut_state"]; ok {
			if dutStateMap[s[0]] > record {
				record = dutStateMap[s[0]]
			}
		}
	}
	suStateMap := map[int]string{
		0: "unknown",
		1: "ready",
		2: "needs_repair",
		3: "repair_failed",
		4: "needs_manual_attention",
	}
	return []string{suStateMap[record]}
}
