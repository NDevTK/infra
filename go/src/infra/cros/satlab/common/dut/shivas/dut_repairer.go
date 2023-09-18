// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shivas

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"

	"infra/cros/satlab/common/paths"
	"infra/cros/satlab/common/utils/executor"
)

type RepairAction string

const (
	// Verify run only verify actions.
	Verify RepairAction = "-verify"
	// DeepRepair use deep-repair task when scheduling a task.
	DeepRepair = "-deep"
	// Normal don't specify `verify` and `deep` flag to shivas CLI
	Normal = ""
)

// DUTRepairer repairs a DUT with the given name.
type DUTRepairer struct {
	Name     string
	Executor executor.IExecCommander
}

type DUTRepairResponse struct {
	BuildLink string
	TaskLink  string
}

var linkRe = regexp.MustCompile(`(?:(?:https?|ftp):\/\/)?[\w/\-?=%.]+\.[\w/\-&?=%.]+`)

// repair invokes shivas with the required arguments to repair a DUT.
func (u *DUTRepairer) Repair(
	ctx context.Context,
	action RepairAction,
) (*DUTRepairResponse, error) {
	args := []string{
		paths.ShivasCLI,
		"repair-duts",
		string(action),
		u.Name,
	}
	command := exec.CommandContext(ctx, args[0], args[1:]...)
	out, err := u.Executor.Exec(command)
	if err != nil {
		return nil, err
	}

	rawData := string(out)
	// extract the urls from the output
	matches := linkRe.FindAllString(rawData, -1)
	// we expected there are two urls
	if len(matches) != 2 {
		return nil, errors.New(fmt.Sprintf("Can't parse the url from the output: %v\n", rawData))
	}

	return &DUTRepairResponse{BuildLink: matches[0], TaskLink: matches[1]}, nil
}
