// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/luci/cipd/client/cipd/platform"
	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/cipkg/base/workflow"
	"go.chromium.org/luci/cipkg/core"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/system/filesystem"
	"go.chromium.org/luci/provenance/api/snooperpb/v1"
	"go.chromium.org/luci/provenance/client"

	"infra/tools/pkgbuild/pkg/spec"
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
	// TODO(fancl): If true, upload packages to CIPD service.
	Upload bool
	// The prefix to use for uploading built packages.
	CIPDPackagePrefix string

	// If true, prepend additional experimental/ to upload path.
	Experiment bool

	// Snoopy service URL for reporting artifact hash.
	SnoopyService string

	// Display help message
	Help bool

	// List of packages to be built. If empty, build all packages in the pool.
	Packages []string

	// Local Package Manager
	PackageManager *workflow.LocalPackageManager
}

// Parse parses input arguments.
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

	fs.StringVar(&a.SnoopyService, "snoopy-service", a.SnoopyService, "Snoopy service URL for reporting artifact hash.")

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

	a.Packages = fs.Args()

	pm, err := workflow.NewLocalPackageManager(a.StorageDir)
	if err != nil {
		return err
	}
	a.PackageManager = pm

	return nil
}

// NewBuilder creates the PackageBuilder used for building packages based on
// the configuration and platform.
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
		return cipdPackage(pkg).download(ctx, a.CIPDService)
	})
	// TODO: set executor to luciexe.

	return &PackageBuilder{
		Packages: a.PackageManager,
		Platforms: generators.Platforms{
			Build:  generators.CurrentPlatform(),
			Host:   target,
			Target: target,
		},

		CIPDService: a.CIPDService,
		CIPDHost:    platform.CurrentPlatform(),
		CIPDTarget:  a.TargetPlatform,
		SpecLoader:  loader,

		BuildTempDir: filepath.Join(a.StorageDir, "temp"),
		Builder:      builder,
	}, nil
}

// TryUpload build and register the cipd if Application.Upload set to true.
func (a *Application) TryUpload(ctx context.Context, pkgs []actions.Package) error {
	if !a.Upload {
		return nil
	}

	clt, err := client.MakeProvenanceClient(ctx, a.SnoopyService)
	if err != nil {
		return errors.Annotate(err, "failed to create provenance client").Err()
	}

	tmp, err := os.MkdirTemp("", "pkgbuild-")
	if err != nil {
		return errors.Annotate(err, "failed to create tmp dir").Err()
	}
	defer filesystem.RemoveAll(tmp)

	for _, pkg := range pkgs {
		if err := a.tryUploadOne(ctx, func(ctx context.Context, in *snooperpb.ReportCipdRequest) error {
			_, err := clt.ReportCipd(ctx, in)
			return err
		}, tmp, pkg); err != nil {
			return err
		}
	}

	return nil
}

// We can't pass ProvenanceClient since it's private. Use a wrapper function
// instead.
type cipdReporter func(ctx context.Context, in *snooperpb.ReportCipdRequest) error

// tryUploadOne uploads the package provided. If reporter function is not nil,
// it will be called after the cipd file generated in tmp, to report the
// cipd package to snoopy service.
func (a *Application) tryUploadOne(ctx context.Context, reporter cipdReporter, tmp string, pkg actions.Package) error {
	cipdPkg := cipdPackage(pkg)

	// Package is available in cipd
	if err := cipdPkg.check(ctx, a.CIPDService); err == nil {
		// TODO(fancl): add tags and refs
		cipdPkg.setTags(ctx, a.CIPDService, nil)
		return nil
	} else if !errors.Is(err, errPackgeNotExist) {
		return err
	}

	// Skip if Package is not available locally.
	if err := cipdPkg.Handler.IncRef(); err != nil {
		return nil
	}
	defer cipdPkg.Handler.DecRef()

	var prefix string
	if a.Experiment {
		prefix = "experimental"
	}

	// TODO(fancl): add tags and refs
	name, iid, err := cipdPkg.upload(ctx, tmp, a.CIPDService, prefix, nil)
	if err != nil {
		return errors.Annotate(err, "failed to upload package").Err()
	}

	// Recursively upload package's dependencies
	var deps []actions.Package
	deps = append(deps, cipdPkg.BuildDependencies...)
	deps = append(deps, cipdPkg.RuntimeDependencies...)
	for _, dep := range deps {
		if err := a.tryUploadOne(ctx, reporter, tmp, dep); err != nil {
			return err
		}
	}

	if reporter != nil && iid != "" {
		// Report package info to server to trigger provenance generation.
		if err := reporter(ctx, &snooperpb.ReportCipdRequest{
			CipdReport: &snooperpb.CipdReport{
				PackageName: name,
				Iid:         iid,
				EventTs:     timestamppb.New(time.Now()),
			},
		}); err != nil {
			// Error during reporting won't block the package build.
			logging.Warningf(ctx, "report cipd package to snoopy failed: %s: %s", pkg.Action.Metadata.Cipd.Name, err)
		}
	}

	return nil
}

type PackageBuilder struct {
	Packages  core.PackageManager
	Platforms generators.Platforms

	CIPDService string
	CIPDHost    string
	CIPDTarget  string
	SpecLoader  *spec.SpecLoader

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

// BuildAll builds loaded packages.
// - If skipUploaded is true and a package we loaded is available in cipd, we
// don't need to rebuild or make it available locally.
// - If a package we added depends on a package available in cipd, we can
// download the prebuilt package using PreExecuteHook.
// - Otherwise, we will build the package locally.
func (b *PackageBuilder) BuildAll(ctx context.Context, skipUploaded bool) ([]actions.Package, error) {
	// TODO(fancl): we should use a real temp directory.
	if err := filesystem.RemoveAll(b.BuildTempDir); err != nil {
		return nil, err
	}
	if err := os.Mkdir(b.BuildTempDir, os.ModePerm); err != nil {
		return nil, err
	}

	pkgs, err := b.Builder.GeneratePackages(ctx, b.loaded)
	if err != nil {
		return nil, err
	}

	var newPkgs []actions.Package
	if skipUploaded {
		// Check if package has been built and available in cipd.
		for _, pkg := range pkgs {
			if err := cipdPackage(pkg).check(ctx, b.CIPDService); err != nil {
				if errors.Is(err, errPackgeNotExist) {
					newPkgs = append(newPkgs, pkg)
				} else {
					return nil, err
				}
			}
		}
	} else {
		newPkgs = slices.Clone(pkgs)
	}

	// Make packages available. We still return all packages regardless of the
	// error.
	if err := b.Builder.BuildPackages(ctx, b.BuildTempDir, newPkgs); err != nil {
		return pkgs, err
	}

	return pkgs, nil
}
