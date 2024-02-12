// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
)

var (
	TtcpContainerName       = "cros-ddd-filter" // ttcp-demo
	LegacyHWContainerName   = "cros-legacy-hw-filter"
	ProvisionContainerName  = "provision-filter"
	TestFinderContainerName = "cros-test-finder"

	// DefaultKarbonFilterNames defines Default karbon filters (SetDefaultFilters may add/remove)
	DefaultKarbonFilterNames = []string{ProvisionContainerName, TestFinderContainerName}

	// DefaultKoffeeFilterNames defines Default koffee filters (SetDefaultFilters may add/remove)
	DefaultKoffeeFilterNames = []string{}

	// Default shas for backwards compatibility
	defaultLegacyHWSha  = "695ae7d6eabe82ba197c8a5c0db6b4292cd3ec940b3bdfaf85378d5ac3910e2b"
	defaultTTCPSha      = "4c5af419e9ded8b270f9ebfea6cd8686f360c3aefc13c8bfe11ab3ee7d66eeee"
	defaultProvisionSha = "0010028a54f3d72c41a93b0f1248dbe9061f25a488efd8d113f73d05b4052c2f"

	prodShas = map[string]string{
		TtcpContainerName:      defaultTTCPSha,
		LegacyHWContainerName:  defaultLegacyHWSha,
		ProvisionContainerName: defaultProvisionSha}

	binaryLookup = map[string]string{
		LegacyHWContainerName:   "legacy_hw_filter",
		TtcpContainerName:       "solver_service",
		ProvisionContainerName:  "provision-filter",
		TestFinderContainerName: "test_finder_filter",
	}
)

// SetDefaultFilters sets/appends proper default filters
func SetDefaultFilters(ctx context.Context, suiteReq *api.SuiteRequest) {
	suiteName := strings.ToLower(suiteReq.GetTestSuite().GetName())
	if strings.HasPrefix(suiteName, "3d") || strings.HasPrefix(suiteName, "ddd") {
		DefaultKarbonFilterNames = append(DefaultKarbonFilterNames, TtcpContainerName)
	} else {
		DefaultKarbonFilterNames = append(DefaultKarbonFilterNames, LegacyHWContainerName)
	}
}

// GetDefaultFilters constructs ctp filters for provided default filters.
func GetDefaultFilters(ctx context.Context, defaultFilterNames []string, contMetadataMap map[string]*buildapi.ContainerImageInfo, build int) ([]*api.CTPFilter, error) {
	defaultFilters := make([]*api.CTPFilter, 0)

	logging.Infof(ctx, "Inside Default Filters: %s", defaultFilterNames)
	for _, filterName := range defaultFilterNames {
		// Attempt to map the filter from the known container metadata.
		ctpFilter, err := CreateCTPFilterWithContainerName(filterName, contMetadataMap, build)
		if err == nil {
			defaultFilters = append(defaultFilters, ctpFilter)
			continue
		}

		logging.Infof(ctx, "Inside backwards compat check.")
		// Test-Finder must always come from the contMetadataMap. Thus if we do not have the "filter" version,
		// We will setup to run the legacy test-finder.
		if filterName == TestFinderContainerName {
			TFFilter, err := CreateCTPFilterWithContainerName(TestFinderContainerName, contMetadataMap, build)
			if err != nil {
				return nil, errors.Annotate(err, "failed to create test-finder default filter").Err()
			}
			defaultFilters = append(defaultFilters, TFFilter)
			continue
		}

		// Otherwise, build the other default filters off prod containers.
		digest, ok := prodShas[filterName]
		if ok {
			logging.Infof(ctx, "Making default container for: %s", filterName)
			ctpFilter, err = CreateCTPDefaultWithContainerName(filterName, digest, build)
			if err != nil {
				return nil, errors.Annotate(err, "failed to create default default filter: ").Err()
			}
			defaultFilters = append(defaultFilters, ctpFilter)
			continue
		}
		return nil, errors.Annotate(err, "failed to create default filter: ").Err()
	}

	return defaultFilters, nil
}

func CreateCTPDefaultWithContainerName(name string, digest string, build int) (*api.CTPFilter, error) {
	c := &buildapi.ContainerImageInfo{
		Repository: &buildapi.GcrRepository{
			Hostname: "us-docker.pkg.dev",
			Project:  "cros-registry/test-services",
		},
		Name:   name,
		Digest: fmt.Sprintf("sha256:%s", digest),
		Tags:   []string{"prod"},
	}

	binaryName := binaryName(name, build)

	return &api.CTPFilter{ContainerInfo: &api.ContainerInfo{Container: c, BinaryName: binaryName}}, nil

}

// CreateCTPFilterWithContainerName creates ctp filter for provided container name through provided container metadata.
func CreateCTPFilterWithContainerName(name string, contMetadataMap map[string]*buildapi.ContainerImageInfo, build int) (*api.CTPFilter, error) {
	if _, ok := contMetadataMap[name]; !ok {
		return nil, errors.Reason("could not find container image info for %s in provided map", name).Err()
	}
	binaryName := binaryName(name, build)
	return &api.CTPFilter{ContainerInfo: &api.ContainerInfo{Container: contMetadataMap[name], BinaryName: binaryName}}, nil
}

// ConstructCtpFilters constructs default and non-default ctp filters.
func ConstructCtpFilters(ctx context.Context, defaultFilterNames []string, contMetadataMap map[string]*buildapi.ContainerImageInfo, filtersToAdd []*api.CTPFilter, build int) ([]*api.CTPFilter, error) {
	filters := make([]*api.CTPFilter, 0)

	// Add default filters
	logging.Infof(ctx, "Inside ConstructCtpFilters.")

	for _, filter := range filtersToAdd {
		// Only add non-default containers.
		if !slices.Contains(defaultFilterNames, filter.GetContainerInfo().GetContainer().GetName()) {
			defaultFilterNames = append(defaultFilterNames, filter.GetContainerInfo().GetContainer().GetName())
		}
	}

	defFilters, err := GetDefaultFilters(ctx, defaultFilterNames, contMetadataMap, build)
	if err != nil {
		return filters, errors.Annotate(err, "failed to get default filters: ").Err()
	}
	logging.Infof(ctx, "After GetDefaultFilters. %s", defFilters)

	filters = append(filters, defFilters...)

	return filters, nil
}

func binaryName(name string, build int) string {
	if name == TestFinderContainerName && needBackwardsCompatibility(build) {
		return "cros-test-finder"
	}

	binName, ok := binaryLookup[name]
	// If no name is found, then assume the container name is the same as the binary.
	// TODO expose the binary name and connect it from the input request.
	if !ok {
		return name
	}
	return binName
}

// CreateContainerRequest creates container request from provided ctp filter.
func CreateContainerRequest(requestedFilter *api.CTPFilter, build int) *skylab_test_runner.ContainerRequest {
	return &skylab_test_runner.ContainerRequest{
		DynamicIdentifier: requestedFilter.GetContainerInfo().GetContainer().GetName(),
		Container: &api.Template{
			Container: &api.Template_Generic{
				Generic: &api.GenericTemplate{
					// TODO (azrahman): Finalize the format of the this dir. Ideally, it should be /tmp/<container_name>.
					// So keeping it as comment for now.
					//DockerArtifactDir: fmt.Sprintf("/tmp/%s", filter.GetContainer().GetName()),
					DockerArtifactDir: "/tmp/filters",
					BinaryArgs: []string{
						"server", "-port", "0",
					},
					// TODO (azrahman): Get binary name from new field of CTPFilter proto.
					BinaryName: requestedFilter.GetContainerInfo().GetBinaryName(),
				},
			},
		},
		// TODO (azrahman): figure this out (not being used right now).
		ContainerImageKey: requestedFilter.GetContainerInfo().GetContainer().GetName(),
		Network:           "host",
	}
}

func needBackwardsCompatibility(build int) bool {
	return build < 15769
}

// CreateTTCPContainerRequest creates container request from provided ctp filter.
// TODO (azrahman): Merge this into a generic container request creator that will
// work for all containers.
func CreateTTCPContainerRequest(requestedFilter *api.CTPFilter) *skylab_test_runner.ContainerRequest {
	return &skylab_test_runner.ContainerRequest{
		DynamicIdentifier: requestedFilter.GetContainerInfo().GetContainer().GetName(),
		Container: &api.Template{
			Container: &api.Template_Generic{
				Generic: &api.GenericTemplate{
					// TODO (azrahman): Finalize the format of the this dir. Ideally, it should be /tmp/<container_name>.
					// So keeping it as comment for now.
					//DockerArtifactDir: fmt.Sprintf("/tmp/%s", filter.GetContainer().GetName()),
					DockerArtifactDir: "/tmp/filters",
					BinaryArgs: []string{
						"-port", "0",
						"-log", "/tmp/filters",
						"-creds", "/creds/service_accounts/service-account-chromeos.json",
					},
					// TODO (azrahman): Get binary name from new field of CTPFilter proto.
					BinaryName:        "/solver_service",
					AdditionalVolumes: []string{"/creds/service_accounts/:/creds/service_accounts/"},
				},
			},
		},
		// TODO (azrahman): figure this out (not being used right now).
		ContainerImageKey: requestedFilter.GetContainerInfo().GetContainer().GetName(),
		Network:           "host",
	}
}

// ListToJson creates json bytes from provided list.
func ListToJson(list *list.List) []byte {
	retBytes := make([]byte, 0)
	for e := list.Front(); e != nil; e = e.Next() {
		bytes, _ := json.MarshalIndent(e, "", "\t")
		retBytes = append(retBytes, bytes...)
	}

	return retBytes
}
