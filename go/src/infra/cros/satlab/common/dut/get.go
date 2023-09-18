// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/common/commands"
	"infra/cros/satlab/common/dns"
	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/satlabcommands"
	"infra/cros/satlab/common/site"
	"infra/cros/satlab/common/utils/executor"
)

// GetDUT contains the information
// that we want to run with `Shivas CLI`
type GetDUT struct {
	SatlabId      string
	Namespace     string
	HostInfoStore bool

	Zones      []string
	Racks      []string
	Machines   []string
	Prototypes []string
	Tags       []string
	States     []string
	Servos     []string
	Servotypes []string
	Switches   []string
	Rpms       []string
	Pools      []string
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

// GetAssetCLIResopnse contains the infromation from
// `shivasCLI get asset`. We can add the information we needed
type GetAssetCLIResopnse struct {
	Name string    `json:"name"`
	Info AssetInfo `json:"info"`
}

// AssetInfo contains the infromation from
// `shivasCLI get asset`. We can add the information we needed
type AssetInfo struct {
	Model string `json:"model"`
	Board string `json:"buildTarget"`
}

// GetDUTResponse contains the information that we need.
type GetDUTResponse struct {
	GetDUTCLIResponse
	AssetInfo

	// Address local IP
	Address string
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
	out["json"] = []string{}

	return out
}

func makeGetAssetShivasFlags(namespace string) Flagmap {
	out := make(Flagmap)

	out["namespace"] = []string{namespace}
	out["json"] = []string{}
	out["assettype"] = []string{"dut"}

	return out
}

func (g *GetDUT) TriggerRun(
	ctx context.Context,
	executor executor.IExecCommander,
) ([]GetDUTResponse, error) {
	resp := []GetDUTResponse{}
	var err error
	if g.SatlabId == "" {
		g.SatlabId, err = satlabcommands.GetDockerHostBoxIdentifier(ctx, executor)
		if err != nil {
			return resp, errors.Annotate(err, "get dut").Err()
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
		return resp, errors.Annotate(err, "get dut - exec command failed").Err()
	}

	var data []GetDUTCLIResponse
	err = json.Unmarshal(out, &data)
	if err != nil {
		return resp, errors.Annotate(err, "get dut - deserialized DUT json failed").Err()
	}

	flags = makeGetAssetShivasFlags(g.Namespace)
	args = (&commands.CommandWithFlags{
		Commands: []string{paths.ShivasCLI, "get", "asset"},
		Flags:    flags,
	}).ToCommand()
	command = exec.CommandContext(ctx, args[0], args[1:]...)
	out, err = executor.Exec(command)

	if err != nil {
		return resp, errors.Annotate(err, "get dut - exec command failed").Err()
	}

	var assetData []GetAssetCLIResopnse
	err = json.Unmarshal(out, &assetData)
	if err != nil {
		return resp, errors.Annotate(err, "get dut - deserialized asset json failed").Err()
	}

	dnsMap, err := dns.ReadHostsToHostMap(ctx, executor)
	if err != nil {
		return resp, errors.Annotate(err, "get dut - can't get DNS information").Err()
	}

	for _, d := range data {
		assetInfo := AssetInfo{
			Model: "unknown",
			Board: "unknown",
		}
		for _, info := range assetData {
			if len(d.Machines) == 0 {
				return resp, errors.Annotate(err, fmt.Sprintf("get dut - machine is empty, %v\n", d)).
					Err()
			}
			//TODO: consider we have a lot of machines with
			// different board and model. If this possible?
			if info.Name == d.Machines[0] {
				assetInfo = info.Info
				break
			}
		}
		address := dnsMap[d.Hostname]

		resp = append(resp, GetDUTResponse{
			GetDUTCLIResponse: d,
			AssetInfo:         assetInfo,
			Address:           address,
		})
	}

	return resp, nil
}
