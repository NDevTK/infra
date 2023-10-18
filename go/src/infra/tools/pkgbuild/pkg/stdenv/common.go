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
	cipdVersionGit     = "version:2@2.36.1.chromium.8"
	cipdVersionCPython = "version:2@3.8.10.chromium.24"
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
	FindBinary     generators.FindBinaryFunc
	BuildPlatform  generators.Platform
}

func DefaultConfig() *Config {
	return &Config{
		FindBinary:    exec.LookPath,
		BuildPlatform: generators.CurrentPlatform(),
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
		resourcesGen.SubDir(os).WithModeOverride(func(info fs.FileInfo) (fs.FileMode, error) {
			if path.Dir(info.Name()) == "bin" {
				return info.Mode() | fs.ModePerm, nil
			}
			return info.Mode(), nil
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
		switch os {
		case "linux":
			posixUtils = append(posixUtils, "cpio", "egrep", "fgrep")
			return importLinux(cfg, posixUtils...)
		case "darwin":
			posixUtils = append(posixUtils, "cpio", "egrep", "fgrep")
			return importDarwin(cfg, posixUtils...)
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
	tmpl := &workflow.Generator{
		Name: g.Name,
		Metadata: &core.Action_Metadata{
			Cipd: &core.Action_Metadata_CIPD{
				Name:    g.CIPDName,
				Version: g.Version,
			},
		},
		Args:         []string{"{{.stdenv_python3}}/bin/python3", "-I", "-B", "{{.stdenv}}/setup/main.py"},
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
