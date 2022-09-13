// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"infra/tools/pkgbuild/pkg/stdenv"

	"go.chromium.org/luci/cipd/client/cipd/platform"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

func main() {
	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Error)

	app := Application{
		LoggingLevel:   logging.Error,
		TargetPlatform: platform.CurrentPlatform(),
		CIPDService:    chromeinfra.CIPDServiceURL,
		Upload:         false,
		Experiment:     false,
	}
	if err := app.Parse(os.Args[1:]); err != nil {
		logging.WithError(err).Errorf(ctx, "failed to parse options")
		os.Exit(1)
	}
	ctx = logging.SetLevel(ctx, app.LoggingLevel)

	if err := stdenv.Init(); err != nil {
		logging.WithError(err).Errorf(ctx, "failed to init stdenv")
		os.Exit(1)
	}

	b, err := app.NewBuilder(ctx)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "failed to init builder")
		os.Exit(1)
	}

	for _, name := range app.Packages {
		pkg, err := b.Build(ctx, name)
		if err != nil {
			logging.WithError(err).Errorf(ctx, "failed to build %s", name)
			os.Exit(1)
		}

		fmt.Println(pkg.Directory()) // (TODO): Upload package here
	}

	b.Storage.Prune(ctx, time.Hour*24, 256)
}
