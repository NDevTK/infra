// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

// roVPDKeys is the list of keys from RO_VPD that should be persisted and flashed in the recovery process.
var roVPDKeys = []string{
	"wifi_sar",
}

const (
	// readROVPDValuesCmdGlob reads the value of VPD key from RO_VPD partition by name.
	readROVPDValuesCmd = "vpd -i RO_VPD -g %s"
	writeVPDValuesCmd  = "vpd -s %s=%s"
)

type DSMVPD struct {
	Rdc  []int32 `json:"r0"`
	Temp []int32 `json:"temp"`
}

// isROVPDDSMCalibRequired confirms that this device is required to have sku_number in RO_VPD.
func isROVPDDSMCalibRequired(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()

	speakerAmp, err := r(ctx, time.Minute, "cros_config /audio/main/ speaker-amp")
	if err != nil {
		return errors.Annotate(err, "is RO_VPD dsm_calib_r0 required").Err()
	}
	// If speaker-amp is unspecified, dsm_calib_r0 is not required.
	if speakerAmp == "" {
		return errors.Reason("dsm_calib_r0 is not required in RO_VPD.").Err()
	}

	soundCardInitConf, err := r(ctx, time.Minute, "cros_config /audio/main/ sound-card-init-conf")
	if err != nil {
		return errors.Annotate(err, "is RO_VPD dsm_calib_r0 required").Err()
	}
	// If sound-card-init-conf is unspecified, dsm_calib_r0 is not required.
	if soundCardInitConf == "" {
		return errors.Reason("dsm_calib_r0 is not required in RO_VPD.").Err()
	}
	return nil
}

func parseSoundCardID(dump string) (string, error) {
	re := regexp.MustCompile(`card 0: ([a-z0-9]+) `)
	m := re.FindStringSubmatch(dump)

	if len(m) != 2 {
		return "", errors.New("no sound card")
	}
	return m[1], nil
}

// verifyROVPDDSMCalib confirms that the key 'dsm_calib_r0_0' is present in RO_VPD.
func verifyROVPDDSMCalib(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	if _, err := r(ctx, time.Minute, "vpd -i RO_VPD -g dsm_calib_r0_0"); err != nil {
		return errors.Annotate(err, "verify dsm_calib_r0 in RO_VPD").Err()
	}

	log.Infof(ctx, "dsm_calib_r0_0 is present in RO_VPD")
	return nil
}

// setFakeROVPDDSMCalib sets a fake dsm_calib_r{}, dsm_calib_temp{} in RO_VPD.
func setFakeROVPDDSMCalib(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	speakerAmp, err := r(ctx, time.Minute, "cros_config /audio/main/ speaker-amp")
	if err != nil {
		return errors.Annotate(err, "cros_config /audio/main/ speaker-amp").Err()
	}

	soundCardInitConf, err := r(ctx, time.Minute, "cros_config /audio/main/ sound-card-init-conf")
	if err != nil {
		return errors.Annotate(err, "cros_config /audio/main/ sound-card-init-conf").Err()
	}

	dump, err := r(ctx, time.Minute, "aplay -l")
	if err != nil {
		return errors.Annotate(err, "aplay -l").Err()
	}
	soundCardID, err := parseSoundCardID(string(dump))
	if err != nil {
		return errors.Annotate(err, "Failed to parse sound card name").Err()
	}

	soundCardInitCmd := fmt.Sprintf("/usr/bin/sound_card_init fake_vpd --json --id %s --amp %s --conf %s", soundCardID, speakerAmp, soundCardInitConf)

	fakeDSMVPDJson, err := r(ctx, time.Minute, soundCardInitCmd)
	if err != nil {
		return errors.Annotate(err, "cannot get fake DSM vpd").Err()
	}

	var fakeDSMVPD DSMVPD
	if err = json.Unmarshal([]byte(fakeDSMVPDJson), &fakeDSMVPD); err != nil {
		return errors.Annotate(err, "cannot parse fake DSM vpd json: "+string(fakeDSMVPDJson)).Err()
	}

	for ch := 0; ch < len(fakeDSMVPD.Rdc); ch++ {
		cmd := fmt.Sprintf("vpd -s dsm_calib_r0_%d=%d", ch, fakeDSMVPD.Rdc[ch])
		if _, err = r(ctx, time.Minute, cmd); err != nil {
			return errors.Annotate(err, "cannot set fake dsm_calib_r0").Err()
		}
		cmd = fmt.Sprintf("vpd -s dsm_calib_temp_%d=%d", ch, fakeDSMVPD.Temp[ch])
		if _, err := r(ctx, time.Minute, cmd); err != nil {
			return errors.Annotate(err, "cannot set fake dsm_calib_temp").Err()
		}
	}

	log.Infof(ctx, "set fake dsm_calib_r0 successfully")
	return nil
}

// isROVPDSkuNumberRequired confirms that this device is required to have sku_number in RO_VPD.
func isROVPDSkuNumberRequired(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	hasSkuNumber, err := r(ctx, time.Minute, "cros_config /cros-healthd/cached-vpd has-sku-number")
	if err != nil {
		return errors.Annotate(err, "is RO_VPD sku_number required").Err()
	}
	// If has-sku-number is not true (e.g. unspecified or false), sku_number is not required.
	if hasSkuNumber != "true" {
		return errors.Reason("sku_number is not required in RO_VPD.").Err()
	}
	return nil
}

// verifyROVPDSkuNumber confirms that the key 'sku_number' is present in RO_VPD.
func verifyROVPDSkuNumber(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	if _, err := r(ctx, time.Minute, "vpd -i RO_VPD -g sku_number"); err != nil {
		return errors.Annotate(err, "verify sku_number in RO_VPD").Err()
	}
	log.Infof(ctx, "sku_number is present in RO_VPD")
	return nil
}

// setFakeROVPDSkuNumber sets a fake sku_number in RO_VPD.
//
// @params: actionArgs should be in the format:
// ["sku_number:FAKE-SKU"]
func setFakeROVPDSkuNumber(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	if !argsMap.Has("sku_number") {
		return errors.Reason("set fake sku_number: fake value is not specified.").Err()
	}
	skuNumber := argsMap.AsString(ctx, "sku_number", "")

	r := info.DefaultRunner()
	cmd := fmt.Sprintf(writeVPDValuesCmd, "sku_number", skuNumber)
	if _, err := r(ctx, time.Minute, cmd); err != nil {
		return errors.Annotate(err, "cannot set fake sku_number").Err()
	}
	log.Infof(ctx, "set fake sku_number successfully")
	return nil
}

// updateROVPDToInv reads RO_VPD values from the resource listed in roVPDKeys into the inventory.
func updateROVPDToInv(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	if info.GetChromeos().GetRoVpdMap() == nil {
		info.GetChromeos().RoVpdMap = make(map[string]string)
	}
	for _, key := range roVPDKeys {
		cmd := fmt.Sprintf(readROVPDValuesCmd, key)
		if value, err := r(ctx, time.Minute, cmd); err == nil {
			info.GetChromeos().RoVpdMap[key] = value
		}
	}
	log.Infof(ctx, "recorded RO_VPD values successfully")
	return nil
}

// matchROVPDToInv matches RO_VPD values from resource to inventory.
func matchROVPDToInv(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	for k, v := range info.GetChromeos().GetRoVpdMap() {
		cmd := fmt.Sprintf(readROVPDValuesCmd, k)
		value, err := r(ctx, time.Minute, cmd)
		if err != nil {
			return errors.Annotate(err, "cannot read RO_VPD key").Err()
		}
		if value != v {
			return errors.Annotate(err, "RO_VPD had a bad value").Err()
		}
	}
	log.Infof(ctx, "RO_VPD values are correct")
	return nil
}

// setROVPD sets RO_VPD values from inventory to resource.
func setROVPD(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	for k, v := range info.GetChromeos().GetRoVpdMap() {
		cmd := fmt.Sprintf(writeVPDValuesCmd, k, v)
		if _, err := r(ctx, time.Minute, cmd); err != nil {
			return errors.Annotate(err, "cannot set RO_VPD key").Err()
		}
	}
	log.Infof(ctx, "set RO_VPD values successfully")
	return nil
}

func init() {
	execs.Register("cros_is_ro_vpd_sku_number_required", isROVPDSkuNumberRequired)
	execs.Register("cros_verify_ro_vpd_sku_number", verifyROVPDSkuNumber)
	execs.Register("cros_set_fake_ro_vpd_sku_number", setFakeROVPDSkuNumber)
	execs.Register("cros_is_ro_vpd_dsm_calib_required", isROVPDDSMCalibRequired)
	execs.Register("cros_verify_ro_vpd_dsm_calib", verifyROVPDDSMCalib)
	execs.Register("cros_set_fake_ro_vpd_dsm_calib", setFakeROVPDDSMCalib)
	execs.Register("cros_update_ro_vpd_inventory", updateROVPDToInv)
	execs.Register("cros_match_ro_vpd_inventory", matchROVPDToInv)
	execs.Register("cros_set_ro_vpd", setROVPD)
}
