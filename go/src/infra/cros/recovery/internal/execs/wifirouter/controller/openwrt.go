// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"path/filepath"
	"sort"
	"strings"
	"time"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

const (
	deviceInfoFilePath        = "/etc/device_info"
	buildInfoFilePath         = "/etc/cros/cros_openwrt_image_build_info.json"
	deviceInfoMatchIfOpenWrt  = "(?m)^DEVICE_MANUFACTURER='OpenWrt'$"
	openWrtArchiveBaseGCSPath = wifiRouterArtifactsGCSBasePath + "/openwrt_images/"
	openWrtArchiveFileExt     = ".tar.xz"
	openWrtImageFileExt       = ".bin"
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
	if c.state == nil || c.state.DeviceBuildInfo == nil {
		return "", errors.Reason("state.DeviceBuildInfo must not be nil").Err()
	}
	return buildModelName(c.DeviceType(), c.state.DeviceBuildInfo.StandardBuildConfig.DeviceName), nil
}

// Features returns the router features that this router supports.
//
// For OpenWrt routers, features are retrieved from the build info file placed
// on the router by the custom ChromeOS OpenWrt image builder.
func (c *OpenWrtRouterController) Features() ([]labapi.WifiRouterFeature, error) {
	if c.state == nil || c.state.DeviceBuildInfo == nil {
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
	if buildInfo.ImageUuid == "" {
		return errors.Reason("failed to get ImageUUID from OpenWrt build info file").Err()
	}
	if buildInfo.StandardBuildConfig == nil || buildInfo.StandardBuildConfig.BuildProfile == "" {
		return errors.Reason("failed to get StandardBuildConfig.BuildProfile from OpenWrt build info file").Err()
	}
	if buildInfo.StandardBuildConfig.DeviceName == "" {
		return errors.Reason("failed to get StandardBuildConfig.DeviceName from OpenWrt build info file").Err()
	}
	return nil
}

// FetchGlobalImageConfig retrieves the global image config from GCS for this
// router and stores it in the controller state.
func (c *OpenWrtRouterController) FetchGlobalImageConfig(ctx context.Context) error {
	if c.state == nil || c.state.DeviceBuildInfo == nil {
		return errors.Reason("state.DeviceBuildInfo must not be nil").Err()
	}
	deviceName := c.state.DeviceBuildInfo.StandardBuildConfig.DeviceName

	// Fetch and unmarshal config from GCS.
	wifiRouterConfig, err := fetchWifiRouterConfig(ctx, c.sshRunner, c.cacheAccess, c.resource)
	if err != nil {
		return err
	}
	if wifiRouterConfig.Openwrt == nil {
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
	if c.state == nil || c.state.DeviceBuildInfo == nil || c.state.Config == nil {
		return errors.Reason("state.DeviceBuildInfo and state.Config must not be nil").Err()
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

func (c *OpenWrtRouterController) selectImageByUUID(deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig, imageUUID string) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	for _, imageConfig := range deviceConfig.Images {
		if strings.EqualFold(imageConfig.ImageUuid, imageUUID) {
			return imageConfig, nil
		}
	}
	return nil, errors.Reason("found no image with ImageUUID %q configured for device", imageUUID).Err()
}

func (c *OpenWrtRouterController) selectCurrentImage(deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	if deviceConfig.CurrentImageUuid == "" {
		return nil, errors.New("no current image configured for device")
	}
	imageConfig, err := c.selectImageByUUID(deviceConfig, deviceConfig.CurrentImageUuid)
	if err != nil {
		return nil, errors.Annotate(err, "failed to select image set as current image").Err()
	}
	return imageConfig, nil
}
func (c *OpenWrtRouterController) selectNextImage(deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	if deviceConfig.NextImageUuid == "" {
		return nil, errors.New("no next image configured for device")
	}
	imageConfig, err := c.selectImageByUUID(deviceConfig, deviceConfig.NextImageUuid)
	if err != nil {
		return nil, errors.Annotate(err, "failed to select image set as next image").Err()
	}
	return imageConfig, nil
}

func (c *OpenWrtRouterController) selectImageByCrosReleaseVersion(ctx context.Context, deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig, dutCrosReleaseVersion string, useCurrentIfNoMatches bool) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	if len(deviceConfig.Images) == 0 {
		return nil, errors.Reason("no images are configured for this device").Err()
	}
	dutVersion, err := cros.ParseChromeOSReleaseVersion(dutCrosReleaseVersion)
	if err != nil {
		return nil, errors.Annotate(err, "invalid CHROMEOS_RELEASE_VERSION %q", dutCrosReleaseVersion).Err()
	}

	// Collect all matching versions.
	var allMatchingImageVersions []cros.ChromeOSReleaseVersion
	imageVersionToConfig := make(map[string]*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage)
	for i, imageConfig := range deviceConfig.Images {
		imageMinVersion, err := cros.ParseChromeOSReleaseVersion(imageConfig.MinDutReleaseVersion)
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
			log.Warningf(ctx, "None of the %d images configured for this device have a MinDutReleaseVersion greater than or equal to %q; selecting current version instead", len(deviceConfig.Images), dutVersion.String())
			return c.selectCurrentImage(deviceConfig)
		}
		return nil, errors.Reason("none of the %d images configured for this device have a MinDutReleaseVersion greater than or equal to %q", len(deviceConfig.Images), dutVersion.String()).Err()
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

func (c *OpenWrtRouterController) selectImageForDut(ctx context.Context, deviceConfig *labapi.OpenWrtWifiRouterDeviceConfig, dutHostname, dutCrosReleaseVersion string) (*labapi.OpenWrtWifiRouterDeviceConfig_OpenWrtOSImage, error) {
	var dutVersion cros.ChromeOSReleaseVersion
	if dutCrosReleaseVersion != "" {
		var err error
		dutVersion, err = cros.ParseChromeOSReleaseVersion(dutCrosReleaseVersion)
		if err != nil {
			return nil, errors.Annotate(err, "invalid CHROMEOS_RELEASE_VERSION %q", dutCrosReleaseVersion).Err()
		}
	}
	for _, hostname := range deviceConfig.NextImageVerificationDutPool {
		if hostname == dutHostname {
			log.Debugf(ctx, "DUT %q is in the NextImageVerificationDutPool, selecting next image\n", dutHostname)
			if dutVersion == nil {
				return c.selectNextImage(deviceConfig)
			}
			nextImage, err := c.selectNextImage(deviceConfig)
			if err != nil {
				return nil, err
			}
			nextImageMinVersion, err := cros.ParseChromeOSReleaseVersion(nextImage.MinDutReleaseVersion)
			if err != nil {
				return nil, errors.Annotate(err, "failed to parse MinDutReleaseVersion for next image with UUID %q", nextImage.ImageUuid).Err()
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
	if c.state == nil || c.state.DeviceBuildInfo == nil {
		return errors.Reason("state.DeviceBuildInfo must not be nil").Err()
	}
	if c.state.DeviceBuildInfo.ImageUuid == "" {
		return errors.Reason("device does not have an image UUID in its OpenWrt build info").Err()
	}
	if c.state.ExpectedImageUuid == "" {
		return errors.Reason("expected OpenWrt image not yet identified for device").Err()
	}
	if !strings.EqualFold(c.state.DeviceBuildInfo.ImageUuid, c.state.ExpectedImageUuid) {
		return errors.Reason("device is expected to have an OpenWrt image with UUID %q, but its installed image has %q", c.state.ExpectedImageUuid, c.state.DeviceBuildInfo.ImageUuid).Err()
	}
	return nil
}

// UpdateToExpectedImage uses the `sysupgrade` command on the router to update
// the image installed on the router with the expected image. This can take a
// few minutes, as it does not return until the new image is installed and the
// router has come back up. Once it is back up, the build info file is retrieved
// from the device again and the ImageUUID is checked to ensure that the image
// that was installed is the expected ImageUUID.
//
// The image binary file used is retrieved from archives stored in GCS, which
// are uploaded manually by developers using the cros_openwrt_image_builder CLI
// utility. The infra cache server downloads the archive from GCS (if it is not
// cached), then the router downloads the archive from the cache server,
// extracts the image binary from the archive locally, then runs `sysupgrade`.
//
// This function may only be called after the expected image has been identified
// with IdentifyExpectedImage.
func (c *OpenWrtRouterController) UpdateToExpectedImage(ctx context.Context) error {
	if c.state == nil || c.state.ExpectedImageUuid == "" {
		return errors.Reason("expected OpenWrt image not yet identified for device").Err()
	}
	return c.updateImage(ctx, c.state.ExpectedImageUuid)
}

func (c *OpenWrtRouterController) updateImage(ctx context.Context, imageUUID string) error {
	if c.state == nil || c.state.Config == nil {
		return errors.Reason("state.Config must not be nil").Err()
	}
	if c.state.DeviceBuildInfo == nil {
		return errors.Reason("state.DeviceBuildInfo must not be nil").Err()
	}
	originalImageUUID := c.state.DeviceBuildInfo.ImageUuid
	log.Infof(ctx, "Updating router OpenWrt image from %q to %q", originalImageUUID, imageUUID)

	// Get the corresponding image binary from the stored archive to the device.
	desiredImage, err := c.selectImageByUUID(c.state.Config, imageUUID)
	if err != nil {
		return err
	}
	log.Infof(ctx, "Downloading to OpenWrt image from %q to router", desiredImage.ArchivePath)
	remoteImageBinaryPath, err := c.downloadImageToDevice(ctx, desiredImage.ArchivePath)
	if err != nil {
		return errors.Annotate(err, "failed to download image to device").Err()
	}

	// Flash the device with the new image binary.
	log.Infof(ctx, "Flashing OpenWrt device with new image binary at %q", remoteImageBinaryPath)
	sysupgradeRunResult := c.sshRunner.RunForResult(ctx, 0, false, "sysupgrade", remoteImageBinaryPath)
	if sysupgradeRunResult.GetExitCode() != -2 {
		return errors.Reason("sysupgrade did not cause remote command to exit as expected: ExitCode=%d, Stdout=%s", sysupgradeRunResult.GetExitCode(), sysupgradeRunResult.GetStdout()).Err()
	}
	log.Infof(ctx, "Waiting 1m before reconnecting to give time for OpenWrt sysupgrade to complete")
	time.Sleep(1 * time.Minute)
	log.Infof(ctx, "Attempting to reconnect to OpenWrt router after sysupgrade")
	if err := cros.WaitUntilSSHable(ctx, 4*time.Minute, 10*time.Second, c.sshRunner.Run, log.Get(ctx)); err != nil {
		return errors.Annotate(err, "failed to reconnect to OpenWrt device over ssh after flashing new image binary").Err()
	}
	log.Infof(ctx, "Successfully reconnected to OpenWrt router after sysupgrade")

	// Update cached device build info.
	if err := c.FetchDeviceBuildInfo(ctx); err != nil {
		return errors.Annotate(err, "failed to fetch build info after flashing image with uuid %q", imageUUID).Err()
	}

	// Verify device image matches expected image.
	if c.state.DeviceBuildInfo.ImageUuid != imageUUID {
		return errors.Annotate(err, "unexpected image uuid after flashing new image; got %q, expected %q", c.state.DeviceBuildInfo.ImageUuid, imageUUID).Err()
	}

	log.Infof(ctx, "Successfully updated router OpenWrt image from %q to %q", originalImageUUID, imageUUID)
	return nil
}

// downloadImageToDevice downloads the image archive from GCS through the
// cache server to the router directly, extracts the image binary file from it,
// and then deletes the original image archive that was downloaded. The
// absolute path to the image binary file on the router is returned.
func (c *OpenWrtRouterController) downloadImageToDevice(ctx context.Context, imageArchiveGCSPath string) (string, error) {
	// Download archive to router.
	if err := c.validateImageArchivePath(imageArchiveGCSPath); err != nil {
		return "", errors.Annotate(err, "invalid image archive GCS path %q", imageArchiveGCSPath).Err()
	}
	tmpDir := filepath.Join("/tmp/cros_infra_image")
	if err := PrepareCleanDir(ctx, c.sshRunner, tmpDir); err != nil {
		return "", errors.Annotate(err, "failed to create tmp image dir %q on router", tmpDir).Err()
	}
	archiveRouterPath := filepath.Join(tmpDir, filepath.Base(imageArchiveGCSPath))
	if err := DownloadFileFromCacheServer(ctx, c.sshRunner, c.cacheAccess, c.resource, 5, 1*time.Minute, imageArchiveGCSPath, archiveRouterPath); err != nil {
		return "", err
	}

	// Extract binary from archive and get filepath from output.
	tarRunOutput, err := c.sshRunner.Run(ctx, 0,
		// Run tar in tmp dir, as it can only extract to current dir.
		"cd",
		tmpDir,
		"&&",
		"tar",

		// Extract from the image archive file.
		"-x",
		"-f",
		archiveRouterPath,

		// Use xz decompression.
		"-J",
		"--xz",

		// Only extract image binary file.
		"--wildcards",
		"*"+openWrtImageFileExt,

		// List extracted files in stdout (paths will be relative).
		"-v",
	)
	if err != nil {
		return "", errors.Annotate(err, "failed to extract binary from archive %q on router", archiveRouterPath).Err()
	}
	imageBinaryRouterPath := filepath.Join(tmpDir, strings.TrimSpace(tarRunOutput))
	if !strings.HasSuffix(imageBinaryRouterPath, openWrtImageFileExt) {
		return "", errors.Reason("invalid image binary file %q extracted from image archive %q on router", imageBinaryRouterPath, imageArchiveGCSPath).Err()
	}
	exists, err := ssh.TestFileExists(ctx, c.sshRunner, imageBinaryRouterPath)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.Reason("extracted image binary file %q does not exist on router as expected", imageBinaryRouterPath).Err()
	}

	// Delete archive from router to conserve space.
	if _, err := c.sshRunner.Run(ctx, 0, "rm", archiveRouterPath); err != nil {
		return "", errors.Annotate(err, "failed to delete archive %q on router after image binary extraction", archiveRouterPath).Err()
	}

	return imageBinaryRouterPath, nil
}

func (c *OpenWrtRouterController) validateImageArchivePath(imageArchiveGCSPath string) error {
	if !strings.HasPrefix(imageArchiveGCSPath, openWrtArchiveBaseGCSPath) {
		return errors.Reason("must start with %q", openWrtArchiveBaseGCSPath).Err()
	}
	if !strings.HasSuffix(imageArchiveGCSPath, openWrtArchiveFileExt) {
		return errors.Reason("must end with %q", openWrtArchiveFileExt).Err()
	}
	return nil
}

// Reboot will reboot the router and wait for it to come back up. A non-nil
// error indicates that the router was rebooted and is ssh-able again.
func (c *OpenWrtRouterController) Reboot(ctx context.Context) error {
	return ssh.Reboot(ctx, c.sshRunner)
}
