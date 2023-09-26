// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package chameleond

import (
	"context"
	"sort"
	"strings"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/components/cros"
	"infra/cros/recovery/internal/log"
)

// SelectChameleondBundleByChameleondCommit returns the bundle config from the
// chameleond config that has a matching chameleond commit. Returns a non-nil
// error if no matching bundle config is found.
func SelectChameleondBundleByChameleondCommit(config *labapi.BluetoothPeerChameleondConfig, chameleondCommit string) (*labapi.BluetoothPeerChameleondConfig_ChameleondBundle, error) {
	for _, bundleConfig := range config.GetBundles() {
		if strings.EqualFold(bundleConfig.GetChameleondCommit(), chameleondCommit) {
			return bundleConfig, nil
		}
	}
	return nil, errors.Reason("select chameleond bundle by chameleond commit: found no bundle with ChameleondCommit %q configured", chameleondCommit).Err()
}

// SelectChameleondBundleByNextCommit returns the bundle config that is
// specified as the next bundle by commit in the chameleond config. Returns a
// non-nil error if the chameleond config has no next bundle configured or if
// the chameleond commit it specifies does not match any configured bundles.
func SelectChameleondBundleByNextCommit(config *labapi.BluetoothPeerChameleondConfig) (*labapi.BluetoothPeerChameleondConfig_ChameleondBundle, error) {
	if config.GetNextChameleondCommit() == "" {
		return nil, errors.Reason("select chameleon bundle by next commit: next commit is empty").Err()
	}
	bundleConfig, err := SelectChameleondBundleByChameleondCommit(config, config.GetNextChameleondCommit())
	if err != nil {
		return nil, errors.Annotate(err, "select chameleon bundle by next commit").Err()
	}
	return bundleConfig, nil
}

// SelectChameleondBundleByCrosReleaseVersion returns the bundle config that has
// the latest MinDutReleaseVersion that is less than or equal to the provided
// dutCrosReleaseVersion among all bundles configured in the chameleond config,
// excluding the bundle configured as the next bundle.
func SelectChameleondBundleByCrosReleaseVersion(config *labapi.BluetoothPeerChameleondConfig, dutCrosReleaseVersion string) (*labapi.BluetoothPeerChameleondConfig_ChameleondBundle, error) {
	if len(config.GetBundles()) == 0 {
		return nil, errors.Reason("select chameleond bundle by release version: no bundles are configured").Err()
	}
	dutVersion, err := cros.ParseChromeOSReleaseVersion(dutCrosReleaseVersion)
	if err != nil {
		return nil, errors.Annotate(err, "select chameleond bundle by release version").Err()
	}

	// Collect all matching versions.
	var allMatchingBundleVersions []cros.ChromeOSReleaseVersion
	bundleVersionToConfig := make(map[string]*labapi.BluetoothPeerChameleondConfig_ChameleondBundle)
	for i, bundleConfig := range config.GetBundles() {
		if strings.EqualFold(bundleConfig.GetChameleondCommit(), config.GetNextChameleondCommit()) {
			// Do not include next bundle when matching by version.
			continue
		}
		bundleMinVersion, err := cros.ParseChromeOSReleaseVersion(bundleConfig.GetMinDutReleaseVersion())
		if err != nil {
			return nil, errors.Annotate(err, "select chameleond bundle by release version: parse for config.Bundles[%d].MinDutReleaseVersion", i).Err()
		}
		if !cros.IsChromeOSReleaseVersionLessThan(dutVersion, bundleMinVersion) {
			allMatchingBundleVersions = append(allMatchingBundleVersions, bundleMinVersion)
			bundleVersionToConfig[bundleMinVersion.String()] = bundleConfig
		}
	}
	if len(allMatchingBundleVersions) == 0 {
		return nil, errors.Reason("select chameleond bundle by release version: none of the %d bundles configured have a MinDutReleaseVersion greater than or equal to %q", len(config.GetBundles()), dutVersion.String()).Err()
	}

	// Sort them and use the highest matching min version.
	sort.SliceStable(allMatchingBundleVersions, func(i, j int) bool {
		return cros.IsChromeOSReleaseVersionLessThan(allMatchingBundleVersions[i], allMatchingBundleVersions[j])
	})
	highestMatchingVersion := allMatchingBundleVersions[len(allMatchingBundleVersions)-1]
	if len(allMatchingBundleVersions) > 1 {
		secondHighestMatchingVersion := allMatchingBundleVersions[len(allMatchingBundleVersions)-2]
		if !cros.IsChromeOSReleaseVersionLessThan(secondHighestMatchingVersion, highestMatchingVersion) {
			// Versions are the same, and thus we cannot pick between the two bundles
			// they belong to (this is a config error we'd need to fix manually).
			return nil, errors.Reason("select chameleond bundle by release version: config error: unable to choose bundle for CHROMEOS_RELEASE_VERSION %q, as multiple matching bundles were found with the same MinDutReleaseVersion %q", dutVersion.String(), highestMatchingVersion.String()).Err()
		}
	}
	return bundleVersionToConfig[highestMatchingVersion.String()], nil
}

// SelectChameleondBundleForDut returns the bundle config that the btpeer should
// be using for chameleond based on the chameleond config, the dut's hostname,
// and the dut's ChromeOS release version.
//
// If there is a next bundle configured, the dut's hostname is in the
// NextDutHosts, and the dut's ChromeOS release version is in the
// NextDutReleaseVersions, the next chameleond bundle is selected via
// SelectChameleondBundleByNextCommit. Otherwise, the bundle is selected
// via SelectChameleondBundleByCrosReleaseVersion.
func SelectChameleondBundleForDut(ctx context.Context, config *labapi.BluetoothPeerChameleondConfig, dutHostname, dutCrosReleaseVersion string) (*labapi.BluetoothPeerChameleondConfig_ChameleondBundle, error) {
	dutVersion, err := cros.ParseChromeOSReleaseVersion(dutCrosReleaseVersion)
	if err != nil {
		return nil, errors.Annotate(err, "select chameleond bundle for dut").Err()
	}

	// Determine if next bundle should be used.
	nextBundleConfigured := config.GetNextChameleondCommit() != ""
	dutInNextHosts := false
	dutVersionInNextVersions := false
	for _, hostname := range config.GetNextDutHosts() {
		if hostname == dutHostname {
			dutInNextHosts = true
			break
		}
	}
	for _, version := range config.GetNextDutReleaseVersions() {
		if version == dutCrosReleaseVersion {
			dutVersionInNextVersions = true
			break
		}
	}
	shouldUseNextBundle := nextBundleConfigured && dutInNextHosts && dutVersionInNextVersions
	log.Debugf(
		ctx,
		"Should DUT %q with version %q use the next bundle? nextBundleConfigured=%t, dutInNextHosts=%t, dutVersionInNextVersions=%t => shouldUseNextBundle=%t",
		dutHostname,
		dutCrosReleaseVersion,
		nextBundleConfigured,
		dutInNextHosts,
		dutVersionInNextVersions,
		shouldUseNextBundle,
	)

	if shouldUseNextBundle {
		// Select next bundle.
		log.Debugf(ctx, "Selecting next bundle for DUT %q with version %q", dutHostname, dutCrosReleaseVersion)
		nextBundle, err := SelectChameleondBundleByNextCommit(config)
		if err != nil {
			return nil, errors.Annotate(err, "select chameleond bundle for dut").Err()
		}
		nextBundleMinVersion, err := cros.ParseChromeOSReleaseVersion(nextBundle.GetMinDutReleaseVersion())
		if err != nil {
			return nil, errors.Annotate(err, "select chameleond bundle for dut: failed to parse MinDutReleaseVersion for next bundle with ChameleondCommit %q", nextBundle.GetChameleondCommit()).Err()
		}
		if cros.IsChromeOSReleaseVersionLessThan(dutVersion, nextBundleMinVersion) {
			log.Warningf(ctx, "The DUT CHROMEOS_RELEASE_VERSION provided, %q, is less than the selected next bundle MinDutReleaseVersion %q", dutVersion.String(), nextBundle.String())
		}
		return nextBundle, nil
	}

	// Not using next bundle, so select by version.
	log.Debugf(ctx, "Selecting bundle by version for DUT %q with version %q", dutHostname, dutCrosReleaseVersion)
	return SelectChameleondBundleByCrosReleaseVersion(config, dutCrosReleaseVersion)
}
