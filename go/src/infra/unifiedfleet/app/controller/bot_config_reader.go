// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/proto/gitiles"
	"go.chromium.org/luci/gae/service/datastore"
	configpb "go.chromium.org/luci/swarming/proto/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/testing/protocmp"

	"infra/libs/git"
	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/external"
	"infra/unifiedfleet/app/model/inventory"
	"infra/unifiedfleet/app/model/registration"

	ufspb "infra/unifiedfleet/api/v1/models"
	api "infra/unifiedfleet/api/v1/rpc"
)

const (
	POOL_PREFIX = "pool:"
)

type ConfigSha struct {
	mu   sync.Mutex
	sha1 string
}

var prevConfig = ConfigSha{sha1: ""}

// ImportBotsConfig gets the OwnershipConfig and git client and passes them to functions for importing bot configs
func ImportBotConfigs(ctx context.Context) error {
	ownershipConfig, gitClient, err := GetConfigAndGitClient(ctx)
	if err != nil {
		return err
	}
	if ownershipConfig == nil {
		return nil
	}

	err = ImportENCBotConfig(ctx, ownershipConfig, gitClient)
	if err != nil {
		return err
	}

	err = ImportSecurityConfig(ctx, ownershipConfig, gitClient)
	return err
}

// GetConfigAndGitClient reads the OwnershipConfig and creates a corresponding git client
func GetConfigAndGitClient(ctx context.Context) (*config.OwnershipConfig, git.ClientInterface, error) {
	es, err := external.GetServerInterface(ctx)
	if err != nil {
		return nil, nil, err
	}
	ownershipConfig := config.Get(ctx).GetOwnershipConfig()
	gitTilesClient, err := es.NewGitTilesInterface(ctx, ownershipConfig.GetGitilesHost())
	currentSha1, err := fetchLatestSHA1(ctx, gitTilesClient, ownershipConfig.GetProject(), ownershipConfig.GetBranch())
	isSameSha1 := prevConfig.compareAndSetSha(currentSha1)
	if isSameSha1 {
		logging.Infof(ctx, "Nothing changed for enc/security config files - lastest SHA1 : %s", currentSha1)
		return nil, nil, nil
	}

	gitClient, err := es.NewGitInterface(ctx, ownershipConfig.GetGitilesHost(), ownershipConfig.GetProject(), ownershipConfig.GetBranch())
	if err != nil {
		logging.Errorf(ctx, "Got Error for git client : %s", err.Error())
		return nil, nil, fmt.Errorf("failed to initialize connection to Gitiles while importing enc bot configs")
	}
	if ownershipConfig == nil {
		logging.Errorf(ctx, "No config found to read ownership data")
		return nil, nil, fmt.Errorf("no config found to read ownership data")
	}
	logging.Infof(ctx, "Importing new enc/security config files - lastest SHA1 : %s", currentSha1)
	return ownershipConfig, gitClient, nil
}

// ImportENCBotConfig imports Bot Config files and stores the bot configs for ownership data in the DataStore.
func ImportENCBotConfig(ctx context.Context, ownershipConfig *config.OwnershipConfig, gitClient git.ClientInterface) error {
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

// ImportSecurityConfig imports Security Config files and stores security data for each bot in the DataStore.
func ImportSecurityConfig(ctx context.Context, ownershipConfig *config.OwnershipConfig, gitClient git.ClientInterface) error {
	logging.Infof(ctx, "Parsing Security config for %d files", len(ownershipConfig.GetSecurityConfig()))
	for _, cfg := range ownershipConfig.GetSecurityConfig() {
		start := time.Now()
		logging.Debugf(ctx, "########### Parse %s ###########", cfg.GetName())
		conf, err := gitClient.GetFile(ctx, cfg.GetRemotePath())
		if err != nil {
			return err
		}
		content := &ufspb.SecurityInfos{}
		err = prototext.Unmarshal([]byte(conf), content)
		if err != nil {
			return err
		}
		ParseSecurityConfig(ctx, content)
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
		err := updateBotConfigForBotIdPrefix(ctx, botGroup.BotIdPrefix, ownershipData)
		if err != nil {
			logging.Debugf(ctx, "Got errors while parsing bot id prefix config for %s - %v", swarmingInstance, err)
		}

		// Update the ownership data for the botIds collected so far.
		err = updateBotConfigForBotIds(ctx, botsIds, ownershipData)
		if err != nil {
			logging.Debugf(ctx, "Got errors while parsing bot id config for %s - %v", swarmingInstance, err)
		}
	}
}

// ParseSecurityConfig parses the Security Config files and stores the security data in the DataStore for every bot in the config.
func ParseSecurityConfig(ctx context.Context, config *ufspb.SecurityInfos) {
	for _, pool := range config.Pools {
		if len(pool.Hosts) == 0 && len(pool.HostPrefixes) == 0 {
			continue
		}
		hosts := []string{}
		for _, host := range pool.Hosts {
			if strings.Contains(host, "{") {
				// Parse the Host Range
				hosts = append(hosts, parseBotIds(host)...)
			} else {
				hosts = append(hosts, host)
			}
		}

		ownershipData := &ufspb.OwnershipData{
			SecurityLevel:    pool.SecurityLevel,
			PoolName:         pool.PoolName,
			SwarmingInstance: pool.SwarmingServerId,
			MibaRealm:        pool.MibaRealm,
			Customer:         pool.Customer,
			Builders:         pool.Builders,
		}

		// Update the ownership for the botIdPrefixes (ie. HostPrefixes)
		err := updateBotConfigForBotIdPrefix(ctx, pool.HostPrefixes, ownershipData)
		if err != nil {
			logging.Debugf(ctx, "Got errors while parsing bot id prefix config for %s - %v", pool.SwarmingServerId, err)
		}

		// Update the ownership data for the botIds (ie. Hosts) collected so far.
		err = updateBotConfigForBotIds(ctx, hosts, ownershipData)
		if err != nil {
			logging.Debugf(ctx, "Got errors while parsing bot id config for %s - %v", pool.SwarmingServerId, err)
		}
	}
}

// ListOwnershipConfigs lists the ownerships based on the specified parameters.
func ListOwnershipConfigs(ctx context.Context, pageSize int32, pageToken, filter string, keysOnly bool) ([]*api.OwnershipByHost, string, error) {
	var filterMap map[string][]interface{}
	var err error
	if filter != "" {
		filterMap, err = getFilterMap(filter, inventory.GetOwnershipIndexedFieldName)
		if err != nil {
			return nil, "", errors.Annotate(err, "failed to read filter for listing Ownerships").Err()
		}
	}
	res, pageToken, err := inventory.ListOwnerships(ctx, pageSize, pageToken, filterMap, keysOnly)
	if err != nil {
		return nil, "", err
	}
	var entities []*api.OwnershipByHost

	for _, entity := range res {
		p, err := entity.GetProto()
		if err != nil {
			logging.Errorf(ctx, "Error parsing entity for ListOwnershipConfigs : %s", err)
		} else {
			pm := p.(*ufspb.OwnershipData)
			ownership := &api.OwnershipByHost{
				Hostname:  entity.Name,
				Ownership: pm,
			}
			entities = append(entities, ownership)
		}
	}
	return entities, pageToken, nil
}

// Updates the Ownership config for the bot ids collected from the config.
func updateBotConfigForBotIds(ctx context.Context, botIds []string, ownershipData *ufspb.OwnershipData) error {
	var errs errors.MultiError
	for _, botId := range botIds {
		updated, assetType, err := isBotOwnershipUpdated(ctx, botId, ownershipData)
		if err != nil && status.Code(err) != codes.NotFound {
			logging.Debugf(ctx, "Failed to check if ownership is updated %s - %v", botId, err)
			errs = append(errs, err)
		} else if !updated {
			logging.Debugf(ctx, "Nothing to update for bot id %s", botId)
		}
		if updated {
			err = updateOwnership(ctx, botId, ownershipData, assetType)
			if err != nil {
				logging.Debugf(ctx, "Failed to update ownership for bot id %s - %v", botId, err)
				errs = append(errs, err)
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
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
	if isOwnershipFieldUpdated(pm.GetCustomer(), newOwnership.GetCustomer()) ||
		isOwnershipFieldUpdated(pm.GetMibaRealm(), newOwnership.GetMibaRealm()) ||
		isOwnershipFieldUpdated(pm.GetSecurityLevel(), newOwnership.GetSecurityLevel()) ||
		isOwnershipFieldUpdated(pm.GetPoolName(), newOwnership.GetPoolName()) ||
		isOwnershipFieldUpdated(pm.GetSwarmingInstance(), newOwnership.GetSwarmingInstance()) ||
		isOwnershipBuildersFieldUpdated(pm.GetBuilders(), newOwnership.GetBuilders()) {
		diff := cmp.Diff(pm, newOwnership, protocmp.Transform())
		logging.Debugf(ctx, "Found ownership diff for bot  %s - %s", botId, diff)
		return true, entity.AssetType, nil
	}
	return false, "", nil
}

// Checks if the Ownership.Builder field was updated, ignoring an empty slice
func isOwnershipBuildersFieldUpdated(oldVal []string, newVal []string) bool {
	if len(newVal) != 0 && len(oldVal) != len(newVal) {
		return true
	}
	for i := range newVal {
		if isOwnershipFieldUpdated(oldVal[i], newVal[i]) {
			return true
		}
	}
	return false
}

// Checks if the ownership field was updated, ignoring empty values
func isOwnershipFieldUpdated(oldVal string, newVal string) bool {
	if (oldVal == "" && newVal != "") || (oldVal != "" && newVal != "" && oldVal != newVal) {
		return true
	}
	return false
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
func updateBotConfigForBotIdPrefix(ctx context.Context, botIdPrefixes []string, ownershipData *ufspb.OwnershipData) error {
	var errs errors.MultiError
	for _, prefix := range botIdPrefixes {
		updated, assetType, err := isBotOwnershipUpdated(ctx, prefix, ownershipData)
		if err != nil && status.Code(err) != codes.NotFound {
			logging.Debugf(ctx, "Failed to check if ownership is updated for prefix %s - %v", prefix, err)
			errs = append(errs, err)
		} else if !updated {
			logging.Debugf(ctx, "Nothing to update for bot id prefix %s", prefix)
		}
		if updated {
			err = updateOwnershipForAssetByPrefix(ctx, prefix, ownershipData, assetType)
			if err != nil {
				logging.Debugf(ctx, "Failed to update ownership for bot id prefix %s - %v", prefix, err)
				errs = append(errs, err)
			} else {
				logging.Debugf(ctx, "updated ownership for bot id prefix %s - %v", prefix, err)
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// Updates Ownership for the given asset type and name prefix
func updateOwnershipForAssetByPrefix(ctx context.Context, prefix string, ownership *ufspb.OwnershipData, assetType string) (err error) {
	return datastore.RunInTransaction(ctx, func(c context.Context) error {
		// First Update the Ownership for the Asset
		switch assetType {
		case inventory.AssetTypeMachine:
			_, err = findAndUpdateMachineOwnershipForPrefix(ctx, prefix, ownership)
		case inventory.AssetTypeMachineLSE:
			_, err = findAndUpdateMachineLSEOwnershipForPrefix(ctx, prefix, ownership)
		case inventory.AssetTypeVM:
			_, err = findAndUpdateVMOwnershipForPrefix(ctx, prefix, ownership)
		default:
			assetType, err = findAndUpdateOwnershipForPrefix(ctx, prefix, ownership)
		}
		if err != nil {
			return err
		}

		// Then update the ownership data table
		_, err = inventory.PutOwnershipData(ctx, ownership, prefix, assetType)
		return err
	}, &datastore.TransactionOptions{})

}

// Searches for entities with names starting with the bot id prefix and
// updates their ownership config
func findAndUpdateOwnershipForPrefix(ctx context.Context, prefix string, ownershipData *ufspb.OwnershipData) (string, error) {
	errs := make(errors.MultiError, 0)
	found, err := findAndUpdateMachineOwnershipForPrefix(ctx, prefix, ownershipData)
	if err != nil {
		errs = append(errs, err)
	}
	if found {
		return inventory.AssetTypeMachine, errs.AsError()
	}
	found, err = findAndUpdateMachineLSEOwnershipForPrefix(ctx, prefix, ownershipData)
	if err != nil {
		errs = append(errs, err)
	}
	if found {
		return inventory.AssetTypeMachineLSE, errs.AsError()
	}
	found, err = findAndUpdateVMOwnershipForPrefix(ctx, prefix, ownershipData)
	if err != nil {
		errs = append(errs, err)
	}
	if found {
		return inventory.AssetTypeVM, errs.AsError()
	}
	return "", errs.AsError()
}

// Searches for machine entities with names starting with the bot id prefix and
// updates their ownership config
func findAndUpdateMachineOwnershipForPrefix(ctx context.Context, prefix string, ownershipData *ufspb.OwnershipData) (bool, error) {
	entities, _, err := registration.ListMachinesByIdPrefixSearch(ctx, -1, "", prefix, true)
	if err != nil || len(entities) == 0 {
		return false, err
	}
	logging.Infof(ctx, "Found %d machines with id prefix %s", len(entities), prefix)
	botsIds := []string{}
	for _, machine := range entities {
		botsIds = append(botsIds, machine.GetName())
	}
	err = updateBotConfigForBotIds(ctx, botsIds, ownershipData)
	return true, err
}

// Searches for machineLSE entities with names starting with the bot id prefix and
// updates their ownership config
func findAndUpdateMachineLSEOwnershipForPrefix(ctx context.Context, prefix string, ownershipData *ufspb.OwnershipData) (bool, error) {
	entities, _, err := inventory.ListMachineLSEsByIdPrefixSearch(ctx, -1, "", prefix, true)
	if err != nil || len(entities) == 0 {
		return false, err
	}
	logging.Infof(ctx, "Found %d machineLSEs with id prefix %s", len(entities), prefix)
	botsIds := []string{}
	for _, machineLSE := range entities {
		botsIds = append(botsIds, machineLSE.GetName())
	}
	err = updateBotConfigForBotIds(ctx, botsIds, ownershipData)
	return true, err
}

// Searches for machineLSE entities with names starting with the bot id prefix and
// updates their ownership config
func findAndUpdateVMOwnershipForPrefix(ctx context.Context, prefix string, ownershipData *ufspb.OwnershipData) (bool, error) {
	entities, _, err := inventory.ListVMsByIdPrefixSearch(ctx, -1, -1, "", prefix, true, nil)
	if err != nil || len(entities) == 0 {
		return false, err
	}
	logging.Infof(ctx, "Found %d VMs with id prefix %s", len(entities), prefix)
	var botIds []string
	for _, vm := range entities {
		botIds = append(botIds, vm.GetName())
	}
	err = updateBotConfigForBotIds(ctx, botIds, ownershipData)
	return true, err
}

// GetOwnershipData gets the ownership data in the Data store for the requested bot in the config.
func GetOwnershipData(ctx context.Context, hostName string) (*ufspb.OwnershipData, error) {
	host, err := inventory.GetOwnershipData(ctx, hostName)
	if err != nil {
		return nil, err
	}
	proto, err := host.GetProto()
	if err != nil {
		return nil, err
	}
	return proto.(*ufspb.OwnershipData), err
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

// Checks if the hashes are equal, updates s.sha1 to the current hash, and returns the comparison result
func (s *ConfigSha) compareAndSetSha(currentSha1 string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	isSameSha := s.sha1 == currentSha1
	s.sha1 = currentSha1
	return isSameSha
}

// Gets the latest SHA1 for the given project and branch.
func fetchLatestSHA1(ctx context.Context, gitilesC external.GitTilesClient, project, branch string) (string, error) {
	resp, err := gitilesC.Log(ctx, &gitiles.LogRequest{
		Project:    project,
		Committish: fmt.Sprintf("refs/heads/%s", branch),
		PageSize:   1,
	})
	if err != nil {
		return "", errors.Annotate(err, "fetch sha1 for %s branch of %s", branch, project).Err()
	}
	if len(resp.Log) == 0 {
		return "", fmt.Errorf("fetch sha1 for %s branch of %s: empty git-log", branch, project)
	}
	return resp.Log[0].GetId(), nil
}
