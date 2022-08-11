// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package python

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/vpython_ng/pkg/common"

	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/common/errors"
)

type Environment struct {
	Executable string
	CPython    cipkg.Generator
	Virtualenv cipkg.Generator
}

func CPython3FromCIPD(version string) cipkg.Generator {
	return &builtins.CIPDEnsure{
		Name: "cpython",
		Ensure: ensure.File{
			PackagesBySubdir: map[string]ensure.PackageSlice{
				"": {
					{PackageTemplate: "infra/3pp/tools/cpython3/${platform}", UnresolvedVersion: version},
				},
			},
		},
	}
}

func VirtualenvFromCIPD(version string) cipkg.Generator {
	return &builtins.CIPDEnsure{
		Name: "virtualenv",
		Ensure: ensure.File{
			PackagesBySubdir: map[string]ensure.PackageSlice{
				"": {
					{PackageTemplate: "infra/3pp/tools/virtualenv", UnresolvedVersion: version},
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
		Builder: common.Python("{{.cpython}}", e.Executable),
		Args:    []string{"-c", pythonVenvBootstrapScript},
		Dependencies: []utilities.BaseDependency{
			{Type: cipkg.DepsHostTarget, Generator: e.CPython, Runtime: true},
			{Type: cipkg.DepsHostTarget, Generator: e.Virtualenv},
		},
	}
	return &utilities.BaseGenerator{
		Name:    "python_pep425tags",
		Builder: common.PythonVENV("{{.python_venv}}", e.Executable),
		Args:    []string{"-c", pythonPep425TagsScript},
		Dependencies: []utilities.BaseDependency{
			{Type: cipkg.DepsHostTarget, Generator: empty},
		},
	}
}

//go:embed bootstrap.py
var pythonVenvBootstrapScript string

func (e *Environment) WithWheels(wheels cipkg.Generator) cipkg.Generator {
	return &utilities.BaseGenerator{
		Name:    "python_venv",
		Builder: common.Python("{{.cpython}}", e.Executable),
		Args:    []string{"-c", pythonVenvBootstrapScript},
		Dependencies: []utilities.BaseDependency{
			{Type: cipkg.DepsHostTarget, Generator: e.CPython, Runtime: true},
			{Type: cipkg.DepsHostTarget, Generator: e.Virtualenv},
			{Type: cipkg.DepsHostTarget, Generator: wheels},
		},
		Env: []string{
			"wheels={{.wheels}}",
		},
	}
}

func CPythonFromRelativePath(subdir, cipdName string) (cipkg.Generator, error) {
	path, err := os.Executable()
	if err != nil {
		return nil, errors.Annotate(err, "failed to get executable").Err()
	}
	cpythonDir := filepath.Join(filepath.Dir(path), subdir)
	v, err := os.Open(filepath.Join(cpythonDir, ".versions", fmt.Sprintf("%s.cipd_version", cipdName)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Reason("Bundled Python %s not found. Use VPYTHON_BYPASS if prebuilt cpython not available on this platform.", subdir).Err()
		}
		return nil, errors.Annotate(err, "failed to open version file").Err()
	}
	defer v.Close()
	version, err := io.ReadAll(v)
	if err != nil {
		return nil, errors.Annotate(err, "failed to read version file").Err()
	}
	return &builtins.Import{
		Name: "cpython",
		Targets: []builtins.ImportTarget{
			{Source: cpythonDir, Version: string(version), Type: builtins.ImportDirectory},
		},
	}, nil
}
