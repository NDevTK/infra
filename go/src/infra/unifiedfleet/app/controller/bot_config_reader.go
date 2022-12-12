// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/gae/service/datastore"
	configpb "go.chromium.org/luci/swarming/proto/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/testing/protocmp"

	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/external"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"

	ufspb "infra/unifiedfleet/api/v1/models"
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
	if ownershipConfig == nil {
		logging.Errorf(ctx, "No config found to read ownership data")
		return fmt.Errorf("no config found to read ownership data")
	}

	logging.Infof(ctx, "Parsing Ownership config for %d files", len(ownershipConfig.GetEncConfig()))
	for _, cfg := range ownershipConfig.GetEncConfig() {
		start := time.Now()
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
		ParseBotConfig(ctx, content, cfg.GetName())
		duration := time.Since(start)
		logging.Debugf(ctx, "########### Done Parsing %s; Time taken %s ###########", cfg.GetName(), fmt.Sprintf(duration.String()))
	}
	return nil
}

// ParseBotConfig parses the Bot Config files and stores the ownership data in the Data store for every bot in the config.
func ParseBotConfig(ctx context.Context, config *configpb.BotsCfg, swarmingInstance string) {
	for _, botGroup := range config.BotGroup {
		if len(botGroup.BotId) == 0 && len(botGroup.BotIdPrefix) == 0 {
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

		// Update the ownership for the botIdPrefixes
		updateBotConfigForBotIdPrefix(ctx, botGroup.BotIdPrefix, ownershipData)

		// Update the ownership data for the botIds collected so far.
		updateBotConfigForBotIds(ctx, botsIds, ownershipData)
	}
}

// Updates the Ownership config for the bot ids collected from the config.
func updateBotConfigForBotIds(ctx context.Context, botIds []string, ownershipData *ufspb.OwnershipData) {
	for _, botId := range botIds {
		updated, assetType, err := isBotOwnershipUpdated(ctx, botId, ownershipData)
		if err != nil && status.Code(err) != codes.NotFound {
			logging.Debugf(ctx, "Failed to check if ownership is updated %s - %v", botId, err)
		}
		if updated {
			err = updateOwnership(ctx, botId, ownershipData, assetType)
			if err != nil {
				logging.Debugf(ctx, "Failed to update ownership for bot id %s - %v", botId, err)
			}
		}
	}
}

// Checks if the bot ownership is updated from the last time we read the configs.
func isBotOwnershipUpdated(ctx context.Context, botId string, newOwnership *ufspb.OwnershipData) (bool, string, error) {
	entity, err := inventory.GetOwnershipData(ctx, botId)
	// Update ownership for bot if it does not exist in the ownership table or if there is error in retrieving the entity
	if err != nil {
		return true, "", err
	}
	p, err := entity.GetProto()
	if err != nil {
		return true, entity.AssetType, err
	}
	pm := p.(*ufspb.OwnershipData)
	if diff := cmp.Diff(pm, newOwnership, protocmp.Transform()); diff != "" {
		return true, entity.AssetType, nil
	}
	return false, "", nil
}

// Updates the ownership for the given assetType and name
func updateOwnership(ctx context.Context, botId string, ownership *ufspb.OwnershipData, assetType string) (err error) {
	return datastore.RunInTransaction(ctx, func(c context.Context) error {
		// First Update the Ownership for the Asset
		switch assetType {
		case inventory.AssetTypeMachine:
			_, err = registration.UpdateMachineOwnership(ctx, botId, ownership)
		case inventory.AssetTypeMachineLSE:
			_, err = inventory.UpdateMachineLSEOwnership(ctx, botId, ownership)
		case inventory.AssetTypeVM:
			_, err = inventory.UpdateVMOwnership(ctx, botId, ownership)
		default:
			assetType, err = findAndUpdateOwnershipForAsset(ctx, botId, ownership)
		}
		if err != nil {
			return err
		}

		// Then update the ownership data table
		_, err = inventory.PutOwnershipData(ctx, ownership, botId, assetType)
		return err
	}, &datastore.TransactionOptions{})
}

// Updates the ownership for the given id, searches through machine, machineLSE	 and VM entities as asset type is unknown
func findAndUpdateOwnershipForAsset(ctx context.Context, botId string, ownershipData *ufspb.OwnershipData) (string, error) {
	errs := make(errors.MultiError, 0)
	_, err := registration.UpdateMachineOwnership(ctx, botId, ownershipData)
	if err == nil {
		return inventory.AssetTypeMachine, nil
	}
	if status.Code(err) != codes.NotFound {
		errs = append(errs, err)
	}
	_, err = inventory.UpdateVMOwnership(ctx, botId, ownershipData)
	if err == nil {
		return inventory.AssetTypeVM, nil
	}
	if status.Code(err) != codes.NotFound {
		errs = append(errs, err)
	}
	_, err = inventory.UpdateMachineLSEOwnership(ctx, botId, ownershipData)
	if err == nil {
		return inventory.AssetTypeMachineLSE, nil
	}
	if status.Code(err) != codes.NotFound {
		errs = append(errs, err)
	}
	if errs.First() != nil {
		return "", errs
	}
	return "", nil
}

// Update the Ownership config for the bot id prefixes collected from the config.
func updateBotConfigForBotIdPrefix(ctx context.Context, botIdPrefixes []string,
	ownershipData *ufspb.OwnershipData) {
	for _, prefix := range botIdPrefixes {
		assetType := findAssetTypeForPrefix(ctx, prefix)
		updateOwnershipForAssetByPrefix(ctx, prefix, ownershipData, assetType)
	}
}

// Updates Ownership for the given asset type and name prefix
func updateOwnershipForAssetByPrefix(ctx context.Context, prefix string, ownership *ufspb.OwnershipData, assetType string) (err error) {
	// First Update the Ownership for the Asset
	switch assetType {
	case inventory.AssetTypeMachine:
		findAndUpdateMachineOwnershipForPrefix(ctx, prefix, ownership)
	case inventory.AssetTypeMachineLSE:
		findAndUpdateMachineLSEOwnershipForPrefix(ctx, prefix, ownership)
	default:
		findAndUpdateOwnershipForPrefix(ctx, prefix, ownership)
	}
	return err
}

// Searches the Ownership table to find the asset type for this prefix
func findAssetTypeForPrefix(ctx context.Context,
	prefix string) string {
	entities, _, err := inventory.ListHostsByIdPrefixSearch(ctx, 1, "", prefix, false)
	// Did not find any entries in the ownership table, return empty assetType
	if err != nil || len(entities) == 0 {
		return ""
	}
	return entities[0].AssetType
}

// Searches for entities with names starting with the bot id prefix and
// updates their ownership config
func findAndUpdateOwnershipForPrefix(ctx context.Context,
	prefix string, ownershipData *ufspb.OwnershipData) bool {
	found := findAndUpdateMachineOwnershipForPrefix(ctx, prefix, ownershipData)
	if !found {
		found = findAndUpdateMachineLSEOwnershipForPrefix(ctx, prefix, ownershipData)
	}
	return found
}

// Searches for machine entities with names starting with the bot id prefix and
// updates their ownership config
func findAndUpdateMachineOwnershipForPrefix(ctx context.Context,
	prefix string, ownershipData *ufspb.OwnershipData) bool {
	entities, _, err := registration.ListMachinesByIdPrefixSearch(ctx, -1, "", prefix, true)
	if err != nil || len(entities) == 0 {
		return false
	}
	logging.Infof(ctx, "Found %d machines with id prefix %s", len(entities), prefix)
	botsIds := []string{}
	for _, machine := range entities {
		botsIds = append(botsIds, machine.GetName())
	}
	updateBotConfigForBotIds(ctx, botsIds, ownershipData)
	return true
}

// Searches for machineLSE entities with names starting with the bot id prefix and
// updates their ownership config
func findAndUpdateMachineLSEOwnershipForPrefix(ctx context.Context,
	prefix string, ownershipData *ufspb.OwnershipData) bool {
	entities, _, err := inventory.ListMachineLSEsByIdPrefixSearch(ctx, -1, "", prefix, true)
	if err != nil || len(entities) == 0 {
		return false
	}
	logging.Infof(ctx, "Found %d machineLSEs with id prefix %s", len(entities), prefix)
	botsIds := []string{}
	for _, machineLSE := range entities {
		botsIds = append(botsIds, machineLSE.GetName())
	}
	updateBotConfigForBotIds(ctx, botsIds, ownershipData)
	return true
}

// GetOwnershipData gets the ownership data in the Data store for the requested bot in the config.
func GetOwnershipData(ctx context.Context, hostName string) (*ufspb.OwnershipData, error) {
	// Check if the host is a machine
	host, err := registration.GetMachine(ctx, hostName)
	if err == nil {
		return host.GetOwnership(), nil
	} else if status.Code(err) != codes.NotFound {
		return nil, err
	}
	vm, err := inventory.GetVM(ctx, hostName)
	if err == nil {
		return vm.GetOwnership(), nil
	} else if status.Code(err) != codes.NotFound {
		return nil, err
	}

	machineLse, err := inventory.GetMachineLSE(ctx, hostName)
	if err == nil {
		return machineLse.GetOwnership(), nil
	}
	return nil, err
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
