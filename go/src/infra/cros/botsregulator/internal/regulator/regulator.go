// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package regulator defines the service main flow.
package regulator

import (
	"context"
	"fmt"

	"go.chromium.org/luci/common/errors"

	"infra/cros/botsregulator/internal/clients"
	"infra/cros/botsregulator/internal/provider"
	"infra/cros/botsregulator/internal/util"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

type regulator struct {
	opts      *RegulatorOptions
	ufsClient clients.UFSClient
}

func NewRegulator(ctx context.Context, opts *RegulatorOptions) (*regulator, error) {
	uc, err := clients.NewUFSClient(ctx, opts.ufs, opts.namespace)
	if err != nil {
		return nil, err
	}
	return &regulator{
		opts:      opts,
		ufsClient: uc,
	}, nil
}

// FetchDUTsByHive fetches the available DUTs from UFS by hive
// and returns a slice of hostname.
func (r *regulator) FetchDUTsByHive(ctx context.Context) ([]*ufspb.MachineLSE, error) {
	// TODO(b/328443703): Handle pagination. Current max value: 1000.
	res, err := r.ufsClient.ListMachineLSEs(ctx, &ufsAPI.ListMachineLSEsRequest{
		Filter: fmt.Sprintf("hive=%s", r.opts.hive),
		// KeysOnly returns the entities' ID only. It is faster than a full query.
		KeysOnly: true,
	})
	if err != nil {
		return nil, errors.Annotate(err, "could not list machinesLSEs").Err()
	}
	lses := res.GetMachineLSEs()
	return lses, nil
}

// UpdateConfig creates a provider.BPI based on the server running environment.
// The provider is responsible for the actual implementation in the provider package.
// Providers currently supported are GCE Provider and Satlab(WIP).
func (r *regulator) UpdateConfig(ctx context.Context, hns []string) error {
	var bc provider.BPI
	var err error
	switch util.GetEnv() {
	case util.GCP:
		bc, err = provider.NewGCEPClient(ctx, r.opts.bpi, r.opts.cfID)
	case util.Satlab:
		err = errors.New("Satlab flow not implemented")
	default:
		panic("unrecognized running environment")
	}
	if err != nil {
		return err
	}
	err = bc.UpdateConfig(ctx, hns)
	if err != nil {
		return err
	}
	return nil
}
