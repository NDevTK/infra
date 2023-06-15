// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/internal/retry"
	"infra/cros/recovery/tlw"
)

const wifiRouterArtifactsGCSBasePath = "gs://chromeos-connectivity-test-artifacts/wifi_router"
const wifiRouterConfigFileGCSPath = wifiRouterArtifactsGCSBasePath + "/wifi_router_config_prod.json"

func cleanDeviceName(deviceName string) string {
	replaceRegex := regexp.MustCompile(`([^a-zA-Z0-9\-]|_)+`)
	return replaceRegex.ReplaceAllString(deviceName, "_")
}

func buildModelName(deviceType labapi.WifiRouterDeviceType, deviceName string) string {
	return fmt.Sprintf("%s[%s]", strings.TrimPrefix(deviceType.String(), "WIFI_ROUTER_DEVICE_TYPE_"), cleanDeviceName(deviceName))
}

// CleanWifiRouterFeatures removes duplicate router features, sorts them by
// name, and replaces invalid router features with
// labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID.
func CleanWifiRouterFeatures(features []labapi.WifiRouterFeature) []labapi.WifiRouterFeature {
	result := removeDuplicateWifiRouterFeatures(features)
	SortWifiRouterFeaturesByName(result)
	replaceInvalidWifiRouterFeatures(result)
	return result
}

// removeDuplicateWifiRouterFeatures returns a new list of features with any
// duplicates removed.
func removeDuplicateWifiRouterFeatures(features []labapi.WifiRouterFeature) []labapi.WifiRouterFeature {
	var result []labapi.WifiRouterFeature
	includedFeatures := make(map[int32]bool)
	for _, feature := range features {
		value := int32(feature.Number())
		if includedFeatures[value] {
			continue
		}
		includedFeatures[value] = true
		result = append(result, feature)
	}
	return result
}

// replaceInvalidWifiRouterFeatures replaces any features that are not declared
// in the labapi.WifiRouterFeature Enum with labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID.
//
// This is to highlight errors in feature lists that are pulled from external
// sources that may or may not have had their values checked against the enum
// so that they may be fixed.
func replaceInvalidWifiRouterFeatures(features []labapi.WifiRouterFeature) {
	for i, feature := range features {
		if _, ok := labapi.WifiRouterFeature_name[int32(feature)]; !ok {
			features[i] = labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID
		}
	}
}

// SortWifiRouterFeaturesByName sorts the list of features by their proto enum
// names. Unknown value names sorted by value at the end of the list.
func SortWifiRouterFeaturesByName(features []labapi.WifiRouterFeature) {
	sort.SliceStable(features, func(i, j int) bool {
		aValue := int32(features[i].Number())
		bValue := int32(features[j].Number())
		aName, aKnown := labapi.WifiRouterFeature_name[aValue]
		bName, bKnown := labapi.WifiRouterFeature_name[bValue]
		if aKnown && bKnown {
			return aName < bName
		}
		if !aKnown && !bKnown {
			return aValue < bValue
		}
		return aKnown && !bKnown
	})
}

// collectCommonWifiRouterFeatures returns a new feature set that only includes
// features present in all featureSets. Features in excludedFeatures will be
// ignored and not present in the result.
func collectCommonWifiRouterFeatures(featureSets [][]labapi.WifiRouterFeature, excludedFeatures []labapi.WifiRouterFeature) []labapi.WifiRouterFeature {
	if len(featureSets) == 0 {
		return nil
	}
	excludedFeatureLookup := make(map[int32]bool)
	for _, feature := range excludedFeatures {
		excludedFeatureLookup[int32(feature)] = true
	}
	totalFeatureCounts := make(map[int32]int32)
	for _, featureSet := range featureSets {
		featuresInSet := make(map[int32]bool)
		for _, feature := range featureSet {
			featureValue := int32(feature)
			if excludedFeatureLookup[featureValue] || featuresInSet[featureValue] {
				continue // Excluded or already counted.
			}
			featuresInSet[featureValue] = true
			totalFeatureCounts[featureValue] = totalFeatureCounts[featureValue] + 1
		}
	}
	totalFeatureSets := len(featureSets)
	var commonFeatures []labapi.WifiRouterFeature
	for featureValue, totalSetsWithFeature := range totalFeatureCounts {
		if int(totalSetsWithFeature) != totalFeatureSets {
			continue // Not in all sets, so not common.
		}
		commonFeatures = append(commonFeatures, labapi.WifiRouterFeature(featureValue))
	}
	return commonFeatures
}

// CollectOverallTestbedWifiRouterFeatures returns a single list of router
// features supported by all routers in the testbed.
//
// If any router does not have features or has an unknown feature, a list with
// just one unknown feature is returned since it is not possible to know what
// features all routers support. It is expected that all routers support at
// least one valid feature.
//
// If any router has an invalid feature, the returned list will have one invalid
// feature included to denote this for maintenance purposes.
//
// If there are no common features across all routers, a list with just one
// invalid feature is returned to denote this for maintenance purposes.
func CollectOverallTestbedWifiRouterFeatures(routers []*tlw.WifiRouterHost) []labapi.WifiRouterFeature {
	if len(routers) == 0 {
		// There are no routers, so there are no features.
		return nil
	}
	// Collect features of each router, taking note of unknown/invalid features
	// present in any set.
	var allRouterFeatureSets [][]labapi.WifiRouterFeature
	anyRouterHasAnInvalidFeature := false
	for _, router := range routers {
		routerFeaturesAreUnknown := false
		if len(router.Features) == 0 {
			// Treat an unset list of features as unknown, as it is expected that all
			// routers have at least one feature. This would mean the router has not
			// been evaluated yet, therefore the features are unknown.
			routerFeaturesAreUnknown = true
		} else {
			for _, feature := range router.Features {
				if feature == labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_UNKNOWN {
					routerFeaturesAreUnknown = true
					break
				}
				if feature == labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID {
					anyRouterHasAnInvalidFeature = true
				}
			}
		}
		if routerFeaturesAreUnknown {
			// Since one router has unknown features, we cannot be sure of the
			// overall testbed's router features.
			return []labapi.WifiRouterFeature{
				labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_UNKNOWN,
			}
		}
		allRouterFeatureSets = append(allRouterFeatureSets, router.Features)
	}
	// Collect the common router features across all routers, ignoring invalid
	// features as that is handled below.
	commonFeatures := collectCommonWifiRouterFeatures(allRouterFeatureSets, []labapi.WifiRouterFeature{
		labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
	})
	if len(commonFeatures) == 0 {
		// There are no common, valid features. Return just a single invalid entry
		// to highlight this for testbed maintenance since it means these routers
		// are unusable.
		return []labapi.WifiRouterFeature{
			labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID,
		}
	}
	if anyRouterHasAnInvalidFeature {
		// Include a single invalid feature entry to note that there is at least one
		// invalid feature among the testbed's routers to highlight this for testbed
		// maintenance. These routers are still usable for their common, valid
		// supported testing features.
		commonFeatures = append(commonFeatures, labapi.WifiRouterFeature_WIFI_ROUTER_FEATURE_INVALID)
	}
	return commonFeatures
}

// RemoteFileContentsMatch checks if the file at remoteFilePath on the remote
// host exists and that its contents match using the regex string matchRegex.
//
// Returns true if the file exists and its contents matches. Returns false
// with a nil error if the file does not exist.
// Returns:
// - true, nil: the file exists and its contents match.
// - false, nil: the file does not exist.
// - false, <error>: failed to check file or the file exists and its contents do not match.
func RemoteFileContentsMatch(ctx context.Context, sshRunner ssh.Runner, remoteFilePath, matchRegex string) (bool, error) {
	// Verify that the file exists.
	fileExists, err := ssh.TestFileExists(ctx, sshRunner, remoteFilePath)
	if err != nil {
		return false, errors.Annotate(err, "failed to check for the existence of file %q", remoteFilePath).Err()
	}
	if !fileExists {
		return false, nil
	}

	// Verify that the file contents match.
	matcher, err := regexp.Compile(matchRegex)
	if err != nil {
		return false, errors.Annotate(err, "failed to compile regex string %q", matchRegex).Err()
	}
	fileContents, err := ssh.CatFile(ctx, sshRunner, remoteFilePath)
	if err != nil {
		return false, err
	}
	return matcher.MatchString(fileContents), nil
}

// CacheAccess is a subset of tlw.Access that just has the ability to access the
// cache server.
type CacheAccess interface {
	// GetCacheUrl provides URL to download requested path to file.
	// URL will use to download image to USB-drive and provisioning.
	GetCacheUrl(ctx context.Context, resourceName, filePath string) (string, error)
}

// fetchWifiRouterConfig downloads the production WifiRouterConfig JSON file
// from GCS via the cache server through the router and returns its unmarshalled
// contents.
func fetchWifiRouterConfig(ctx context.Context, sshRunner ssh.Runner, cacheAccess CacheAccess, hostResource string) (*labapi.WifiRouterConfig, error) {
	wifiRouterConfigJSON, err := ReadFileFromCacheServer(ctx, sshRunner, cacheAccess, hostResource, 5, 30*time.Second, wifiRouterConfigFileGCSPath)
	if err != nil {
		return nil, err
	}
	config := &labapi.WifiRouterConfig{}
	if err := protojson.Unmarshal([]byte(wifiRouterConfigJSON), config); err != nil {
		return nil, errors.Annotate(err, "failed to unmarshal WifiRouterConfig from %q", wifiRouterConfigFileGCSPath).Err()
	}
	return config, nil
}

// WgetURL will run "wget <downloadURL> [additionalWgetArgs...]" on the router
// host. If the wget command fails, it will retry in 1 second intervals until
// the maxAttempts or timeout has been reached.
func WgetURL(ctx context.Context, sshRunner ssh.Runner, maxAttempts int, timeout time.Duration, downloadURL string, additionalWgetArgs ...string) (*tlw.RunResult, error) {
	wgetArgs := append([]string{downloadURL}, additionalWgetArgs...)
	var wgetResult *tlw.RunResult
	currentAttempt := 1
	if err := retry.WithTimeout(ctx, time.Second, timeout, func() error {
		if currentAttempt > maxAttempts {
			return errors.Reason("max attempts (%d) reached", maxAttempts).Err()
		}
		if _, err := sshRunner.Run(ctx, 0, "wget", wgetArgs...); err != nil {
			return err
		}
		currentAttempt += 1
		return nil
	}, fmt.Sprintf("router host wget %s", strings.Join(wgetArgs, " "))); err != nil {
		return nil, errors.Annotate(err, "failed to download file from %q on router host", downloadURL).Err()
	}
	return wgetResult, nil
}

// ReadFileFromCacheServer downloads a file from the cache server through
// the router host and then returns its contents as a string. No temporary file
// is used on the router host, as its contents are taken from the stdout of wget.
//
// The cache server will download the file from GCS if it is not already cached.
func ReadFileFromCacheServer(ctx context.Context, sshRunner ssh.Runner, cacheAccess CacheAccess, hostResource string, maxDownloadAttempts int, downloadTimeout time.Duration, srcFilePath string) (string, error) {
	// Prepare file for download from cache server.
	downloadURL, err := cacheAccess.GetCacheUrl(ctx, hostResource, srcFilePath)
	if err != nil {
		return "", errors.Annotate(err, "failed to get download URL from cache server for file path %q", srcFilePath).Err()
	}
	// Download file from cache server to router with wget to stdout.
	wgetResult, err := WgetURL(ctx, sshRunner, maxDownloadAttempts, downloadTimeout, downloadURL, "-q", "-O", "-")
	if err != nil {
		return "", err
	}
	return wgetResult.Stdout, nil
}

// DownloadFileFromCacheServer downloads a file from the cache server to the
// router host.
//
// The cache server will download the file from GCS if it is not already cached.
func DownloadFileFromCacheServer(ctx context.Context, sshRunner ssh.Runner, cacheAccess CacheAccess, hostResource string, maxDownloadAttempts int, downloadTimeout time.Duration, srcFilePath, dstFilePath string) error {
	// Prepare file for download from cache server.
	downloadURL, err := cacheAccess.GetCacheUrl(ctx, hostResource, srcFilePath)
	if err != nil {
		return errors.Annotate(err, "failed to get download URL from cache server for file path %q", srcFilePath).Err()
	}
	// Download file from cache server to router with wget to dstFilePath.
	if _, err := WgetURL(ctx, sshRunner, maxDownloadAttempts, downloadTimeout, downloadURL, "-O", dstFilePath); err != nil {
		return err
	}
	return nil
}

// PrepareCleanDir will delete (if it exists) and then recreate the directory
// at remoteDirPath on the remote host.
func PrepareCleanDir(ctx context.Context, sshRunner ssh.Runner, remoteDirPath string) error {
	exists, err := ssh.TestPath(ctx, sshRunner, "-d", remoteDirPath)
	if err != nil {
		return errors.Annotate(err, "failed to check if remote dir %q exists", remoteDirPath).Err()
	}
	if exists {
		if _, err := sshRunner.Run(ctx, 0, "rm", "-r", remoteDirPath); err != nil {
			return errors.Annotate(err, "failed to remove existing remote dir %q", remoteDirPath).Err()
		}
	}
	if _, err := sshRunner.Run(ctx, 0, "mkdir", "-p", remoteDirPath); err != nil {
		return errors.Reason("failed to create new remote dir %q", remoteDirPath).Err()
	}
	return nil
}
