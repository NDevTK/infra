// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executor

import (
	"fmt"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
)

func DefaultFilters(req *api.CTPv2Request, containerMd map[string]*buildapi.ContainerImageInfo) ([]*api.CTPFilter, []*api.CTPFilter, error) {
	karbons := []*api.CTPFilter{}
	koffees := []*api.CTPFilter{}

	// Can expand these as needed
	defaultKarbonServices := make(map[string]bool)

	// TODO and note: We will use only the first given GCSPath for test-finder.
	// This means autotest will *NOT* WAI, and this must be resolved on either this side
	// or the autotest side before putting prod traffic in.
	defaultKarbonServices["cros-test-finder"] = true

	// Currently none for default Koffee. I am sure some will come.
	defaultKoffeeServices := make(map[string]bool)

	for _, filter := range req.KarbonFilters {
		_, ok := defaultKarbonServices[filter.Container.Name]
		if ok {
			defaultKarbonServices[filter.Container.Name] = false
		}
	}

	for _, filter := range req.KoffeeFilters {
		_, ok := defaultKoffeeServices[filter.Container.Name]
		if ok {
			defaultKoffeeServices[filter.Container.Name] = false
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

func ResolvedContainer(name string, containerMd map[string]*buildapi.ContainerImageInfo) (*api.CTPFilter, error) {
	image, ok := containerMd[name]
	if !ok {
		return nil, fmt.Errorf("Container image not found: %s", name)
	}
	return &api.CTPFilter{
		Container: &buildapi.ContainerImageInfo{
			Name:   name,
			Digest: image.Digest,
		},
	}, nil
}
