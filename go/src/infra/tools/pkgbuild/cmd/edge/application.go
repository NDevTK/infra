package main

import (
	"context"
	"crypto"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/pkgbuild/pkg/packages"
	"infra/tools/pkgbuild/pkg/spec"

	"go.chromium.org/luci/cipd/client/cipd/platform"
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
	packages *utilities.LocalPackageManager
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

	return nil
}

func (a *Application) NewBuilder(ctx context.Context) (*PackageBuilder, error) {
	var err error
	a.packages, err = utilities.NewLocalPackageManager(a.StorageDir)
	if err != nil {
		return nil, errors.Annotate(err, "failed to load storage").Err()
	}
	pm := packages.NewCIPDPackageManager(ctx, a.CIPDService, a.packages)

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
		Packages: pm,
		Platforms: cipkg.Platforms{
			Build:  utilities.CurrentPlatform(),
			Host:   target,
			Target: target,
		},

		CIPDHost:   platform.CurrentPlatform(),
		CIPDTarget: a.TargetPlatform,
		SpecLoader: loader,

		BuildTempDir:      filepath.Join(a.StorageDir, "temp"),
		DerivationBuilder: utilities.NewBuilder(pm),

		StatisticsDir: a.StatisticsDir,
	}, nil
}

func (a *Application) Prune(ctx context.Context, ttl time.Duration, max int) {
	a.packages.Prune(ctx, ttl, max)
}

type PackageBuilder struct {
	Packages  cipkg.PackageManager
	Platforms cipkg.Platforms

	CIPDHost   string
	CIPDTarget string
	SpecLoader *spec.SpecLoader

	BuildTempDir      string
	DerivationBuilder *utilities.Builder

	StatisticsDir string

	// Override the default build func
	BuildFunc func(p cipkg.Package) error
}

// Add(...) loads 3pp spec by name and convert it into a cipkg.Package. If the
// 3pp spec depends on other specs, they will also be loaded and added.
// The package is added to the builder so its content will be available after
// BuildAll(...) executed.
func (b *PackageBuilder) Add(ctx context.Context, name string) (cipkg.Package, error) {
	g, err := b.SpecLoader.FromSpec(name, b.CIPDHost, b.CIPDTarget)
	if err != nil {
		return nil, err
	}

	// Generate derivations
	bctx := &cipkg.BuildContext{
		Platforms: b.Platforms,
		Packages:  b.Packages,
		Context:   ctx,
	}

	drv, meta, err := g.Generate(bctx)
	if err != nil {
		return nil, errors.Annotate(err, "failed to generate derivation").Err()
	}
	pkg := b.Packages.Add(drv, meta)

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

	if b.StatisticsDir != "" {
		if err := os.MkdirAll(b.StatisticsDir, os.ModePerm); err != nil {
			return err
		}
		buildF := f
		f = func(p cipkg.Package) error {
			id := p.Derivation().ID()

			startTime := time.Now()
			if err := buildF(p); err != nil {
				return err
			}
			endTime := time.Now()

			h := crypto.SHA256.New()
			if err := builtins.WalkDir(os.DirFS(p.Directory()), ".", h, func(string, fs.DirEntry, error) error { return nil }); err != nil {
				return err
			}

			stat, err := os.Create(filepath.Join(b.StatisticsDir, id+".json"))
			if err != nil {
				return err
			}
			defer stat.Close()

			return json.NewEncoder(stat).Encode(map[string]any{
				"buildTimeSeconds": endTime.Sub(startTime).Seconds(),
				"resultSHA256":     fmt.Sprintf("%x", h.Sum(nil)),
			})
		}
	}

	if b.BuildFunc != nil {
		f = b.BuildFunc
	}

	if err := b.DerivationBuilder.BuildAll(f); err != nil {
		return errors.Annotate(err, "failed to build package").Err()
	}

	return nil
}
