// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"encoding/json"
	"os/exec"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/common/commands"
	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/satlabcommands"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/executor"
)

// GetDUT contains fields used to control behavior when fetching DUTs
type GetDUT struct {
	SatlabId      string
	Namespace     string
	HostInfoStore bool

	// Zone value Name(s) of a zone to filter by.
	Zones []string
	// Rack Name(s) of a rack to filter by.
	Racks []string
	// Machines Name(s) of a machine/asset to filter by.
	Machines []string
	// Prototypes Name(s) of a host prototype to filter by.
	Prototypes []string
	// Tags Name(s) of a tag to filter by.
	Tags []string
	// States Name(s) of a state to filter by.
	// [deployed_testing, needs_repair, decommissioned, ready, missing, deployed_pre_serving, needs_reset, repair_failed, reserved, registered, serving, disabled, deploying, build]
	States []string
	// Servos Name(s) of a servo:port to filter by.
	Servos []string
	// Servotypes Name(s) of a servo type to filter by.
	Servotypes []string
	// Switches Name(s) of a switch to filter by.
	Switches []string
	// Rpms Name(s) of a rpm to filter by.
	Rpms []string
	// Pools Name(s) of a pool to filter by.
	Pools []string

	// OutputFlags decide which format which we want to be.
	OutputFlags site.OutputFlags
}

// ChromeosMachineLse contains the infromation from
// `shivasCLI get dut`. We can add the information we needed
type ChromeosMachineLse struct {
	DeviceLse `json:"deviceLse"`
}

// DeviceLse contains the infromation from
// `shivasCLI get dut`. We can add the information we needed
type DeviceLse struct {
	Dut `json:"dut"`
}

// Dut contains the infromation from
// `shivasCLI get dut`. We can add the information we needed
type Dut struct {
	Pools []string `json:"pools"`
}

// GetDUTCLIResponse contains the infromation from
// `shivasCLI get dut`. We can add the information we needed
type GetDUTCLIResponse struct {
	Name               string   `json:"name"`
	Hostname           string   `json:"hostname"`
	Machines           []string `json:"machines"`
	ChromeosMachineLse `         json:"chromeosMachineLse"`
}

func makeGetDUTShivasFlags(f *GetDUT) Flagmap {
	out := make(Flagmap)

	if len(f.Zones) != 0 {
		out["zone"] = f.Zones
	}
	if len(f.Racks) != 0 {
		out["rack"] = f.Racks
	}
	if len(f.Machines) != 0 {
		out["machine"] = f.Machines
	}
	if len(f.Prototypes) != 0 {
		out["prototype"] = f.Prototypes
	}
	if len(f.Servos) != 0 {
		out["servo"] = f.Servos
	}
	if len(f.Servotypes) != 0 {
		out["servotype"] = f.Servotypes
	}
	if len(f.Switches) != 0 {
		out["switch"] = f.Switches
	}
	if len(f.Rpms) != 0 {
		out["rpms"] = f.Rpms
	}
	if len(f.Pools) != 0 {
		out["pools"] = f.Pools
	}
	if f.HostInfoStore {
		out["host-info-store"] = []string{}
	}
	if f.Namespace != "" {
		out["namespace"] = []string{f.Namespace}
	} else {
		flags := site.EnvFlags{}
		out["namespace"] = []string{flags.GetNamespace()}
	}
	if f.OutputFlags.JSON() {
		out["json"] = []string{}
	}
	if f.OutputFlags.Tsv() {
		out["tsv"] = []string{}
	}
	if f.OutputFlags.Full() {
		out["full"] = []string{}
	}
	if f.OutputFlags.NoEmit() {
		out["noemit"] = []string{}
	}

	return out
}

// TriggerRun trigger `shivas get dut` CLI to get the machines information.
func (g *GetDUT) TriggerRun(ctx context.Context, executor executor.IExecCommander) (string, error) {
	var err error
	if g.SatlabId == "" {
		g.SatlabId, err = satlabcommands.GetDockerHostBoxIdentifier(ctx, executor)
		if err != nil {
			return "", errors.Annotate(err, "get dut").Err()
		}
	}

	if g.Namespace == "" {
		flags := site.EnvFlags{}
		g.Namespace = flags.GetNamespace()
	}

	flags := makeGetDUTShivasFlags(g)

	args := (&commands.CommandWithFlags{
		Commands: []string{paths.ShivasCLI, "get", "dut"},
		Flags:    flags,
	}).ToCommand()
	command := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := executor.Exec(command)

	if err != nil {
		return "", errors.Annotate(err, "get dut - exec command failed").Err()
	}

	return string(out), nil
}

// ParseMachineListOutput parse the `shivas get dut` CLI json output to
// our struct that we are insterested.
func ParseMachineListOutput(output string) ([]GetDUTCLIResponse, error) {
	var result []GetDUTCLIResponse

	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
