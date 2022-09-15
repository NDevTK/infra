// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cros

import (
	"context"
	"fmt"
	"time"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
)

const (
	readHWIDCommand = "crossystem hwid"
)

// updateHWIDToInvExec read HWID from the resource and update DUT info.
func updateHWIDToInvExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	hwid, err := run(ctx, time.Minute, readHWIDCommand)
	if err != nil {
		return errors.Annotate(err, "update HWID in DUT-info").Err()
	}
	if hwid == "" {
		return errors.Reason("update HWID in DUT-info: is empty").Err()
	}
	info.GetChromeos().Hwid = hwid
	log.Debugf(ctx, "Update HWID: %q", info.GetChromeos().GetHwid())
	return nil
}

// matchHWIDToInvExec matches HWID from the resource to value in the Inventory.
func matchHWIDToInvExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	actualHWID, err := run(ctx, time.Minute, readHWIDCommand)
	if err != nil {
		return errors.Annotate(err, "match HWID to inventory").Err()
	}
	expectedHWID := info.GetChromeos().GetHwid()
	if actualHWID != expectedHWID {
		return errors.Reason("match HWID to inventory: failed, expected: %q, but got %q", expectedHWID, actualHWID).Err()
	}
	return nil
}

// updateHWIDFromInvExec updates HWID from inventory to host.
func updateHWIDFromInvExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	originalHwid := info.GetChromeos().GetHwid()
	tempFileName := fmt.Sprintf("/tmp/bios_%s.bin", info.GetActiveResource())
	log.Debugf(ctx, "Update HWID from host: Try to read AP to file: %q", tempFileName)
	if out, err := run(ctx, time.Minute, "flashrom", "-p", "host", "-r", "-i", fmt.Sprintf("GBB:%s", tempFileName)); err != nil {
		return errors.Annotate(err, "update HWID from host").Err()
	} else {
		log.Debugf(ctx, "Update HWID from host: read AP (output): %s", out)
	}
	// Update HWID to the file inside temp file.
	log.Debugf(ctx, "Update HWID from host: update HWID %q in AP to file %q", originalHwid, tempFileName)
	if out, err := run(ctx, time.Minute, "futility", "gbb", "--set", "--hwid", fmt.Sprintf("%q", originalHwid), tempFileName); err != nil {
		return errors.Annotate(err, "update HWID from host").Err()
	} else {
		log.Debugf(ctx, "Update HWID from host: updated HWID in AP file (output): %s", out)
	}
	// Write updated AP file back to host.
	log.Debugf(ctx, "Update HWID from host: flash AP file %q back", tempFileName)
	if out, err := run(ctx, time.Minute, "flashrom", "-p", "host", "-w", "-i", fmt.Sprintf("GBB:%s", tempFileName)); err != nil {
		return errors.Annotate(err, "update HWID from host").Err()
	} else {
		log.Debugf(ctx, "Update HWID from host: AP file flashed to the host (output): %s", out)
	}
	return nil
}

func init() {
	execs.Register("cros_update_hwid_to_inventory", updateHWIDToInvExec)
	execs.Register("cros_match_hwid_to_inventory", matchHWIDToInvExec)
	execs.Register("cros_update_hwid_from_inventory_to_host", updateHWIDFromInvExec)
}
