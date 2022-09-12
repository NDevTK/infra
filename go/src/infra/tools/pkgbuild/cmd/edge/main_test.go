// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"embed"
	"errors"
	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/stdenv"
	"io/fs"
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
	storageDir := t.TempDir()

	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Error)

	specs, err := fs.Sub(tests, "tests")
	if err != nil {
		t.Fatalf("failed to get test data: %v", err)
	}
	l := spec.NewSpecLoader(specs, nil)

	Convey("Select platform", t, func() {
		cipdPlat := platform.CurrentPlatform()
		plat, err := parseCIPDPlatform(cipdPlat)
		So(err, ShouldBeNil)

		Convey("Build ninja", func() {
			g, err := l.FromSpec("ninja", cipdPlat)
			So(err, ShouldBeNil)
			err = build(ctx, storageDir, plat, g)
			So(err, ShouldBeNil)
		})

		// It takes too long (10+ mins) and downloads code from the internet.
		// Disable the test until we vendored the code.
		SkipConvey("Build curl", func() {
			g, err := l.FromSpec("curl", cipdPlat)
			So(err, ShouldBeNil)
			err = build(ctx, storageDir, plat, g)
			So(err, ShouldBeNil)
		})
	})
}
