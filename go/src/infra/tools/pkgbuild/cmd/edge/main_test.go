// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"embed"
	"errors"
	"infra/libs/cipkg"
	"infra/libs/cipkg/utilities"
	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/stdenv"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

// TODO(fancl): We should generate all the commands without executing them
// similar to recipe tests.
func TestBuildPackagesFromSpec(t *testing.T) {
	storageDir := t.TempDir()

	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Error)

	storage, err := utilities.NewLocalStorage(storageDir)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}

	loader, err := spec.NewSpecLoader(tests, nil)
	if err != nil {
		t.Fatalf("failed to init spec loader: %v", err)
	}

	Convey("Native platform", t, func() {
		b := &PackageBuilder{
			Storage: storage,
			Platforms: cipkg.Platforms{
				Build:  utilities.CurrentPlatform(),
				Host:   utilities.CurrentPlatform(),
				Target: utilities.CurrentPlatform(),
			},
			CIPDTarget:        platform.CurrentPlatform(),
			SpecLoader:        loader,
			BuildTempDir:      filepath.Join(storageDir, "temp"),
			DerivationBuilder: utilities.NewBuilder(storage),
		}

		Convey("Build packages", func() {
			_, err := b.Add(ctx, "tools/ninja")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)
		})

		// It takes too long (10+ mins) and downloads code from the internet.
		// Disable the test until we vendored the code.
		SkipConvey("Build curl", func() {
			_, err := b.Add(ctx, "static_libs/curl")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)
		})
	})

	Convey("Cross-compile platform", t, func() {
		build := utilities.CurrentPlatform()
		if build.OS() != "linux" || build.Arch() != "amd64" {
			return
		}

		host := utilities.NewPlatform("linux", "arm64")
		cipd := "linux-arm64"

		b := &PackageBuilder{
			Storage: storage,
			Platforms: cipkg.Platforms{
				Build:  build,
				Host:   host,
				Target: host,
			},
			CIPDTarget:        cipd,
			SpecLoader:        loader,
			BuildTempDir:      filepath.Join(storageDir, "temp"),
			DerivationBuilder: utilities.NewBuilder(storage),
		}

		Convey("Build packages", func() {
			_, err := b.Add(ctx, "tools/ninja")
			So(err, ShouldBeNil)
			err = b.BuildAll(ctx)
			So(err, ShouldBeNil)
		})
	})
}
