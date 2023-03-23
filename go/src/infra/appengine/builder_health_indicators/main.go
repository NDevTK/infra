// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/cron"
	"go.chromium.org/luci/server/module"
	"infra/appengine/builder_health_indicators/internal/generate"
)

func main() {
	modules := []module.Module{
		cron.NewModuleFromFlags(),
	}
	server.Main(nil, modules, func(srv *server.Server) error {
		cron.RegisterHandler("generate", generate.Generate)
		return nil
	})
}
