// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"

	"infra/appengine/drone-queen/internal/config"
	icron "infra/appengine/drone-queen/internal/cron"
	"infra/appengine/drone-queen/internal/frontend"
)

func main() {
	modules := []module.Module{
		gaeemulation.NewModuleFromFlags(),
		cron.NewModuleFromFlags(),
		cfgmodule.NewModuleFromFlags(),
	}
	server.Main(nil, modules, func(srv *server.Server) error {
		icron.InstallHandlers()
		srv.RegisterUnaryServerInterceptors(config.UnaryConfig)
		frontend.RegisterServers(srv)
		return nil
	})
}
