// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"embed"
	"errors"
	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	. "infra/libs/cipkg/utilities/testing"
	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/stdenv"
	"log"
	"os"
	"runtime"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/cipd/client/cipd/platform"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
)

//go:embed tests
var tests embed.FS

func TestMain(m *testing.M) {
	if runtime.GOOS == "windows" {
		log.Println("Skip pkgbuild tests: not implemented.")
		return
	}

	if err := stdenv.Init(func(bin string) (string, error) {
		// We don't need to import binaries from host.
		return bin, nil
	}); err != nil {
		log.Fatalf("failed to init stdenv: %v", err)
	}
	os.Exit(m.Run())
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

		mockBuild := NewMockBuild()
		mockStorage := NewMockStorage()
		b := &PackageBuilder{
			Storage: mockStorage,
			Platforms: cipkg.Platforms{
				Build:  buildPlatform,
				Host:   buildPlatform,
				Target: buildPlatform,
			},
			CIPDTarget:        cipdPlatform,
			SpecLoader:        loader,
			BuildTempDir:      buildTemp,
			DerivationBuilder: utilities.NewBuilder(mockStorage),

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

		Convey("Build curl", func() {
			pkg, err := b.Add(ctx, "static_libs/curl")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)

			drv := pkg.Derivation()
			So(drv.Name, ShouldEqual, "curl")
			So(drv.Platform, ShouldEqual, buildPlatform.String())
			So(builtins.GetEnv("_3PP_PLATFORM", drv.Env), ShouldEqual, cipdPlatform)
		})
	})

	Convey("Cross-compile platform", t, func() {
		buildPlatform := utilities.CurrentPlatform()
		if buildPlatform.OS() != "linux" || buildPlatform.Arch() != "amd64" {
			return
		}

		hostPlatform := utilities.NewPlatform("linux", "arm64")
		cipdPlatform := "linux-arm64"

		mockBuild := NewMockBuild()
		mockStorage := NewMockStorage()
		b := &PackageBuilder{
			Storage: mockStorage,
			Platforms: cipkg.Platforms{
				Build:  buildPlatform,
				Host:   hostPlatform,
				Target: hostPlatform,
			},
			CIPDTarget:        cipdPlatform,
			SpecLoader:        loader,
			BuildTempDir:      buildTemp,
			DerivationBuilder: utilities.NewBuilder(mockStorage),

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
			So(builtins.GetEnv("_3PP_PLATFORM", drv.Env), ShouldEqual, cipdPlatform)
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

type MockSourceResolver struct{}

func (*MockSourceResolver) ResolveGitSource(*spec.GitSource) (string, string, error) {
	return "tag", "commit", nil
}
func (*MockSourceResolver) ResolveScriptSource(*spec.ScriptSource) (string, error) {
	panic("not implemented")
}

func MockSpecLoaderConfig() *spec.SpecLoaderConfig {
	return &spec.SpecLoaderConfig{
		CIPDPackagePrefix:     "mock",
		CIPDSourceCachePrefix: "sources",
		SourceResolver:        &MockSourceResolver{},
	}
}
