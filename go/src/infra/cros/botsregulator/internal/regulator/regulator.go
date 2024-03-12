// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package regulator defines the service main flow.
package regulator

import (
	"context"
	"flag"
	"fmt"
	"net/http"

	"google.golang.org/grpc/metadata"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/grpc/prpc"
	"go.chromium.org/luci/server/auth"

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

	t, err := auth.GetRPCTransport(ctx, auth.AsSelf, auth.WithScopes(auth.CloudOAuthScopes...))
	if err != nil {
		return nil, errors.Annotate(err, "could not create http.RoundTripper").Err()
	}
	// TODO(b/328443703): Handle pagination. Current max value: 1000.
	ic := ufsAPI.NewFleetPRPCClient(&prpc.Client{
		C:    &http.Client{Transport: t},
		Host: r.opts.ufs,
		Options: &prpc.Options{
			UserAgent: "bots-regulator/0.1.0",
		},
	})
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
