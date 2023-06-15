// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"
	"regexp"
	"sort"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"go.chromium.org/luci/common/errors"
	"infra/cros/recovery/internal/execs/wifirouter/ssh"
	"infra/cros/recovery/tlw"
)

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
