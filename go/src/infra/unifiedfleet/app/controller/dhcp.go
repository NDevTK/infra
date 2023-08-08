// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"context"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/configuration"
)

// GetDHCPConfig returns a dhcp reecord based on hostname from datastore.
func GetDHCPConfig(ctx context.Context, hostname string) (*ufspb.DHCPConfig, error) {
	return configuration.GetDHCPConfig(ctx, hostname)
}

// BatchGetDhcpConfigs returns a batch of dhcp records
func BatchGetDhcpConfigs(ctx context.Context, hostnames []string) ([]*ufspb.DHCPConfig, error) {
	return configuration.BatchGetDHCPConfigs(ctx, hostnames)
}
