// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"infra/tools/pkgbuild/pkg/spec"

	"go.chromium.org/luci/cipd/client/cipd/platform"
	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/cipkg/base/workflow"
	"go.chromium.org/luci/cipkg/core"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/system/filesystem"
)

type Application struct {
	LoggingLevel logging.Level

	// Target CIPD platform for packages.
	TargetPlatform string

	// Storage directory for build and cache packages.
	StorageDir string
	// Spec pool directory for finding 3pp specs.
	SpecPoolDir string

	// If not empty, statistics data will be stored.
	StatisticsDir string

	// CIPD service URL for downloading and uploading packages.
	CIPDService string
	// (TODO): If true, upload packages to CIPD service.
	Upload bool
	// The prefix to use for uploading built packages.
	CIPDPackagePrefix string

	// If true, prepend additional experimental/ to upload path.
	Experiment bool

	// Display help message
	Help bool

	// List of packages to be built. If empty, build all packages in the pool.
	Packages []string

	// Local Package Manager
	PackageManager *workflow.LocalPackageManager
}

func (a *Application) Parse(args []string) error {
	fs := flag.NewFlagSet("pkgbuild", flag.ContinueOnError)

	fs.Var(&a.LoggingLevel, "logging-level", "Logging level for pkgbuild.")

	fs.StringVar(&a.TargetPlatform, "target-platform", a.TargetPlatform, "Target CIPD platform.")

	fs.StringVar(&a.StorageDir, "storage-dir", a.StorageDir, "Required; Local storage directory for build and cache packages.")
	fs.StringVar(&a.SpecPoolDir, "spec-pool", a.SpecPoolDir, "Required; Spec pool directory for finding 3pp specs.")

	fs.StringVar(&a.StatisticsDir, "statistics-dir", a.StatisticsDir, "If statistics-dir is not empty, statistics data will be stored. This may slow down the build.")

	fs.StringVar(&a.CIPDService, "cipd-service", a.CIPDService, "CIPD service URL for downloading and uploading packages.")
	fs.BoolVar(&a.Upload, "upload", a.Upload, "If upload is true, packages will be uploaded to CIPD.")
	fs.StringVar(&a.CIPDPackagePrefix, "cipd-package-prefix", a.CIPDPackagePrefix, "Required; The prefix to use for uploading built packages.")

	fs.BoolVar(&a.Experiment, "experiment", a.Experiment, "If experiment is true, packages will be uploaded to experimental/.")

	fs.BoolVar(&a.Help, "help", false, "Display help message.")

	fs.Usage = func() {
		fmt.Fprint(fs.Output(), "Usage: pkgbuild [OPTION]... [PKG_NAME]...\n")
		fmt.Fprint(fs.Output(), "Build listed packages. If no package name (e.g. tools/ninja) is provided, build all packages in the pool.\n\n")
		fmt.Fprint(fs.Output(), "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}
	if a.Help {
		fs.Usage()
		return nil
	}

	if a.StorageDir == "" || a.SpecPoolDir == "" {
		fs.Usage()
		return fmt.Errorf("storage-dir and spec-pool are required")
	}

	if a.CIPDPackagePrefix == "" {
		fs.Usage()
		return fmt.Errorf("cipd-package-prefix is required")
	}

	if a.Experiment {
		a.CIPDPackagePrefix = path.Join("experimental", a.CIPDPackagePrefix)
	}

	a.Packages = fs.Args()

	pm, err := workflow.NewLocalPackageManager(a.StorageDir)
	if err != nil {
		return err
	}
	a.PackageManager = pm

	return nil
}

func (a *Application) NewBuilder(ctx context.Context) (*PackageBuilder, error) {
	vpythonSpecPath := filepath.Join(a.SpecPoolDir, ".vpython3")
	if _, err := os.Stat(vpythonSpecPath); err != nil {
		return nil, errors.Annotate(err, "failed to find vpython3 specs").Err()
	}
	specLoaderCfg := spec.DefaultSpecLoaderConfig(vpythonSpecPath)
	specLoaderCfg.CIPDPackagePrefix = a.CIPDPackagePrefix
	loader, err := spec.NewSpecLoader(a.SpecPoolDir, specLoaderCfg)
	if err != nil {
		return nil, errors.Annotate(err, "failed to load specs").Err()
	}

	ap := actions.NewActionProcessor()

	target := generators.PlatformFromCIPD(a.TargetPlatform)
	plats := generators.Platforms{
		Build:  generators.CurrentPlatform(),
		Host:   target,
		Target: target,
	}

	builder := workflow.NewBuilder(plats, a.PackageManager, ap)
	builder.SetPreExecuteHook(func(ctx context.Context, pkg actions.Package) error {
		// TODO: Fetch package from cipd, if available
		return nil
	})

	return &PackageBuilder{
		Packages: a.PackageManager,
		Platforms: generators.Platforms{
			Build:  generators.CurrentPlatform(),
			Host:   target,
			Target: target,
		},

		CIPDHost:   platform.CurrentPlatform(),
		CIPDTarget: a.TargetPlatform,
		SpecLoader: loader,

		BuildTempDir: filepath.Join(a.StorageDir, "temp"),
		Builder:      builder,
	}, nil
}

type PackageBuilder struct {
	Packages  core.PackageManager
	Platforms generators.Platforms

	CIPDHost   string
	CIPDTarget string
	SpecLoader *spec.SpecLoader

	BuildTempDir string
	Builder      *workflow.Builder

	loaded []generators.Generator
}

// Load 3pp spec by name and convert it into a cipkg.Package. If the 3pp spec
// depends on other specs, they will also be loaded and added. The package is
// added to the builder so its content will be available after BuildAll
// executed.
func (b *PackageBuilder) Load(ctx context.Context, name string) error {
	g, err := b.SpecLoader.FromSpec(name, b.CIPDHost, b.CIPDTarget)
	if err != nil {
		return err
	}
	b.loaded = append(b.loaded, g)
	return nil
}

// BuildAll builds all added packages.
func (b *PackageBuilder) BuildAll(ctx context.Context) ([]actions.Package, error) {
	if err := filesystem.RemoveAll(b.BuildTempDir); err != nil {
		return nil, err
	}
	if err := os.Mkdir(b.BuildTempDir, os.ModePerm); err != nil {
		return nil, err
	}

	return b.Builder.BuildAll(ctx, b.BuildTempDir, b.loaded)
}
