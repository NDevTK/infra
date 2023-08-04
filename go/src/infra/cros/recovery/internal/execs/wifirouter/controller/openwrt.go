// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"sort"
	"strings"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/internal/log"
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

// Reboot will reboot the router and wait for it to come back up. A non-nil
// error indicates that the router was rebooted and is ssh-able again.
func (c *OpenWrtRouterController) Reboot(ctx context.Context) error {
	return ssh.Reboot(ctx, c.sshRunner)
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

// FetchGlobalImageConfig retrieves the global image config from GCS for this
// router and stores it in the controller state.
func (c *OpenWrtRouterController) FetchGlobalImageConfig(ctx context.Context) error {
	if c.state.GetDeviceBuildInfo() == nil {
		return errors.Reason("state.DeviceBuildInfo must not be nil").Err()
	}
	deviceName := c.state.GetDeviceBuildInfo().GetStandardBuildConfig().GetDeviceName()

	// Fetch and unmarshal config from GCS.
	wifiRouterConfig, err := fetchWifiRouterConfig(ctx, c.sshRunner, c.cacheAccess, c.resource)
	if err != nil {
		return err
	}
	if wifiRouterConfig.GetOpenwrt() == nil {
		return errors.Reason("wifi router config missing OpenWrt configuration").Err()
	}

	// Select the config for this device.
	deviceConfig, ok := wifiRouterConfig.Openwrt[deviceName]
	if !ok || deviceConfig == nil {
		return errors.Reason("no config available for OpenWrt device %q", deviceName).Err()
	}

	// Validate required fields.
	if len(deviceConfig.Images) == 0 {
		return errors.Reason("OpenWrt device config for %q does not have any images", deviceName).Err()
	}
	for i, image := range deviceConfig.Images {
		if err := c.validateOpenWrtOSImage(image); err != nil {
			return errors.Annotate(err, "OpenWrt device config for %q has an invalid image at OpenWrtWifiRouterDeviceConfig.images[%d]", deviceName, i).Err()
		}
	}

	c.state.Config = deviceConfig
	return nil
}

// validateOpenWrtOSImage will check all properties of an image and throw an
// error if any are empty or if MinDutReleaseVersion is invalid. All images
// are expected to have all of these fields populated in the config file.
func (c *OpenWrtRouterController) validateOpenWrtOSImage(image *labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage) error {
	if image == nil {
		return errors.Reason("image is nil").Err()
	}
	if image.ImageUuid == "" {
		return errors.Reason("ImageUuid is empty").Err()
	}
	if image.ArchivePath == "" {
		return errors.Reason("ArchivePath is empty").Err()
	}
	if image.MinDutReleaseVersion == "" {
		return errors.Reason("MinDutReleaseVersion is empty").Err()
	}
	if _, err := cros.ParseChromeOSReleaseVersion(image.MinDutReleaseVersion); err != nil {
		return errors.Annotate(err, "MinDutReleaseVersion is invalid").Err()
	}
	return nil
}

// IdentifyExpectedImage will select an appropriate OpenWrt OS image from the
// available images in the config based on its model and the provided
// dutHostname and dutChromeOSReleaseVersion. The selected image's UUID is
// stored for later reference. A non-nil error is returned if an image was not
// selected.
//
// Each supported OpenWrt device has its own set of images, as defined by the
// corresponding labapi.OpenWrtWifiRouterDeviceConfig. See the proto definition
// for more details on how this config defines which image should be used.
func (c *OpenWrtRouterController) IdentifyExpectedImage(ctx context.Context, dutHostname, dutChromeOSReleaseVersion string) error {
	if c.state.GetDeviceBuildInfo() == nil {
		return errors.Reason("state.DeviceBuildInfo must not be nil").Err()
	}
	if c.state.GetConfig() == nil {
		return errors.Reason("state.Config must not be nil").Err()
	}
	log.Debugf(ctx, "Identifying expected router OpenWrt OS image for dut %q with CHROMEOS_RELEASE_VERSION %q", dutHostname, dutChromeOSReleaseVersion)
	expectedImageConfig, err := c.selectImageForDut(ctx, c.state.Config, dutHostname, dutChromeOSReleaseVersion)
	if err != nil {
		return errors.Annotate(err, "failed to select image for dut %q with CHROMEOS_RELEASE_VERSION %q", dutHostname, dutChromeOSReleaseVersion).Err()
	}
	log.Debugf(ctx, "Identified image %q as expected router OpenWrt OS image for dut %q with CHROMEOS_RELEASE_VERSION %q", expectedImageConfig.ImageUuid, dutHostname, dutChromeOSReleaseVersion)
	c.state.ExpectedImageUuid = expectedImageConfig.ImageUuid
	return nil
}

// selectImageByUUID returns the image config from the deviceConfig that has
// a matching image UUID. Returns a non-nil error if no matching image config
// is found.
func (c *OpenWrtRouterController) selectImageByUUID(deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig, imageUUID string) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	for _, imageConfig := range deviceConfig.GetImages() {
		if strings.EqualFold(imageConfig.GetImageUuid(), imageUUID) {
			return imageConfig, nil
		}
	}
	return nil, errors.Reason("found no image with ImageUUID %q configured for device", imageUUID).Err()
}

// selectCurrentImage returns the image config from the device config that is
// specified as the current image in the device config. Returns a non-nil error
// if the device config has no current image specified or the image UUID it
// specifies does not match any image configs in the device config.
func (c *OpenWrtRouterController) selectCurrentImage(deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	if deviceConfig.GetCurrentImageUuid() == "" {
		return nil, errors.Reason("no current image configured for device").Err()
	}
	imageConfig, err := c.selectImageByUUID(deviceConfig, deviceConfig.GetCurrentImageUuid())
	if err != nil {
		return nil, errors.Annotate(err, "failed to select image set as current image").Err()
	}
	return imageConfig, nil
}

// selectNextImage returns the image config from the device config that is
// specified as the next image in the device config. Returns a non-nil error
// if the device config has no next image specified or the image UUID it
// specifies does not match any image configs in the device config.
func (c *OpenWrtRouterController) selectNextImage(deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	if deviceConfig.GetNextImageUuid() == "" {
		return nil, errors.Reason("no next image configured for device").Err()
	}
	imageConfig, err := c.selectImageByUUID(deviceConfig, deviceConfig.GetNextImageUuid())
	if err != nil {
		return nil, errors.Annotate(err, "failed to select image set as next image").Err()
	}
	return imageConfig, nil
}

// selectImageByCrosReleaseVersion returns the image config that has the
// latest MinDutReleaseVersion that is less than or equal to the provided
// dutCrosReleaseVersion among all image configs in the device config. If the
// dutCrosReleaseVersion is less than all images' MinDutReleaseVersions and
// useCurrentIfNoMatches is true, the current image config is returned, as
// determined by selectCurrentImage.
func (c *OpenWrtRouterController) selectImageByCrosReleaseVersion(ctx context.Context, deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig, dutCrosReleaseVersion string, useCurrentIfNoMatches bool) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	if len(deviceConfig.GetImages()) == 0 {
		return nil, errors.Reason("no images are configured for this device").Err()
	}
	dutVersion, err := cros.ParseChromeOSReleaseVersion(dutCrosReleaseVersion)
	if err != nil {
		return nil, errors.Annotate(err, "invalid CHROMEOS_RELEASE_VERSION %q", dutCrosReleaseVersion).Err()
	}

	// Collect all matching versions.
	var allMatchingImageVersions []cros.ChromeOSReleaseVersion
	imageVersionToConfig := make(map[string]*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage)
	for i, imageConfig := range deviceConfig.GetImages() {
		imageMinVersion, err := cros.ParseChromeOSReleaseVersion(imageConfig.GetMinDutReleaseVersion())
		if err != nil {
			return nil, errors.Annotate(err, "failed to parse deviceConfig.Images[%d].MinDutReleaseVersion", i).Err()
		}
		if !cros.IsChromeOSReleaseVersionLessThan(dutVersion, imageMinVersion) {
			allMatchingImageVersions = append(allMatchingImageVersions, imageMinVersion)
			imageVersionToConfig[imageMinVersion.String()] = imageConfig
		}
	}
	if len(allMatchingImageVersions) == 0 {
		if useCurrentIfNoMatches {
			log.Warningf(ctx, "None of the %d images configured for this device have a MinDutReleaseVersion greater than or equal to %q; selecting current version instead", len(deviceConfig.GetImages()), dutVersion.String())
			return c.selectCurrentImage(deviceConfig)
		}
		return nil, errors.Reason("none of the %d images configured for this device have a MinDutReleaseVersion greater than or equal to %q", len(deviceConfig.GetImages()), dutVersion.String()).Err()
	}

	// Sort them and use the highest matching min version.
	sort.SliceStable(allMatchingImageVersions, func(i, j int) bool {
		return cros.IsChromeOSReleaseVersionLessThan(allMatchingImageVersions[i], allMatchingImageVersions[j])
	})
	highestMatchingVersion := allMatchingImageVersions[len(allMatchingImageVersions)-1]
	if len(allMatchingImageVersions) > 1 {
		secondHighestMatchingVersion := allMatchingImageVersions[len(allMatchingImageVersions)-2]
		if !cros.IsChromeOSReleaseVersionLessThan(secondHighestMatchingVersion, highestMatchingVersion) {
			// Versions are the same, and thus we cannot pick between the two images
			// they belong to (this is a config error we'd need to fix manually).
			return nil, errors.Reason("config error: unable to choose image for CHROMEOS_RELEASE_VERSION %q, as multiple matching images were found with the same MinDutReleaseVersion %q", dutVersion.String(), highestMatchingVersion.String()).Err()
		}
	}
	return imageVersionToConfig[highestMatchingVersion.String()], nil
}

// selectImageForDut returns the config of the OpenWrt OS image that the router
// should be using, as specified by the rules of the device config, the dut's
// hostname, and the dut's ChromeOS release version.
//
// If the dut's hostname is in the NextImageVerificationDutPool, the next image
// is returned, as specified by selectNextImage. These are testbeds that we use
// to validate that new router images work as expected before releasing them to
// all the other testbeds. Their related test results are monitored separately.
//
// If the dutCrosReleaseVersion is provided (not empty), then the image is
// selected with selectImageByCrosReleaseVersion. Otherwise, the current image
// is selected using selectCurrentImage.
func (c *OpenWrtRouterController) selectImageForDut(ctx context.Context, deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig, dutHostname, dutCrosReleaseVersion string) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	var dutVersion cros.ChromeOSReleaseVersion
	if dutCrosReleaseVersion != "" {
		var err error
		dutVersion, err = cros.ParseChromeOSReleaseVersion(dutCrosReleaseVersion)
		if err != nil {
			return nil, errors.Annotate(err, "invalid CHROMEOS_RELEASE_VERSION %q", dutCrosReleaseVersion).Err()
		}
	}
	for _, hostname := range deviceConfig.GetNextImageVerificationDutPool() {
		if hostname == dutHostname {
			log.Debugf(ctx, "DUT %q is in the NextImageVerificationDutPool, selecting next image\n", dutHostname)
			if dutVersion == nil {
				return c.selectNextImage(deviceConfig)
			}
			nextImage, err := c.selectNextImage(deviceConfig)
			if err != nil {
				return nil, err
			}
			nextImageMinVersion, err := cros.ParseChromeOSReleaseVersion(nextImage.GetMinDutReleaseVersion())
			if err != nil {
				return nil, errors.Annotate(err, "failed to parse MinDutReleaseVersion for next image with UUID %q", nextImage.GetImageUuid()).Err()
			}
			if cros.IsChromeOSReleaseVersionLessThan(dutVersion, nextImageMinVersion) {
				log.Warningf(ctx, "The DUT CHROMEOS_RELEASE_VERSION provided, %q, is less than the selected next OpenWrt OS image MinDutReleaseVersion %q\n", dutVersion.String(), nextImage.String())
			}
			return nextImage, nil
		}
	}
	if dutVersion == nil {
		log.Debugf(ctx, "DUT %q is not in the NextImageVerificationDutPool and no DUT CHROMEOS_RELEASE_VERSION specified, selecting current image\n", dutHostname)
		return c.selectCurrentImage(deviceConfig)
	}
	currentImage, err := c.selectCurrentImage(deviceConfig)
	if err != nil {
		return nil, err
	}
	currentImageMinVersion, err := cros.ParseChromeOSReleaseVersion(currentImage.MinDutReleaseVersion)
	if err != nil {
		return nil, errors.Annotate(err, "failed to parse MinDutReleaseVersion for current image with UUID %q", currentImage.ImageUuid).Err()
	}
	if cros.IsChromeOSReleaseVersionLessThan(dutVersion, currentImageMinVersion) {
		log.Warningf(ctx, "DUT %q is not in the NextImageVerificationDutPool and the current image's MinDutReleaseVersion is higher than requested, selecting image by CHROMEOS_RELEASE_VERSION from available images\n", dutHostname)
		return c.selectImageByCrosReleaseVersion(ctx, deviceConfig, dutCrosReleaseVersion, true)
	}
	return currentImage, nil
}

// AssertHasExpectedImage asserts that router has the expected image installed
// on it.
//
// This function may only be called after the expected image has been identified
// with IdentifyExpectedImage.
func (c *OpenWrtRouterController) AssertHasExpectedImage() error {
	if c.state.GetDeviceBuildInfo().GetImageUuid() == "" {
		return errors.Reason("the device's installed OpenWrt image's UUID is not known").Err()
	}
	if c.state.GetExpectedImageUuid() == "" {
		return errors.Reason("expected OpenWrt image not yet identified for device").Err()
	}
	if !strings.EqualFold(c.state.DeviceBuildInfo.ImageUuid, c.state.ExpectedImageUuid) {
		return errors.Reason("device is expected to have an OpenWrt image with UUID %q, but its installed image has %q", c.state.ExpectedImageUuid, c.state.DeviceBuildInfo.ImageUuid).Err()
	}
	return nil
}
