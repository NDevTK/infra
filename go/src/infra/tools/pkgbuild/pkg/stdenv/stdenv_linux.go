// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"embed"
	"fmt"

	"infra/libs/cipkg"
	"infra/libs/cipkg/utilities"
)

var (
	//go:embed resources/setup_linux.py
	//go:embed resources/bin/realpath
	setupFiles embed.FS
	setup      cipkg.Generator
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

func (g *Generator) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	src, err := g.fetchSource()
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
	}

	containers := containers(ctx.Platforms.Host)
	if containers == "" {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, fmt.Errorf("containers not available for %s", ctx.Platforms.Host)
	}

	base := &utilities.BaseGenerator{
		Name:    g.Name,
		Builder: "{{.stdenv_python3}}/bin/python3",
		Args:    []string{"-I", "-B", "{{.setup}}/setup_linux.py", "{{.stdenv}}"},
		Env: append([]string{
			"buildFlags=",
			"installFlags=",
			fmt.Sprintf("srcs={{.%s_source}}", g.Name),
			fmt.Sprintf("dockerImage=%s", containers),
		}, g.Env...),
		Dependencies: append([]utilities.BaseDependency{
			{Type: cipkg.DepsBuildHost, Generator: src},
			{Type: cipkg.DepsBuildHost, Generator: common.Stdenv},
			{Type: cipkg.DepsBuildHost, Generator: common.Git},
			{Type: cipkg.DepsBuildHost, Generator: common.Python3},
			{Type: cipkg.DepsBuildHost, Generator: common.PosixUtils},
			{Type: cipkg.DepsBuildHost, Generator: common.Docker},
			{Type: cipkg.DepsBuildHost, Generator: setup},
		}, g.Dependencies...),
	}
	return base.Generate(ctx)
}
