// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cros/firmware"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/logger/metrics"
	"infra/cros/recovery/tlw"
)

// collectFirmwareTargetExec read fw target from DUT to inventory.
//
// If inventory has data then data will not be overwritten.
func collectFirmwareTargetExec(ctx context.Context, info *execs.ExecInfo) error {
	argsMap := info.GetActionArgs(ctx)
	isForceOverride := argsMap.AsBool(ctx, "force_override", false)
	if info.GetChromeos() == nil {
		return errors.Reason("collect firmware target: only for chromeos devices").Err()
	}
	fi := info.GetChromeos().GetFirmwareInfo()
	if fi == nil {
		fi = &tlw.FirmwareInfo{}
		info.GetChromeos().FirmwareInfo = fi
	}
	log.Debugf(ctx, "Fw targets before update ec:%q, ap:%q", fi.GetEcTarget(), fi.GetApTarget())
	run := info.NewRunner(info.GetDut().Name)
	fwTarget, err := firmware.GetFirmwareManifestKeyFromDUT(ctx, run, log.Get(ctx))
	if err != nil {
		return errors.Annotate(err, "collect firmware target").Err()
	}
	fwTarget = strings.TrimSpace(fwTarget)
	metrics.DefaultActionAddObservations(ctx, metrics.NewStringObservation("collected_fw_target", fwTarget))
	if isForceOverride || fi.GetApTarget() == "" {
		fi.ApTarget = fwTarget
		fi.EcTarget = fwTarget
	}
	log.Debugf(ctx, "Fw targets after update ec:%q, ap:%q", fi.GetEcTarget(), fi.GetApTarget())
	return nil
}

func init() {
	execs.Register("cros_collect_firmware_target", collectFirmwareTargetExec)
}
