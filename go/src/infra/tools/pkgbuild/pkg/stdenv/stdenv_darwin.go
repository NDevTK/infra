// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"embed"

	"infra/libs/cipkg"
	"infra/libs/cipkg/utilities"
)

var (
	//go:embed resources/darwin
	setupFiles embed.FS
	setup      cipkg.Generator
)

func (g *Generator) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	src, srcsEnv, err := g.fetchSource()
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
	}

	base := &utilities.BaseGenerator{
		Name:    g.Name,
		Builder: "{{.stdenv_python3}}/bin/python3",
		Args:    []string{"-I", "-B", "{{.stdenv}}/setup/main.py"},
		Env: append([]string{
			"buildFlags=",
			"installFlags=",
			srcsEnv,

			// Env GREP and SED added here to skip the configure testing them.
			// TODO(fancl): Update the specs to include gnu grep/sed in the tools if
			// configure.ac expects gnu tools.
			"GREP={{.posixUtils_import}}/bin/grep",
			"SED={{.posixUtils_import}}/bin/sed",
		}, g.Env...),
		Dependencies: append([]utilities.BaseDependency{
			{Type: cipkg.DepsBuildHost, Generator: src},
			{Type: cipkg.DepsBuildHost, Generator: common.Stdenv},
			{Type: cipkg.DepsBuildHost, Generator: common.Git},
			{Type: cipkg.DepsBuildHost, Generator: common.Python3},
			{Type: cipkg.DepsBuildHost, Generator: common.PosixUtils},
			{Type: cipkg.DepsBuildHost, Generator: common.XCode},
			{Type: cipkg.DepsBuildHost, Generator: setup},
		}, g.Dependencies...),
		CacheKey: g.CacheKey,
		Version:  g.Version,
	}
	return base.Generate(ctx)
}
