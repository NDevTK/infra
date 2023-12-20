// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cros/vpd"
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

// isROVPDDSMCalibRequiredExec confirms that this device is required to have sku_number in RO_VPD.
func isROVPDDSMCalibRequiredExec(ctx context.Context, info *execs.ExecInfo) error {
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

// verifyROVPDDSMCalibExec confirms that the key 'dsm_calib_r0_0' is present in RO_VPD.
func verifyROVPDDSMCalibExec(ctx context.Context, info *execs.ExecInfo) error {
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
		cmd := fmt.Sprintf("vpd -i RO_VPD -g dsm_calib_r0_%d", ch)
		if _, err := r(ctx, time.Minute, cmd); err != nil {
			return errors.Annotate(err, cmd).Err()
		}
		cmd = fmt.Sprintf("vpd -i RO_VPD -g dsm_calib_temp_%d", ch)
		if _, err := r(ctx, time.Minute, cmd); err != nil {
			return errors.Annotate(err, cmd).Err()
		}
	}
	return nil
}

// setFakeROVPDDSMCalibExec sets a fake dsm_calib_r{}, dsm_calib_temp{} in RO_VPD.
func setFakeROVPDDSMCalibExec(ctx context.Context, info *execs.ExecInfo) error {
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

// isROVPDSkuNumberRequiredExec confirms that this device is required to have sku_number in RO_VPD.
func isROVPDSkuNumberRequiredExec(ctx context.Context, info *execs.ExecInfo) error {
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

// verifyROVPDSkuNumberExec confirms that the key 'sku_number' is present in RO_VPD.
func verifyROVPDSkuNumberExec(ctx context.Context, info *execs.ExecInfo) error {
	r := info.DefaultRunner()
	if _, err := r(ctx, time.Minute, "vpd -i RO_VPD -g sku_number"); err != nil {
		return errors.Annotate(err, "verify sku_number in RO_VPD").Err()
	}
	log.Infof(ctx, "sku_number is present in RO_VPD")
	return nil
}

// setFakeROVPDSkuNumberExec sets a fake sku_number in RO_VPD.
//
// @params: actionArgs should be in the format:
// ["sku_number:FAKE-SKU"]
func setFakeROVPDSkuNumberExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	skuNumber := argsMap.AsString(ctx, "sku_number", "")
	if skuNumber == "" {
		return errors.Reason("set fake sku_number: fake value is not specified.").Err()
	}
	r := info.DefaultRunner()
	cmd := fmt.Sprintf(writeVPDValuesCmd, "sku_number", skuNumber)
	if _, err := r(ctx, time.Minute, cmd); err != nil {
		return errors.Annotate(err, "cannot set fake sku_number").Err()
	}
	log.Infof(ctx, "set fake sku_number successfully")
	return nil
}

// updateROVPDToInvExec reads RO_VPD values from the resource listed in roVPDKeys into the inventory.
func updateROVPDToInvExec(ctx context.Context, info *execs.ExecInfo) error {
	ha := info.DefaultHostAccess()
	if info.GetChromeos().GetRoVpdMap() == nil {
		info.GetChromeos().RoVpdMap = make(map[string]string)
	}
	for _, key := range roVPDKeys {
		if value, err := vpd.ReadRO(ctx, ha, time.Minute, key); err == nil {
			info.GetChromeos().RoVpdMap[key] = value
		}
	}
	log.Infof(ctx, "recorded RO_VPD values successfully")
	return nil
}

// matchROVPDToInvExec matches RO_VPD values from resource to inventory.
func matchROVPDToInvExec(ctx context.Context, info *execs.ExecInfo) error {
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

// setROVPDExec sets RO_VPD values from inventory to resource.
func setROVPDExec(ctx context.Context, info *execs.ExecInfo) error {
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

// setVPDValueExec sets VPD value by provided key.
func setVPDValueExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	key := strings.TrimSpace(argsMap.AsString(ctx, "key", ""))
	value := strings.TrimSpace(argsMap.AsString(ctx, "value", ""))
	if key == "" {
		return errors.Reason("set VPD value: key is empty").Err()
	} else if value == "" {
		return errors.Reason("set VPD value by key %q: value is empty", key).Err()
	}
	err := vpd.Set(ctx, info.DefaultHostAccess(), info.GetExecTimeout(), key, value)
	return errors.Annotate(err, "set VPD value %q:%q", key, value).Err()
}

// checkVPDValueExec checks VPD to read and present of value by provided key.
func checkVPDValueExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	key := strings.TrimSpace(argsMap.AsString(ctx, "key", ""))
	if key == "" {
		return errors.Reason("check VPD value: key is empty").Err()
	}
	value, err := vpd.Read(ctx, info.DefaultHostAccess(), info.GetExecTimeout(), key)
	if err != nil {
		return errors.Annotate(err, "check VPD value: by key:%q failed", key).Err()
	}
	if value == "" {
		return errors.Reason("check VPD value by key:%q: has empty value", key).Err()
	}
	log.Debugf(ctx, "VPD has %q:%q", key, value)
	return nil
}

// setRandomStableDeviceSecretExec generates a random 32 byte string and stores it in vpd under
// "stable_device_secret_DO_NOT_SHARE".
func setRandomStableDeviceSecretExec(ctx context.Context, info *execs.ExecInfo) error {
	key := "stable_device_secret_DO_NOT_SHARE"

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return errors.Annotate(err, "failed to generate random string").Err()
	}
	value := hex.EncodeToString(bytes)

	err := vpd.Set(ctx, info.DefaultHostAccess(), info.GetExecTimeout(), key, value)
	return errors.Annotate(err, "set VPD value %q:%q", key, value).Err()
}

func init() {
	execs.Register("cros_is_ro_vpd_sku_number_required", isROVPDSkuNumberRequiredExec)
	execs.Register("cros_verify_ro_vpd_sku_number", verifyROVPDSkuNumberExec)
	execs.Register("cros_set_fake_ro_vpd_sku_number", setFakeROVPDSkuNumberExec)
	execs.Register("cros_is_ro_vpd_dsm_calib_required", isROVPDDSMCalibRequiredExec)
	execs.Register("cros_verify_ro_vpd_dsm_calib", verifyROVPDDSMCalibExec)
	execs.Register("cros_set_fake_ro_vpd_dsm_calib", setFakeROVPDDSMCalibExec)
	execs.Register("cros_update_ro_vpd_inventory", updateROVPDToInvExec)
	execs.Register("cros_match_ro_vpd_inventory", matchROVPDToInvExec)
	execs.Register("cros_set_ro_vpd", setROVPDExec)
	execs.Register("cros_set_vpd_value", setVPDValueExec)
	execs.Register("cros_check_vpd_value", checkVPDValueExec)
	execs.Register("cros_set_random_ro_vpd_stable_device_secret", setRandomStableDeviceSecretExec)
}
