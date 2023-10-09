// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package shivas

import (
	"bytes"
	"os/exec"

	"go.chromium.org/luci/common/errors"

	"infra/cros/satlab/common/commands"
	"infra/cros/satlab/common/paths"
	e "infra/cros/satlab/common/utils/errors"
	"infra/cros/satlab/common/utils/executor"
	"infra/cros/satlab/common/utils/misc"
)

// Asset is a group of parameters needed to add an asset to UFS.
type Asset struct {
	Asset     string
	Rack      string
	Zone      string
	Model     string
	Board     string
	Namespace string
	Type      string
}

// CheckAndAdd adds the asset if it does not already exist.
func (a *Asset) CheckAndAdd(executor executor.IExecCommander) (string, error) {
	assetMsg, err := a.check(executor)
	if err != nil {
		return "", errors.Annotate(err, "check and update").Err()
	}
	if len(assetMsg) == 0 {
		return a.add(executor)
	} else {
		return "", e.AssetExist
	}
}

// Check checks for the existence of the UFS asset.
func (a *Asset) check(executor executor.IExecCommander) (string, error) {
	args := (&commands.CommandWithFlags{
		Commands:       []string{paths.ShivasCLI, "get", "asset"},
		PositionalArgs: []string{a.Asset},
		Flags: map[string][]string{
			"rack":      {a.Rack},
			"zone":      {a.Zone},
			"model":     {a.Model},
			"board":     {a.Board},
			"namespace": {a.Namespace},
			// Type cannot be provided when getting a DUT.
		},
	}).ToCommand()

	var b bytes.Buffer
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = &b
	assetMsgBytes, err := executor.Exec(command)

	if err != nil {
		return "", errors.Annotate(err, "check asset - %s", b.String()).Err()
	}
	assetMsg := misc.TrimOutput(assetMsgBytes)

	return assetMsg, nil
}

// Add adds an asset unconditionally to UFS.
func (a *Asset) add(executor executor.IExecCommander) (string, error) {
	// Add the asset.
	args := (&commands.CommandWithFlags{
		Commands: []string{paths.ShivasCLI, "add", "asset"},
		Flags: map[string][]string{
			"model":     {a.Model},
			"board":     {a.Board},
			"rack":      {a.Rack},
			"zone":      {a.Zone},
			"name":      {a.Asset},
			"namespace": {a.Namespace},
			"type":      {a.Type},
		},
	}).ToCommand()

	var b bytes.Buffer
	command := exec.Command(args[0], args[1:]...)
	command.Stderr = &b
	out, err := executor.Exec(command)
	if err != nil {
		return "", errors.Annotate(err, "add asset - %s", b.String()).Err()
	}

	return string(out), nil
}
