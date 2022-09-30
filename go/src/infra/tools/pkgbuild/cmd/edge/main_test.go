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
	"os/exec"
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
	if runtime.GOOS == "linux" {
		// Docker is required for running pkgbuild on Linux. Skip tests if it's
		// not available.
		if _, err := exec.LookPath("docker"); errors.Is(err, exec.ErrNotFound) {
			log.Println("Skip pkgbuild tests: docker not found in the PATH.")
			return
		}
	}
	if runtime.GOOS == "windows" {
		log.Println("Skip pkgbuild tests: not implemented.")
		return
	}

	if err := stdenv.Init(); err != nil {
		log.Fatalf("failed to init stdenv: %v", err)
	}
	os.Exit(m.Run())
}

func TestBuildPackagesFromSpec(t *testing.T) {
	buildTemp := t.TempDir()

	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Error)

	loader, err := spec.NewSpecLoader(tests, nil)
	if err != nil {
		t.Fatalf("failed to init spec loader: %v", err)
	}

	Convey("Native platform", t, func() {
		build := utilities.CurrentPlatform()
		cipd := platform.CurrentPlatform()

		mockBuild := NewMockBuild()
		mockStorage := NewMockStorage()
		b := &PackageBuilder{
			Storage: mockStorage,
			Platforms: cipkg.Platforms{
				Build:  build,
				Host:   build,
				Target: build,
			},
			CIPDTarget:        cipd,
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
			So(drv.Platform, ShouldEqual, build.String())
			So(builtins.GetEnv("_3PP_PLATFORM", drv.Env), ShouldEqual, cipd)
		})

		Convey("Build curl", func() {
			pkg, err := b.Add(ctx, "static_libs/curl")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)

			drv := pkg.Derivation()
			So(drv.Name, ShouldEqual, "curl")
			So(drv.Platform, ShouldEqual, build.String())
			So(builtins.GetEnv("_3PP_PLATFORM", drv.Env), ShouldEqual, cipd)
		})
	})

	Convey("Cross-compile platform", t, func() {
		build := utilities.CurrentPlatform()
		if build.OS() != "linux" || build.Arch() != "amd64" {
			return
		}

		host := utilities.NewPlatform("linux", "arm64")
		cipd := "linux-arm64"

		mockBuild := NewMockBuild()
		mockStorage := NewMockStorage()
		b := &PackageBuilder{
			Storage: mockStorage,
			Platforms: cipkg.Platforms{
				Build:  build,
				Host:   host,
				Target: host,
			},
			CIPDTarget:        cipd,
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
			So(drv.Platform, ShouldEqual, build.String())
			So(builtins.GetEnv("_3PP_PLATFORM", drv.Env), ShouldEqual, cipd)
		})
	})
}
