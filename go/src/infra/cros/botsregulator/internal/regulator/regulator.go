// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package regulator defines the service main flow.
package regulator

import (
	"flag"

	"infra/cros/botsregulator/internal/util"
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
func (r *RegulatorOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&r.bpi, "bpi", util.GCEPDev, "URI endpoint of the service used to scale bots.")
	fs.StringVar(&r.cfID, "config", util.ConfigID, "CloudBots config prefix.")
	fs.StringVar(&r.hive, "hive", "cloudbots", "hive used for UFS filtering.")
	fs.StringVar(&r.namespace, "ufs-namespace", ufsUtil.OSNamespace, "UFS namespace.")
	fs.StringVar(&r.ufs, "ufs", util.UFSDev, "UFS endpoint.")
}
