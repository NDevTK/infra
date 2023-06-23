// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package cache

import (
	"fmt"
	"sort"
	"strings"

	ufsmodels "infra/unifiedfleet/api/v1/models"
)

// NewPreferredEnv creates a new preferred caching service environment.
// In this environment, we skip further server selection based on either UFS
// zone or subnets.
func NewPreferredEnv(services string) (Environment, error) {
	svc := parseCSVAndSort(services)
	if len(svc) == 0 {
		return nil, fmt.Errorf("new preferred caching service environment: no preferred service specified")
	}
	var ss []CachingService
	for _, s := range svc {
		ss = append(ss, CachingService(s))
	}
	return &preferedEnv{services: ss}, nil
}

type preferedEnv struct {
	services []CachingService
}

func (p *preferedEnv) Subnets() []Subnet {
	return nil
}

func (p *preferedEnv) CacheZones() map[ufsmodels.Zone][]CachingService {
	return map[ufsmodels.Zone][]CachingService{ufsmodels.Zone_ZONE_UNSPECIFIED: p.services}
}

func (p *preferedEnv) GetZoneForServer(string) (ufsmodels.Zone, error) {
	return ufsmodels.Zone_ZONE_UNSPECIFIED, nil
}

func (p *preferedEnv) GetZoneForDUT(string) (ufsmodels.Zone, error) {
	return ufsmodels.Zone_ZONE_UNSPECIFIED, nil
}

// parseCSVAndSort parse input comma separated string into a string slice and
// sort it.
func parseCSVAndSort(value string) []string {
	if value == "" {
		return nil
	}
	ss := strings.Split(value, ",")
	sort.Strings(ss)
	return ss
}
