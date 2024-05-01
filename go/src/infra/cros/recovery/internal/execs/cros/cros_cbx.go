// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Temporary check to see if device chromebook X state matches UFS. b/282236972
package cros

import (
	"context"
	"strings"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger/metrics"
)

const (
	featureExplorerCbxCmd = "feature_explorer --feature_level"
)

func verifyDeviceCbxMatchesUFS(ctx context.Context, info *execs.ExecInfo) error {
	ufsCbx := info.GetDut().GetChromeos().GetCbx()
	r := info.DefaultRunner()
	cbxFeatureOutput, err := r(ctx, time.Minute, featureExplorerCbxCmd)
	if err != nil {
		return errors.Annotate(err, "Error when reading cbx feature level").Err()
	}

	deviceCbx := (strings.TrimSpace(cbxFeatureOutput) == "1")
	if deviceCbx != ufsCbx {
		log.Debugf(ctx, "CBX mismatch: DUT %v UFS %v", cbxFeatureOutput, ufsCbx)
		info.AddObservation(metrics.NewStringObservation("cbx_mismatch_sku", info.GetDut().GetChromeos().GetDeviceSku()))
		return errors.Reason("UFS and device CBX feature level are different").Err()
	}
	return nil
}

// TODO(b/293492109): Remove hard-coded values once we can pull the data from UFS
// b/284334769 tracks the UFS updates
var hbDevices = map[string]map[string]bool{
	"brya":   {"omnigul": true},
	"skyrim": {"markarth": true},
}

func deviceIsHB(ctx context.Context, info *execs.ExecInfo) error {
	board := info.GetDut().GetChromeos().GetBoard()
	validModels, ok := hbDevices[board]
	if !ok {
		return errors.Reason("Device board is not one of known HB devices").Err()
	}
	model := info.GetDut().GetChromeos().GetModel()
	if !validModels[model] {
		return errors.Reason("Device board and model are not one of known HB devices").Err()
	}
	return nil
}

func init() {
	execs.Register("cros_verify_cbx_matches_ufs", verifyDeviceCbxMatchesUFS)
	execs.Register("cros_check_cbx_device_is_hb", deviceIsHB)
}
