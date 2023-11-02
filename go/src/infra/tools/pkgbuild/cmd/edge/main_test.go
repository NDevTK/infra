// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// We only test the builder on a subset of platforms we support.
// Other platforms should be cross-compiled.
//go:build amd64 || (arm64 && darwin)

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"go.chromium.org/luci/cipd/client/cipd/platform"
	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/cipkg/base/workflow"
	"go.chromium.org/luci/cipkg/core"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/common/testing/assertions"

	"infra/tools/pkgbuild/pkg/spec"
	"infra/tools/pkgbuild/pkg/stdenv"
)

func initStdenv(build generators.Platform) {
	So(stdenv.Init(&stdenv.Config{
		XcodeDeveloper: &generators.ImportTargets{Name: "xcode_import"},
		WinSDK:         &generators.ImportTargets{Name: "winsdk_files"},
		FindBinary: func(bin string) (string, error) {
			// We don't need to import binaries from host.
			return filepath.FromSlash(fmt.Sprintf("//bin/%s", bin)), nil
		},
		BuildPlatform: build,
	}), ShouldBeNil)
}

func TestBuildPackagesFromSpec(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	ctx := gologger.StdConfig.Use(context.Background())

	Convey("native platform", t, func() {
		tempBase := t.TempDir()
		buildTemp := filepath.Join(tempBase, "build")
		storeTemp := filepath.Join(tempBase, "store")
		specs := filepath.Join(cwd, "testdata")

		loader, err := spec.NewSpecLoader(specs, MockSpecLoaderConfig())
		if err != nil {
			t.Fatalf("failed to init spec loader: %v", err)
		}

		buildPlatform := generators.CurrentPlatform()
		cipdPlatform := platform.CurrentPlatform()

		initStdenv(buildPlatform)

		pm, err := workflow.NewLocalPackageManager(storeTemp)
		So(err, ShouldBeNil)
		ap := actions.NewActionProcessor()

		plats := generators.Platforms{
			Build:  buildPlatform,
			Host:   buildPlatform,
			Target: buildPlatform,
		}
		b := &PackageBuilder{
			Packages:     pm,
			Platforms:    plats,
			CIPDHost:     cipdPlatform,
			CIPDTarget:   cipdPlatform,
			SpecLoader:   loader,
			BuildTempDir: buildTemp,
			Builder:      workflow.NewBuilder(plats, pm, ap),
		}
		b.Builder.SetExecutor(func(context.Context, *workflow.ExecutionConfig, *core.Derivation) error { return nil })

		Convey("build ninja", func() {
			err := b.Load(ctx, "tools/ninja")
			So(err, ShouldBeNil)
			pkgs, err := b.BuildAll(ctx)
			So(err, ShouldBeNil)

			pkg := pkgs[len(pkgs)-1]
			env := environ.New(pkg.Derivation.Env)
			So(pkg.Derivation.Name, ShouldEqual, "ninja")
			So(pkg.Derivation.Platform, ShouldEqual, buildPlatform.String())
			So(env.Get("_3PP_PLATFORM"), ShouldEqual, cipdPlatform)
		})

		Convey("Build go", func() {
			err := b.Load(ctx, "tools/go")
			So(err, ShouldBeNil)
			pkgs, err := b.BuildAll(ctx)
			So(err, ShouldBeNil)

			pkg := pkgs[len(pkgs)-1]
			env := environ.New(pkg.Derivation.Env)
			So(pkg.Derivation.Name, ShouldEqual, "go")
			So(pkg.Derivation.Platform, ShouldEqual, buildPlatform.String())
			So(env.Get("_3PP_PLATFORM"), ShouldEqual, cipdPlatform)
		})
	})

	Convey("cross-compile platform", t, func() {
		tempBase := t.TempDir()
		buildTemp := filepath.Join(tempBase, "build")
		storeTemp := filepath.Join(tempBase, "store")
		specs := filepath.Join(cwd, "testdata")

		loader, err := spec.NewSpecLoader(specs, MockSpecLoaderConfig())
		if err != nil {
			t.Fatalf("failed to init spec loader: %v", err)
		}

		buildPlatform := generators.NewPlatform("linux", "amd64")
		hostPlatform := generators.NewPlatform("linux", "arm64")
		cipdHost := "linux-amd64"
		cipdTarget := "linux-arm64"

		initStdenv(buildPlatform)

		pm, err := workflow.NewLocalPackageManager(storeTemp)
		So(err, ShouldBeNil)
		ap := actions.NewActionProcessor()

		plats := generators.Platforms{
			Build:  buildPlatform,
			Host:   hostPlatform,
			Target: hostPlatform,
		}
		b := &PackageBuilder{
			Packages:     pm,
			Platforms:    plats,
			CIPDHost:     cipdHost,
			CIPDTarget:   cipdTarget,
			SpecLoader:   loader,
			BuildTempDir: buildTemp,
			Builder:      workflow.NewBuilder(plats, pm, ap),
		}
		b.Builder.SetExecutor(func(context.Context, *workflow.ExecutionConfig, *core.Derivation) error { return nil })

		Convey("build packages", func() {
			err := b.Load(ctx, "tools/ninja")
			So(err, ShouldBeNil)
			pkgs, err := b.BuildAll(ctx)
			So(err, ShouldBeNil)

			pkg := pkgs[len(pkgs)-1]
			env := environ.New(pkg.Derivation.Env)
			So(pkg.Derivation.Name, ShouldEqual, "ninja")
			So(pkg.Derivation.Platform, ShouldEqual, buildPlatform.String())
			So(env.Get("_3PP_PLATFORM"), ShouldEqual, cipdTarget)
		})

		// If a dependency is not available, ErrPackageNotAvailable should be the
		// inner error.
		Convey("unavailable dependency", func() {
			err := b.Load(ctx, "tests/unavailable_depends")
			So(err, ShouldNotBeNil)
			So(err, ShouldNotEqual, spec.ErrPackageNotAvailable)
			So(errors.Is(err, spec.ErrPackageNotAvailable), ShouldBeTrue)
		})

		// If a package itself is not available, ErrPackageNotAvailable should be
		// the direct error.
		Convey("unavailable", func() {
			err := b.Load(ctx, "tests/unavailable_arm64")
			So(err, ShouldNotBeNil)
			So(err, ShouldEqual, spec.ErrPackageNotAvailable)
		})
	})
}

func TestPackageSources(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	ctx := gologger.StdConfig.Use(context.Background())
	ctx = logging.SetLevel(ctx, logging.Error)

	Convey("native platform", t, func() {
		tempBase := t.TempDir()
		buildTemp := filepath.Join(tempBase, "build")
		storeTemp := filepath.Join(tempBase, "store")
		specs := filepath.Join(cwd, "testdata")

		loader, err := spec.NewSpecLoader(specs, MockSpecLoaderConfig())
		if err != nil {
			t.Fatalf("failed to init spec loader: %v", err)
		}

		buildPlatform := generators.NewPlatform("linux", "amd64")
		cipdPlatform := "linux-amd64"

		initStdenv(buildPlatform)

		pm, err := workflow.NewLocalPackageManager(storeTemp)
		So(err, ShouldBeNil)
		ap := actions.NewActionProcessor()

		plats := generators.Platforms{
			Build:  buildPlatform,
			Host:   buildPlatform,
			Target: buildPlatform,
		}
		b := &PackageBuilder{
			Packages:     pm,
			Platforms:    plats,
			CIPDHost:     cipdPlatform,
			CIPDTarget:   cipdPlatform,
			SpecLoader:   loader,
			BuildTempDir: buildTemp,
			Builder:      workflow.NewBuilder(plats, pm, ap),
		}
		b.Builder.SetExecutor(func(context.Context, *workflow.ExecutionConfig, *core.Derivation) error { return nil })

		Convey("git source", func() {
			err := b.Load(ctx, "tools/ninja")
			So(err, ShouldBeNil)
			pkgs, err := b.BuildAll(ctx)
			So(err, ShouldBeNil)

			verifySource(t, pkgs, &core.Action_Metadata{
				Cipd: &core.Action_Metadata_CIPD{
					Name:    "mock/sources/git/github.com/ninja-build/ninja",
					Version: "3@git-tag",
				},
			})
		})

		Convey("url source", func() {
			err := b.Load(ctx, "static_libs/curl")
			So(err, ShouldBeNil)
			pkgs, err := b.BuildAll(ctx)
			So(err, ShouldBeNil)

			verifySource(t, pkgs, &core.Action_Metadata{
				Cipd: &core.Action_Metadata_CIPD{
					Name:    "mock/sources/url/static_libs/curl/" + cipdPlatform,
					Version: "3@7.59.0",
				},
			})
		})

		Convey("script source", func() {
			err := b.Load(ctx, "tools/go")
			So(err, ShouldBeNil)
			pkgs, err := b.BuildAll(ctx)
			So(err, ShouldBeNil)

			verifySource(t, pkgs, &core.Action_Metadata{
				Cipd: &core.Action_Metadata_CIPD{
					Name:    "mock/sources/script/tools/go/" + cipdPlatform,
					Version: "3@script-version",
				},
			})
		})
	})
}

func verifySource(t *testing.T, pkgs []actions.Package, metadata *core.Action_Metadata) {
	t.Helper()
	pkg := pkgs[len(pkgs)-1]
	name := fmt.Sprintf("%s_source", pkg.Derivation.Name)
	for _, p := range pkg.BuildDependencies {
		if p.Derivation.Name == name {
			So(p.Action.Metadata, assertions.ShouldResembleProto, metadata)
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
func (*MockSourceResolver) ResolveScriptSource(cipdHostPlatform, dir string, script *spec.ScriptSource) (spec.ScriptSourceInfo, error) {
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
