// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package main is the entrypoint to the fleet cost server.
package main

import (
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"

	"infra/cros/fleetcost/internal/costserver"
	ufspb "infra/unifiedfleet/api/v1/rpc"
)

// main starts the fleet cost server.
func main() {
	mods := []module.Module{
		gaeemulation.NewModuleFromFlags(),
		cfgmodule.NewModuleFromFlags(),
	}

	server.Main(nil, mods, func(srv *server.Server) error {
		fleetCostFrontend := costserver.NewFleetCostFrontend().(*costserver.FleetCostFrontend)
		ufsClient, err := ufspb.NewClient(srv.Context)
		if err != nil {
			return errors.Annotate(err, "setting up UFS client").Err()
		}
		costserver.SetUFSClient(fleetCostFrontend, ufsClient)
		costserver.InstallServices(fleetCostFrontend, srv)
		logging.Infof(srv.Context, "Initialization finished.")
		return nil
	})
}
