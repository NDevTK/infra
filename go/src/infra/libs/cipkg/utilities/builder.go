// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utilities

import (
	"fmt"
	"infra/libs/cipkg"
	"os/exec"
)

type Builder struct {
	pkgs    []cipkg.Package
	added   map[string]struct{}
	storage cipkg.Storage
}

func NewBuilder(storage cipkg.Storage) *Builder {
	return &Builder{
		added:   make(map[string]struct{}),
		storage: storage,
	}
}

func (b *Builder) Add(pkg cipkg.Package) error {
	id := pkg.Derivation().ID()
	if _, ok := b.added[id]; ok {
		return nil
	}

	for _, dep := range pkg.Derivation().Inputs {
		dpkg := b.storage.Get(dep)
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

func (b *Builder) BuildAll(builder func(cipkg.Package) error) error {
	for _, pkg := range b.pkgs {
		// if package has been built, we don't need to build it again - but we still
		// need to RLock the package since its content may be used by others.
		if ok, _ := pkg.Available(); ok {
			if err := pkg.RLock(); err != nil {
				return fmt.Errorf("failed to acquire read lock for package: %#v: %w", pkg.Derivation(), err)
			}
			continue
		}

		if err := pkg.Build(builder); err != nil {
			return fmt.Errorf("failed to build package : %#v: %w", pkg.Derivation(), err)
		}
		if err := pkg.RLock(); err != nil {
			return fmt.Errorf("failed to acquire read lock for package: %#v: %w", pkg.Derivation(), err)
		}
	}

	// Release all locks. Leave it to the user to decide those of which packages
	// will be used at the runtime. There may be a chance that a package is
	// removed during the short amount of time. But since RLock will update the
	// last accessed timestamp and pruning is based on a reasonable TTL, this is
	// highly unlikely. And even if it's happened, we can retry the process.
	for _, pkg := range b.pkgs {
		if err := pkg.RUnlock(); err != nil {
			return fmt.Errorf("failed to release read lock for package: %#v: %w", pkg.Derivation(), err)
		}
	}
	return nil
}

func CommandFromPackage(pkg cipkg.Package) *exec.Cmd {
	drv := pkg.Derivation()
	cmd := exec.Command(drv.Builder, drv.Args...)
	cmd.Env = append([]string{fmt.Sprintf("out=%s", pkg.Directory())}, drv.Env...)
	return cmd
}
