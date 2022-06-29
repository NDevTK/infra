// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package tlslib provides the canonical implementation of a common TLS server.
package tlslib

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/ssh"
)

const (
	firmwareUpdaterPath = "/usr/sbin/chromeos-firmwareupdate"
)

type FirmwareManifest map[string]FirmwareManifestData

type FirmwareManifestData struct {
	Host struct {
		Versions struct {
			Ro string `json:"ro"`
			Rw string `json:"rw"`
		} `json:"versions"`
		Keys struct {
			Root     string `json:"root"`
			Recovery string `json:"recovery"`
		} `json:"keys"`
		Image string `json:"image"`
	} `json:"host"`
	Ec struct {
		Versions struct {
			Ro string `json:"ro"`
			Rw string `json:"rw"`
		} `json:"versions"`
		Image string `json:"image"`
	} `json:"ec"`
	SignatureId string `json:"signature_id"`
}

// getAvailableFirmwareVersion read firmware manifest from current OS and extract available firmware version based on model.
func getAvailableFirmwareVersion(c *ssh.Client) (string, error) {
	out, err := runCmdOutput(c, fmt.Sprintf("%s --manifest", firmwareUpdaterPath))
	if err != nil {
		return "", fmt.Errorf("getAvailableFirmwareVersion: failed to get firmware manifest, %s", err)
	}
	var manifest FirmwareManifest
	if err := json.Unmarshal([]byte(out), &manifest); err != nil {
		return "", fmt.Errorf("getAvailableFirmwareVersion: failed to unmarshal firmware manifest, %s", err)
	}
	fwModel, err := runCmdOutput(c, "crosid | grep FIRMWARE_MANIFEST_KEY | cut -d\"'\" -f2")
	if err != nil {
		return "", fmt.Errorf("getAvailableFirmwareVersion: failed to get FIRMWARE_MANIFEST_KEY, %s", err)
	}
	fwModel = strings.TrimSuffix(fwModel, "\n")
	if data, ok := manifest[fwModel]; ok {
		return data.Host.Versions.Rw, nil
	}
	return "", fmt.Errorf("getAvailableFirmwareVersion: failed to get firmware data of key %s from manifest, %s", fwModel, err)
}

// getCurrentFirmwareVersion read current system firmware version on the DUT.
func getCurrentFirmwareVersion(c *ssh.Client) (string, error) {
	out, err := runCmdOutput(c, "crossystem fwid")
	if err != nil {
		return "", fmt.Errorf("getCurrentFirmwareVersion: failed to read current system firmware, %s", err)
	}
	return out, nil
}

func isFirmwareUpdateToDate(c *ssh.Client) (bool, error) {
	currentFw, err := getCurrentFirmwareVersion(c)
	if err != nil {
		return false, err
	}
	availableFw, err := getAvailableFirmwareVersion(c)
	if err != nil {
		return false, err
	}
	log.Printf("Current firmware on DUT: %s, available firmware from new OS: %s.", currentFw, availableFw)
	return currentFw == availableFw, nil
}

// updateFirmware update DUT's firmware(RW) to current available version from OS image.
func (p *provisionState) updateFirmware(ctx context.Context) (bool, error) {
	if err := runCmd(p.c, fmt.Sprintf("%s --wp=1 --mode=autoupdate", firmwareUpdaterPath)); err != nil {
		return false, fmt.Errorf("updateFirmware: failed to execute chromeos-firmwareupdate, %s", err)
	}
	fwChanged, err := isFirmwareSlotChanged(p.c)
	if err != nil {
		return false, err
	}
	if p.preventReboot {
		log.Printf("updateFirmware: reboot prevented by request")
		return fwChanged, nil
	}
	if fwChanged {
		log.Printf("Firmware slot changed on next boot, rebooting the DUT.")
		if err := rebootDUT(ctx, p.c); err != nil {
			return fwChanged, fmt.Errorf("updateFirmware: failed to reboot DUT, %s", err)
		}
	}
	return fwChanged, nil
}

func isFirmwareSlotChanged(c *ssh.Client) (bool, error) {
	current, err := runCmdOutput(c, "crossystem mainfw_act")
	if err != nil {
		return false, fmt.Errorf("isFirmwareSlotChanged: failed to get current active main firmware slot, %s", err)
	}
	next, err := runCmdOutput(c, "crossystem fw_try_next")
	if err != nil {
		return false, fmt.Errorf("isFirmwareSlotChanged: failed to get next main firmware slot, %s", err)
	}
	log.Printf("Current active firmware slot: %s, next boot firmware slot: %s", current, next)
	return current != next, nil
}
