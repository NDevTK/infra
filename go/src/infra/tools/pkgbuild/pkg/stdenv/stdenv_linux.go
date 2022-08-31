// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"embed"
	"fmt"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
)

var (
	//go:embed setup_linux.py
	setupLinuxFiles embed.FS
	setupLinux      = &builtins.CopyFiles{
		Name:  "setup_linux",
		Files: setupLinuxFiles,
	}
)

// Return the dockcross image for the platform.
// TODO(fancl): build the container using pkgbuild.
func containers(plat cipkg.Platform) string {
	const prefix = "gcr.io/chromium-container-registry/infra-dockerbuild/"
	const version = ":v1.4.18"
	if plat.OS() != "linux" {
		return ""
	}
	switch plat.Arch() {
	case "amd64":
		return prefix + "manylinux-x64-py3" + version
	case "arm64":
		return prefix + "linux-arm64-py3" + version
	case "arm":
		return prefix + "linux-armv6-py3" + version
	default:
		return ""
	}
}

func (g *Generator) fetchSource() (gen cipkg.Generator, envs []string, err error) {
	name := fmt.Sprintf("%s_source", g.Name)
	switch s := g.Source.(type) {
	case *SourceGit:
		const gitCommand = `cd "${out}" && "$0" clone "$1" src && cd src && "$0" checkout "$2" `
		return &utilities.BaseGenerator{
				Name:    name,
				Builder: "{{.posixUtils_import}}/bin/bash",
				Args:    []string{"-c", gitCommand, "{{.stdenv_git}}/bin/git", s.URL, s.Ref},
				Dependencies: append([]utilities.BaseDependency{
					{Type: cipkg.DepsBuildHost, Generator: common.PosixUtils},
					{Type: cipkg.DepsBuildHost, Generator: common.Git},
				}),
			}, []string{
				// We don't need unpacking the source for git
				"skipUnpack=1",
				fmt.Sprintf("sourceRoot={{.%s_source}}/src", g.Name),
			}, nil
	case *SourceURL:
		return &builtins.FetchURL{
				Name:          name,
				URL:           s.URL,
				Filename:      s.Filename,
				HashAlgorithm: s.HashAlgorithm,
				HashString:    s.HashString,
			}, []string{
				fmt.Sprintf("srcs={{.%s_source}}", g.Name),
			}, nil
	default:
		return nil, nil, fmt.Errorf("unknown source type %#v:", s)
	}
}

func (g *Generator) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	src, srcEnvs, err := g.fetchSource()
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
	}

	containers := containers(ctx.Platforms.Host)
	if containers == "" {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, fmt.Errorf("containers not available for %s", ctx.Platforms.Host)
	}

	envs := []string{
		"buildFlags=",
		"installFlags=",
		fmt.Sprintf("dockerImage=%s", containers),
	}
	envs = append(envs, srcEnvs...)
	envs = append(envs, g.Env...)

	base := &utilities.BaseGenerator{
		Name:    g.Name,
		Builder: "{{.stdenv_python3}}/bin/python3",
		Args:    []string{"-I", "-B", "{{.setup_linux}}/setup_linux.py", "{{.stdenv}}"},
		Env:     envs,
		Dependencies: append([]utilities.BaseDependency{
			{Type: cipkg.DepsBuildHost, Generator: src},
			{Type: cipkg.DepsBuildHost, Generator: common.Stdenv},
			{Type: cipkg.DepsBuildHost, Generator: common.PosixUtils},
			{Type: cipkg.DepsBuildHost, Generator: common.Docker},
			{Type: cipkg.DepsBuildHost, Generator: common.Git},
			{Type: cipkg.DepsBuildHost, Generator: common.Python3},
			{Type: cipkg.DepsBuildHost, Generator: setupLinux},
		}, g.Dependencies...),
	}
	return base.Generate(ctx)
}
