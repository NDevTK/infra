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

type LocalStorage struct {
	storagePath string
	packages    map[string]cipkg.Package
}

func NewLocalStorage(path string) (cipkg.Storage, error) {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, fmt.Errorf("initialize local storage failed: %s: %w", path, err)
	}
	s := &LocalStorage{
		storagePath: path,
		packages:    make(map[string]cipkg.Package),
	}
	return s, nil
}

func (s *LocalStorage) Get(id string) cipkg.Package {
	if pkg := s.packages[id]; pkg != nil {
		return pkg
	}

	// The ill-formed package is returned for (maybe) cleanup.
	return &LocalStoragePackage{
		baseDirectory: filepath.Join(s.storagePath, id),
		lockFile:      filepath.Join(s.storagePath, fmt.Sprintf(".%s.lock", id)),
	}
}

func (s *LocalStorage) Add(drv cipkg.Derivation, m cipkg.PackageMetadata) cipkg.Package {
	id := drv.ID()
	pkg := &LocalStoragePackage{
		baseDirectory: filepath.Join(s.storagePath, id),
		derivation:    &drv,
		metadata:      &m,
		lockFile:      filepath.Join(s.storagePath, fmt.Sprintf(".%s.lock", id)),
	}
	s.packages[id] = pkg
	return pkg
}

func (s *LocalStorage) Prune(c context.Context, ttl time.Duration, max int) {
	deadline := time.Now().Add(-ttl)
	locks, err := fs.Glob(os.DirFS(s.storagePath), ".*.lock")
	if err != nil {
		logging.WithError(err).Warningf(c, "failed to list locks")
	}
	pruned := 0
	for _, l := range locks {
		id := l[1 : len(l)-5] // remove prefix "." and suffix ".lock"
		pkg := s.Get(id)
		if ok, mtime := pkg.Available(); !ok || mtime.Before(deadline) {
			if removed, err := pkg.TryRemove(); err != nil {
				logging.WithError(err).Warningf(c, "failed to remove package")
			} else if removed {
				logging.Debugf(c, "prune: remove package (not used since %s): %s", mtime, id)
				if pruned++; pruned == max {
					logging.Debugf(c, "prune: hit prune limit of %d ", max)
					break
				}
			}
		} else {
			logging.Debugf(c, "prune: skip package (not used since %s): %s", mtime, id)
		}
	}
}

type LocalStoragePackage struct {
	baseDirectory string
	derivation    *cipkg.Derivation
	metadata      *cipkg.PackageMetadata
	lockFile      string
	rlockHandle   fslock.Handle
}

func (p *LocalStoragePackage) Derivation() cipkg.Derivation {
	return *p.derivation
}

func (p *LocalStoragePackage) Metadata() cipkg.PackageMetadata {
	return *p.metadata
}

func (p *LocalStoragePackage) Directory() string {
	return filepath.Join(p.baseDirectory, "contents")
}

func (p *LocalStoragePackage) Build(builder func(cipkg.Package) error) error {
	if p.rlockHandle != nil {
		return fmt.Errorf("can't build package when read lock is held")
	}

	return fslock.WithBlocking(p.lockFile, blocker, func() error {
		if ok, _ := p.Available(); ok {
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

func (p *LocalStoragePackage) TryRemove() (ok bool, err error) {
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

func (p *LocalStoragePackage) Available() (bool, time.Time) {
	if s, err := os.Stat(p.stampPath()); err == nil {
		return true, s.ModTime()
	}
	return false, time.Time{}
}

func (p *LocalStoragePackage) RLock() error {
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
		if ok, _ := p.Available(); !ok {
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

func (p *LocalStoragePackage) RUnlock() error {
	if err := p.rlockHandle.Unlock(); err != nil {
		return fmt.Errorf("failed to release read lock: %w", err)
	}
	p.rlockHandle = nil
	return nil
}

func (p *LocalStoragePackage) stampPath() string {
	return filepath.Join(p.baseDirectory, "derivation.json")
}

func (p *LocalStoragePackage) touch() error {
	if ok, _ := p.Available(); !ok {
		return nil
	}
	return filesystem.Touch(p.stampPath(), time.Time{}, 0644)
}

// RLockRecursive will RLock the package with all its dependencies recursively.
// If an error happened, it may end up with only part of the packages are
// locked.
func RLockRecursive(s cipkg.Storage, pkg cipkg.Package) error {
	return doPackageRecursive(s, pkg, make(map[string]struct{}),
		func(pkg cipkg.Package) error { return pkg.RLock() })
}

// RUnlockRecursive will RUnlock the package with all its dependencies
// recursively. If an error happened, it may end up with only part of the
// packages are unlocked.
func RUnlockRecursive(s cipkg.Storage, pkg cipkg.Package) error {
	return doPackageRecursive(s, pkg, make(map[string]struct{}),
		func(pkg cipkg.Package) error { return pkg.RUnlock() })
}

func doPackageRecursive(s cipkg.Storage, pkg cipkg.Package, visited map[string]struct{}, f func(cipkg.Package) error) error {
	if _, ok := visited[pkg.Derivation().ID()]; ok {
		return nil
	}
	if err := f(pkg); err != nil {
		return err
	}
	visited[pkg.Derivation().ID()] = struct{}{}

	for _, id := range pkg.Metadata().Dependencies {
		dep := s.Get(id)
		if err := doPackageRecursive(s, dep, visited, f); err != nil {
			return err
		}
	}
	return nil
}
