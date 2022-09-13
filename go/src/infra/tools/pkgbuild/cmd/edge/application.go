package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/storage"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
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

	// (TODO): If true, append additional experimental/ to upload path.
	Experiment bool

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
	fs.BoolVar(&a.Experiment, "experiment", a.Experiment, "If experiment is true, packages will be uploaded to experimental/.")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if a.StorageDir == "" || a.SpecPoolDir == "" {
		fs.Usage()
		return fmt.Errorf("storage-dir and spec-pool are required")
	}

	a.Packages = fs.Args()
	if len(a.Packages) == 0 {
		// TODO(fancl): Support 3pp/ and packages defined in subdir.
		fs, err := os.ReadDir(a.SpecPoolDir)
		if err != nil {
			return err
		}
		for _, f := range fs {
			if f.IsDir() {
				a.Packages = append(a.Packages, f.Name())
			}
		}
	}
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

	return &PackageBuilder{
		Storage: s,
		Platforms: cipkg.Platforms{
			Build:  utilities.CurrentPlatform(),
			Host:   target,
			Target: target,
		},

		CIPDTarget: a.TargetPlatform,
		SpecLoader: spec.NewSpecLoader(os.DirFS(a.SpecPoolDir), nil),

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
}

func (b *PackageBuilder) Build(ctx context.Context, name string) (cipkg.Package, error) {
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

	// Build derivations
	if err := b.DerivationBuilder.Add(pkg); err != nil {
		return nil, errors.Annotate(err, "failed to add package to builder").Err()
	}
	if err := os.RemoveAll(b.BuildTempDir); err != nil {
		return nil, err
	}
	if err := os.Mkdir(b.BuildTempDir, os.ModePerm); err != nil {
		return nil, err
	}
	if err := b.DerivationBuilder.BuildAll(func(p cipkg.Package) error {
		id := p.Derivation().ID()
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
	}); err != nil {
		return nil, errors.Annotate(err, "failed to build package").Err()
	}

	return pkg, nil
}
