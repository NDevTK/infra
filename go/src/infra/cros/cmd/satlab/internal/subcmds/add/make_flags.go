// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package add

import (
	"fmt"
	"strings"
)

type flagmap = map[string][]string

func makeShivasFlags(c *addDUT) flagmap {
	out := make(flagmap)

	// These other flags are inherited from shivas.
	if c.newSpecsFile != "" {
		out["f"] = []string{c.newSpecsFile}
	}
	if c.zone != "" {
		out["zone"] = []string{c.zone}
	}
	if c.rack != "" {
		out["rack"] = []string{c.rack}
	}
	if c.hostname != "" {
		out["name"] = []string{c.hostname}
	}
	if c.asset != "" {
		out["asset"] = []string{c.asset}
	}
	if c.servo != "" {
		out["servo"] = []string{c.servo}
	}
	if c.servoSerial != "" {
		out["servo-serial"] = []string{c.servoSerial}
	}
	if c.servoSetupType != "" {
		out["servo-setup"] = []string{c.servoSetupType}
	}
	if len(c.pools) != 0 {
		out["pools"] = []string{strings.Join(c.pools, ",")}
	}
	if len(c.licenseTypes) != 0 {
		out["licensetype"] = []string{strings.Join(c.licenseTypes, ",")}
	}
	if c.rpm != "" {
		out["rpm"] = []string{c.rpm}
	}
	if c.rpmOutlet != "" {
		out["rpm-outlet"] = []string{c.rpmOutlet}
	}
	if c.deployTaskTimeout != 0 {
		out["deploy-timeout"] = []string{fmt.Sprintf("%d", c.deployTaskTimeout)}
	}
	if c.ignoreUFS {
		out["ignore-ufs"] = []string{}
	}
	if len(c.deployTags) != 0 {
		out["deploy-tags"] = []string{strings.Join(c.deployTags, ",")}
	}
	if c.deploySkipDownloadImage {
		out["deploy-skip-download-image"] = []string{}
	}
	if c.deploySkipInstallFirmware {
		out["deploy-skip-install-firmware"] = []string{}
	}
	if c.deploySkipInstallOS {
		out["deploy-skip-install-os"] = []string{}
	}
	if len(c.tags) != 0 {
		out["tags"] = []string{strings.Join(c.tags, ",")}
	}
	if c.state != "" {
		out["state"] = []string{c.state}
	}
	if c.description != "" {
		out["desc"] = []string{c.description}
	}
	if len(c.chameleons) != 0 {
		out["chameleons"] = []string{strings.Join(c.chameleons, ",")}
	}
	if len(c.cameras) != 0 {
		out["cameras"] = []string{strings.Join(c.cameras, ",")}
	}
	if len(c.cables) != 0 {
		out["cables"] = []string{strings.Join(c.cables, ",")}
	}
	if c.antennaConnection != "" {
		out["antennaconnection"] = []string{c.antennaConnection}
	}
	if c.router != "" {
		out["router"] = []string{c.router}
	}
	if c.facing != "" {
		out["facing"] = []string{c.facing}
	}
	if c.light != "" {
		out["light"] = []string{c.light}
	}
	if c.carrier != "" {
		out["carrier"] = []string{c.carrier}
	}
	if c.audioBoard {
		out["audioboard"] = []string{}
	}
	if c.audioBox {
		out["audiobox"] = []string{}
	}
	if c.atrus {
		out["atrus"] = []string{}
	}
	if c.wifiCell {
		out["wificell"] = []string{}
	}
	if c.touchMimo {
		out["touchmimo"] = []string{}
	}
	if c.cameraBox {
		out["camerabox"] = []string{}
	}
	if c.chaos {
		out["chaos"] = []string{}
	}
	if c.audioCable {
		out["audiocable"] = []string{}
	}
	if c.smartUSBHub {
		out["smartusbhub"] = []string{}
	}
	if c.model != "" {
		out["model"] = []string{}
	}
	if c.board != "" {
		out["board"] = []string{}
	}
	return out
}

func appendFlag(arr []string, flag string, values ...string) []string {
	arr = append(arr, fmt.Sprintf("-%s", flag))
	for _, v := range values {
		arr = append(arr, v)
	}
	return arr
}
