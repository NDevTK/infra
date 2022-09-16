// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"strconv"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/dutstate"
	"infra/cros/recovery/internal/components/cros/storage"
	"infra/cros/recovery/internal/components/linux"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	readStorageInfoCMD = ". /usr/share/misc/storage-info-common.sh; get_storage_info"
)

// storageStateMap maps state from storageState type to tlw.HardwareState type
var storageStateMap = map[storage.StorageState]tlw.HardwareState{
	storage.StorageStateNormal:    tlw.HardwareState_HARDWARE_NORMAL,
	storage.StorageStateWarning:   tlw.HardwareState_HARDWARE_ACCEPTABLE,
	storage.StorageStateCritical:  tlw.HardwareState_HARDWARE_NEED_REPLACEMENT,
	storage.StorageStateUndefined: tlw.HardwareState_HARDWARE_UNSPECIFIED,
}

// auditStorageSMARTExec confirms that it is able to audi smartStorage info and mark the dut if it needs replacement.
func auditStorageSMARTExec(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	if info.GetChromeos().GetStorage() == nil {
		return errors.Reason("audit storage smart: data is not present in dut info").Err()
	}
	rawOutput, err := r(ctx, time.Minute, readStorageInfoCMD)
	if err != nil {
		return errors.Annotate(err, "audit storage smart").Err()
	}
	ss, err := storage.ParseSMARTInfo(ctx, rawOutput)
	if err != nil {
		return errors.Annotate(err, "audit storage smart").Err()
	}
	log.Debugf(ctx, "Detected storage type: %q", ss.StorageType)
	log.Debugf(ctx, "Detected storage state: %q", ss.StorageState)
	convertedHardwareState, ok := storageStateMap[ss.StorageState]
	if !ok {
		return errors.Reason("audit storage smart: cannot find corresponding hardware state match in the map").Err()
	}
	if convertedHardwareState == tlw.HardwareState_HARDWARE_UNSPECIFIED {
		return errors.Reason("audit storage smart: DUT storage did not detected or state cannot extracted").Err()
	}
	if convertedHardwareState == tlw.HardwareState_HARDWARE_NEED_REPLACEMENT {
		log.Debugf(ctx, "Detected issue with storage on the DUT")
		info.GetChromeos().GetStorage().State = tlw.HardwareState_HARDWARE_NEED_REPLACEMENT
		log.Debugf(ctx, "Audit Storage Smart: setting the dut state to :%s", string(dutstate.NeedsReplacement))
		info.RunArgs.DUT.State = dutstate.NeedsReplacement
		return errors.Reason("audit storage smart: hardware state need replacement").Err()
	}
	return nil
}

// hasEnoughStorageSpaceExec confirms the given path has at least the amount of free space specified by the actionArgs arguments.
// provides arguments should be in the formart of:
// ["path:x"]
// x is the number of GB of the disk space.
// input will only consist of one path and its corresponding value for storage.
func hasEnoughStorageSpaceExec(ctx context.Context, info *execs.ExecInfo) error {
	// TODO: recheck it and simplify. Also do it for hasEnoughStoragePercentageExec
	if len(info.ActionArgs) != 1 {
		return errors.Reason("has enough storage space: input in wrong format").Err()
	}
	inputs := strings.Split(info.ActionArgs[0], ":")
	if len(inputs) != 2 {
		return errors.Reason("has enough storage space: input in wrong format").Err()
	}
	path := inputs[0]
	pathMinSpaceInGB, convertErr := strconv.ParseFloat(inputs[1], 64)
	if convertErr != nil {
		return errors.Annotate(convertErr, "has enough storage space: convert stateful path min space").Err()
	}
	if err := linux.PathHasEnoughValue(ctx, info.DefaultRunner(), info.RunArgs.ResourceName, path, linux.SpaceTypeDisk, pathMinSpaceInGB); err != nil {
		return errors.Annotate(err, "has enough storage space").Err()
	}
	return nil
}

// hasEnoughFreeIndexNodesExec confirms the given path has at least the amount of free index nodes specified by the actionArgs arguments.
// provides arguments should be in the formart of:
// ["path:x"]
// x is the number of kilos of index nodes.
// input will only consist of one path and its corresponding value for storage.
func hasEnoughFreeIndexNodesExec(ctx context.Context, info *execs.ExecInfo) error {
	if len(info.ActionArgs) != 1 {
		return errors.Reason("has enough index nodes: input in wrong format").Err()
	}
	inputs := strings.Split(info.ActionArgs[0], ":")
	if len(inputs) != 2 {
		return errors.Reason("has enough index nodes: input in wrong format").Err()
	}
	path := inputs[0]
	pathMinKiloIndexNodes, convertErr := strconv.ParseFloat(inputs[1], 64)
	if convertErr != nil {
		return errors.Annotate(convertErr, "has enough storage index nodes: convert stateful path min kilo nodes").Err()
	}
	err := linux.PathHasEnoughValue(ctx, info.DefaultRunner(), info.RunArgs.ResourceName, path, linux.SpaceTypeInode, pathMinKiloIndexNodes*1000)
	return errors.Annotate(err, "has enough storage index nodes").Err()
}

// hasEnoughStorageSpacePercentageExec confirms the given path has at least the percentage of free space specified by the actionArgs arguments.
// provides arguments should be in the formart of:
// ["path:x"]
// x is the percentage of the disk space.
// input will only consist of one path and its corresponding percentage for storage.
func hasEnoughStorageSpacePercentageExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	path := argsMap.AsString(ctx, "path", "")
	pathMinSpaceInPercentage := argsMap.AsFloat64(ctx, "expected", -1)
	if path == "" {
		return errors.Reason("has enough storage space percentage: missing path argument").Err()
	}
	if pathMinSpaceInPercentage < 0 || pathMinSpaceInPercentage > 100 {
		return errors.Reason("has enough storage space percentage: invalid value for expected argument %e", pathMinSpaceInPercentage).Err()
	}
	if occupiedSpacePercentage, err := linux.PathOccupiedSpacePercentage(ctx, info.DefaultRunner(), path); err != nil {
		return errors.Annotate(err, "has enough storage space percentage").Err()
	} else if actualFreePercentage := (100 - occupiedSpacePercentage); pathMinSpaceInPercentage > actualFreePercentage {
		return errors.Reason("path have enough free space percentage: %s/%s, expect %v%%, actual %v%%", info.RunArgs.ResourceName, path, pathMinSpaceInPercentage, actualFreePercentage).Err()
	}
	return nil
}

func init() {
	execs.Register("cros_audit_storage_smart", auditStorageSMARTExec)
	execs.Register("cros_has_enough_storage_space", hasEnoughStorageSpaceExec)
	execs.Register("cros_has_enough_storage_space_percentage", hasEnoughStorageSpacePercentageExec)
	execs.Register("cros_has_enough_index_nodes", hasEnoughFreeIndexNodesExec)
}
