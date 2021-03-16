// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cache

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"

	"google.golang.org/grpc/metadata"

	ufsmodels "infra/unifiedfleet/api/v1/models"
	ufsapi "infra/unifiedfleet/api/v1/rpc"
)

type sshProxyServerEnv struct {
	subnets []Subnet
}

// NewSSHProxyServerEnv creates an instance of Environment for getting SSH Proxy
// server information.
func NewSSHProxyServerEnv() (Environment, error) {
	ctx := context.Background()
	client, err := ufsapi.NewClient(
		ctx,
		ufsapi.ServiceName("staging.ufs.api.cr.dev"),
		ufsapi.ServiceAccountJSONPath("/creds/service_accounts/skylab-drone.json"),
		ufsapi.UserAgent("fleet-tlw/3.0.0"),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting UFS client: %s", err)
	}
	servers, err := fetchSSHProxyServerList(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("error getting SSH proxy servers: %s", err)
	}
	s, err := fetchSubnets(servers)
	if err != nil {
		return nil, fmt.Errorf("error getting SSH proxy servers' subnets: %s", err)
	}
	return sshProxyServerEnv{subnets: s}, nil
}

func (e sshProxyServerEnv) Subnets() []Subnet {
	// Make a copy of 'e.subnets' to prevent being modified by a caller.
	s := make([]Subnet, len(e.subnets))
	copy(s, e.subnets)
	return s
}

func (e sshProxyServerEnv) IsBackendHealthy(s string) bool {
	// Assume all backends are healthy.
	return true
}

func fetchSubnets(servers []*ufsmodels.CachingService) ([]Subnet, error) {
	info := make(map[string][]string)
	for _, s := range servers {
		if state := s.GetState(); state != ufsmodels.State_STATE_SERVING {
			continue
		}
		servingSubnet := s.GetServingSubnet()
		// Remove prefix "cachingservice/" from the Name.
		splitName := strings.Split(s.GetName(), "/")
		serverIP := splitName[len(splitName)-1]
		info[servingSubnet] = append(info[servingSubnet], serverIP)
	}
	// Register all subnets.
	var ss []Subnet
	for subnet, IPs := range info {
		_, ipNet, err := net.ParseCIDR(subnet)
		if err != nil {
			return nil, err
		}
		sort.Strings(IPs)
		ss = append(ss, Subnet{IPNet: ipNet, Backends: IPs})
	}
	return ss, nil
}

func fetchSSHProxyServerList(ctx context.Context, c ufsapi.FleetClient) ([]*ufsmodels.CachingService, error) {
	md := metadata.Pairs("namespace", "os")
	ctx = metadata.NewOutgoingContext(ctx, md)
	resp, err := c.ListCachingServices(ctx, &ufsapi.ListCachingServicesRequest{})
	if err != nil {
		return nil, err
	}
	return resp.GetCachingServices(), nil
}
