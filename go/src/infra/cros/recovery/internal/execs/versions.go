// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package execs

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components"
	"infra/cros/recovery/tlw"
)

// Local implementation of Versioner interface.
type versioner struct {
	a tlw.Access
}

// Versioner returns versioner to read any version.
func (ei *ExecInfo) Versioner() components.Versioner {
	return &versioner{
		a: ei.runArgs.Access,
	}
}

// Cros return version info for request Chrome OS device.
// Deprecated. please use GetVersion.
func (v *versioner) Cros(ctx context.Context, resource string) (*components.VersionInfo, error) {
	res, err := v.GetVersion(ctx, components.VersionDeviceCros, resource, "", "")
	return res, errors.Annotate(err, "cros version").Err()
}

// GetVersion returns version info for the requested device.
func (v *versioner) GetVersion(ctx context.Context, deviceType components.VersionDeviceType, resource, board, model string) (*components.VersionInfo, error) {
	req := &tlw.VersionRequest{
		Resource: resource,
		Board:    board,
		Model:    model,
	}
	switch deviceType {
	case components.VersionDeviceCros:
		req.Type = tlw.VersionRequest_CROS
	case components.VersionDeviceWifiRouter:
		req.Type = tlw.VersionRequest_WIFI_ROUTER
	default:
		return nil, errors.Reason("get version: device type %q is not supported", deviceType).Err()
	}
	r, err := v.a.Version(ctx, req)
	if err != nil {
		return nil, errors.Annotate(err, "get version").Err()
	}
	if len(r.GetValue()) < 1 {
		return nil, errors.Reason("get version: no version received").Err()
	}
	res := &components.VersionInfo{}
	if v, ok := r.GetValue()["os_image"]; ok {
		res.OSImage = v
	}
	if v, ok := r.GetValue()["fw_image"]; ok {
		res.FwImage = v
	}
	if v, ok := r.GetValue()["fw_version"]; ok {
		res.FwVersion = v
	}
	return res, nil
}
