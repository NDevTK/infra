package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/pkgbuild/pkg/stdenv"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
)

func main() {
	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Error)

	if err := stdenv.Init(); err != nil {
		log.Fatal(err)
	}

	if err := build(ctx, os.Args[1], &stdenv.Generator{
		Name: "pcre2",
		Source: stdenv.SourceURL{
			URL:           "https://github.com/PCRE2Project/pcre2/releases/download/pcre2-10.23/pcre2-10.23.tar.gz",
			HashAlgorithm: builtins.HashIgnore,
		},
	}); err != nil {
		log.Fatal(err)
	}
}

func build(ctx context.Context, path string, out cipkg.Generator) error {
	s, err := utilities.NewLocalStorage(path)
	if err != nil {
		return errors.Annotate(err, "failed to load storage").Err()
	}

	s.Prune(ctx, time.Hour*24, 256)

	// Generate derivations
	bctx := &cipkg.BuildContext{
		Platform: cipkg.Platform{
			Build:  fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
			Host:   fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
			Target: fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH),
		},
		Storage: s,
		Context: ctx,
	}

	drv, meta, err := out.Generate(bctx)
	if err != nil {
		return errors.Annotate(err, "failed to generate venv derivation").Err()
	}
	pkg := s.Add(drv, meta)

	// Build derivations
	b := utilities.NewBuilder(s)
	if err := b.Add(pkg); err != nil {
		return errors.Annotate(err, "failed to add venv to builder").Err()
	}

	var temp = filepath.Join(path, "temp")
	if err := os.RemoveAll(temp); err != nil {
		return err
	}
	if err := os.Mkdir(temp, os.ModePerm); err != nil {
		return err
	}

	if err := b.BuildAll(func(p cipkg.Package) error {
		id := p.Derivation().ID()
		d, err := os.MkdirTemp(temp, fmt.Sprintf("%s-", id))
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
		return nil
	}); err != nil {
		return errors.Annotate(err, "failed to build venv").Err()
	}

	return nil
}
