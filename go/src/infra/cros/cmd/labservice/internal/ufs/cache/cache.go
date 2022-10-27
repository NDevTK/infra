// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package cache provides functionality to map DUTs to caching servers.
package cache

import (
	"fmt"
	"hash/fnv"
	"log"
	"net"

	ufsapi "infra/unifiedfleet/api/v1/rpc"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewLocator() *Locator {
	return &Locator{
		subnets: newSubnetsFinder(),
	}
}

// Locator helps to find a caching server for any given DUT.
// It caches ip addresses and corresponding subnets of caching servers.
type Locator struct {
	subnets *subnetsFinder
}

// FindCacheServer returns the ip address of a cache server mapped to a dut.
func (l *Locator) FindCacheServer(dutName string, client ufsapi.FleetClient) (*labapi.IpEndpoint, error) {
	cs, err := l.findCacheServerBySubnet(dutName, client)
	if err != nil {
		return nil, fmt.Errorf("find cache server for %q: %s", dutName, err)
	}
	return cs, nil
}

func (l *Locator) findCacheServerBySubnet(dutName string, client ufsapi.FleetClient) (*labapi.IpEndpoint, error) {
	subnets, err := l.subnets.getSubnets(client)
	if err != nil {
		return nil, fmt.Errorf("find cache server by subnet: %s", err)
	}

	sn, err := findSubnet(dutName, subnets)
	if err != nil {
		return nil, fmt.Errorf("find cache server by subnet: %s", err)
	}

	be := chooseBackend(sn.Backends, dutName)
	return &labapi.IpEndpoint{
		Address: be.Ip,
		Port:    be.Port,
	}, nil
}

func findSubnet(dutName string, subnets []Subnet) (Subnet, error) {
	addr, err := lookupHost(dutName)
	if err != nil {
		return Subnet{}, status.Errorf(codes.NotFound, fmt.Sprintf("FindCacheServer: lookup IP of %q: %s", dutName, err.Error()))
	}
	log.Printf("FindCacheServer: the IP of %q is %q", dutName, addr)

	ip := net.ParseIP(addr)
	for _, s := range subnets {
		if s.IPNet.Contains(ip) {
			return s, nil
		}
	}
	return Subnet{}, fmt.Errorf("%q is not in any cache subnets (all subnets: %v)", addr, subnets)
}

// chooseBackend finds one healthy backend from given backends according to
// the requested `hostname` using 'mod N' algorithm.
func chooseBackend(backends []address, hostname string) address {
	return backends[hash(hostname)%len(backends)]
}

// hash returns integer hash value of the input string.
// We use the hash value to map to a backend according to a specified algorithm.
// We choose FNV hashing because we concern more on computation speed, not for
// cryptography.
func hash(s string) int {
	h := fnv.New32a()
	h.Write([]byte(s))
	return int(h.Sum32())
}
