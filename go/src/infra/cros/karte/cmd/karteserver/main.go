// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

// This is the entrypoint for the Karte service in production and dev.
// Control is transferred here, inside the Docker container, when the
// application starts.

import (
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"

	"infra/cros/karte/internal/frontend"
	"infra/cros/karte/internal/idstrategy"
)

// Transfer control to the LUCI server
func main() {
	modules := []module.Module{
		gaeemulation.NewModuleFromFlags(),
		cfgmodule.NewModuleFromFlags(),
		cron.NewModuleFromFlags(),
	}

	server.Main(nil, modules, func(srv *server.Server) error {
		logging.Infof(srv.Context, "Installing dependencies into context")
		srv.Context = idstrategy.Use(srv.Context, idstrategy.NewDefault())
		logging.Infof(srv.Context, "Starting server.")
		logging.Infof(srv.Context, "Installing Services.")
		k := frontend.NewKarteFrontend()
		frontend.InstallServices(k, srv.PRPC)
		logging.Infof(srv.Context, "Initialization finished.")
		return nil
	})
}
