// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package provider

import "os"

// Environments supported.
type environment = int

const (
	gcp environment = iota
	satlab
)

// getEnv determines where the server is running.
// TODO(b/329682593): Support more environments.
func getEnv() environment {
	// K_SERVICE environment variable is set on Cloud Run instances.
	// 7446b7c04264699f6a1c4990a3126e0769081999:luci/luci-go/server/server.go;l=688
	if os.Getenv("K_SERVICE") == "" {
		return satlab
	}
	return gcp
}
