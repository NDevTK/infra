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

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
)

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

// CollectCommonWifiRouterFeatures returns a new feature set that only includes
// features present in all featureSets.
func CollectCommonWifiRouterFeatures(featureSets [][]labapi.WifiRouterFeature) []labapi.WifiRouterFeature {
	if len(featureSets) == 0 {
		return nil
	}
	totalFeatureCounts := make(map[int32]int32)
	for _, featureSet := range featureSets {
		featuresInSet := make(map[int32]bool)
		for _, feature := range featureSet {
			featureValue := int32(feature)
			if featuresInSet[featureValue] {
				continue // Already counted.
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
