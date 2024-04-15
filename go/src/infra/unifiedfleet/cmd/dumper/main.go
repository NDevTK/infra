// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/profiler"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/limiter"
	"go.chromium.org/luci/server/module"

	"infra/unifiedfleet/app/config"
	"infra/unifiedfleet/app/dumper"
	"infra/unifiedfleet/app/external"
	"infra/unifiedfleet/app/util"
)

func main() {
	// skip the realms check for dumper service
	util.SkipRealmsCheck = true
	modules := []module.Module{
		gaeemulation.NewModuleFromFlags(),
		limiter.NewModuleFromFlags(),
	}

	cfgLoader := config.Loader{}
	cfgLoader.RegisterFlags(flag.CommandLine)

	server.Main(nil, modules, func(srv *server.Server) error {
		// We closely follow the profiler-enabling documentation available at the following URL:
		// https://cloud.google.com/profiler/docs/profiling-go#enabling-profiler-api
		cfg := profiler.Config{
			Service: "ufs-dumper",
			// TODO(gregorynisbet): replace with commit hash or some other smarter way of getting
			//                      the UFS version.
			ServiceVersion:     "1.0.0",
			EnableOCTelemetry:  false,
			ProjectID:          srv.Options.CloudProject,
			DebugLogging:       true,
			DebugLoggingOutput: nil, // stderr by default
		}

		profilerStartErr := profiler.Start(cfg)
		// TODO(gregorynisbet): Upgrade this to a panic once enabling the profiler is reliable enough in prod.
		if profilerStartErr == nil {
			logging.Infof(srv.Context, "profiler started successfully: de3e33ed-b5c9-40c6-b0ea-52a3ca5d6138")
		} else {
			logging.Errorf(srv.Context, "%s\n", errors.Annotate(profilerStartErr, "error encountered when setting up profiler during startup").Err())
		}

		// Load service config form a local file (deployed via GKE),
		// periodically reread it to pick up changes without full restart.
		if _, err := cfgLoader.Load(srv.Context); err != nil {
			return err
		}
		srv.RunInBackground("ufs.config", cfgLoader.ReloadLoop)
		srv.Context = config.Use(srv.Context, cfgLoader.Config())
		srv.Context = external.WithServerInterface(srv.Context)

		client, err := bigquery.NewClient(srv.Context, srv.Options.CloudProject)
		if err != nil {
			return err
		}
		srv.Context = dumper.Use(srv.Context, client)
		srv.Context = dumper.UseProject(srv.Context, srv.Options.CloudProject)
		dumper.InstallCronServices(srv)
		dumper.InitServer(srv)
		return nil
	})
}
