// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"infra/libs/cipkg"
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

	if app.Help {
		os.Exit(0)
	}

	if err := stdenv.Init(); err != nil {
		logging.WithError(err).Errorf(ctx, "failed to init stdenv")
		os.Exit(1)
	}

	b, err := app.NewBuilder(ctx)
	if err != nil {
		logging.WithError(err).Errorf(ctx, "failed to init builder")
		os.Exit(1)
	}

	// Build all packages by default
	names := app.Packages
	if len(names) == 0 {
		names = b.SpecLoader.ListAllByFullName()
	}

	var pkgs []cipkg.Package
	for _, name := range names {
		pkg, err := b.Add(ctx, name)
		if err != nil {
			logging.WithError(err).Errorf(ctx, "failed to add %s", name)
			os.Exit(1)
		}
		pkgs = append(pkgs, pkg)
	}

	if err := b.BuildAll(ctx); err != nil {
		logging.WithError(err).Errorf(ctx, "failed to build packages")
		os.Exit(1)
	}

	for _, pkg := range pkgs {
		fmt.Println(pkg.Metadata().CacheKey, pkg.Metadata().Version) // (TODO): Upload package here
	}

	b.Storage.Prune(ctx, time.Hour*24, 256)
}
