// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utilities

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"infra/libs/cipkg"

	"github.com/danjacques/gofslock/fslock"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/system/filesystem"
)

func blocker() error { return clock.Sleep(context.Background(), time.Millisecond*10).Err }

// LocalPackageManager is a PackageManager implementation that stores packages
// locally. It supports recording package references acrossing multiple
// instances using fslock.
type LocalPackageManager struct {
	storagePath string
	packages    map[string]cipkg.Package
}

func NewLocalPackageManager(path string) (*LocalPackageManager, error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("initialize local storage failed: %s: %w", path, err)
	}
	s := &LocalPackageManager{
		storagePath: path,
		packages:    make(map[string]cipkg.Package),
	}
	return s, nil
}

func (pm *LocalPackageManager) Get(id string) cipkg.Package {
	if pkg := pm.packages[id]; pkg != nil {
		return pkg
	}

	// The ill-formed package is returned for (maybe) cleanup.
	return &LocalPackage{
		baseDirectory: filepath.Join(pm.storagePath, id),
		lockFile:      filepath.Join(pm.storagePath, fmt.Sprintf(".%s.lock", id)),
	}
}

func (pm *LocalPackageManager) Add(drv cipkg.Derivation, m cipkg.PackageMetadata) cipkg.Package {
	id := drv.ID()
	pkg := &LocalPackage{
		baseDirectory: filepath.Join(pm.storagePath, id),
		derivation:    &drv,
		metadata:      &m,
		lockFile:      filepath.Join(pm.storagePath, fmt.Sprintf(".%s.lock", id)),
	}
	pm.packages[id] = pkg
	return pkg
}

func (pm *LocalPackageManager) Prune(c context.Context, ttl time.Duration, max int) {
	deadline := time.Now().Add(-ttl)
	locks, err := fs.Glob(os.DirFS(pm.storagePath), ".*.lock")
	if err != nil {
		logging.WithError(err).Warningf(c, "failed to list locks")
	}
	pruned := 0
	for _, l := range locks {
		id := l[1 : len(l)-5] // remove prefix "." and suffix ".lock"
		pkg := pm.Get(id)
		if st := pkg.Status(); !st.Available || st.LastUsed.Before(deadline) {
			if removed, err := pkg.TryRemove(); err != nil {
				logging.WithError(err).Warningf(c, "failed to remove package")
			} else if removed {
				logging.Debugf(c, "prune: remove package (not used since %s): %s", st.LastUsed, id)
				if pruned++; pruned == max {
					logging.Debugf(c, "prune: hit prune limit of %d ", max)
					break
				}
			}
		} else {
			logging.Debugf(c, "prune: skip package (not used since %s): %s", st.LastUsed, id)
		}
	}
}

type LocalPackage struct {
	baseDirectory string
	derivation    *cipkg.Derivation
	metadata      *cipkg.PackageMetadata
	lockFile      string
	rlockHandle   fslock.Handle
}

func (p *LocalPackage) Derivation() cipkg.Derivation {
	return *p.derivation
}

func (p *LocalPackage) Metadata() cipkg.PackageMetadata {
	return *p.metadata
}

func (p *LocalPackage) Directory() string {
	return filepath.Join(p.baseDirectory, "contents")
}

func (p *LocalPackage) Build(builder func(cipkg.Package) error) error {
	if p.rlockHandle != nil {
		return fmt.Errorf("can't build package when read lock is held")
	}

	return fslock.WithBlocking(p.lockFile, blocker, func() error {
		if st := p.Status(); st.Available {
			return nil
		}

		if err := filesystem.RemoveAll(p.baseDirectory); err != nil {
			return fmt.Errorf("failed to remove package dir: %s: %w", p.Directory(), err)
		}
		if err := os.MkdirAll(p.Directory(), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create package dir: %s: %w", p.Directory(), err)
		}

		if err := builder(p); err != nil {
			return fmt.Errorf("failed to build derivation: %w", err)
		}

		f, err := os.Create(p.stampPath())
		if err != nil {
			return fmt.Errorf("failed to create stamp file: %s: %w", p.stampPath(), err)
		}
		defer f.Close()
		if err := json.NewEncoder(f).Encode(p.Derivation()); err != nil {
			return fmt.Errorf("failed to encode stamp file: %s: %w", p.stampPath(), err)
		}
		return nil
	})
}

func (p *LocalPackage) TryRemove() (ok bool, err error) {
	switch err := fslock.With(p.lockFile, func() error {
		if err := filesystem.RemoveAll(p.baseDirectory); err != nil {
			return fmt.Errorf("failed to remove package dir: %s: %w", p.Directory(), err)
		}
		return nil
	}); err {
	case nil:
		if err := filesystem.RemoveAll(p.lockFile); err != nil {
			return false, nil
		}
		return true, nil
	case fslock.ErrLockHeld:
		return false, nil
	default:
		return false, err
	}
}

func (p *LocalPackage) Status() cipkg.PackageStatus {
	if s, err := os.Stat(p.stampPath()); err == nil {
		return cipkg.PackageStatus{
			Available: true,
			LastUsed:  s.ModTime(),
		}
	}

	return cipkg.PackageStatus{
		Available: false,
	}
}

func (p *LocalPackage) IncRef() error {
	if p.rlockHandle != nil {
		return fmt.Errorf("acquire read lock multiple times on same package")
	}

	h, err := fslock.LockSharedBlocking(p.lockFile, blocker)
	if err != nil {
		return fmt.Errorf("failed to acquire read lock: %w", err)
	}
	if err := func() error {
		if err := h.PreserveExec(); err != nil {
			return fmt.Errorf("failed to perserve lock: %w", err)
		}
		if st := p.Status(); !st.Available {
			return fmt.Errorf("package not available")
		}

		// Update mtime of the stamp since at this point we ensured:
		// 1. Package is locked and won't be removed
		// 2. Stamp is presented in the package
		if err := filesystem.Touch(p.stampPath(), time.Time{}, 0644); err != nil {
			return fmt.Errorf("failed to touch the the stamp: %w", err)
		}
		return nil
	}(); err != nil {
		h.Unlock()
		return err
	}
	p.rlockHandle = h
	return nil
}

func (p *LocalPackage) DecRef() error {
	if err := p.rlockHandle.Unlock(); err != nil {
		return fmt.Errorf("failed to release read lock: %w", err)
	}
	p.rlockHandle = nil
	return nil
}

func (p *LocalPackage) stampPath() string {
	return filepath.Join(p.baseDirectory, "derivation.json")
}

func (p *LocalPackage) touch() error {
	if st := p.Status(); !st.Available {
		return nil
	}
	return filesystem.Touch(p.stampPath(), time.Time{}, 0644)
}

// IncRefRecursive will IncRef the package with all its dependencies
// recursively. If an error happened, it may end up with only part of the
// packages are referenced.
func IncRefRecursive(pm cipkg.PackageManager, pkg cipkg.Package) error {
	return doPackageRecursive(pm, pkg, make(map[string]struct{}),
		func(pkg cipkg.Package) error { return pkg.IncRef() })
}

// DecRefRecursive will DecRef the package with all its dependencies
// recursively. If an error happened, it may end up with only part of the
// packages are dereferenced.
func DecRefRecursive(pm cipkg.PackageManager, pkg cipkg.Package) error {
	return doPackageRecursive(pm, pkg, make(map[string]struct{}),
		func(pkg cipkg.Package) error { return pkg.DecRef() })
}

func doPackageRecursive(pm cipkg.PackageManager, pkg cipkg.Package, visited map[string]struct{}, f func(cipkg.Package) error) error {
	if _, ok := visited[pkg.Derivation().ID()]; ok {
		return nil
	}
	if err := f(pkg); err != nil {
		return err
	}
	visited[pkg.Derivation().ID()] = struct{}{}

	for _, id := range pkg.Metadata().Dependencies {
		dep := pm.Get(id)
		if err := doPackageRecursive(pm, dep, visited, f); err != nil {
			return err
		}
	}
	return nil
}
