// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executor

import (
	"fmt"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
)

func DefaultFilters(req *api.CTPRequest2, containerMd map[string]*buildapi.ContainerImageInfo) ([]*api.Filter, []*api.Filter, error) {
	karbons := []*api.Filter{}
	koffees := []*api.Filter{}

	// Can expand these as needed
	defaultKarbonServices := make(map[string]bool)
	defaultKarbonServices["cros-test-finder"] = true

	// Currently none for default Koffee. I am sure some will come.
	defaultKoffeeServices := make(map[string]bool)

	for _, filter := range req.KarbonFilters {
		_, ok := defaultKarbonServices[filter.Container.ServiceName]
		if ok {
			defaultKarbonServices[filter.Container.ServiceName] = false
		}
	}

	for _, filter := range req.KoffeeFilters {
		_, ok := defaultKoffeeServices[filter.Container.ServiceName]
		if ok {
			defaultKoffeeServices[filter.Container.ServiceName] = false
		}
	}

	for k, v := range defaultKarbonServices {
		if v {
			container, err := ResolvedContainer(k, containerMd)
			if err != nil {
				return nil, nil, err
			}
			karbons = append(karbons, container)
		}
	}

	for k, v := range defaultKoffeeServices {
		if v {
			container, err := ResolvedContainer(k, containerMd)
			if err != nil {
				return nil, nil, err
			}
			koffees = append(koffees, container)
		}
	}
	return karbons, koffees, nil
}

func ResolvedContainer(name string, containerMd map[string]*buildapi.ContainerImageInfo) (*api.Filter, error) {
	image, ok := containerMd[name]
	if !ok {
		return nil, fmt.Errorf("Container image not found: %s", name)
	}
	return &api.Filter{
		Container: &api.ContainerInfo{
			ServiceName:   name,
			ContainerPath: image.Digest,
		},
	}, nil
}
