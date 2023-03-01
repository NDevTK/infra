package testing

import (
	"fmt"
	"infra/libs/cipkg"
	"time"
)

// Testing utilities for mocking common cipkg interfaces.

// MockBuild is used as the build function for utilities.Builder.
// MockBuild.Build can be passed to Builder.BuildAll for collecting packages'
// information.
type MockBuild struct {
	Packages []cipkg.Package
}

func NewMockBuild() *MockBuild {
	return &MockBuild{
		Packages: make([]cipkg.Package, 0),
	}
}

func (b *MockBuild) Build(p cipkg.Package) error {
	b.Packages = append(b.Packages, p)
	return nil
}
func (b *MockBuild) Reset() { b.Packages = b.Packages[:0] }

// MockStorage and MockPackage implements cipkg.Storage interface. It stores
// metadata and derivation in the memory. It doesn't allocate any "real" storage
// in the filesystem.
type MockPackageManager struct {
	pkgs map[string]cipkg.Package
}

func NewMockPackageManage() cipkg.PackageManager {
	return &MockPackageManager{
		pkgs: make(map[string]cipkg.Package),
	}
}

func (pm *MockPackageManager) Get(id string) cipkg.Package { return pm.pkgs[id] }
func (pm *MockPackageManager) Add(drv cipkg.Derivation, metadata cipkg.PackageMetadata) cipkg.Package {
	pkg := &MockPackage{
		derivation: drv,
		metadata:   metadata,
		available:  false,
	}
	pm.pkgs[drv.ID()] = pkg
	return pkg
}

type MockPackage struct {
	derivation cipkg.Derivation
	metadata   cipkg.PackageMetadata
	available  bool

	lastUsed time.Time
	ref      int
}

func (p *MockPackage) Derivation() cipkg.Derivation    { return p.derivation }
func (p *MockPackage) Metadata() cipkg.PackageMetadata { return p.metadata }
func (p *MockPackage) Directory() string               { return p.derivation.ID() }

func (p *MockPackage) Build(f func(cipkg.Package) error) error {
	if err := f(p); err != nil {
		return err
	}
	p.available = true
	return nil
}

func (p *MockPackage) TryRemove() (ok bool, err error) {
	if !p.available || p.ref != 0 {
		return false, nil
	}
	p.available = false
	return true, nil
}

func (p *MockPackage) Status() cipkg.PackageStatus {
	return cipkg.PackageStatus{
		Available: p.available,
		LastUsed:  p.lastUsed,
	}
}

func (p *MockPackage) IncRef() error {
	p.ref += 1
	p.lastUsed = time.Now()
	return nil
}

func (p *MockPackage) DecRef() error {
	if p.ref == 0 {
		return fmt.Errorf("no reference to the package")
	}
	p.ref -= 1
	return nil
}
