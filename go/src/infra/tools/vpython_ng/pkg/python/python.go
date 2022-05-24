// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package python

import (
	_ "embed"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/vpython_ng/pkg/common"

	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/hardcoded/chromeinfra"
)

type Versions struct {
	CPython    string
	VirtualENV string
}

type Environment struct {
	cpython    cipkg.Generator
	virtualenv cipkg.Generator
}

func NewEnvironment(v Versions) *Environment {
	return &Environment{
		cpython: &builtins.CIPDEnsure{
			Name: "cpython",
			Ensure: ensure.File{
				ServiceURL: chromeinfra.CIPDServiceURL,
				PackagesBySubdir: map[string]ensure.PackageSlice{
					"": {
						{PackageTemplate: "infra/3pp/tools/cpython3/${platform}", UnresolvedVersion: v.CPython},
					},
				},
			},
		},
		virtualenv: &builtins.CIPDEnsure{
			Name: "virtualenv",
			Ensure: ensure.File{
				ServiceURL: chromeinfra.CIPDServiceURL,
				PackagesBySubdir: map[string]ensure.PackageSlice{
					"": {
						{PackageTemplate: "infra/3pp/tools/virtualenv", UnresolvedVersion: v.VirtualENV},
					},
				},
			},
		},
	}
}

//go:embed pep425tags.py
var pythonPep425TagsScript string

func (e *Environment) Pep425Tags() cipkg.Generator {
	// Generate an empty virtual environment to probe the pep425tags
	empty := &utilities.BaseGenerator{
		Name:    "python_venv",
		Builder: common.Python3("{{.cpython}}"),
		Args:    []string{"-c", pythonVenvBootstrapScript},
		Dependencies: []cipkg.Dependency{
			{Type: cipkg.DepsHostTarget, Generator: e.cpython},
			{Type: cipkg.DepsHostTarget, Generator: e.virtualenv},
		},
	}
	return &utilities.BaseGenerator{
		Name:    "python_pep425tags",
		Builder: common.Python3VENV("{{.python_venv}}"),
		Args:    []string{"-c", pythonPep425TagsScript},
		Dependencies: []cipkg.Dependency{
			{Type: cipkg.DepsHostTarget, Generator: empty},
		},
	}
}

//go:embed bootstrap.py
var pythonVenvBootstrapScript string

func (e *Environment) WithWheels(wheels cipkg.Generator) cipkg.Generator {
	return &utilities.BaseGenerator{
		Name:    "python_venv",
		Builder: common.Python3("{{.cpython}}"),
		Args:    []string{"-c", pythonVenvBootstrapScript},
		Dependencies: []cipkg.Dependency{
			{Type: cipkg.DepsHostTarget, Generator: e.cpython},
			{Type: cipkg.DepsHostTarget, Generator: e.virtualenv},
			{Type: cipkg.DepsHostTarget, Generator: wheels},
		},
	}
}
