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
	//go:embed resources/linux
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
	src, srcsEnv, err := g.fetchSource()
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
	}

	containers := containers(ctx.Platforms.Host)
	if containers == "" {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, fmt.Errorf("containers not available for %s", ctx.Platforms.Host)
	}

	var deps []utilities.BaseDependency
	deps = append(deps, g.Dependencies...)
	deps = append(deps,
		utilities.BaseDependency{Type: cipkg.DepsBuildHost, Generator: src},
		utilities.BaseDependency{Type: cipkg.DepsBuildHost, Generator: setup},
		utilities.BaseDependency{Type: cipkg.DepsBuildHost, Generator: common.Stdenv},
		utilities.BaseDependency{Type: cipkg.DepsBuildHost, Generator: common.Git},
		utilities.BaseDependency{Type: cipkg.DepsBuildHost, Generator: common.Python3},
		utilities.BaseDependency{Type: cipkg.DepsBuildHost, Generator: common.PosixUtils},
		utilities.BaseDependency{Type: cipkg.DepsBuildHost, Generator: common.Docker},
	)

	base := &utilities.BaseGenerator{
		Name:    g.Name,
		Builder: "{{.stdenv_python3}}/bin/python3",
		Args:    []string{"-I", "-B", "{{.stdenv}}/setup/main.py"},
		Env: append([]string{
			"buildFlags=",
			"installFlags=",
			srcsEnv,
			fmt.Sprintf("dockerImage=%s", containers),
		}, g.Env...),
		Dependencies: deps,
		CacheKey:     g.CacheKey,
		Version:      g.Version,
	}
	return base.Generate(ctx)
}
