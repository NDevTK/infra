// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

// This is the entrypoint for the Karte service in production and dev.
// Control is transferred here, inside the Docker container, when the
// application starts.

import (
	"time"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"

	"infra/cros/karte/internal/frontend"
	"infra/cros/karte/internal/identifiers"
)

// Transfer control to the LUCI server
func main() {
	modules := []module.Module{
		gaeemulation.NewModuleFromFlags(),
		cfgmodule.NewModuleFromFlags(),
		cron.NewModuleFromFlags(),
	}

	options := &server.Options{
		// TODO(gregorynisbet): extract to config file.
		// Allow for long-running cron jobs like those persisting datastore records to BigQuery.
		DefaultRequestTimeout: 10 * time.Minute,
		// TODO(gregorynisbet): extract to config file.
		// Explicitly set our internal timeout to GAE's maximum value.
		InternalRequestTimeout: 10 * time.Minute,
	}

	server.Main(options, modules, func(srv *server.Server) error {
		logging.Infof(srv.Context, "Installing dependencies into context")
		srv.Context = identifiers.Use(srv.Context, identifiers.NewDefault())
		logging.Infof(srv.Context, "Starting server.")
		logging.Infof(srv.Context, "Installing Services.")
		k := frontend.NewKarteFrontend()
		frontend.InstallServices(k, srv.PRPC)
		logging.Infof(srv.Context, "Initialization finished.")
		return nil
	})
}
