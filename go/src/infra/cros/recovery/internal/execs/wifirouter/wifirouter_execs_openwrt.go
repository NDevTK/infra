// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wifirouter

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"infra/cros/recovery/internal/execs"
)

func fetchOpenWrtBuildInfoExec(ctx context.Context, info *execs.ExecInfo) error {
	c, err := activeHostOpenWrtRouterController(ctx, info)
	if err != nil {
		return errors.Annotate(err, "fetch openwrt build info").Err()
	}
	if err := c.FetchDeviceBuildInfo(ctx); err != nil {
		return errors.Annotate(err, "failed to fetch device build info").Err()
	}
	return nil
}

func fetchOpenWrtConfigExec(ctx context.Context, info *execs.ExecInfo) error {
	c, err := activeHostOpenWrtRouterController(ctx, info)
	if err != nil {
		return errors.Annotate(err, "fetch openwrt config").Err()
	}
	if err := c.FetchGlobalImageConfig(ctx); err != nil {
		return errors.Annotate(err, "failed to fetch device build info").Err()
	}
	return nil
}

func init() {
	execs.Register("wifi_router_openwrt_fetch_build_info", fetchOpenWrtBuildInfoExec)
	execs.Register("wifi_router_openwrt_fetch_config", fetchOpenWrtConfigExec)
}
