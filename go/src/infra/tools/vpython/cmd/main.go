// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"os"
	"time"

	"infra/tools/vpython/api/env"
	"infra/tools/vpython/application"

	"golang.org/x/net/context"
)

var defaultApplication = application.A{
	VENVPackage: env.Spec_Package{
		Path:    "infra/python/virtualenv",
		Version: "version:15.1.0",
	},
	PruneThreshold: 7 * 24 * time.Hour, // One week.
	PruneLimit:     3,
}

func main() {
	rv := defaultApplication.Main(context.Background())
	os.Exit(rv)
}
