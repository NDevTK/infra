package packages

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"

	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/system/filesystem"
)

type CIPDPackageManager struct {
	serviceURL string
	logger     logging.Logger

	cipkg.PackageManager
}

// An overlay storage implementation for cipd package caches. It will check
// the cipd service before build the package locallly.
func NewCIPDPackageManager(ctx context.Context, serviceURL string, pm cipkg.PackageManager) *CIPDPackageManager {
	return &CIPDPackageManager{
		serviceURL:     serviceURL,
		logger:         logging.Get(ctx),
		PackageManager: pm,
	}
}

func (pm *CIPDPackageManager) Get(id string) cipkg.Package {
	return &CIPDStoragePackage{
		serviceURL: pm.serviceURL,
		logger:     pm.logger,
		Package:    pm.PackageManager.Get(id),
	}
}

func (pm *CIPDPackageManager) Add(drv cipkg.Derivation, m cipkg.PackageMetadata) cipkg.Package {
	return &CIPDStoragePackage{
		serviceURL: pm.serviceURL,
		logger:     pm.logger,
		Package:    pm.PackageManager.Add(drv, m),
	}
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

	key, err := url.Parse(m.CacheKey)
	if err != nil {
		return fmt.Errorf("failed to parse cache key")
	}

	rootDir := pkg.Directory()
	if d := key.Query().Get("subdir"); d != "" {
		rootDir = filepath.Join(rootDir, d)
	}

	tag := pkg.Derivation().ID()
	if t := key.Query().Get("tag"); t != "" {
		tag = t
	}

	cmd := builtins.CIPDCommand("export", "-service-url", serviceURL, "-root", rootDir, "-ensure-file", "-")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("%s version:%s", key.Path, tag))
	if err := cmd.Run(); err != nil {
		if err := filesystem.RemoveAll(pkg.Directory()); err != nil {
			return fmt.Errorf("failed to clean up export directory: %w", err)
		}
		if err := os.Mkdir(pkg.Directory(), os.ModePerm); err != nil {
			return fmt.Errorf("failed to recreate package directory: %w", err)
		}
		return err
	}

	return nil
}
