// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package main is the entrypoint to BotsRegulator.
package main

import (
	"context"
	"flag"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server"
	scron "go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/module"

	"infra/cros/botsregulator/internal/cron"
	"infra/cros/botsregulator/internal/regulator"
)

func main() {
	mods := []module.Module{
		scron.NewModuleFromFlags(),
	}

	r := regulator.RegulatorOptions{}
	r.RegisterFlags(flag.CommandLine)

	server.Main(nil, mods, func(srv *server.Server) error {
		logging.Infof(srv.Context, "starting server")

		scron.RegisterHandler("regulate-bots", func(ctx context.Context) error {
			return cron.Regulate(ctx, &r)
		})
		scron.RegisterHandler("migrate-bots", func(ctx context.Context) error {
			return cron.Migrate(ctx, &r)
		})
		return nil
	})
}
