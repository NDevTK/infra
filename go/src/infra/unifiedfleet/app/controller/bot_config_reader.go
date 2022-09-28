// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.chromium.org/luci/common/logging"
	"google.golang.org/protobuf/encoding/prototext"

	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/external"
	"infra/unifiedfleet/app/model/registration"

	ufspb "infra/unifiedfleet/api/v1/models"

	configpb "go.chromium.org/luci/swarming/proto/config"
)

const (
	POOL_PREFIX = "pool:"
)

// ImportENCBotConfig imports Bot Config files and stores the bot configs for ownership data in the DataStore.
func ImportENCBotConfig(ctx context.Context) error {
	es, err := external.GetServerInterface(ctx)
	if err != nil {
		return err
	}
	ownershipConfig := config.Get(ctx).GetOwnershipConfig()
	gitClient, err := es.NewGitInterface(ctx, ownershipConfig.GetGitilesHost(), ownershipConfig.GetProject(), ownershipConfig.GetBranch())
	if err != nil {
		logging.Errorf(ctx, "Got Error for git client : %s", err.Error())
		return fmt.Errorf("failed to initialize connection to Gitiles while importing enc bot configs")
	}

	for _, cfg := range ownershipConfig.GetEncConfig() {
		logging.Debugf(ctx, "########### Parse %s ###########", cfg.GetName())
		conf, err := gitClient.GetFile(ctx, cfg.GetRemotePath())
		if err != nil {
			return err
		}
		content := &configpb.BotsCfg{}
		err = prototext.Unmarshal([]byte(conf), content)
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseBotConfig parses the Bot Config files and stores the ownership data in the Data store for every bot in the config.
func ParseBotConfig(ctx context.Context, config *configpb.BotsCfg, swarmingInstance string) {
	for _, botGroup := range config.BotGroup {
		if len(botGroup.BotId) == 0 {
			continue
		}
		botsIds := []string{}
		for _, id := range botGroup.BotId {
			if strings.Contains(id, "{") {
				// Parse the BotId Range
				botsIds = append(botsIds, parseBotIds(id)...)
			} else {
				botsIds = append(botsIds, id)
			}
		}

		pool := ""
		for _, dim := range botGroup.GetDimensions() {
			// Extract pool from the bot dimensions
			if strings.HasPrefix(dim, POOL_PREFIX) {
				pool = strings.TrimPrefix(dim, POOL_PREFIX)
				break
			}
		}
		ownershipData := &ufspb.OwnershipData{
			PoolName:         pool,
			SwarmingInstance: swarmingInstance,
		}

		// Update the ownership data for the botIds collected so far.
		for _, botId := range botsIds {
			// TODO(b/248039146) - Bot can be one of machine, machine_lse or vm. Store ownership accordingly
			_, err := registration.UpdateMachineOwnership(ctx, botId, ownershipData)
			logging.Debugf(ctx, "Failed to update ownership for bot id %s - %v", botId, err)
		}
	}
}

// parseBotIds parses a range of bot Ids from the input string and returns an array of bot Ids
func parseBotIds(idExpr string) []string {
	prefix := ""
	suffix := ""
	botsIds := []string{}
	if strings.Contains(idExpr, "{") && strings.Contains(idExpr, "}") && strings.Index(idExpr, "{") < strings.Index(idExpr, "}") {
		// Get the prefix and suffix but trimming string after and before '{' and '}'
		prefix = strings.Split(idExpr, "{")[0]
		suffix = strings.Split(idExpr, "}")[1]

		// Get the number range by trimming prefix and suffix
		numRange := strings.TrimSuffix(strings.TrimPrefix(idExpr, prefix+"{"), "}"+suffix)
		nums := strings.Split(numRange, ",")

		for _, num := range nums {
			if !strings.Contains(num, "..") {
				botsIds = append(botsIds, prefix+num+suffix)
			} else {
				// Expand the range of numbers. Ex 1..10 will be expanded to number 1 to 10 inclusive
				rangeIds := strings.Split(num, "..")
				if len(rangeIds) != 2 {
					continue
				}
				start, starterr := strconv.Atoi(rangeIds[0])
				end, enderr := strconv.Atoi(rangeIds[1])

				// Skip this range is it not correctly formed
				if starterr != nil || enderr != nil || start > end {
					continue
				}

				for i := start; i <= end; i++ {
					botsIds = append(botsIds, fmt.Sprintf("%s%d%s", prefix, i, suffix))
				}
			}
		}
	}
	return botsIds
}
