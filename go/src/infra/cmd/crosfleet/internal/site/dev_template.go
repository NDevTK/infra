// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build dev
// +build dev

// This file is intended for overriding environment values in this package when building for local testing.
//
//   go build -tags='dev'
//
// You can copy this file to dev.go and edit it.

package site

import (
	"fmt"

	"go.chromium.org/luci/grpc/prpc"
)

func init() {
	if false { // Change this to true.
		Dev.AdminService = "0.0.0.0:8082"
		Dev.InventoryService = "0.0.0.0:8081"
		Dev.UFSService = "127.0.0.1:8800"
		Prod = Dev
		DefaultPRPCOptions = &prpc.Options{
			Insecure:  true,
			UserAgent: fmt.Sprintf("crosfleet/%s", VersionNumber),
		}
	}
}
