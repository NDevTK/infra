package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/storage"

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
}

func (a *Application) Parse(args []string) error {
	fs := flag.NewFlagSet("pkgbuild", flag.ContinueOnError)

	fs.Var(&a.LoggingLevel, "logging-level", "Logging level for pkgbuild.")

	fs.StringVar(&a.TargetPlatform, "target-platform", a.TargetPlatform, "Target CIPD platform.")

	fs.StringVar(&a.StorageDir, "storage-dir", a.StorageDir, "Required; Local storage directory for build and cache packages.")
	fs.StringVar(&a.SpecPoolDir, "spec-pool", a.SpecPoolDir, "Required; Spec pool directory for finding 3pp specs.")

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

	return nil
}

func (a *Application) NewBuilder(ctx context.Context) (*PackageBuilder, error) {
	s, err := utilities.NewLocalStorage(a.StorageDir)
	if err != nil {
		return nil, errors.Annotate(err, "failed to load storage").Err()
	}
	s = storage.NewCIPDStorage(ctx, a.CIPDService, s)

	target, err := spec.ParseCIPDPlatform(a.TargetPlatform)
	if err != nil {
		return nil, errors.Annotate(err, "failed to parse cipd platform").Err()
	}

	vpythonSpecPath := filepath.Join(a.SpecPoolDir, ".vpython3")
	if _, err := os.Stat(vpythonSpecPath); err != nil {
		return nil, errors.Annotate(err, "failed to find vpython3 specs").Err()
	}
	specLoaderCfg := spec.DefaultSpecLoaderConfig(vpythonSpecPath)
	specLoaderCfg.CIPDPackagePrefix = a.CIPDPackagePrefix
	loader, err := spec.NewSpecLoader(os.DirFS(a.SpecPoolDir), specLoaderCfg)
	if err != nil {
		return nil, errors.Annotate(err, "failed to load specs").Err()
	}

	return &PackageBuilder{
		Storage: s,
		Platforms: cipkg.Platforms{
			Build:  utilities.CurrentPlatform(),
			Host:   target,
			Target: target,
		},

		CIPDTarget: a.TargetPlatform,
		SpecLoader: loader,

		BuildTempDir:      filepath.Join(a.StorageDir, "temp"),
		DerivationBuilder: utilities.NewBuilder(s),
	}, nil
}

type PackageBuilder struct {
	Storage   cipkg.Storage
	Platforms cipkg.Platforms

	CIPDTarget string
	SpecLoader *spec.SpecLoader

	BuildTempDir      string
	DerivationBuilder *utilities.Builder

	// Override the default build func
	BuildFunc func(p cipkg.Package) error
}

// Add(...) loads 3pp spec by name and convert it into a cipkg.Package. If the
// 3pp spec depends on other specs, they will also be loaded and added.
// The package is added to the builder so its content will be available after
// BuildAll(...) executed.
func (b *PackageBuilder) Add(ctx context.Context, name string) (cipkg.Package, error) {
	g, err := b.SpecLoader.FromSpec(name, b.CIPDTarget)
	if err != nil {
		return nil, err
	}

	// Generate derivations
	bctx := &cipkg.BuildContext{
		Platforms: b.Platforms,
		Storage:   b.Storage,
		Context:   ctx,
	}

	drv, meta, err := g.Generate(bctx)
	if err != nil {
		return nil, errors.Annotate(err, "failed to generate derivation").Err()
	}
	pkg := b.Storage.Add(drv, meta)

	if err := b.DerivationBuilder.Add(pkg); err != nil {
		return nil, errors.Annotate(err, "failed to add package to builder").Err()
	}

	return pkg, nil
}

// BuildAll(...) builds all added packages.
func (b *PackageBuilder) BuildAll(ctx context.Context) error {
	if err := filesystem.RemoveAll(b.BuildTempDir); err != nil {
		return err
	}
	if err := os.Mkdir(b.BuildTempDir, os.ModePerm); err != nil {
		return err
	}

	f := func(p cipkg.Package) error {
		id := p.Derivation().ID()
		logging.Infof(ctx, "build package %s", id)

		d, err := os.MkdirTemp(b.BuildTempDir, fmt.Sprintf("%s-", id))
		if err != nil {
			return err
		}
		var out strings.Builder
		cmd := utilities.CommandFromPackage(p)
		cmd.Stdout = &out
		cmd.Stderr = &out
		cmd.Dir = d
		if err := builtins.Execute(ctx, cmd); err != nil {
			logging.Errorf(ctx, "%s", out.String())
			return err
		}
		logging.Debugf(ctx, "%s", out.String())
		return nil
	}
	if b.BuildFunc != nil {
		f = b.BuildFunc
	}

	if err := b.DerivationBuilder.BuildAll(f); err != nil {
		return errors.Annotate(err, "failed to build package").Err()
	}

	return nil
}
