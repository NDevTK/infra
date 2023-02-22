// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// notInPoolExec verifies that DUT is not used in special pools.
// List of pools should be listed as part of ActionArgs.
func notInPoolExec(ctx context.Context, info *execs.ExecInfo) error {
	if len(info.GetExecArgs()) == 0 {
		log.Debugf(ctx, "Not in pool: no pools passed as arguments.")
		return nil
	}
	poolMap := getDUTPoolMap(ctx, info.GetDut())
	for _, pool := range info.GetExecArgs() {
		pool = strings.TrimSpace(pool)
		if poolMap[pool] {
			return errors.Reason("not in pool: dut is in pool %q", pool).Err()
		}
		log.Debugf(ctx, "Not in pools: %q pool is not matched.", pool)
	}
	log.Debugf(ctx, "Not in pools: no intersection found.")
	return nil
}

// isInPoolExec verifies that DUT is used in special pools.
// List of pools should be listed as part of ActionArgs.
func isInPoolExec(ctx context.Context, info *execs.ExecInfo) error {
	if len(info.GetExecArgs()) == 0 {
		log.Debugf(ctx, "Is in pool: no pools passed as arguments.")
		return nil
	}
	poolMap := getDUTPoolMap(ctx, info.GetDut())
	for _, pool := range info.GetExecArgs() {
		pool = strings.TrimSpace(pool)
		if poolMap[pool] {
			log.Debugf(ctx, "Is in pools: %q pool listed at the DUT.", pool)
			return nil
		}
		log.Debugf(ctx, "Is in pools: %q pool is not matched.", pool)
	}
	return errors.Reason("is in pool: not match found").Err()
}

// getDUTPoolMap extract map of pools listed under DUT.
func getDUTPoolMap(ctx context.Context, d *tlw.Dut) map[string]bool {
	poolMap := make(map[string]bool)
	pools := d.ExtraAttributes[tlw.ExtraAttributePools]
	if len(pools) == 0 {
		log.Debugf(ctx, "device does not have any pools.")
		return poolMap
	}
	for _, pool := range pools {
		poolMap[pool] = true
	}
	return poolMap
}

// notBrowserLegacyDUTExec verifies that if the DUT is a legacy DUT in browser lab.
func notBrowserLegacyDUTExec(ctx context.Context, info *execs.ExecInfo) error {
	// Only legacy DUT migrated from browser lab has assetTag to contain browser prefix.
	// DUTs used for browser testing but purchased by CrOS Warehouse has real asset tag without prefix.
	assetTag := info.GetDut().Id
	if strings.HasPrefix(assetTag, "chrome-") || strings.HasPrefix(assetTag, "chromium-") {
		return errors.Reason("check if not browser legacy DUT: %s is a legacy browser DUT", assetTag).Err()
	}
	log.Debugf(ctx, "check if not browser legacy DUT: %s is not legacy browser DUT.", assetTag)
	return nil
}

func init() {
	execs.Register("dut_not_in_pool", notInPoolExec)
	execs.Register("dut_is_in_pool", isInPoolExec)
	execs.Register("dut_is_not_browser_legacy_duts", notBrowserLegacyDUTExec)
}
