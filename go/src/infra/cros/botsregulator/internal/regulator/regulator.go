// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package regulator defines the service main flow.
package regulator

import (
	"context"
	"flag"
	"fmt"

	"google.golang.org/grpc/metadata"

	"go.chromium.org/luci/common/errors"

	"infra/cros/botsregulator/internal/provider"
	"infra/cros/botsregulator/internal/util"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

type regulator struct {
	opts *RegulatorOptions
}

func NewRegulator(opts *RegulatorOptions) regulator {
	return regulator{
		opts: opts,
	}
}

// FetchDUTsByHive fetches the available DUTs from UFS by hive
// and returns a slice of hostname.
func (r *regulator) FetchDUTsByHive(ctx context.Context) ([]*ufspb.MachineLSE, error) {
	md := metadata.Pairs("namespace", r.opts.namespace)
	ctx = metadata.NewOutgoingContext(ctx, md)
	pc, err := util.RawPRPCClient(ctx, r.opts.ufs)
	if err != nil {
		return nil, err
	}
	ic := ufsAPI.NewFleetPRPCClient(pc)
	// TODO(b/328443703): Handle pagination. Current max value: 1000.
	res, err := ic.ListMachineLSEs(ctx, &ufsAPI.ListMachineLSEsRequest{
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

// RegulatorOptions refers to the flag options needed
// to create a new regulator struct.
type RegulatorOptions struct {
	bpi       string
	cfID      string
	hive      string
	namespace string
	ufs       string
}

// RegisterFlags exposes the command line flags required to run the application.
// We never check for flag emptiness so all options must have defaults.
func (r *RegulatorOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&r.bpi, "bpi", util.GCEPDev, "URI endpoint of the service used to scale bots.")
	fs.StringVar(&r.cfID, "config", util.ConfigID, "CloudBots config prefix.")
	fs.StringVar(&r.hive, "hive", "cloudbots", "hive used for UFS filtering.")
	fs.StringVar(&r.namespace, "ufs-namespace", ufsUtil.OSNamespace, "UFS namespace.")
	fs.StringVar(&r.ufs, "ufs", util.UFSDev, "UFS endpoint.")
}
