// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package swarming contains utilities for skylab swarming tasks.
package util

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/data/strpair"

	swarmingv2 "go.chromium.org/luci/swarming/proto/api_v2"
	"infra/appengine/crosskylabadmin/internal/app/clients"
	"infra/appengine/crosskylabadmin/internal/app/config"
)

const (
	// taskUser is the user for tasks created by Tasker.
	taskUser = "admin-service"
)

// URLForTask returns the task URL for a given task ID.
func URLForTask(ctx context.Context, tid string) string {
	cfg := config.Get(ctx)
	u := url.URL{
		Scheme: "https",
		Host:   cfg.Swarming.Host,
		Path:   "task",
	}
	q := u.Query()
	q.Set("id", tid)
	u.RawQuery = q.Encode()
	return u.String()
}

// AddCommonTags adds some Swarming tags common to all Skylab admin tasks.
func AddCommonTags(ctx context.Context, ts ...string) []string {
	cfg := config.Get(ctx)
	tags := make([]string, 0, len(ts)+2)
	tags = append(tags, cfg.Swarming.LuciProjectTag)
	tags = append(tags, cfg.Swarming.FleetAdminTaskTag)
	tags = append(tags, ts...)
	return tags
}

// SetCommonTaskArgs sets Swarming task arguments common to all Skylab admin tasks.
func SetCommonTaskArgs(ctx context.Context, args *clients.SwarmingCreateTaskArgs) *clients.SwarmingCreateTaskArgs {
	cfg := config.Get(ctx)
	args.User = taskUser
	args.Realm = cfg.Tasker.AdminTaskRealm
	args.ServiceAccount = cfg.Tasker.AdminTaskServiceAccount
	args.Pool = cfg.Swarming.BotPool
	return args
}

// ExtractSingleValuedDimension extracts one specified dimension from a dimension slice.
func ExtractSingleValuedDimension(dims strpair.Map, key string) (string, error) {
	vs, ok := dims[key]
	if !ok {
		return "", fmt.Errorf("failed to find dimension %s", key)
	}
	switch len(vs) {
	case 1:
		return vs[0], nil
	case 0:
		return "", fmt.Errorf("no value for dimension %s", key)
	default:
		return "", fmt.Errorf("multiple values for dimension %s", key)
	}
}

// DimensionsMap converts swarming bot dimensions to a map.
func DimensionsMap(sdims []*swarming.SwarmingRpcsStringListPair) strpair.Map {
	dims := make(strpair.Map)
	for _, sdim := range sdims {
		dims[sdim.Key] = sdim.Value
	}
	return dims
}

// DimensionsMapV2 converts swarming bot dimensions to a map.
func DimensionsMapV2(sdims []*swarmingv2.StringListPair) strpair.Map {
	dims := make(strpair.Map)
	for _, sdim := range sdims {
		dims[sdim.GetKey()] = sdim.GetValue()
	}
	return dims
}

// BotState represents State of the BOT in the swarming
type BotState struct {
	StorageState  []string `json:"storage_state"`
	ServoUSBState []string `json:"servo_usb_state"`
	RpmState      []string `json:"rpm_state"`
}

// ExtractBotState extracts BOTState from BOT info.
func ExtractBotState(botInfo *swarming.SwarmingRpcsBotInfo) BotState {
	state := BotState{}
	if err := json.Unmarshal([]byte(botInfo.State), &state); err != nil {
		fmt.Println(err)
	}
	return state
}
