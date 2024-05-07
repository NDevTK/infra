// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package migrator defines the CloudBots migration main flow.
package migrator

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/config/cfgclient"

	"infra/cros/botsregulator/internal/clients"
	"infra/cros/botsregulator/internal/regulator"
	"infra/cros/botsregulator/protos"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// migrationFile is the the name of the CloudBots migration file.
const migrationFile = "migration.cfg"

type migrator struct {
	ufsClient clients.UFSClient
}

func NewMigrator(ctx context.Context, r *regulator.RegulatorOptions) (*migrator, error) {
	logging.Infof(ctx, "creating migrator \n")
	uc, err := clients.NewUFSClient(ctx, r.UFS)
	if err != nil {
		return nil, err
	}
	return &migrator{
		ufsClient: uc,
	}, nil
}

// GetMigrationConfig fetches CloudBots migration file from luci-config.
func (m *migrator) GetMigrationConfig(ctx context.Context) (*protos.Migration, error) {
	logging.Infof(ctx, "fetching migration file: %s \n", migrationFile)
	out := &protos.Migration{}
	err := cfgclient.Get(ctx, "services/${appid}", migrationFile, cfgclient.ProtoText(out), nil)
	if err != nil {
		return nil, errors.Annotate(err, "could not fetch migration file").Err()
	}
	return out, nil
}

// FetchSFOMachines only returns the machines located in sfo36/em25.
func (m *migrator) FetchSFOMachines(ctx context.Context) ([]*ufspb.Machine, error) {
	logging.Infof(ctx, "fetching machines in SFO36")
	ctx = clients.SetUFSNamespace(ctx, "os")
	// TODO(b/328443703): Handle pagination. Current max value: 1000.
	res, err := m.ufsClient.ListMachines(ctx, &ufsAPI.ListMachinesRequest{
		PageSize: 1000,
		Filter:   "zone=ZONE_SFO36_OS",
	})
	if err != nil {
		return nil, err
	}
	return res.GetMachines(), nil
}

// FetchSFOMachineLSEs only returns the machineLSEs located in sfo36/em25.
func (m *migrator) FetchSFOMachineLSEs(ctx context.Context) ([]*ufspb.MachineLSE, error) {
	logging.Infof(ctx, "fetching machineLSEs in SFO36")
	ctx = clients.SetUFSNamespace(ctx, "os")
	// TODO(b/328443703): Handle pagination. Current max value: 1000.
	res, err := m.ufsClient.ListMachineLSEs(ctx, &ufsAPI.ListMachineLSEsRequest{
		PageSize: 1000,
		Filter:   "zone=ZONE_SFO36_OS",
	})
	if err != nil {
		return nil, err
	}
	return res.GetMachineLSEs(), nil
}
