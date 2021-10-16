// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/cros/storage"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	readStorageInfoCMD = ". /usr/share/misc/storage-info-common.sh; get_storage_info"
)

// storageStateMap maps state from storageState type to tlw.HardwareState type
var storageStateMap map[storage.StorageState]tlw.HardwareState = map[storage.StorageState]tlw.HardwareState{
	storage.StorageStateNormal:    tlw.HardwareStateNormal,
	storage.StorageStateWarning:   tlw.HardwareStateAcceptable,
	storage.StorageStateCritical:  tlw.HardwareStateNeedReplacement,
	storage.StorageStateUndefined: tlw.HardwareStateUnspecified,
}

// auditStorageSMARTExec confirms that it is able to audi smartStorage info and mark the dut if it needs replacement.
func auditStorageSMARTExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	storageRunner := args.NewRunner(args.ResourceName)
	rawOutput, err := storageRunner(ctx, readStorageInfoCMD)
	if err != nil {
		return errors.Annotate(err, "audit storage smart").Err()
	}
	ss, err := storage.ReadSMARTInfo(ctx, rawOutput)
	if err != nil {
		return errors.Annotate(err, "audit storage smart").Err()
	}
	log.Debug(ctx, "Detected storage type: %q", ss.StorageType)
	log.Debug(ctx, "Detected storage state: %q", ss.StorageState)
	convertedHardwareState, ok := storageStateMap[ss.StorageState]
	if !ok {
		return errors.Reason("audit storage smart: cannot find corresponding hardware state match in the map").Err()
	}
	if convertedHardwareState == tlw.HardwareStateUnspecified {
		return errors.Reason("audit storage smart: DUT storage did not detected or state cannot extracted").Err()
	}
	if convertedHardwareState == tlw.HardwareStateNeedReplacement {
		log.Debug(ctx, "Detected issue with storage on the DUT")
		args.DUT.Storage.State = tlw.HardwareStateNeedReplacement
		return errors.Reason("audit storage smart: hardware state need replacement").Err()
	}
	return nil
}

// hasEnoughStorageSpaceExec confirms the given path has at least the amount of free space specified by the actionArgs arguments.
// provides arguments should be in the formart of:
// ["../../path:x"]
// x is the number of GB of the disk space.
func hasEnoughStorageSpaceExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	path := strings.Split(actionArgs[0], ":")[0]
	pathMinSpaceInGB, convertErr := strconv.ParseFloat(strings.Split(actionArgs[0], ":")[1], 64)
	if convertErr != nil {
		return errors.Annotate(convertErr, "has enough storage space: convert stateful path min space").Err()
	}
	if err := pathHasEnoughValue(ctx, args, args.ResourceName, path, "disk space", pathMinSpaceInGB); err != nil {
		return errors.Annotate(err, "has enough storage space").Err()
	}
	return nil
}

// hasEnoughInodesExec confirms the given path has at least the amount of free inodes specified by the actionArgs arguments.
// provides arguments should be in the formart of:
// ["../../path:x"]
// x is the number of kilos of inodes.
func hasEnoughInodesExec(ctx context.Context, args *execs.RunArgs, actionArgs []string) error {
	path := strings.Split(actionArgs[0], ":")[0]
	pathMinKiloInodes, convertErr := strconv.ParseFloat(strings.Split(actionArgs[0], ":")[1], 64)
	if convertErr != nil {
		return errors.Annotate(convertErr, "has enough storage inodes: convert stateful path min kilo inodes").Err()
	}
	err := pathHasEnoughValue(ctx, args, args.ResourceName, path, "inodes", pathMinKiloInodes*1000)
	return errors.Annotate(err, "has enough storage inodes").Err()
}

// pathHasEnoughValue is a helper function that checks the given path's free disk space / inodes is no less than the min disk space /indoes specified.
func pathHasEnoughValue(ctx context.Context, args *execs.RunArgs, dutName string, path string, typeOfSpace string, minSpaceNeeded float64) error {
	if !IsPathExist(ctx, args, path) {
		return errors.Reason("path has enough value: %s: path: %q not exist", typeOfSpace, path).Err()
	}
	var cmd string
	if typeOfSpace == "disk space" {
		oneMB := math.Pow(10, 6)
		log.Info(ctx, "Checking for >= %f (GB/inodes) of %s under %s on machine %s", minSpaceNeeded, typeOfSpace, path, dutName)
		cmd = fmt.Sprintf(`df -PB %.f %s | tail -1`, oneMB, path)
	} else {
		// checking typeOfSpace == "inodes"
		cmd = fmt.Sprintf(`df -Pi %s | tail -1`, path)
	}
	r := args.NewRunner(dutName)
	output, err := r(ctx, cmd)
	if err != nil {
		return errors.Annotate(err, "path has enough value: %s", typeOfSpace).Err()
	}
	outputList := strings.Fields(output)
	free, err := strconv.ParseFloat(outputList[3], 64)
	if err != nil {
		log.Error(ctx, err.Error())
		return errors.Annotate(err, "path has enough value: %s", typeOfSpace).Err()
	}
	if typeOfSpace == "diskspace" {
		mbPerGB := math.Pow(10, 3)
		free = float64(free) / mbPerGB
	}
	if free < minSpaceNeeded {
		return errors.Reason("path has enough value: %s: Not enough free %s on %s - %f (GB/inodes) free, want %f (GB/inodes)", typeOfSpace, typeOfSpace, path, free, minSpaceNeeded).Err()
	}
	log.Info(ctx, "Found %f (GB/inodes) >= %f (GB/inodes) of %s under %s on machine %s", free, minSpaceNeeded, typeOfSpace, path, dutName)
	return nil
}

func init() {
	execs.Register("cros_audit_storage_smart", auditStorageSMARTExec)
	execs.Register("cros_has_enough_storage_space", hasEnoughStorageSpaceExec)
	execs.Register("cros_has_enough_inodes", hasEnoughInodesExec)
}
