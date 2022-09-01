// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"go.chromium.org/luci/server"
	_ "go.chromium.org/luci/server/encryptedcookies/session/datastore"

	analysisserver "go.chromium.org/luci/analysis/server"
)

// Entrypoint for the result-ingestion service.
func main() {
	analysisserver.Main(func(srv *server.Server) error {
		return nil
	})
}
