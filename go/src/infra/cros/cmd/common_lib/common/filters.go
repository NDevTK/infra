// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"container/list"
	"context"
	"encoding/json"
	"strings"

	"golang.org/x/exp/slices"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
)

// DefaultKarbonFilterNames defines Default karbon filters (SetDefaultFilters may add/remove)
var DefaultKarbonFilterNames = []string{"provision-filter", "cros-test-finder"}

// DefaultKoffeeFilterNames defines Default koffee filters (SetDefaultFilters may add/remove)
var DefaultKoffeeFilterNames = []string{}

// SetDefaultFilters sets/appends proper default filters
func SetDefaultFilters(ctx context.Context, suiteReq *api.SuiteRequest) {
	suiteName := strings.ToLower(suiteReq.GetTestSuite().GetName())
	if strings.HasPrefix(suiteName, "3d") || strings.HasPrefix(suiteName, "ddd") {
		DefaultKarbonFilterNames = append(DefaultKarbonFilterNames, "ttcp-demo")
	} else {
		DefaultKarbonFilterNames = append(DefaultKarbonFilterNames, "legacy_hw_filter")
	}
}

// GetDefaultFilters constructs ctp filters for provided default filters.
func GetDefaultFilters(ctx context.Context, defaultFilterNames []string, contMetadataMap map[string]*buildapi.ContainerImageInfo) ([]*api.CTPFilter, error) {
	defaultFilters := make([]*api.CTPFilter, 0)

	for _, filterName := range defaultFilterNames {
		ctpFilter, err := CreateCTPFilterWithContainerName(filterName, contMetadataMap)
		if err != nil {
			return nil, errors.Annotate(err, "failed to create default filter: ").Err()
		}
		defaultFilters = append(defaultFilters, ctpFilter)
	}

	return defaultFilters, nil
}

// GetNonDefaultFilters constructs ctp filters for provided non-default filters.
func GetNonDefaultFilters(ctx context.Context, defaultFilterNames []string, contMetadataMap map[string]*buildapi.ContainerImageInfo, filtersToAdd []*api.CTPFilter) ([]*api.CTPFilter, error) {
	nonDefaultFilters := make([]*api.CTPFilter, 0)

	for _, filter := range filtersToAdd {
		if slices.Contains(defaultFilterNames, filter.GetContainer().GetName()) {
			// skip default filters as they should be constructed separately.
			continue
		}
		ctpFilter, err := CreateCTPFilterWithContainerName(filter.GetContainer().GetName(), contMetadataMap)
		if err != nil {
			return nil, errors.Annotate(err, "failed to create non default filter: ").Err()
		}
		nonDefaultFilters = append(nonDefaultFilters, ctpFilter)
	}

	return nonDefaultFilters, nil
}

// CreateCTPFilterWithContainerName creates ctp filter for provided container name through provided container metadata.
func CreateCTPFilterWithContainerName(name string, contMetadataMap map[string]*buildapi.ContainerImageInfo) (*api.CTPFilter, error) {
	if _, ok := contMetadataMap[name]; !ok {
		return nil, errors.Reason("could not find container image info for %s in provided map", name).Err()
	}

	return &api.CTPFilter{Container: contMetadataMap[name]}, nil
}

// ConstructCtpFilters constructs default and non-default ctp filters.
func ConstructCtpFilters(ctx context.Context, defaultFilterNames []string, contMetadataMap map[string]*buildapi.ContainerImageInfo, filtersToAdd []*api.CTPFilter) ([]*api.CTPFilter, error) {
	filters := make([]*api.CTPFilter, 0)

	// Add default filters
	defFilters, err := GetDefaultFilters(ctx, defaultFilterNames, contMetadataMap)
	if err != nil {
		return filters, errors.Annotate(err, "failed to get default filters: ").Err()
	}

	filters = append(filters, defFilters...)

	// Add non-default  filters
	nonDefFilters, err := GetNonDefaultFilters(ctx, defaultFilterNames, contMetadataMap, filtersToAdd)
	if err != nil {
		return filters, errors.Annotate(err, "failed to get non-default filters: ").Err()
	}

	filters = append(filters, nonDefFilters...)

	return filters, nil
}

// CreateContainerRequest creates container request from provided ctp filter.
func CreateContainerRequest(requestedFilter *api.CTPFilter) *skylab_test_runner.ContainerRequest {
	return &skylab_test_runner.ContainerRequest{
		DynamicIdentifier: requestedFilter.GetContainer().GetName(),
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
					BinaryName: requestedFilter.GetContainer().GetName(),
				},
			},
		},
		// TODO (azrahman): figure this out (not being used right now).
		ContainerImageKey: requestedFilter.GetContainer().GetName(),
		Network:           "host",
	}
}

// CreateTFContainerRequest creates container request from provided ctp filter.
func CreateTFContainerRequest(requestedFilter *api.CTPFilter) *skylab_test_runner.ContainerRequest {
	return &skylab_test_runner.ContainerRequest{
		DynamicIdentifier: requestedFilter.GetContainer().GetName(),
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
					BinaryName: "test_finder_filter",
				},
			},
		},
		// TODO (azrahman): figure this out (not being used right now).
		ContainerImageKey: requestedFilter.GetContainer().GetName(),
		Network:           "host",
	}
}

// CreateTTCPContainerRequest creates container request from provided ctp filter.
// TODO (azrahman): Merge this into a generic container request creator that will
// work for all containers.
func CreateTTCPContainerRequest(requestedFilter *api.CTPFilter) *skylab_test_runner.ContainerRequest {
	return &skylab_test_runner.ContainerRequest{
		DynamicIdentifier: requestedFilter.GetContainer().GetName(),
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
		ContainerImageKey: requestedFilter.GetContainer().GetName(),
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
