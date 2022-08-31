// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/stdenv"

	"go.chromium.org/luci/cipd/client/cipd/platform"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
)

func main() {
	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Error)

	// TODO(fancl): properly parse the args.
	storageDir, specDir := os.Args[1], os.Args[2]
	targetCIPD := platform.CurrentPlatform()
	if len(os.Args) > 3 {
		targetCIPD = os.Args[3]
	}

	if err := stdenv.Init(); err != nil {
		log.Fatal(err)
	}

	l := spec.NewSpecLoader(specDir)

	pkgs, err := packages(l, targetCIPD)
	if err != nil {
		log.Fatal(err)
	}

	target, err := parseCIPDPlatform(targetCIPD)
	if err != nil {
		log.Fatal(err)
	}
	if err := build(ctx, storageDir, target, pkgs...); err != nil {
		log.Fatal(err)
	}
}

func parseCIPDPlatform(plat string) (cipkg.Platform, error) {
	idx := strings.Index(plat, "-")
	if idx == -1 {
		return nil, errors.Reason("invalid cipd target platform: %s", plat).Err()
	}
	os, arch := plat[:idx], plat[idx+1:]
	if os == "mac" {
		os = "darwin"
	}
	if arch == "armv6l" {
		arch = "arm"
	}
	return utilities.NewPlatform(os, arch), nil
}

// Return the list of packages' generators to be built for the target platform
// e.g. linux-amd64.
func packages(l *spec.SpecLoader, target string) ([]cipkg.Generator, error) {
	pkgs := []string{
		"curl",
		"ninja",
	}

	var gs []cipkg.Generator
	for _, pkg := range pkgs {
		g, err := l.FromSpec(pkg, target)
		if err != nil {
			return nil, err
		}
		gs = append(gs, g)
	}

	return gs, nil
}

func build(ctx context.Context, path string, target cipkg.Platform, gens ...cipkg.Generator) error {
	s, err := utilities.NewLocalStorage(path)
	if err != nil {
		return errors.Annotate(err, "failed to load storage").Err()
	}
	// Generate derivations
	bctx := &cipkg.BuildContext{
		Platforms: cipkg.Platforms{
			Build:  utilities.CurrentPlatform(),
			Host:   target,
			Target: target,
		},
		Storage: s,
		Context: ctx,
	}

	for _, g := range gens {
		drv, meta, err := g.Generate(bctx)
		if err != nil {
			return errors.Annotate(err, "failed to generate venv derivation").Err()
		}
		pkg := s.Add(drv, meta)
		// Build derivations
		b := utilities.NewBuilder(s)
		if err := b.Add(pkg); err != nil {
			return errors.Annotate(err, "failed to add package to builder").Err()
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
			return errors.Annotate(err, "failed to build package").Err()
		}
	}
	s.Prune(ctx, time.Hour*24, 256)
	return nil
}
