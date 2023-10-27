// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wifirouter

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/execs/wifirouter/controller"
	"infra/cros/recovery/tlw"
)

// activeHost finds active host related to the executed plan.
func activeHost(info *execs.ExecInfo) (*tlw.WifiRouterHost, error) {
	resource := info.GetActiveResource()
	chromeos := info.GetChromeos()
	if chromeos == nil {
		return nil, errors.Reason("chromeos is not present").Err()
	}
	for _, router := range chromeos.GetWifiRouters() {
		if router.GetName() == resource {
			return router, nil
		}
	}
	return nil, errors.Reason("router: router not found").Err()
}

func activeHostRouterController(ctx context.Context, info *execs.ExecInfo) (controller.RouterController, error) {
	if info.GetDut() == nil {
		return nil, errors.Reason("dut is nil").Err()
	}
	wifiRouterHost, err := activeHost(info)
	if err != nil {
		return nil, err
	}
	return controller.NewRouterDeviceController(ctx, info.GetAccess(), info.GetAccess(), info.GetActiveResource(), info.GetDut().Name, wifiRouterHost)
}

func activeHostAsusWrtRouterController(ctx context.Context, info *execs.ExecInfo) (*controller.AsusWrtRouterController, error) {
	genericController, err := activeHostRouterController(ctx, info)
	if err != nil {
		return nil, err
	}
	asusWrtController, ok := genericController.(*controller.AsusWrtRouterController)
	if !ok {
		return nil, errors.Reason("active host is not an AsusWrt test router").Err()
	}
	return asusWrtController, nil
}

func activeHostOpenWrtRouterController(ctx context.Context, info *execs.ExecInfo) (*controller.OpenWrtRouterController, error) {
	genericController, err := activeHostRouterController(ctx, info)
	if err != nil {
		return nil, err
	}
	openWrtController, ok := genericController.(*controller.OpenWrtRouterController)
	if !ok {
		return nil, errors.Reason("active host is not an OpenWrt test router").Err()
	}
	return openWrtController, nil
}
