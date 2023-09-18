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
	Servo Servo    `json:"servo"`
}

type Servo struct {
	ServoHostname       string `json:"servoHostname"`
	ServoPort           int    `json:"servoPort"`
	ServoSerial         string `json:"servoSerial"`
	ServoType           string `json:"servoType"`
	ServoSetup          string `json:"servoSetup"`
	DockerContainerName string `json:"dockerContainerName"`
}

// GetDUTCLIResponse contains the infromation from
// `shivasCLI get dut`. We can add the information we needed
type MachineLSE struct {
	Name               string `json:"name"`
	ProtoType          string `json:"machineLsePrototype"`
	Hostname           string `json:"hostname"`
	ChromeosMachineLse `         json:"chromeosMachineLse"`
	Machines           []string `json:"machines"`
	Rack               string   `json:"rack"`
	Tags               []string `json:"tags"`
	Zone               string   `json:"zone"`
	Description        string   `json:"description"`
	ResourceState      string   `json:"resourceState"`
	Schedulable        bool     `json:"schedulable"`
	LogicalZone        string   `json:"logicalZone"`
	Realm              string   `json:"realm"`
	// UpdateTime is RFC3339 format
	// We can use `time.Parse(time.RFC3339, t)` to parse it.
	UpdateTime string `json:"updateTime"`
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

	// Default flags
	out["namespace"] = []string{site.GetNamespace(f.Namespace)}
	out["json"] = []string{}

	return out
}

// TriggerRun trigger `shivas get dut` CLI to get the machines information.
func (g *GetDUT) TriggerRun(
	ctx context.Context,
	executor executor.IExecCommander,
) ([]*MachineLSE, error) {
	var err error
	if g.SatlabId == "" {
		g.SatlabId, err = satlabcommands.GetDockerHostBoxIdentifier(ctx, executor)
		if err != nil {
			return nil, errors.Annotate(err, "get dut").Err()
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
		return nil, errors.Annotate(err, "get dut - exec command failed").Err()
	}

	var res []*MachineLSE

	err = json.Unmarshal(out, &res)
	if err != nil {
		return nil, errors.Annotate(err, "get dut - decode to json failed").Err()
	}

	return res, nil
}
