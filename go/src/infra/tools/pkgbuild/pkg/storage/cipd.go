package storage

import (
	"context"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"infra/libs/cipkg"

	"go.chromium.org/luci/common/logging"
)

type CIPDStorage struct {
	serviceURL string
	logger     logging.Logger

	cipkg.Storage
}

// An overlay storage implementation for cipd package caches. It will check
// the cipd service before build the package locallly.
func NewCIPDStorage(ctx context.Context, serviceURL string, s cipkg.Storage) cipkg.Storage {
	return &CIPDStorage{
		serviceURL: serviceURL,
		logger:     logging.Get(ctx),
		Storage:    s,
	}
}

func (s *CIPDStorage) Get(id string) cipkg.Package {
	return &CIPDStoragePackage{
		serviceURL: s.serviceURL,
		logger:     s.logger,
		Package:    s.Storage.Get(id),
	}
}

func (s *CIPDStorage) Add(drv cipkg.Derivation, m cipkg.PackageMetadata) cipkg.Package {
	return &CIPDStoragePackage{
		serviceURL: s.serviceURL,
		logger:     s.logger,
		Package:    s.Storage.Add(drv, m),
	}
}

func (s *CIPDStorage) Prune(c context.Context, ttl time.Duration, max int) {
	s.Storage.Prune(c, ttl, max)
}

type CIPDStoragePackage struct {
	serviceURL string
	logger     logging.Logger

	cipkg.Package
}

func (p *CIPDStoragePackage) Build(builder func(cipkg.Package) error) error {
	return p.Package.Build(func(pkg cipkg.Package) error {
		if err := cipdExport(p.serviceURL, pkg); err == nil {
			p.logger.Infof("cipd storage: copied from cached: %s", pkg.Derivation().Name)
			return nil
		} else {
			p.logger.Debugf("cipd storage: not cached: %s: %v", pkg.Derivation().Name, err)
		}
		return builder(pkg)
	})
}

func cipdExport(serviceURL string, pkg cipkg.Package) error {
	m := pkg.Metadata()
	if m.CacheKey == "" {
		return fmt.Errorf("no cache key available")
	}

	cipd := findCIPD()
	if cipd == "" {
		return fmt.Errorf("cipd binary not available")
	}

	key, err := url.Parse(m.CacheKey)
	if err != nil {
		return fmt.Errorf("failed to parse cache key")
	}

	rootDir := pkg.Directory()
	if d := key.Query().Get("subdir"); d != "" {
		rootDir = filepath.Join(rootDir, d)
	}

	cmd := exec.Command(cipd, "export", "-service-url", serviceURL, "-root", rootDir, "-ensure-file", "-")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s version:%s", key.Path, m.Version))
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func findCIPD() string {
	for _, s := range []string{"cipd", "cipd.exe", "cipd.bat"} {
		if path, err := exec.LookPath(s); err == nil {
			return path
		}
	}
	return ""
}
