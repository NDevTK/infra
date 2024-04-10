// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os/exec"
	"path"
	"path/filepath"

	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/cipkg/base/workflow"
	"go.chromium.org/luci/cipkg/core"
	"go.chromium.org/luci/common/system/environ"
)

var (
	//go:embed all:setup
	stdenvEmbed embed.FS
	stdenvGen   = generators.InitEmbeddedFS(
		"stdenv", stdenvEmbed,
	)

	//go:embed all:resources
	resourcesEmbed embed.FS
	resourcesGen   = generators.InitEmbeddedFS(
		"setup", resourcesEmbed,
	).SubDir("resources")

	baseByOS = map[string][]generators.Generator{}
)

const (
	cipdVersionGit     = "version:2@2.42.0.chromium.11"
	cipdVersionCPython = "version:2@3.8.10.chromium.26"
)

var (
	git = &generators.CIPDExport{
		Name: "stdenv_git",
		Ensure: ensure.File{
			PackagesBySubdir: map[string]ensure.PackageSlice{
				"": {
					{PackageTemplate: "infra/3pp/tools/git/${platform}", UnresolvedVersion: cipdVersionGit},
				},
			},
		},
	}
	cpython = &generators.CIPDExport{
		Name: "stdenv_python3",
		Ensure: ensure.File{
			PackagesBySubdir: map[string]ensure.PackageSlice{
				"": {
					{PackageTemplate: "infra/3pp/tools/cpython3/${platform}", UnresolvedVersion: cipdVersionCPython},
				},
			},
		},
	}
)

type Config struct {
	XcodeDeveloper generators.Generator
	WinSDK         generators.Generator
	BuildPlatform  generators.Platform

	// If true, pull docker image(s) during Init. Docker may require gcloud and
	// other credentials to access private repositories, which are not available
	// inside the derivation.
	DockerPullImage bool

	// For mock testing. Required binaries may not available in test environment.
	FindBinary generators.FindBinaryFunc
}

func DefaultConfig() *Config {
	return &Config{
		BuildPlatform:   generators.CurrentPlatform(),
		DockerPullImage: true,
	}
}

// Initialize the stdenv. If finder is nil, binaries will be imported from
// PATH.
func Init(cfg *Config) error {
	os := cfg.BuildPlatform.OS()
	if _, ok := baseByOS[os]; ok {
		return nil
	}
	var base []generators.Generator

	// Embedded files
	base = append(base,
		stdenvGen,
		resourcesGen.SubDir(os).WithModeOverride(func(name string) (fs.FileMode, error) {
			if path.Dir(name) == "bin" {
				// -r-xr-xr-x
				return 0o555, nil
			}
			// -r--r--r--
			return 0o444, nil
		}),
	)

	// Prebuilt binaries
	base = append(base, git, cpython)

	posixUtils := []string{
		"awk",
		"basename",
		"bash",
		"cat",
		"cut",
		"chmod",
		"cmp",
		"cp",
		"date",
		"dirname",
		"echo",
		"env",
		"expr",
		"false",
		"file",
		"find",
		"grep",
		"gzip",
		"head",
		"hostname",
		"id",
		"install",
		"ls",
		"mkdir",
		"mktemp",
		"mv",
		"ln",
		"od",
		"patch",
		"perl",
		"ps",
		"rm",
		"rmdir",
		"sh",
		"sleep",
		"sort",
		"tail",
		"tar",
		"touch",
		"tr",
		"true",
		"uniq",
		"wc",
		"which",
		"uname",
	}

	// OS specified
	gs, err := func() ([]generators.Generator, error) {
		ctx := context.Background()
		switch os {
		case "linux":
			posixUtils = append(posixUtils, "cpio", "egrep", "fgrep")
			return importLinux(cfg, posixUtils...)
		case "darwin":
			posixUtils = append(posixUtils, "cpio", "egrep", "fgrep")
			return importDarwin(ctx, cfg, posixUtils...)
		case "windows":
			posixUtils = append(posixUtils, "cygpath", "nproc")
			return importWindows(cfg, posixUtils...)
		default:
			return nil, fmt.Errorf("unknown os: %s", os)
		}
	}()
	if err != nil {
		return err
	}
	base = append(base, gs...)

	baseByOS[os] = base

	if os == "linux" && cfg.DockerPullImage {
		for _, c := range containers {
			if out, err := exec.Command("docker", "pull", c).CombinedOutput(); err != nil {
				fmt.Println(string(out))
				return fmt.Errorf("pull docker images failed: %w", err)
			}
		}
	}

	return nil
}

type Generator struct {
	Name         string
	Source       Source
	Env          environ.Env
	Dependencies []generators.Dependency

	CIPDName string
	Version  string
}

func (g *Generator) Generate(ctx context.Context, plats generators.Platforms) (*core.Action, error) {
	src, srcsEnv, err := g.fetchSource()
	if err != nil {
		return nil, err
	}

	deps := append([]generators.Dependency{
		{Type: generators.DepsBuildHost, Generator: src},
	}, g.Dependencies...)
	for _, g := range baseByOS[plats.Build.OS()] {
		deps = append(deps, generators.Dependency{
			Type:      generators.DepsBuildHost,
			Generator: g,
		})
	}

	env := g.Env.Clone()
	env.Set("buildFlags", "")
	env.Set("installFlags", "")
	env.SetEntry(srcsEnv)
	py3 := filepath.Join("{{.stdenv_python3}}", "bin", "python3")
	if plats.Build.OS() == "windows" {
		py3 += ".exe"
	}
	tmpl := &workflow.Generator{
		Name: g.Name,
		Metadata: &core.Action_Metadata{
			Cipd: &core.Action_Metadata_CIPD{
				Name:    g.CIPDName,
				Version: g.Version,
			},
		},
		Args:         []string{py3, "-I", "-B", filepath.Join("{{.stdenv}}", "setup", "main.py")},
		Env:          env,
		Dependencies: deps,
	}

	switch plats.Build.OS() {
	case "linux":
		if err := g.generateLinux(plats, tmpl); err != nil {
			return nil, err
		}
	case "darwin":
		if err := g.generateDarwin(plats, tmpl); err != nil {
			return nil, err
		}
	case "windows":
		if err := g.generateWindows(plats, tmpl); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown build os: %s", plats.Build.OS())
	}

	return tmpl.Generate(ctx, plats)
}
