// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"crypto"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"

	"go.chromium.org/luci/cipd/client/cipd/ensure"
)

// TODO(fancl): Use all:setup after go 1.18.
//
//go:embed setup/*
var stdenv embed.FS

// Initialize resources defined in each platforms.
func init() {
	files, err := fs.Sub(setupFiles, filepath.Join("resources", runtime.GOOS))
	if err != nil {
		panic(err)
	}
	setup = &builtins.CopyFiles{
		Name:  "setup",
		Files: files,
		FileMode: func(f fs.File) (fs.FileMode, error) {
			return fs.ModePerm, nil
		},
	}
}

var common struct {
	// Static files
	Stdenv cipkg.Generator

	// Prebuilt binaries
	Git     cipkg.Generator
	Python3 cipkg.Generator

	// Import from host environment
	PosixUtils cipkg.Generator
	Docker     cipkg.Generator
	XCode      cipkg.Generator
}

var cipdPackages = []ensure.PackageDef{}

const (
	cipdVersionGit     = "version:2@2.36.1.chromium.8"
	cipdVersionCPython = "version:2@3.8.10.chromium.24"
)

// Initialize the stdenv. If finder is nil, binaries will be imported from
// PATH.
func Init(finder builtins.FindBinaryFunc) (err error) {
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

	if common.PosixUtils, err = builtins.FromPathBatch("posixUtils_import", finder,
		"awk",
		"basename",
		"bash",
		"cat",
		"cut",
		"chmod",
		"cp",
		"expr",
		"file",
		"find",
		"grep",
		"id",
		"ls",
		"mkdir",
		"mktemp",
		"mv",
		"od",
		"perl",
		"rm",
		"sed",
		"sh",
		"sleep",
		"sort",
		"touch",
		"tr",
		"which",
		"uname",
	); err != nil {
		return
	}

	// OS specified
	switch runtime.GOOS {
	case "linux":
		if common.Docker, err = builtins.FromPathBatch("docker_import", finder, "docker"); err != nil {
			return
		}
	case "darwin":
		if common.XCode, err = builtins.FromPathBatch("xcode", finder,
			"ar",
			"cc",
			"c++",
			"clang",
			"clang++",
			"gcc",
			"g++",
			"make",
			"xcrun",
		); err != nil {
			return
		}
	}

	return
}

type Generator struct {
	Name         string
	Source       Source
	Env          []string
	Dependencies []utilities.BaseDependency

	CacheKey string
	Version  string
}

func (g *Generator) fetchSource() (cipkg.Generator, string, error) {
	// The name of the source derivation. It's also used in environment variable
	// srcs to pointing to the location of source file(s), which will be expanded
	// to absolute path by utilities.BaseGenerator.
	name := fmt.Sprintf("%s_source", g.Name)
	switch s := g.Source.(type) {
	case *SourceGit:
		const gitCommand = `cd "${out}" && "$0" clone "$1" src && cd src && "$0" checkout "$2" && rm -rf .git`
		return &utilities.BaseGenerator{
			Name:    name,
			Builder: "{{.posixUtils_import}}/bin/bash",
			Args:    []string{"-c", gitCommand, "{{.stdenv_git}}/bin/git", s.URL, s.Ref},
			Dependencies: append([]utilities.BaseDependency{
				{Type: cipkg.DepsBuildHost, Generator: common.PosixUtils},
				{Type: cipkg.DepsBuildHost, Generator: common.Git},
			}),
			Version:  s.Version,
			CacheKey: s.CacheKey,
		}, fmt.Sprintf("srcs={{.%s}}", name), nil
	case *SourceURLs:
		urls := builtins.FetchURLs{
			Name: name,
		}
		var srcs []string
		for _, u := range s.URLs {
			urls.URLs = append(urls.URLs, builtins.FetchURL{
				Name:          name,
				URL:           u.URL,
				Filename:      u.Filename,
				HashAlgorithm: u.HashAlgorithm,
				HashString:    u.HashString,
			})
			srcs = append(srcs, fmt.Sprintf("{{.%s}}/%s", name, u.Filename))
		}
		return &utilities.WithMetadata{
			Generator: &urls,
			Metadata: cipkg.PackageMetadata{
				Version:  s.Version,
				CacheKey: s.CacheKey,
			},
		}, fmt.Sprintf("srcs=%s", strings.Join(srcs, string(filepath.ListSeparator))), nil
	default:
		return nil, "", fmt.Errorf("unknown source type %#v:", s)
	}
}

type Source interface {
	isSourceMethod()
}

type SourceGit struct {
	// The url to the git repository. Support any protocol used by the git
	// command line interfaces.
	URL string

	// The reference for the git repo. Can be anything supported by git checkout.
	// Typical use cases:
	// - refs/tags/xxx
	// - 8e8722e14772727b0e1cd5bd925a0f089611a60b
	Ref string

	CacheKey string
	Version  string
}

func (s *SourceGit) isSourceMethod() {}

type SourceURL struct {
	URL           string
	Filename      string
	HashAlgorithm crypto.Hash
	HashString    string
}

type SourceURLs struct {
	URLs []SourceURL

	CacheKey string
	Version  string
}

func (s *SourceURLs) isSourceMethod() {}
