// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"flag"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/server"

	"infra/vm_leaser/internal/cron"
)

func main() {
	gcpProject := flag.String(
		"gcp-project",
		"chrome-fleet-vm-leaser-dev",
		"The GCP project where VMs are located.",
	)
	server.Main(nil, nil, func(srv *server.Server) error {
		logging.Infof(srv.Context, "Registering cron server.")
		cron.RegisterCronServer(srv, *gcpProject)
		return nil
	})
}
