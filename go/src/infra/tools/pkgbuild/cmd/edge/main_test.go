// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// We only test the builder on a subset of platforms we support.
// Other platforms should be cross-compiled.
//go:build amd64 || (arm64 && darwin)

package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	. "infra/libs/cipkg/utilities/testing"
	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/stdenv"
	"io/fs"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/cipd/client/cipd/platform"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
)

//go:embed tests
var tests embed.FS

func initStdenv(build *utilities.Platform) {
	So(stdenv.Init(&stdenv.Config{
		XcodeDeveloper: &builtins.CopyFiles{
			Name:  "xcode_import",
			Files: embed.FS{},
		},
		WinSDK: &builtins.CopyFiles{
			Name:  "winsdk_files",
			Files: embed.FS{},
		},
		FindBinary: func(bin string) (string, error) {
			// We don't need to import binaries from host.
			return bin, nil
		},
		BuildPlatform: build,
	}), ShouldBeNil)
}

func TestBuildPackagesFromSpec(t *testing.T) {
	buildTemp := t.TempDir()

	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Error)

	loader, err := spec.NewSpecLoader(tests, MockSpecLoaderConfig())
	if err != nil {
		t.Fatalf("failed to init spec loader: %v", err)
	}

	Convey("Native platform", t, func() {
		buildPlatform := utilities.CurrentPlatform()
		cipdPlatform := platform.CurrentPlatform()

		initStdenv(buildPlatform)

		mockBuild := NewMockBuild()
		mockPackageManage := NewMockPackageManage()
		b := &PackageBuilder{
			Packages: mockPackageManage,
			Platforms: cipkg.Platforms{
				Build:  buildPlatform,
				Host:   buildPlatform,
				Target: buildPlatform,
			},
			CIPDHost:          cipdPlatform,
			CIPDTarget:        cipdPlatform,
			SpecLoader:        loader,
			BuildTempDir:      buildTemp,
			DerivationBuilder: utilities.NewBuilder(mockPackageManage),

			BuildFunc: mockBuild.Build,
		}

		Convey("Build ninja", func() {
			pkg, err := b.Add(ctx, "tools/ninja")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)

			drv := pkg.Derivation()
			So(drv.Name, ShouldEqual, "ninja")
			So(drv.Platform, ShouldEqual, buildPlatform.String())
			So(builtins.GetEnv("_3PP_PLATFORM", drv.Env), ShouldEqual, cipdPlatform)
		})

		Convey("Build go", func() {
			pkg, err := b.Add(ctx, "tools/go")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)

			drv := pkg.Derivation()
			So(drv.Name, ShouldEqual, "go")
			So(drv.Platform, ShouldEqual, buildPlatform.String())
			So(builtins.GetEnv("_3PP_PLATFORM", drv.Env), ShouldEqual, cipdPlatform)
		})
	})

	Convey("Cross-compile platform", t, func() {
		buildPlatform := utilities.NewPlatform("linux", "amd64")
		hostPlatform := utilities.NewPlatform("linux", "arm64")
		cipdHost := "linux-amd64"
		cipdTarget := "linux-arm64"

		initStdenv(buildPlatform)

		mockBuild := NewMockBuild()
		mockPackageManage := NewMockPackageManage()
		b := &PackageBuilder{
			Packages: mockPackageManage,
			Platforms: cipkg.Platforms{
				Build:  buildPlatform,
				Host:   hostPlatform,
				Target: hostPlatform,
			},
			CIPDHost:          cipdHost,
			CIPDTarget:        cipdTarget,
			SpecLoader:        loader,
			BuildTempDir:      buildTemp,
			DerivationBuilder: utilities.NewBuilder(mockPackageManage),

			BuildFunc: mockBuild.Build,
		}

		Convey("Build packages", func() {
			pkg, err := b.Add(ctx, "tools/ninja")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)

			drv := pkg.Derivation()
			So(drv.Name, ShouldEqual, "ninja")
			So(drv.Platform, ShouldEqual, buildPlatform.String())
			So(builtins.GetEnv("_3PP_PLATFORM", drv.Env), ShouldEqual, cipdTarget)
		})

		// If a dependency is not available, ErrPackageNotAvailable should be the
		// inner error.
		Convey("Unavailable dependency", func() {
			_, err := b.Add(ctx, "tests/unavailable_depends")
			So(err, ShouldNotBeNil)
			So(err, ShouldNotEqual, spec.ErrPackageNotAvailable)
			So(errors.Is(err, spec.ErrPackageNotAvailable), ShouldBeTrue)
		})

		// If a package itself is not available, ErrPackageNotAvailable should be
		// the direct error.
		Convey("Unavailable", func() {
			_, err := b.Add(ctx, "tests/unavailable_arm64")
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, spec.ErrPackageNotAvailable)
		})
	})
}

func TestPackageSources(t *testing.T) {
	buildTemp := t.TempDir()

	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Error)

	loader, err := spec.NewSpecLoader(tests, MockSpecLoaderConfig())
	if err != nil {
		t.Fatalf("failed to init spec loader: %v", err)
	}

	Convey("Native platform", t, func() {
		buildPlatform := utilities.NewPlatform("linux", "amd64")
		cipdPlatform := "linux-amd64"

		initStdenv(buildPlatform)

		mockBuild := NewMockBuild()
		mockPackageManage := NewMockPackageManage()
		b := &PackageBuilder{
			Packages: mockPackageManage,
			Platforms: cipkg.Platforms{
				Build:  buildPlatform,
				Host:   buildPlatform,
				Target: buildPlatform,
			},
			CIPDHost:          cipdPlatform,
			CIPDTarget:        cipdPlatform,
			SpecLoader:        loader,
			BuildTempDir:      buildTemp,
			DerivationBuilder: utilities.NewBuilder(mockPackageManage),

			BuildFunc: mockBuild.Build,
		}

		Convey("Git Source", func() {
			_, err := b.Add(ctx, "tools/ninja")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)

			verifySource(t, mockBuild.Packages, cipkg.PackageMetadata{
				CacheKey: "mock/sources/git/github.com/ninja-build/ninja?subdir=src&tag=" + url.QueryEscape("3@git-tag"),
			})
		})

		Convey("URL Source", func() {
			_, err := b.Add(ctx, "static_libs/curl")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)

			verifySource(t, mockBuild.Packages, cipkg.PackageMetadata{
				CacheKey: "mock/sources/url/static_libs/curl/" + cipdPlatform + "?tag=" + url.QueryEscape("3@7.59.0"),
			})
		})

		Convey("Script Source", func() {
			_, err := b.Add(ctx, "tools/go")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)

			verifySource(t, mockBuild.Packages, cipkg.PackageMetadata{
				CacheKey: "mock/sources/script/tools/go/" + cipdPlatform + "?tag=" + url.QueryEscape("3@script-version"),
			})
		})
	})
}

func verifySource(t *testing.T, pkgs []cipkg.Package, metadata cipkg.PackageMetadata) {
	t.Helper()
	name := fmt.Sprintf("%s_source", pkgs[len(pkgs)-1].Derivation().Name)
	for _, p := range pkgs {
		if p.Derivation().Name == name {
			So(p.Metadata(), ShouldResemble, metadata)
			return
		}
	}
	t.Fatalf("source not found: %s", name)
}

type MockSourceResolver struct{}

func (*MockSourceResolver) ResolveGitSource(git *spec.GitSource) (spec.GitSourceInfo, error) {
	return spec.GitSourceInfo{
		Tag:    "git-tag",
		Commit: "commit",
	}, nil
}
func (*MockSourceResolver) ResolveScriptSource(cipdHostPlatform string, dir fs.FS, script *spec.ScriptSource) (spec.ScriptSourceInfo, error) {
	return spec.ScriptSourceInfo{
		Version: "script-version",
		URL:     []string{"url"},
		Name:    []string{"name"},
	}, nil
}

func MockSpecLoaderConfig() *spec.SpecLoaderConfig {
	return &spec.SpecLoaderConfig{
		CIPDPackagePrefix:     "mock",
		CIPDSourceCachePrefix: "sources",
		SourceResolver:        &MockSourceResolver{},
	}
}
