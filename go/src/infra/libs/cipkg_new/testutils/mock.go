// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package testutils

import (
	"fmt"
	"path/filepath"
	"time"

	"infra/libs/cipkg_new/core"
)

var (
	_ core.PackageManager = &MockPackageManager{}
	_ core.PackageHandler = &MockPackageHandler{}
)

// MockStorage and MockPackage implements core.PackageManager interface. It stores
// metadata and derivation in the memory. It doesn't allocate any "real" storage
// in the filesystem.
type MockPackageManager struct {
	pkgs    map[string]core.PackageHandler
	baseDir string
}

func NewMockPackageManage(tempDir string) core.PackageManager {
	return &MockPackageManager{
		pkgs:    make(map[string]core.PackageHandler),
		baseDir: tempDir,
	}
}

func (pm *MockPackageManager) Get(id string) core.PackageHandler {
	if h, ok := pm.pkgs[id]; ok {
		return h
	}
	h := &MockPackageHandler{
		id:        id,
		available: false,
		baseDir:   filepath.Join(pm.baseDir, "pkgs", id),
	}
	pm.pkgs[id] = h
	return h
}

type MockPackageHandler struct {
	id        string
	available bool

	lastUsed time.Time
	ref      int

	baseDir string
}

func (p *MockPackageHandler) OutputDirectory() string {
	return filepath.Join(p.baseDir, "content")
}
func (p *MockPackageHandler) LoggingDirectory() string {
	return filepath.Join(p.baseDir, "logs")
}
func (p *MockPackageHandler) Build(f func() error) error {
	if err := f(); err != nil {
		return err
	}
	p.available = true
	return nil
}

func (p *MockPackageHandler) TryRemove() (ok bool, err error) {
	if !p.available || p.ref != 0 {
		return false, nil
	}
	p.available = false
	return true, nil
}

func (p *MockPackageHandler) IncRef() error {
	p.ref += 1
	p.lastUsed = time.Now()
	return nil
}

func (p *MockPackageHandler) DecRef() error {
	if p.ref == 0 {
		return fmt.Errorf("no reference to the package")
	}
	p.ref -= 1
	return nil
}
