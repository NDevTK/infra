// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.chromium.org/luci/cipd/client/cipd/platform"
	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/hardcoded/chromeinfra"

	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/stdenv"
)

func main() {
	ctx := context.Background()
	actions.NewReexecRegistry().Intercept(ctx)

	if err := Main(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func Main(ctx context.Context) error {
	ctx = gologger.StdConfig.Use(ctx)
	ctx = logging.SetLevel(ctx, logging.Error)

	app := Application{
		LoggingLevel:   logging.Error,
		TargetPlatform: platform.CurrentPlatform(),
		CIPDService:    chromeinfra.CIPDServiceURL,
		Upload:         false,
		Experiment:     false,
		SnoopyService:  "http://localhost:11000",
	}
	if err := app.Parse(os.Args[1:]); err != nil {
		return errors.Annotate(err, "failed to parse options").Err()
	}
	ctx = logging.SetLevel(ctx, app.LoggingLevel)

	if app.Help {
		return nil
	}

	if err := stdenv.Init(stdenv.DefaultConfig()); err != nil {
		return errors.Annotate(err, "failed to init stdenv").Err()
	}

	b, err := app.NewBuilder(ctx)
	if err != nil {
		return errors.Annotate(err, "failed to init builder").Err()
	}

	// Build all packages by default
	names := app.Packages
	if len(names) == 0 {
		names = b.SpecLoader.ListAllByFullName()
	}

	for _, name := range names {
		if err := b.Load(ctx, name); err != nil {
			// Only skip a package if it's directly unavailable without checking
			// inner errors. A package marked as available on the target platform has
			// any dependency unavailable shouldn't be skipped.
			if err == spec.ErrPackageNotAvailable {
				logging.Infof(ctx, "skip package %s on %s", name, app.TargetPlatform)
				continue
			}
			return errors.Annotate(err, "failed to add %s", name).Err()
		}
	}

	pkgs, err := b.BuildAll(ctx)
	if err != nil {
		return errors.Annotate(err, "failed to build packages").Err()
	}

	for _, pkg := range pkgs {
		logging.Infof(ctx, "built package %s:%s with %s", pkg.Action.Metadata.Cipd.Name, pkg.Action.Metadata.Cipd.Version, pkg.DerivationID)
	}

	if err := app.TryUpload(ctx, pkgs); err != nil {
		return err
	}

	app.PackageManager.Prune(ctx, time.Hour*24, 256)
	return nil
}
