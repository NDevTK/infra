// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/components/cros/firmware"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

const (
	// Read the AP firmware and dump the sections that we're interested in.
	readAndDumpAPFirmwareCmd = `mkdir /tmp/verify_firmware;` +
		`cd /tmp/verify_firmware; ` +
		`for section in VBLOCK_A VBLOCK_B FW_MAIN_A FW_MAIN_B; ` +
		`do flashrom -p host -r -i $section:$section; ` +
		`done`
	// Verify the firmware blocks A and B.
	verifyFirmwareCmd = `vbutil_firmware --verify /tmp/verify_firmware/VBLOCK_%s` +
		` --signpubkey /usr/share/vboot/devkeys/root_key.vbpubk` +
		` --fv /tmp/verify_firmware/FW_MAIN_%s`
	// Remove the firmware related files we created before.
	removeFirmwareFileCmd = `rm -rf /tmp/verify_firmware`
	// Firmware tarball filename in GS.
	firmwareTarName = "firmware_from_source.tar.bz2"
	// Default mode for chromeos-firmwareupdate when install firmware image.
	defaultFirmwareImageUpdateMode = "recovery"
)

// isFirmwareInGoodState confirms that a host's firmware is in a good state.
//
// For DUTs that run firmware tests, it's possible that the firmware on the DUT can get corrupted.
// This verify action checks whether it appears that firmware should be re-flashed using servo.
func isFirmwareInGoodState(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	_, err := r(ctx, time.Minute, readAndDumpAPFirmwareCmd)
	if err != nil {
		return errors.Annotate(err, "firmware in good state").Err()
	}
	defer func() { r(ctx, time.Minute, removeFirmwareFileCmd) }()
	for _, val := range []string{"A", "B"} {
		_, err := r(ctx, time.Minute, fmt.Sprintf(verifyFirmwareCmd, val, val))
		if err != nil {
			return errors.Annotate(err, "firmware in good state: firmware %s is in a bad state", val).Err()
		}
	}
	return nil
}

// isOnRWFirmwareStableVersionExec confirms that the DUT is currently running the stable version based on its specification.
func isOnRWFirmwareStableVersionExec(ctx context.Context, info *execs.ExecInfo) error {
	sv, err := info.Versioner().Cros(ctx, info.RunArgs.DUT.Name)
	if err != nil {
		return errors.Annotate(err, "on rw firmware stable version").Err()
	}
	err = cros.MatchCrossystemValueToExpectation(ctx, info.DefaultRunner(), "fwid", sv.FwVersion)
	return errors.Annotate(err, "on rw firmware stable version").Err()
}

// isRWFirmwareStableVersionAvailableExec confirms the stable firmware is up to date with the available firmware.
func isRWFirmwareStableVersionAvailableExec(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	sv, err := info.Versioner().Cros(ctx, info.RunArgs.DUT.Name)
	if err != nil {
		return errors.Annotate(err, "rw firmware stable version available").Err()
	}
	modelFirmware, err := ReadFirmwareManifest(ctx, r, info.GetChromeos().GetModel())
	if err != nil {
		return errors.Annotate(err, "rw firmware stable version available").Err()
	}
	availableVersion, err := modelFirmware.AvailableRWFirmware()
	if err != nil {
		return errors.Annotate(err, "rw firmware stable version available").Err()
	}
	stableVersion := sv.FwVersion
	if availableVersion != stableVersion {
		return errors.Reason("rw firmware stable version not available, expected %q, found %q", availableVersion, stableVersion).Err()
	}
	return nil
}

// runFirmwareUpdaterExec run firmware process on the host to flash firmware from installed OS.
//
// Default mode used is autoupdate.
// To reboot device by the end please provide `reboot:by_servo` or `reboot:by_host`.
func runFirmwareUpdaterExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.NewRunner(info.RunArgs.DUT.Name)
	am := info.GetActionArgs(ctx)
	logger := info.NewLogger()
	req := &firmware.FirmwareUpdaterRequest{
		// Options for the mode are: autoupdate, recovery, factory.
		Mode:           am.AsString(ctx, "mode", "autoupdate"),
		Force:          am.AsBool(ctx, "force", false),
		UpdaterTimeout: am.AsDuration(ctx, "updater_timeout", 600, time.Second),
	}
	logger.Debugf("Run firmware update: request to run with %q mode.", req.Mode)
	if err := firmware.RunFirmwareUpdater(ctx, req, run, logger); err != nil {
		logger.Debugf("Run firmware update: fail on run updater. Error: %s", err)
		if am.AsBool(ctx, "no_strict_update", false) {
			logger.Debugf("Run firmware update: continue as allowed updater to fail.")
		} else {
			return errors.Annotate(err, "run firmware update").Err()
		}
	}
	switch am.AsString(ctx, "reboot", "") {
	case "by_servo":
		logger.Debugf("Start DUT reset by servo.")
		if err := info.NewServod().Set(ctx, "power_state", "reset"); err != nil {
			return errors.Annotate(err, "run firmware update: reboot by servo").Err()
		}
	case "by_host":
		logger.Debugf("Start DUT reset by host.")
		if _, err := run(ctx, time.Minute, "reboot && exit"); err != nil {
			logger.Debugf("fail to initiate reboot (not critical): %v", err)
		}
	}
	return nil
}

// runDisableWriteProtectExec disables software-controlled write-protect.
//
// ChromeOS devices have 'host' and 'ec' FPROMs, provide by 'fprom:ec'.
func runDisableFPROMWriteProtectExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.NewRunner(info.RunArgs.DUT.Name)
	am := info.GetActionArgs(ctx)
	fprom := am.AsString(ctx, "fprom", "")
	err := firmware.DisableWriteProtect(ctx, run, info.NewLogger(), info.ActionTimeout, fprom)
	return errors.Annotate(err, "disable fprom: %q write-protect", fprom).Err()
}

func hasDevSignedFirmwareExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.NewRunner(info.RunArgs.DUT.Name)
	if keys, err := firmware.ReadFirmwareKeysFromHost(ctx, run, info.NewLogger()); err != nil {
		return errors.Annotate(err, "has dev signed firmware").Err()
	} else if firmware.IsDevKeys(keys, info.NewLogger()) {
		return nil
	}
	return errors.Reason("has dev signed firmware: dev signed key not found").Err()
}

// updateFirmwareFromFirmwareImage update RW/RO firmware to a given firmwarm image(stable_version by default).
func updateFirmwareFromFirmwareImage(ctx context.Context, info *execs.ExecInfo) error {
	sv, err := info.Versioner().Cros(ctx, info.RunArgs.DUT.Name)
	if err != nil {
		return errors.Annotate(err, "update firmware image").Err()
	}
	actionArgs := info.GetActionArgs(ctx)
	imageName := actionArgs.AsString(ctx, "version_name", sv.FwImage)
	log.Debugf(ctx, "Used fw image name: %s", imageName)
	gsBucket := actionArgs.AsString(ctx, "gs_bucket", gsCrOSImageBucket)
	log.Debugf(ctx, "Used gs bucket name: %s", gsBucket)
	gsImagePath := actionArgs.AsString(ctx, "gs_image_path", fmt.Sprintf("%s/%s", gsBucket, imageName))
	log.Debugf(ctx, "Used fw image path: %s", gsImagePath)
	fwDownloadDir := actionArgs.AsString(ctx, "fw_download_dir", defaultFwFolderPath(info.RunArgs.DUT))
	log.Debugf(ctx, "Used fw image path: %s", gsImagePath)
	// Requesting convert GC path to caches service path.
	// Example: `http://Addr:8082/download/chromeos-image-archive/board-firmware/R99-XXXXX.XX.0`
	downloadPath, err := info.RunArgs.Access.GetCacheUrl(ctx, info.RunArgs.DUT.Name, gsImagePath)
	if err != nil {
		return errors.Annotate(err, "update firmware image").Err()
	}
	fwFileName := actionArgs.AsString(ctx, "fw_filename", firmwareTarName)
	downloadFilename := fmt.Sprintf("%s/%s", downloadPath, fwFileName)
	run := info.DefaultRunner()
	req := &firmware.InstallFirmwareImageRequest{
		DownloadImagePath:    downloadFilename,
		DownloadImageTimeout: actionArgs.AsDuration(ctx, "download_timeout", 120, time.Second),
		DownloadDir:          fwDownloadDir,
		DutRunner:            run,
		Board:                actionArgs.AsString(ctx, "dut_board", info.GetChromeos().GetBoard()),
		Model:                actionArgs.AsString(ctx, "dut_model", info.GetChromeos().GetModel()),
		UpdateEC:             actionArgs.AsBool(ctx, "update_ap", true),
		UpdateAP:             actionArgs.AsBool(ctx, "update_ec", true),
		UpdaterMode:          actionArgs.AsString(ctx, "mode", defaultFirmwareImageUpdateMode),
		UpdaterTimeout:       actionArgs.AsDuration(ctx, "updater_timeout", 600, time.Second),
	}
	if err := firmware.InstallFirmwareImage(ctx, req, info.NewLogger()); err != nil {
		return errors.Annotate(err, "update firmware image").Err()
	}
	if _, err := run(ctx, time.Minute, "reboot && exit"); err != nil {
		return errors.Annotate(err, "update firmware image").Err()
	}
	return nil
}

func init() {
	execs.Register("cros_is_firmware_in_good_state", isFirmwareInGoodState)
	execs.Register("cros_is_on_rw_firmware_stable_version", isOnRWFirmwareStableVersionExec)
	execs.Register("cros_is_rw_firmware_stable_version_available", isRWFirmwareStableVersionAvailableExec)
	execs.Register("cros_has_dev_signed_firmware", hasDevSignedFirmwareExec)
	execs.Register("cros_run_firmware_update", runFirmwareUpdaterExec)
	execs.Register("cros_disable_fprom_write_protect", runDisableFPROMWriteProtectExec)
	execs.Register("cros_update_firmware_from_firmware_image", updateFirmwareFromFirmwareImage)
}
