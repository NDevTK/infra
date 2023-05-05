// Copyright 2021 The Chromium Authors
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
	actionArgs := info.GetActionArgs(ctx)
	allowedOverride := actionArgs.AsBool(ctx, "allowed_override", false)
	invHWID := info.GetChromeos().GetHwid()
	if invHWID != "" {
		if allowedOverride {
			log.Debugf(ctx, "Inventory have HWID %q and override is allowed.", invHWID)
		} else {
			log.Debugf(ctx, "Inventory have HWID %q and it is not allowed to override it.", invHWID)
			return nil
		}
	}
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

// TODO(b/280635852): Remove when stable versions upgraded.
// Special script to read and update GBB removed in favore of futility.
const legacyGBBReadFilename = "/usr/share/vboot/bin/get_gbb_flags.sh"

// updateHWIDFromInvExec updates HWID from inventory to host.
//
// HWID can be checkd on the DUT bu futility.
// Crossystem represent cached data and will wait till reboot to update the value.
func updateHWIDFromInvExec(ctx context.Context, info *execs.ExecInfo) error {
	run := info.DefaultRunner()
	originalHwid := info.GetChromeos().GetHwid()
	if _, gbbScriptErr := run(ctx, 15*time.Second, fmt.Sprintf("test -f %s", legacyGBBReadFilename)); gbbScriptErr == nil {
		// TODO(b/280635852): Remove when stable versions upgraded.
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
	} else {
		// Update HWID to the AP firmware.
		log.Debugf(ctx, "Update HWID from host: update HWID %q in AP firmware.", originalHwid)
		if out, err := run(ctx, 3*time.Minute, "futility", "gbb", "--flash --set", "--hwid", fmt.Sprintf("%q", originalHwid)); err != nil {
			return errors.Annotate(err, "update HWID from host").Err()
		} else {
			log.Debugf(ctx, "Update HWID from host: updated HWID in AP file (output): %s", out)
		}
	}
	return nil
}

func init() {
	execs.Register("cros_update_hwid_to_inventory", updateHWIDToInvExec)
	execs.Register("cros_match_hwid_to_inventory", matchHWIDToInvExec)
	execs.Register("cros_update_hwid_from_inventory_to_host", updateHWIDFromInvExec)
}
