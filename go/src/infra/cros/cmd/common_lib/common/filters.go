// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common

import (
	"container/list"
	"context"
	"encoding/json"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	"go.chromium.org/luci/common/errors"
	"golang.org/x/exp/slices"
)

// Default filters
var DefaultKarbonFilterNames = []string{"provision_filter", "test_finder_filter", "legacy_hw_filter"}
var DefaultKoffeeFilterNames = []string{}

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

// ListToJson creates json bytes from provided list.
func ListToJson(list *list.List) []byte {
	retBytes := make([]byte, 0)
	for e := list.Front(); e != nil; e = e.Next() {
		bytes, _ := json.MarshalIndent(e, "", "\t")
		retBytes = append(retBytes, bytes...)
	}

	return retBytes
}
