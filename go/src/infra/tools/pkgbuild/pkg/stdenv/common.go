// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"embed"
	"fmt"
	"io/fs"
	"os/exec"
	"path"

	"go.chromium.org/luci/cipd/client/cipd/ensure"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
)

var (
	//go:embed all:setup
	stdenv embed.FS

	//go:embed all:resources
	resources embed.FS

	baseByOS = map[string][]cipkg.Generator{}
)

const (
	cipdVersionGit     = "version:2@2.36.1.chromium.8"
	cipdVersionCPython = "version:2@3.8.10.chromium.24"
)

var (
	git = &builtins.CIPDExport{
		Name: "stdenv_git",
		Ensure: ensure.File{
			PackagesBySubdir: map[string]ensure.PackageSlice{
				"": {
					{PackageTemplate: "infra/3pp/tools/git/${platform}", UnresolvedVersion: cipdVersionGit},
				},
			},
		},
	}
	cpython = &builtins.CIPDExport{
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
	XcodeDeveloper cipkg.Generator
	WinSDK         cipkg.Generator
	FindBinary     builtins.FindBinaryFunc
	BuildPlatform  *utilities.Platform
}

func DefaultConfig() *Config {
	return &Config{
		FindBinary:    exec.LookPath,
		BuildPlatform: utilities.CurrentPlatform(),
	}
}

// Initialize the stdenv. If finder is nil, binaries will be imported from
// PATH.
func Init(cfg *Config) error {
	os := cfg.BuildPlatform.OS()
	if _, ok := baseByOS[os]; ok {
		return nil
	}
	var base []cipkg.Generator

	// Prebuilt binaries
	files, err := fs.Sub(resources, path.Join("resources", os))
	if err != nil {
		return err
	}
	base = append(base,
		&builtins.CopyFiles{
			Name:  "stdenv",
			Files: stdenv,
		},
		&builtins.CopyFiles{
			Name: "setup",
			Files: builtins.FSWithMode{
				FS: files,
				ModeOverride: func(info fs.FileInfo) (fs.FileMode, error) {
					if path.Dir(info.Name()) == "bin" {
						return info.Mode() | fs.ModePerm, nil
					}
					return info.Mode(), nil
				},
			},
		},
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
	gs, err := func() ([]cipkg.Generator, error) {
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
	Env          []string
	Dependencies []utilities.BaseDependency

	CacheKey string
	Version  string
}

func (g *Generator) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	src, srcsEnv, err := g.fetchSource()
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
	}

	deps := append([]utilities.BaseDependency{
		{Type: cipkg.DepsBuildHost, Generator: src},
	}, g.Dependencies...)
	for _, g := range baseByOS[ctx.Platforms.Build.OS()] {
		deps = append(deps, utilities.BaseDependency{
			Type:      cipkg.DepsBuildHost,
			Generator: g,
		})
	}

	tmpl := &utilities.BaseGenerator{
		Name:    g.Name,
		Builder: "{{.stdenv_python3}}/bin/python3",
		Args:    []string{"-I", "-B", "{{.stdenv}}/setup/main.py"},
		Env: append([]string{
			"buildFlags=",
			"installFlags=",
			srcsEnv,
		}, g.Env...),
		Dependencies: deps,
		CacheKey:     g.CacheKey,
		Version:      g.Version,
	}

	switch ctx.Platforms.Build.OS() {
	case "linux":
		if err := g.generateLinux(ctx, tmpl); err != nil {
			return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
		}
	case "darwin":
		if err := g.generateDarwin(ctx, tmpl); err != nil {
			return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
		}
	case "windows":
		if err := g.generateWindows(ctx, tmpl); err != nil {
			return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
		}
	default:
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, fmt.Errorf("unknown build os: %s", ctx.Platforms.Build.OS())
	}

	return tmpl.Generate(ctx)
}
