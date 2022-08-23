// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"crypto"
	"embed"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"

	"go.chromium.org/luci/cipd/client/cipd/ensure"
)

// TODO(fancl): Use all:setup after go 1.18.
//go:embed setup/*
var stdenv embed.FS

var common struct {
	// Static files
	Stdenv cipkg.Generator

	// Import from host environment
	PosixUtils cipkg.Generator
	Docker     cipkg.Generator

	// Prebuilt binaries
	Git     cipkg.Generator
	Python3 cipkg.Generator
}

var cipdPackages = []ensure.PackageDef{}

const (
	cipdVersionGit     = "version:2@2.36.1.chromium.8"
	cipdVersionCPython = "version:2@3.8.10.chromium.24"
)

func Init() (err error) {
	common.PosixUtils, err = builtins.FromPathBatch("posixUtils_import",
		"bash",
		"chmod",
		"cp",
		"file",
		"find",
		"grep",
		"id",
		"mkdir",
		"mktemp",
		"rm",
		"sed",
		"touch",
	)

	if common.Docker, err = builtins.FromPath("docker"); err != nil {
		return
	}

	common.Git = &builtins.CIPDEnsure{
		Name: "stdenv_git",
		Ensure: ensure.File{
			PackagesBySubdir: map[string]ensure.PackageSlice{
				"": {
					{PackageTemplate: "infra/3pp/tools/git/${platform}", UnresolvedVersion: cipdVersionGit},
				},
			},
		},
	}
	common.Python3 = &builtins.CIPDEnsure{
		Name: "stdenv_python3",
		Ensure: ensure.File{
			PackagesBySubdir: map[string]ensure.PackageSlice{
				"": {
					{PackageTemplate: "infra/3pp/tools/cpython3/${platform}", UnresolvedVersion: cipdVersionCPython},
				},
			},
		},
	}

	common.Stdenv = &builtins.CopyFiles{
		Name:  "stdenv",
		Files: stdenv,
	}

	return
}

type Generator struct {
	Name         string
	Source       Source
	Env          []string
	Dependencies []utilities.BaseDependency
}

type Source interface {
	isSourceMethod()
}

type SourceGit struct {
	Repository string
}

func (s *SourceGit) isSourceMethod() {}

type SourceURL struct {
	URL           string
	Filename      string
	HashAlgorithm crypto.Hash
	HashString    string
}

func (s *SourceURL) isSourceMethod() {}
