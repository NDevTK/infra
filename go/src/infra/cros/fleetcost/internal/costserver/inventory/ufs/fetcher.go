// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ufs

import (
	"context"

	"google.golang.org/protobuf/protoadapt"

	// UFS and shivas still uses deprecated "github.com/golang/protobuf/proto" package, hence we use the apater here.
	"go.chromium.org/luci/common/errors"

	// TODO, move shared util to a standalone directory.
	shivasUtil "infra/cmd/shivas/utils"
	ufspb "infra/unifiedfleet/api/v1/models"
	ufsAPI "infra/unifiedfleet/api/v1/rpc"
	ufsUtil "infra/unifiedfleet/app/util"
)

// GetAllChromeosDuts gets all DUTs(MachineLSE) in chromeos namespace.
func GetAllChromeosDuts(ctx context.Context, ic ufsAPI.FleetClient) ([]*ufspb.MachineLSE, error) {
	ctx = shivasUtil.SetupContext(ctx, ufsUtil.OSNamespace)
	res, err := shivasUtil.BatchList(ctx, ic, listMachineLSEs, []string{}, 0, false, false)
	if err != nil {
		return nil, errors.Annotate(err, "get all chromeos duts").Err()
	}
	lses := make([]*ufspb.MachineLSE, len(res))
	for i, r := range res {
		lses[i] = r.(*ufspb.MachineLSE)
	}
	return lses, nil
}

// GetChromeosDut gets a single DUT(MachineLSE) based on hostname in chromeos namespace.
func GetChromeosDut(ctx context.Context, ic ufsAPI.FleetClient, name string) (*ufspb.MachineLSE, error) {
	ctx = shivasUtil.SetupContext(ctx, ufsUtil.OSNamespace)
	res, err := shivasUtil.GetSingleMachineLSE(ctx, ic, name)
	if err != nil {
		return nil, errors.Annotate(err, "get chromeos dut").Err()
	}
	return res.(*ufspb.MachineLSE), nil
}

// GetLabstationDutMapping gets a map of labstation to dut hostnames based on provided labstation hostnames.
func GetLabstationDutMapping(ctx context.Context, ic ufsAPI.FleetClient, labs []string) (map[string][]string, error) {
	ctx = shivasUtil.SetupContext(ctx, ufsUtil.OSNamespace)
	res, err := ic.GetDUTsForLabstation(ctx, &ufsAPI.GetDUTsForLabstationRequest{Hostname: labs})
	if err != nil {
		return nil, errors.Annotate(err, "get labstation dut mapping").Err()
	}
	labMap := make(map[string][]string)
	for _, item := range res.GetItems() {
		labMap[item.GetHostname()] = item.GetDutName()
	}
	return labMap, nil
}

// listMachineLSEs calls the list MachineLSE in UFS to get a list of MachineLSEs
func listMachineLSEs(ctx context.Context, ic ufsAPI.FleetClient, pageSize int32, pageToken, filter string, keysOnly, full bool) ([]protoadapt.MessageV1, string, error) {
	req := &ufsAPI.ListMachineLSEsRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
		Filter:    filter,
		KeysOnly:  keysOnly,
		Full:      full,
	}
	res, err := ic.ListMachineLSEs(ctx, req)
	if err != nil {
		return nil, "", err
	}
	protos := make([]protoadapt.MessageV1, len(res.GetMachineLSEs()))
	for i, kvm := range res.GetMachineLSEs() {
		protos[i] = kvm
	}
	return protos, res.GetNextPageToken(), nil
}
