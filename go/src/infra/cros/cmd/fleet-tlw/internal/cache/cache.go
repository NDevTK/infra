// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cache

import (
	"fmt"
	"hash/fnv"
	"net"
)

// Environment is the runtime dependencies, e.g. networking, etc. of the
// implementation. The main goal of it is for unit test.
type Environment interface {
	// Subnets returns the caching subnets.
	// The slice returned may be shared, so do not modify it.
	// This function is concurrency safe.
	Subnets() []Subnet
}

// Subnet is a network in labs (i.e. test VLAN).
// DUTs can only access caching backends in the same Subnet.
type Subnet struct {
	IPNet    *net.IPNet
	Backends []string
}

// Frontend manages caching backends and assigns backends for client requests.
type Frontend struct {
	env Environment
}

// NewFrontend creates a new cache frontend.
func NewFrontend(env Environment) *Frontend {
	return &Frontend{env: env}
}

// AssignBackend assigns a healthy backend to the request from `dutName` on
// `filename`.
// This function is concurrency safe.
func (f *Frontend) AssignBackend(dutName, filename string) (string, error) {
	dutAddr, err := lookupHost(dutName)
	if err != nil {
		return "", fmt.Errorf("assign backend to %q: %s", dutName, err)
	}
	// Get cache backends serving the DUT subnet.
	subnet, ok := f.findSubnet(net.ParseIP(dutAddr))
	if !ok {
		return "", fmt.Errorf("%q is not in any cache subnets (all subnets: %v)", dutAddr, f.env.Subnets())
	}
	// Get a cache backend according to the hash value of 'filename'.
	return findOneBackend(filename, subnet.Backends), nil
}

func (f *Frontend) findSubnet(ip net.IP) (*Subnet, bool) {
	for i := range f.env.Subnets() {
		if f.env.Subnets()[i].IPNet.Contains(ip) {
			return &f.env.Subnets()[i], true
		}
	}
	return nil, false
}

// findOneBackend finds one healthy backend from the current subnet according to
// the requested `filename` using 'mod N' algorithm.
func findOneBackend(filename string, backends []string) string {
	return backends[hash(filename)%len(backends)]
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
