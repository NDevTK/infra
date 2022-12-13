// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
			updated, assetType, err := isBotOwnershipUpdated(ctx, botId, ownershipData)
			if err != nil {
				logging.Debugf(ctx, "Failed to check if ownership is updated %s - %v", botId, err)
			}
			if updated {
				err = updateOwnership(ctx, botId, ownershipData, assetType)
				if err != nil {
					logging.Debugf(ctx, "Failed to update ownership for bot id %s - %v", botId, err)
				}
			}
			logging.Debugf(ctx, "Nothing to update for bot id %s", botId)
		}
	}
}

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
	errs = append(errs, err)
	return "", errs
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
