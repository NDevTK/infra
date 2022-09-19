// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package firmware

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/internal/components/servo"
	"infra/cros/recovery/logger"
)

// ReadAPInfoRequest holds request date to read AP info.
type ReadAPInfoRequest struct {
	FilePath string
	// Force extract AP from the DUT.
	ForceExtractAPFile bool
	GBBFlags           bool
	Keys               bool
}

// ReadAPInfoResponse holds response of AP info.
type ReadAPInfoResponse struct {
	GBBFlagsRaw string
	GBBFlags    int
	Keys        []string
}

// ecExemptedModels holds a map of models that doesn't have EC firmware.
var ecExemptedModels = map[string]bool{
	"drallion360": true,
	"sarien":      true,
	"arcada":      true,
	"drallion":    true,
}

// targetOverrideModels holds a map of models that need to override its firmware target.
var targetOverrideModels = map[string]string{
	// TODO(b/226402941): Read existing ec image name using futility.
	"dragonair": "dratini",
	// Models that use _signed version of firmware.
	"drallion360": "drallion_signed",
	"sarien":      "sarien_signed",
	"arcada":      "arcada_signed",
	"drallion":    "drallion_signed",
	"foob360":     "foob",
	"blooglet":    "bloog",
	"garg360":     "garg",
	"laser14":     "phaser",
	"bluebird":    "casta",
	"vorticon":    "meep",
	"dorp":        "meep",
	"orbatrix":    "fleex",
	"blooguard":   "bloog",
	"grabbiter":   "fleex",
	"apel":        "ampton",
	"nospike":     "ampton",
	"phaser360":   "phaser",
	"blorb":       "bobba",
	"droid":       "bobba",
	"garfour":     "garg",
	"vortininja":  "meep",
	"sparky":      "bobba",
	"sparky360":   "bobba",
	"bobba360":    "bobba",
	"mimrock":     "meep",
}

// ReadAPInfoByServo read AP info from DUT.
//
// AP will be extracted from the DUT to flash back with changes.
func ReadAPInfoByServo(ctx context.Context, req *ReadAPInfoRequest, run components.Runner, servod components.Servod, log logger.Logger) (*ReadAPInfoResponse, error) {
	if run == nil || servod == nil || log == nil {
		return nil, errors.Reason("read ap info: run, servod or logger is not provided").Err()
	}
	p, err := NewProgrammer(ctx, run, servod, log)
	if err != nil {
		return nil, errors.Annotate(err, "read ap info").Err()
	}
	defer func() {
		if cerr := p.Close(ctx); cerr != nil {
			log.Debugf("Close programmer fail: %s", cerr)
		}
	}()
	p.Prepare(ctx)
	if err := p.ExtractAP(ctx, req.FilePath, req.ForceExtractAPFile); err != nil {
		return nil, errors.Annotate(err, "read ap info").Err()
	}
	res := &ReadAPInfoResponse{}
	if req.GBBFlags {
		cmd := fmt.Sprintf("gbb_utility --get --flags %s", req.FilePath)
		gbbOut, err := run(ctx, 30*time.Second, cmd)
		if err != nil {
			return nil, errors.Annotate(err, "read ap info: read flags").Err()
		}
		// Parsing output to extract real GBB value.
		parts := strings.Split(gbbOut, ":")
		if len(parts) < 2 {
			return nil, errors.Annotate(err, "read ap info: gbb not found").Err()
		} else if raw := strings.TrimSpace(parts[1]); raw == "" {
			return nil, errors.Annotate(err, "read ap info: gbb not found").Err()
		} else {
			log.Infof("Read GBB raw: %v", raw)
			res.GBBFlagsRaw = raw
		}
		gbb, err := gbbToInt(res.GBBFlagsRaw)
		if err != nil {
			return nil, errors.Annotate(err, "read ap info").Err()
		}
		log.Debugf("Read GBB flags: %v", gbb)
		res.GBBFlags = gbb
	}
	if req.Keys {
		if keys, err := readAPKeysFromFile(ctx, req.FilePath, run, log); err != nil {
			return nil, errors.Annotate(err, "read ap info").Err()
		} else {
			res.Keys = keys
		}
	}
	return res, nil
}

const (
	DevSignedFirmwareKeyPrefix = "b11d"
)

// IsDevKeys checks if any of provided keys are dev signed.
func IsDevKeys(keys []string, log logger.Logger) bool {
	for _, key := range keys {
		if strings.HasPrefix(key, DevSignedFirmwareKeyPrefix) {
			log.Debugf("Found dev signed key: %q !", key)
			return true
		}
	}
	return false
}

// readAPKeysFromFile read firmware keys from the AP image.
func readAPKeysFromFile(ctx context.Context, filePath string, run components.Runner, log logger.Logger) ([]string, error) {
	cmd := fmt.Sprintf("futility show %s |grep \"Key sha1sum:\" |awk '{print $3}'", filePath)
	out, err := run(ctx, time.Minute, cmd)
	if err != nil {
		return nil, errors.Annotate(err, "read ap keys").Err()
	}
	log.Debugf("Read firmware keys: %v", out)
	return strings.Split(out, "\n"), nil
}

// SetApInfoByServoRequest holds and provides info to update AP.
type SetApInfoByServoRequest struct {
	// Path to where AP used or will be extracted
	FilePath string
	// Force extract AP from the DUT.
	ForceExtractAPFile bool
	// Indicates if --force flag should be specified when invoke AP programmer.
	ForceUpdate bool
	// GBB flags value need to be set to AP.
	// Example: 0x18
	GBBFlags       string
	UpdateGBBFlags bool
}

// SetApInfoByServo sets info to AP on the DUT by servo.
//
// AP will be extracted from the DUT to flash back with changes.
func SetApInfoByServo(ctx context.Context, req *SetApInfoByServoRequest, run components.Runner, servod components.Servod, log logger.Logger) error {
	if run == nil || servod == nil || log == nil {
		return errors.Reason("set ap info: run, servod or logger is not provided").Err()
	}
	p, err := NewProgrammer(ctx, run, servod, log)
	if err != nil {
		return errors.Annotate(err, "set ap info").Err()
	}
	defer func() {
		if cerr := p.Close(ctx); cerr != nil {
			log.Debugf("Close programmer fail: %s", cerr)
		}
	}()
	p.Prepare(ctx)
	if err := p.ExtractAP(ctx, req.FilePath, req.ForceExtractAPFile); err != nil {
		return errors.Annotate(err, "set ap info").Err()
	}
	log.Debugf("Set AP info: starting flashing AP to the DUT")
	err = p.ProgramAP(ctx, req.FilePath, req.GBBFlags, req.ForceUpdate, false)
	return errors.Annotate(err, "set ap info: read flags").Err()
}

const (
	extractFileTimeout = 600 * time.Second
	ecMonitorFileName  = "npcx_monitor.bin"
)

// InstallFirmwareImageRequest holds info for InstallFirmwareImage method to flash EC/AP on the DUT.
type InstallFirmwareImageRequest struct {
	// Board and model of the DUT.
	Board string
	Model string

	// Dir where we download the fw image file and then extracted.
	DownloadDir string
	// Path to the fw-Image file and timeout to download it.
	DownloadImagePath    string
	DownloadImageTimeout time.Duration

	// Indicates if --force flag should be specified when invoke chromeos-firmwareupdate or AP programmer.
	ForceUpdate bool

	// Specify how many time to attempt when update EC, where 0 means don't not update EC firmware.
	// Please note attempt count more than 1 only applies when flash via servo.
	// When recover a DUT from corrupted EC use servo, there may be flakiness and we may need flash a couple of times to get a success.
	UpdateEcAttemptCount int
	// Specify how many time to attempt when update AP, where 0 means don't not update AP firmware.
	// Please note attempt count more than 1 only applies when flash via servo.
	// When recover a DUT from corrupted AP use servo, there may be flakiness and we may need flash a couple of times to get a success.
	UpdateApAttemptCount int

	// GBB flags value need to be set to AP.
	// Example: 0x18
	GBBFlags string

	// Flash firmware via servo if true, otherwise flash firmware on DUT itself use chromeos-firmwareupdate.
	FlashThroughServo bool

	// Runner to execute command on the DUT side.
	DutRunner components.Runner
	// Runner to execute command on the servohost.
	ServoHostRunner components.Runner

	// servod instance will be used to collect firmware target info and/or flash image.
	Servod components.Servod

	// Indicates --mode flag used for chromeos-firwmareupdate, only has effect when flash on the DUT side.
	// Possible values is: autoupdate, recovery, factory.
	UpdaterMode string

	// Timeout value for run firmware updater on the DUT, only has effect when flash on the DUT side.
	UpdaterTimeout time.Duration
}

// targetHostRunner returns a runner should be used based on FlashThroughServo flag.
// If flashed via servo, we download firmware and execute flash from servohost(labstation).
// Otherwise we download firmware and execute flash from the DUT directly.
func (req *InstallFirmwareImageRequest) targetHostRunner() components.Runner {
	if req.FlashThroughServo {
		return req.ServoHostRunner
	}
	return req.DutRunner
}

// Helper function to validate InstallFirmwareImage.
func validateInstallFirmwareImageRequest(req *InstallFirmwareImageRequest) error {
	prefix := "validate InstallFirmwareImageRequest: "
	if req == nil {
		return errors.Reason(prefix + "the request is nil").Err()
	} else if req.Board == "" || req.Model == "" {
		return errors.Reason(prefix + "both Board and Model needs to be provided.").Err()
	} else if req.DownloadDir == "" || req.DownloadImagePath == "" || req.DownloadImageTimeout == 0 {
		return errors.Reason(prefix + "both DownloadDir, DownloadImagePath and DownloadImageTimeout needs to be provided.").Err()
	} else if req.UpdateEcAttemptCount == 0 && req.UpdateApAttemptCount == 0 {
		return errors.Reason("validate InstallFirmwareImageRequest both EC and AP attempt count are set to 0, at least one need to be larger than 0.").Err()
	}
	if req.FlashThroughServo {
		// Validating request in the case flash via servo.
		template := prefix + "flash via servo selected but %s is not provided."
		if req.Servod == nil {
			return errors.Reason(fmt.Sprintf(template, "Servod")).Err()
		} else if req.ServoHostRunner == nil {
			return errors.Reason(fmt.Sprintf(template, "ServoHostRunner")).Err()
		}
	} else {
		// Validating request in the case flash from the DUT itself.
		template := prefix + "lash from the DUT selected but %s is not provided."
		if req.UpdaterMode == "" {
			return errors.Reason(fmt.Sprintf(template, "UpdaterMode")).Err()
		} else if req.DutRunner == nil {
			return errors.Reason(fmt.Sprintf(template, "DutRunner")).Err()
		} else if req.UpdaterTimeout == 0 {
			return errors.Reason(fmt.Sprintf(template, "UpdaterTimeout")).Err()
		}
	}
	return nil
}

// InstallFirmwareImage updates a specific AP or/and EC firmware image on the DUT.
func InstallFirmwareImage(ctx context.Context, req *InstallFirmwareImageRequest, log logger.Logger) error {
	log.Debugf("Received request:\n%+v\n", req)
	if err := validateInstallFirmwareImageRequest(req); err != nil {
		return errors.Annotate(err, "install firmware image").Err()
	}
	const (
		// Specify the name used for download file.
		downloadFilename = "fw_image.tar.bz2"
	)
	run := req.targetHostRunner()
	clearDirectory := func() {
		if _, err := run(ctx, time.Minute, "rm", "-rf", req.DownloadDir); err != nil {
			log.Debugf("Failed to remove download directory %q, Error: %s", req.DownloadDir, err)
		}
	}
	// Remove directory in case something left from last times.
	clearDirectory()
	if _, err := run(ctx, time.Minute, "mkdir", "-p", req.DownloadDir); err != nil {
		return errors.Annotate(err, "install firmware image").Err()
	}
	// Always clean up after creating folder as host has limit storage space.
	defer clearDirectory()
	// construct filename for file to download.
	tarballPath := filepath.Join(req.DownloadDir, downloadFilename)
	if out, err := run(ctx, req.DownloadImageTimeout, "curl", req.DownloadImagePath, "--output", tarballPath); err != nil {
		log.Debugf("Output to download fw-image: %s", out)
		return errors.Annotate(err, "install firmware image").Err()
	}
	log.Infof("Successful download tarbar %q from %q", tarballPath, req.DownloadImagePath)
	if ecExemptedModels[req.Model] {
		log.Debugf("Override UpdateEcAttemptCount to 0 as model %s doesn't have EC firmware", req.Model)
		req.UpdateEcAttemptCount = 0
	}
	if req.FlashThroughServo {
		return installFirmwareViaServo(ctx, req, tarballPath, log)
	}
	return installFirmwareImageViaUpdater(ctx, req, tarballPath, log)
}

// installFirmwareImageViaUpdater extract AP or/and EC firmware image from provided tarball and install it via chromeos-firwmareupdate on DUT.
func installFirmwareImageViaUpdater(ctx context.Context, req *InstallFirmwareImageRequest, tarballPath string, log logger.Logger) error {
	updaterReq := FirmwareUpdaterRequest{
		Mode:           req.UpdaterMode,
		UpdaterTimeout: req.UpdaterTimeout,
		Force:          req.ForceUpdate,
	}
	if req.UpdateEcAttemptCount > 0 {
		log.Debugf("Start extraction EC image from %q", tarballPath)
		ecImage, err := extractECImage(ctx, req, tarballPath, log)
		if err != nil {
			return errors.Annotate(err, "install firmware via updater").Err()
		}
		updaterReq.EcImage = ecImage
	}
	if req.UpdateApAttemptCount > 0 {
		log.Debugf("Start extraction AP image from %q", tarballPath)
		apImage, err := extractAPImage(ctx, req, tarballPath, log)
		if err != nil {
			return errors.Annotate(err, "install firmware via updater").Err()
		}
		updaterReq.ApImage = apImage
	}
	return RunFirmwareUpdater(ctx, &updaterReq, req.DutRunner, log)
}

// installFirmwareViaServo extract AP or/and EC firmware image from provided tarball and flash it via servo.
func installFirmwareViaServo(ctx context.Context, req *InstallFirmwareImageRequest, tarballPath string, log logger.Logger) error {
	p, err := NewProgrammer(ctx, req.ServoHostRunner, req.Servod, log)
	if err != nil {
		return errors.Annotate(err, "install firmware via servo").Err()
	}
	if req.UpdateEcAttemptCount > 0 {
		log.Debugf("Start extraction EC image from %q", tarballPath)
		ecImage, err := extractECImage(ctx, req, tarballPath, log)
		if err != nil {
			return errors.Annotate(err, "install firmware via servo").Err()
		}
		log.Debugf("Start program EC image %q", ecImage)
		ecRetryCount := req.UpdateEcAttemptCount
		var ecErr error
		for ecRetryCount > 0 {
			ecRetryCount -= 1
			log.Debugf("Program EC attempt %d, maximum retry: %d", req.UpdateEcAttemptCount-ecRetryCount, req.UpdateEcAttemptCount)
			ecErr = p.ProgramEC(ctx, ecImage)
			if ecErr == nil {
				break
			} else if ecRetryCount > 0 {
				time.Sleep(10 * time.Second)
			}
		}
		if ecErr != nil {
			return errors.Annotate(ecErr, "install firmware via servo").Err()
		}
		log.Infof("Finished program EC image %q", ecImage)
	}
	if req.UpdateApAttemptCount > 0 {
		log.Debugf("Start extraction AP image from %q", tarballPath)
		apImage, err := extractAPImage(ctx, req, tarballPath, log)
		if err != nil {
			return errors.Annotate(err, "install firmware via servo").Err()
		}
		log.Debugf("Start program AP image %q", apImage)
		apRetryCount := req.UpdateApAttemptCount
		var apErr error
		for apRetryCount > 0 {
			apRetryCount -= 1
			log.Debugf("Program AP attempt %d, maximum retry: %d", req.UpdateApAttemptCount-apRetryCount, req.UpdateApAttemptCount)
			// If flash failed in the first attempt, switch to external flashrom instead of libflashrom for remaining attempts.
			useExternalFlashrom := req.UpdateApAttemptCount-apRetryCount > 1
			apErr = p.ProgramAP(ctx, apImage, req.GBBFlags, req.ForceUpdate, useExternalFlashrom)
			if apErr == nil {
				break
			} else if apRetryCount > 0 {
				time.Sleep(10 * time.Second)
			}
		}
		if apErr != nil {
			return errors.Annotate(err, "install firmware via servo").Err()
		}
		log.Infof("Finished program AP image %q", apImage)
	}
	return nil
}

// Helper function to extract EC image from downloaded tarball.
// A ChromeOS device may use firmware image name other than its own board/model, a.k.a firmware target.
// For extract firmware image, we're following below orders to decide firmware target on the DUT:
//
//	(1) Use data in targetOverrideModels if a model appears in the map.
//	(2) Use response from `ec_board` control if available, except when it equal to board/model name.
//	(3) Use name parsed from DUT crossystem_fwid, except when it equal to board/model name.
//	(4) Use model name of the DUT.
//	(5) Use board name of the DUT.
func extractECImage(ctx context.Context, req *InstallFirmwareImageRequest, tarballPath string, log logger.Logger) (string, error) {
	destDir := filepath.Join(filepath.Dir(tarballPath), "EC")
	candidatesFiles := []string{}
	// Handle special case where some model use non-regular firmware mapping.
	if m, ok := targetOverrideModels[req.Model]; ok {
		log.Debugf("Firmware target override detected, DUT model: %s, new firmware target: %s", req.Model, m)
		candidatesFiles = append(candidatesFiles, fmt.Sprintf("%s/ec.bin", m))
	}
	if req.Servod != nil {
		fwBoard, err := servo.GetString(ctx, req.Servod, "ec_board")
		if err != nil {
			log.Debugf("Fail to read `ec_board` value from servo. Skipping.")
		}
		// Based on b:220157423 some board report name is upper case.
		fwBoard = strings.ToLower(fwBoard)
		if fwBoard != "" && fwBoard != req.Model && fwBoard != req.Board {
			candidatesFiles = append(candidatesFiles, fmt.Sprintf("%s/ec.bin", fwBoard))
		}
	}
	run := req.targetHostRunner()
	if !req.FlashThroughServo {
		fwTarget, err := getFirmwareTargetFromDUT(ctx, run, log)
		if err != nil {
			log.Debugf("Failed to get firmware target info from DUT.")
		}
		if fwTarget != "" && fwTarget != req.Model && fwTarget != req.Board {
			candidatesFiles = append(candidatesFiles, fmt.Sprintf("%s/ec.bin", fwTarget))
		}
	}
	candidatesFiles = append(candidatesFiles,
		fmt.Sprintf("%s/ec.bin", req.Model),
		fmt.Sprintf("%s/ec.bin", req.Board),
		"ec.bin",
	)
	imagePath, err := extractFromTarball(ctx, tarballPath, destDir, candidatesFiles, run, log)
	if err != nil {
		return "", errors.Annotate(err, "extract ec files").Err()
	}
	// Extract subsidiary binaries for EC
	// Find a monitor binary for NPCX_UUT chip type, if any.
	var monitorFiles []string
	for _, f := range candidatesFiles {
		monitorFiles = append(monitorFiles, strings.Replace(f, "ec.bin", ecMonitorFileName, 1))
	}
	if _, err := extractFromTarball(ctx, tarballPath, destDir, monitorFiles, run, log); err != nil {
		log.Debugf("Extract EC files: fail to extract %q file. Error: %s", ecMonitorFileName, err)
	}
	return filepath.Join(destDir, imagePath), nil
}

// Helper function to extract BIOS image from downloaded tarball.
// A ChromeOS device may use firmware image name other than its own board/model, a.k.a firmware target.
// For extract firmware image, we're following below orders to decide firmware target on the DUT:
//
//	(1) Use data in targetOverrideModels if a model appears in the map.
//	(2) Use response from `ec_board` control if available, except when it equal to board/model name.
//	(3) Use name parsed from DUT crossystem_fwid, except when it equal to board/model name.
//	(4) Use model name of the DUT.
//	(5) Use board name of the DUT.
func extractAPImage(ctx context.Context, req *InstallFirmwareImageRequest, tarballPath string, log logger.Logger) (string, error) {
	destDir := filepath.Join(filepath.Dir(tarballPath), "AP")
	candidatesFiles := []string{}
	// Handle special case where some model use non-regular firmware mapping.
	if m, ok := targetOverrideModels[req.Model]; ok {
		log.Debugf("Firmware target override detected, DUT model: %s, new firmware target: %s", req.Model, m)
		candidatesFiles = append(candidatesFiles, fmt.Sprintf("image-%s.bin", m))
	}
	if req.Servod != nil {
		fwBoard, err := servo.GetString(ctx, req.Servod, "ec_board")
		if err != nil {
			log.Debugf("Fail to read `ec_board` value from servo. Skipping.")
		}
		// Based on b:220157423 some board report name is upper case.
		fwBoard = strings.ToLower(fwBoard)
		if fwBoard != "" && fwBoard != req.Model && fwBoard != req.Board {
			candidatesFiles = append(candidatesFiles, fmt.Sprintf("image-%s.bin", fwBoard))
		}
	}
	run := req.targetHostRunner()
	if !req.FlashThroughServo {
		fwTarget, err := getFirmwareTargetFromDUT(ctx, run, log)
		if err != nil {
			log.Debugf("Failed to get firmware target info from DUT.")
		}
		if fwTarget != "" && fwTarget != req.Model && fwTarget != req.Board {
			candidatesFiles = append(candidatesFiles, fmt.Sprintf("image-%s.bin", fwTarget))
		}
	}
	candidatesFiles = append(candidatesFiles,
		fmt.Sprintf("image-%s.bin", req.Model),
		fmt.Sprintf("image-%s.bin", req.Board),
		"image.bin",
	)
	imagePath, err := extractFromTarball(ctx, tarballPath, destDir, candidatesFiles, run, log)
	if err != nil {
		return "", errors.Annotate(err, "extract ec files").Err()
	}
	return filepath.Join(destDir, imagePath), nil
}

// Try extracting the image_candidates from the tarball.
func extractFromTarball(ctx context.Context, tarballPath, destDirPath string, candidates []string, run components.Runner, log logger.Logger) (string, error) {
	const (
		// Extract list of files present in archive.
		// To avoid extraction of all files we can limit it t the list of files we interesting in by provide them as arguments at the end.
		tarballListTheFileGlob = "tar tf %s %s"
		// Extract file from the archive.
		tarballExtractTheFileGlob = "tar xf %s -C %s %s"
	)
	// Create the firmware_name subdirectory if it doesn't exist
	if _, err := run(ctx, extractFileTimeout, "mkdir", "-p", destDirPath); err != nil {
		return "", errors.Annotate(err, "extract from tarball: fail to create a destination directory %s", destDirPath).Err()
	}
	// Generate a list of all tarball files
	tarballFiles := make(map[string]bool, 50)
	cmd := fmt.Sprintf(tarballListTheFileGlob, tarballPath, strings.Join(candidates, " "))
	out, err := run(ctx, extractFileTimeout, cmd)
	if err != nil {
		log.Debugf("Fail with error: %s", err)
	}
	log.Debugf("Found candidates: %q", out)
	for _, fn := range strings.Split(out, "\n") {
		tarballFiles[fn] = true
	}
	// Check if image candidates are in the list of tarball files.
	for _, cf := range candidates {
		if !tarballFiles[cf] {
			log.Debugf("Extract from tarball: candidate file %q is not in tarball.", cf)
			continue
		}
		cmd := fmt.Sprintf(tarballExtractTheFileGlob, tarballPath, destDirPath, cf)
		if _, err := run(ctx, extractFileTimeout, cmd); err != nil {
			log.Debugf("Extract from tarball: candidate %q fail to be extracted from tarball.", cf)
		} else {
			log.Infof("Extract from tarball: candidate file %q extracted.", cf)
			return cf, nil
		}
	}
	return "", errors.Reason("extract from tarball: no candidate file found").Err()
}

// getFirmwareTargetFromDUT determine firmware target based on output of crossystem from the DUT.
func getFirmwareTargetFromDUT(ctx context.Context, run components.Runner, log logger.Logger) (string, error) {
	const (
		// An example output of `crossystem fwid` is Google_Fizz.10139.172.0, and what we want is "Fizz" part.
		getFirmwareTargetCmd = "crossystem fwid | awk -F. '{print $1}' | awk -F_ '{print $2}'"
	)
	out, err := run(ctx, time.Second*60, getFirmwareTargetCmd)
	if err != nil {
		return "", errors.Annotate(err, "get firmware target from DUT").Err()
	}
	log.Debugf("Firmware target info from DUT: %s", out)
	// The first letter of firmware target read from DUT is capitalized, so convert to lower case here.
	return strings.ToLower(out), nil
}
