package testing

import (
	"context"
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
// To simplify the implementation, there are some behavior differences:
// - Prune is a no-op.
// - No locking provided. RLock and RUnlock don't actually lock anything.
// - Available always returns an empty timestamp.
type MockStorage struct {
	pkgs map[string]cipkg.Package
}

func NewMockStorage() cipkg.Storage {
	return &MockStorage{
		pkgs: make(map[string]cipkg.Package),
	}
}

func (s *MockStorage) Get(id string) cipkg.Package { return s.pkgs[id] }
func (s *MockStorage) Add(drv cipkg.Derivation, metadata cipkg.PackageMetadata) cipkg.Package {
	pkg := &MockPackage{
		derivation: drv,
		metadata:   metadata,
		available:  false,
	}
	s.pkgs[drv.ID()] = pkg
	return pkg
}
func (s *MockStorage) Prune(ctx context.Context, ttl time.Duration, max int) {}

type MockPackage struct {
	derivation cipkg.Derivation
	metadata   cipkg.PackageMetadata
	available  bool
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
	if !p.available {
		return false, nil
	}
	p.available = false
	return true, nil
}
func (p *MockPackage) Available() (ok bool, mtime time.Time) { return p.available, time.Time{} }
func (p *MockPackage) RLock() error                          { return nil }
func (p *MockPackage) RUnlock() error                        { return nil }
