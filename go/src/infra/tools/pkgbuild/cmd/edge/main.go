// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/stdenv"

	"go.chromium.org/luci/cipd/client/cipd/platform"
	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/common/system/filesystem"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"go.chromium.org/luci/provenance/api/snooperpb/v1"
	"go.chromium.org/luci/provenance/client"
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
	}
	if err := app.Parse(os.Args[1:]); err != nil {
		return errors.Annotate(err, "failed to parse options").Err()
	}
	ctx = logging.SetLevel(ctx, app.LoggingLevel)

	if app.Help {
		// TODO: print help message?
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

	if app.Upload {
		if err := upload(ctx, app.Experiment, pkgs); err != nil {
			return err
		}
	}

	app.PackageManager.Prune(ctx, time.Hour*24, 256)
	return nil
}

func upload(ctx context.Context, experimental bool, pkgs []actions.Package) error {
	clt, err := client.MakeProvenanceClient(ctx, "http://localhost:11000")
	if err != nil {
		return errors.Annotate(err, "failed to create provenance client").Err()
	}

	tmp, err := os.MkdirTemp("", "pkgbuild-")
	if err != nil {
		return errors.Annotate(err, "failed to create tmp dir").Err()
	}
	defer filesystem.RemoveAll(tmp)

	for _, pkg := range pkgs {
		metadata := pkg.Action.Metadata
		logging.Infof(ctx, "creating cipd package %s:%s", metadata.Cipd.Name, metadata.Cipd.Version)

		name := metadata.Cipd.Name
		if experimental {
			name = path.Join("experimental", name)
		}

		out := filepath.Join(tmp, pkg.DerivationID+".cipd")
		Iid, err := createCIPD(name, pkg.Handler.OutputDirectory(), out)
		if err != nil {
			return errors.Annotate(err, "failed to build cipd package").Err()
		}

		// Report package info to server to trigger provenance generation.
		if _, err := clt.ReportCipd(ctx, &snooperpb.ReportCipdRequest{
			CipdReport: &snooperpb.CipdReport{
				PackageName: name,
				Iid:         Iid,
			},
		}); err != nil {
			// Error during reporting won't block the package build.
			logging.Warningf(ctx, "report cipd package to snoopy failed: %s: %s", pkg.Action.Metadata.Cipd.Name, err)
		}
	}

	return nil
}

func createCIPD(name, src, dst string) (Iid string, err error) {
	resultFile := dst + ".json"
	cmd := CIPDCommand("pkg-build",
		"-name", name,
		"-in", src,
		"-out", dst,
		"-json-output", resultFile,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	f, err := os.Open(resultFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var result struct {
		Result struct {
			Package    string
			InstanceID string `json:"instance_id"`
		}
	}
	if err := json.NewDecoder(f).Decode(&result); err != nil {
		return "", err
	}

	return result.Result.InstanceID, nil
}
