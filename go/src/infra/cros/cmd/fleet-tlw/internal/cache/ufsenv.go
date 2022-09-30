// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cache

import (
	"context"
	"fmt"
	"log"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"

	ufsmodels "infra/unifiedfleet/api/v1/models"
	ufsapi "infra/unifiedfleet/api/v1/rpc"
)

const refreshInterval = time.Hour

// NewUFSEnv creates an instance of Environment for caching services registered
// in UFS.
// It caches the result to prevent frequent access to UFS. It updates the cache
// regularly.
func NewUFSEnv(c ufsapi.FleetClient) (Environment, error) {
	e := &ufsEnv{client: c}
	if err := e.refreshSubnets(); err != nil {
		return nil, fmt.Errorf("NewUFSEnv: %s", err)
	}
	return e, nil
}

type ufsEnv struct {
	client   ufsapi.FleetClient
	expireMu sync.Mutex
	expire   time.Time
	subnets  []Subnet
}

func (e *ufsEnv) Subnets() []Subnet {
	if err := e.refreshSubnets(); err != nil {
		log.Printf("UFSEnv: fallback to cached data due to refresh failure: %s", err)
	}
	return e.subnets
}

func (e *ufsEnv) refreshSubnets() error {
	n := time.Now()
	e.expireMu.Lock()
	defer e.expireMu.Unlock()
	if n.Before(e.expire) {
		return nil
	}
	e.expire = n.Add(refreshInterval)

	s, err := e.fetchCachingSubnets()
	if err != nil {
		return fmt.Errorf("refresh subnets: %s", err)
	}
	e.subnets = s
	return nil
}

// fetchCachingSubnets fetches caching services info from UFS and constructs
// caching subnets.
func (e *ufsEnv) fetchCachingSubnets() ([]Subnet, error) {
	log.Printf("Fetching caching subnets from UFS")
	cachingServices, err := fetchCachingServicesFromUFS(e.client)
	if err != nil {
		return nil, fmt.Errorf("fetch caching subnets: %s", err)
	}
	log.Printf("Fetching caching subnets from UFS done")

	var result []Subnet
	m := make(map[string][]string)
	for _, s := range cachingServices {
		if state := s.GetState(); state != ufsmodels.State_STATE_SERVING {
			continue
		}
		svc, err := cachingServiceName(s)
		if err != nil {
			return nil, err
		}
		subnets := s.GetServingSubnets()
		for _, s := range subnets {
			m[s] = append(m[s], svc)
		}
	}
	for k, v := range m {
		_, ipNet, err := net.ParseCIDR(k)
		if err != nil {
			return nil, fmt.Errorf("fetch caching subnets: parse subnet %q: %s", k, err)
		}
		sort.Strings(v)
		result = append(result, Subnet{IPNet: ipNet, Backends: v})
		log.Printf("Caching subnet: %q: %#v", k, v)
	}
	return result, nil
}

// cachingServiceName returns a caching service name in format of
// 'http://host:port'.
func cachingServiceName(s *ufsmodels.CachingService) (string, error) {
	// The name ufsmodels.CachingService has a prefix of "cachingservice/".
	nameParts := strings.Split(s.GetName(), "/")
	if len(nameParts) != 2 {
		return "", fmt.Errorf("caching service name: %q isn't in format of 'cachingservice/<name>'", s.GetName())
	}
	port := strconv.Itoa(int(s.GetPort()))
	ip, err := lookupHost(nameParts[1])
	if err != nil {
		return "", fmt.Errorf("caching service name: %s", err)
	}
	return fmt.Sprintf("http://%s", net.JoinHostPort(ip, port)), nil
}

func fetchCachingServicesFromUFS(c ufsapi.FleetClient) ([]*ufsmodels.CachingService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	md := metadata.Pairs("namespace", "os")
	ctx = metadata.NewOutgoingContext(ctx, md)

	resp, err := c.ListCachingServices(ctx, &ufsapi.ListCachingServicesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list caching service from UFS: %s", err)
	}
	return resp.GetCachingServices(), nil
}

// lookupHost looks up the IP address of the provided host by using the local
// resolver.
func lookupHost(hostname string) (string, error) {
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return "", fmt.Errorf("look up IP of %q: %s", hostname, err)
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("look up IP of %q: No addresses found", hostname)
	}
	return addrs[0], nil
}
