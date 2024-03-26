// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package regulator defines the service main flow.
package regulator

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"

	"go.chromium.org/luci/common/errors"

	"infra/cros/botsregulator/internal/clients"
	"infra/cros/botsregulator/internal/provider"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

type regulator struct {
	bpiClient provider.BPI
	ufsClient clients.UFSClient
	opts      *RegulatorOptions
}

func NewRegulator(ctx context.Context, opts *RegulatorOptions) (*regulator, error) {
	fmt.Printf("opts: %v\n", opts)
	uc, err := clients.NewUFSClient(ctx, opts.ufs, opts.namespace)
	if err != nil {
		return nil, err
	}
	bc, err := provider.NewProviderFromEnv(ctx, opts.bpi)
	if err != nil {
		return nil, err
	}
	return &regulator{
		bpiClient: bc,
		ufsClient: uc,
		opts:      opts,
	}, nil
}

// FetchDUTsByHive fetches the available DUTs from UFS by hive
// and returns a slice of hostname.
func (r *regulator) FetchDUTsByHive(ctx context.Context) ([]*ufspb.MachineLSE, error) {
	md := metadata.Pairs("namespace", r.opts.namespace)
	ctx = metadata.NewOutgoingContext(ctx, md)
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

// UpdateConfig is a wrapper around the current provider UpdateConfig method.
func (r *regulator) UpdateConfig(ctx context.Context, hns []string) error {
	return r.bpiClient.UpdateConfig(ctx, hns, r.opts.cfID)
}
