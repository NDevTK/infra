// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"go.chromium.org/luci/common/errors"
)

// SwarmingBotProvider is the host that runs a swarming bot, e.g. GCE or Drone.
type SwarmingBotProvider string

const (
	// swarmingBotIdEnvName is the env variable name for bot ID
	swarmingBotIdEnvName = "SWARMING_BOT_ID"
	// swarmingBotPrefixDrone is the prefix used in all Drone bots.
	swarmingBotPrefixDrone = "crossk-"
	// swarmingBotPrefixGce is the prefix used in all Gce bots for VMLab.
	// The vmlab PROD pool has a more specific prefix: "chromeos-test-vmlab-"
	// See https://crrev.com/i/5266273. Note that the staging pool is shared.
	swarmingBotPrefixGce = "chromeos-"

	// List of supported SwarmingBotProvider types.
	// TODO(mingkong): consider moving these values to a proto enum
	BotProviderGce     SwarmingBotProvider = "GCE"
	BotProviderDrone   SwarmingBotProvider = "Drone"
	BotProviderUnknown SwarmingBotProvider = "Unknown"
)

// GetHostIp returns the IP address that is accessible from outside the host
func GetHostIp() (string, error) {
	cmd := exec.Command("hostname", "-I")
	stdout, stderr, err := RunCommand(context.Background(), cmd, "hostname", nil, true)
	if err != nil {
		return "", errors.Annotate(err, "Unable to find localhost IP: "+stderr).Err()
	}
	if strings.TrimSpace(stdout) == "" {
		return "", errors.New("Unable to find localhost IP: hostname -I returns no results")
	}
	return strings.Fields(stdout)[0], nil
}

// GetBotProvider detects the SwarmingBotProvider by examining env variable.
func GetBotProvider() SwarmingBotProvider {
	if lookup, found := os.LookupEnv(swarmingBotIdEnvName); found {
		if strings.HasPrefix(lookup, swarmingBotPrefixDrone) {
			return BotProviderDrone
		}
		if strings.HasPrefix(lookup, swarmingBotPrefixGce) {
			return BotProviderGce
		}
	}
	return BotProviderUnknown
}
