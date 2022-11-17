// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"go.chromium.org/luci/config/server/cfgmodule"
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/gaeemulation"
	"go.chromium.org/luci/server/module"

	icron "infra/appengine/drone-queen/internal/cron"
	"infra/appengine/drone-queen/internal/frontend"
	"infra/appengine/drone-queen/internal/middleware"
)

func main() {
	modules := []module.Module{
		gaeemulation.NewModuleFromFlags(),
		cron.NewModuleFromFlags(),
		cfgmodule.NewModuleFromFlags(),
	}
	server.Main(nil, modules, func(srv *server.Server) error {
		srv.RegisterUnaryServerInterceptor(middleware.UnaryTrace)
		icron.InstallHandlers(srv)
		frontend.InstallHandlers(srv)
		return nil
	})
}
