// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wifirouter

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/execs"
)

func fetchOpenWrtBuildInfoExec(ctx context.Context, info *execs.ExecInfo) error {
	c, err := activeHostOpenWrtRouterController(ctx, info)
	if err != nil {
		return err
	}
	if err := c.FetchDeviceBuildInfo(ctx); err != nil {
		return errors.Annotate(err, "failed to fetch device build info").Err()
	}
	return nil
}

func fetchOpenWrtConfigExec(ctx context.Context, info *execs.ExecInfo) error {
	c, err := activeHostOpenWrtRouterController(ctx, info)
	if err != nil {
		return err
	}
	if err := c.FetchGlobalImageConfig(ctx); err != nil {
		return errors.Annotate(err, "failed to fetch device build info").Err()
	}
	return nil
}

func identifyExpectedOpenWrtImageExec(ctx context.Context, info *execs.ExecInfo) error {
	c, err := activeHostOpenWrtRouterController(ctx, info)
	if err != nil {
		return err
	}
	dut := info.GetDut()
	if dut == nil {
		return errors.Reason("dut is not present").Err()
	}
	provisionedInfo := info.GetDut().ProvisionedInfo
	if provisionedInfo == nil {
		return errors.Reason("Dut.ProvisionedInfo is not present").Err()
	}
	if provisionedInfo.CrosVersion == "" {
		return errors.Reason("Dut.ProvisionedInfo.CrosVersion is empty").Err()
	}
	crosReleaseVersion, err := cros.ParseReleaseVersionFromBuilderPath(provisionedInfo.CrosVersion)
	if err != nil {
		return errors.Annotate(err, "failed to parse release version from Dut.ProvisionedInfo.CrosVersion %q", provisionedInfo.CrosVersion).Err()
	}
	if err := c.IdentifyExpectedImage(ctx, dut.Name, crosReleaseVersion); err != nil {
		return errors.Annotate(err, "failed to identify expected OpenWrt OS image for device").Err()
	}
	return nil
}

func assertOpenWrtRouterHasExpectedImageExec(ctx context.Context, info *execs.ExecInfo) error {
	c, err := activeHostOpenWrtRouterController(ctx, info)
	if err != nil {
		return err
	}
	return c.AssertHasExpectedImage()
}

func updateOpenWrtRouterToExpectedImageExec(ctx context.Context, info *execs.ExecInfo) error {
	c, err := activeHostOpenWrtRouterController(ctx, info)
	if err != nil {
		return err
	}
	return c.UpdateToExpectedImage(ctx)
}

func init() {
	execs.Register("wifi_router_openwrt_fetch_build_info", fetchOpenWrtBuildInfoExec)
	execs.Register("wifi_router_openwrt_fetch_config", fetchOpenWrtConfigExec)
	execs.Register("wifi_router_openwrt_identify_expected_image", identifyExpectedOpenWrtImageExec)
	execs.Register("wifi_router_openwrt_has_expected_image", assertOpenWrtRouterHasExpectedImageExec)
	execs.Register("wifi_router_openwrt_update_to_expected_image", updateOpenWrtRouterToExpectedImageExec)
}
