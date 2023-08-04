// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/tlw"
)

const (
	// deviceInfoFilePath is the path to the device info file on OpenWrt routers.
	deviceInfoFilePath = "/etc/device_info"

	// deviceInfoMatchIfOpenWrt is the regex that will match the contents of the
	// device info file on the host if it is an OpenWrt device.
	deviceInfoMatchIfOpenWrt = "(?m)^DEVICE_MANUFACTURER='OpenWrt'$"

	// buildInfoFilePath is the path to the build info file on OpenWrt routers.
	buildInfoFilePath = "/etc/cros/cros_openwrt_image_build_info.json"
)

// hostIsOpenWrtRouter checks if the remote host is an OpenWrt router.
func hostIsOpenWrtRouter(ctx context.Context, sshRunner ssh.Runner) (bool, error) {
	matches, err := RemoteFileContentsMatch(ctx, sshRunner, deviceInfoFilePath, deviceInfoMatchIfOpenWrt)
	if err != nil {
		return false, errors.Annotate(err, "failed to check if remote file %q contents match %q", deviceInfoFilePath, deviceInfoMatchIfOpenWrt).Err()
	}
	if !matches {
		return false, nil
	}
	hasBuildInfoFile, err := ssh.TestFileExists(ctx, sshRunner, buildInfoFilePath)
	if err != nil {
		return false, err
	}
	return hasBuildInfoFile, nil
}

// OpenWrtRouterController is the RouterController implementation for
// OpenWrt router devices.
//
// This is intended to support any router device with a custom ChromeOS OpenWrt
// OS test image installed on it. These custom images, built with the
// cros_openwrt_image_builder CLI tool, include a build info file that is read
// for image and device identification.
type OpenWrtRouterController struct {
	sshRunner      ssh.Runner
	wifiRouterHost *tlw.WifiRouterHost
	state          *tlw.OpenWrtRouterControllerState
	cacheAccess    CacheAccess
	resource       string
}

func newOpenWrtRouterController(sshRunner ssh.Runner, wifiRouterHost *tlw.WifiRouterHost, state *tlw.OpenWrtRouterControllerState, cacheAccess CacheAccess, resource string) *OpenWrtRouterController {
	return &OpenWrtRouterController{
		wifiRouterHost: wifiRouterHost,
		sshRunner:      sshRunner,
		state:          state,
		cacheAccess:    cacheAccess,
		resource:       resource,
	}
}

// WifiRouterHost returns the corresponding tlw.WifiRouterHost instance for
// this router. Changes to this instance are persisted across execs.
func (c *OpenWrtRouterController) WifiRouterHost() *tlw.WifiRouterHost {
	return c.wifiRouterHost
}

// DeviceType returns the labapi.WifiRouterDeviceType of the router.
func (c *OpenWrtRouterController) DeviceType() labapi.WifiRouterDeviceType {
	return labapi.WifiRouterDeviceType_WIFI_ROUTER_DEVICE_TYPE_OPENWRT
}

// Model returns a unique name for the router model.
//
// For OpenWrt routers, this is a combination of the DeviceType and the device
// name retrieved from the router. The device name is retrieved from the build
// info file placed on the router by the ChromeOS OpenWrt image builder
//
// OpenWrt device names are a combination of
// the router manufacturer and model, and in most cases the names we use are
// the same (sanitized) names set by the OpenWrt community.
func (c *OpenWrtRouterController) Model() (string, error) {
	if c.state.GetDeviceBuildInfo() == nil {
		return "", errors.Reason("state.DeviceBuildInfo must not be nil").Err()
	}
	return buildModelName(c.DeviceType(), c.state.DeviceBuildInfo.StandardBuildConfig.DeviceName), nil
}

// Features returns the router features that this router supports.
//
// For OpenWrt routers, features are retrieved from the build info file placed
// on the router by the custom ChromeOS OpenWrt image builder.
func (c *OpenWrtRouterController) Features() ([]labapi.WifiRouterFeature, error) {
	if c.state.GetDeviceBuildInfo() == nil {
		return nil, errors.Reason("state.DeviceBuildInfo must not be nil").Err()
	}
	return c.state.DeviceBuildInfo.RouterFeatures, nil
}

// FetchDeviceBuildInfo retrieves the build info from the router and stores
// it in the controller state.
func (c *OpenWrtRouterController) FetchDeviceBuildInfo(ctx context.Context) error {
	if c.state == nil {
		return errors.Reason("state must not be nil").Err()
	}

	// Fetch and unmarshal build info from host.
	buildInfoFileContents, err := ssh.CatFile(ctx, c.sshRunner, buildInfoFilePath)
	if err != nil {
		return err
	}
	buildInfo := &labapi.CrosOpenWrtImageBuildInfo{}
	if err := protojson.Unmarshal([]byte(buildInfoFileContents), buildInfo); err != nil {
		return errors.Annotate(err, "failed to unmarshal build info file %q from host", buildInfoFilePath).Err()
	}
	if len(buildInfo.RouterFeatures) == 0 {
		buildInfo.RouterFeatures = []labapi.WifiRouterFeature{
			labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_UNKNOWN,
		}
	}
	c.state.DeviceBuildInfo = buildInfo

	// Validate required fields.
	if buildInfo.GetImageUuid() == "" {
		return errors.Reason("failed to get ImageUUID from OpenWrt build info file").Err()
	}
	if buildInfo.GetStandardBuildConfig().GetBuildProfile() == "" {
		return errors.Reason("failed to get StandardBuildConfig.BuildProfile from OpenWrt build info file").Err()
	}
	if buildInfo.GetStandardBuildConfig().GetDeviceName() == "" {
		return errors.Reason("failed to get StandardBuildConfig.DeviceName from OpenWrt build info file").Err()
	}
	return nil
}

// Reboot will reboot the router and wait for it to come back up. A non-nil
// error indicates that the router was rebooted and is ssh-able again.
func (c *OpenWrtRouterController) Reboot(ctx context.Context) error {
	return ssh.Reboot(ctx, c.sshRunner)
}
