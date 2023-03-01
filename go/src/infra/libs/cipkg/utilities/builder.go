// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utilities

import (
	"fmt"
	"infra/libs/cipkg"
	"os/exec"

	"go.chromium.org/luci/common/errors"
)

type Builder struct {
	pkgs     []cipkg.Package
	added    map[string]struct{}
	packages cipkg.PackageManager
}

func NewBuilder(pm cipkg.PackageManager) *Builder {
	return &Builder{
		added:    make(map[string]struct{}),
		packages: pm,
	}
}

func (b *Builder) Add(pkg cipkg.Package) error {
	id := pkg.Derivation().ID()
	if _, ok := b.added[id]; ok {
		return nil
	}

	for _, dep := range pkg.Derivation().Inputs {
		dpkg := b.packages.Get(dep)
		if dpkg == nil {
			return fmt.Errorf("package not found: %s", id)
		}
		if err := b.Add(dpkg); err != nil {
			return fmt.Errorf("add package failed: %#v: %w", pkg.Derivation(), err)
		}
	}

	b.pkgs = append(b.pkgs, pkg)
	b.added[id] = struct{}{}
	return nil
}

// BuildAll builds packages added to the builder and all their dependencies.
// All packages will be dereferenced after the build. Leave it to the user
// to decide those of which packages will be used at the runtime. There may
// be a chance that a package is removed during the short amount of time.
// But since IncRef will update the last accessed timestam, this is highly
// unlikely. And even if it's happened, we can retry the process.
func (b *Builder) BuildAll(builder func(cipkg.Package) error) (err error) {
	var cleanupFuncs []func() error
	defer func() {
		// TODO(fancl): use errors.Join after Go 1.20
		var merr errors.MultiError
		merr.MaybeAdd(err)
		for _, f := range cleanupFuncs {
			merr.MaybeAdd(f())
		}
		err = merr.AsError()
	}()

	for _, pkg := range b.pkgs {
		// if package has been built, we don't need to build it again - but we still
		// need to refer the package since its content may be used by others.
		if st := pkg.Status(); st.Available {
			if err := pkg.IncRef(); err != nil {
				return fmt.Errorf("failed to reference the package: %#v: %w", pkg.Derivation(), err)
			}
			cleanupFuncs = append(cleanupFuncs, pkg.DecRef)
			continue
		}

		if err := pkg.Build(builder); err != nil {
			return fmt.Errorf("failed to build package : %#v: %w", pkg.Derivation(), err)
		}
		if err := pkg.IncRef(); err != nil {
			return fmt.Errorf("failed to reference the package: %#v: %w", pkg.Derivation(), err)
		}
		cleanupFuncs = append(cleanupFuncs, pkg.DecRef)
	}

	return nil
}

func CommandFromPackage(pkg cipkg.Package) *exec.Cmd {
	drv := pkg.Derivation()
	cmd := exec.Command(drv.Builder, drv.Args...)
	cmd.Env = append([]string{fmt.Sprintf("out=%s", pkg.Directory())}, drv.Env...)
	return cmd
}
