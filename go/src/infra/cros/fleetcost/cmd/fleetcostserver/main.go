// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package main is the entrypoint to the fleet cost server.
package main

import (
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"

	"infra/cros/fleetcost/internal/costserver"
)

// main starts the fleet cost server.
func main() {
	mods := []module.Module{
		gaeemulation.NewModuleFromFlags(),
		cfgmodule.NewModuleFromFlags(),
	}

	server.Main(nil, mods, func(srv *server.Server) error {
		fleetCostFrontend := costserver.NewFleetCostFrontend()
		costserver.InstallServices(fleetCostFrontend, srv)
		logging.Infof(srv.Context, "Initialization finished.")
		return nil
	})
}
