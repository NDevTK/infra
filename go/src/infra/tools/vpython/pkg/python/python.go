// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package python

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"infra/tools/vpython/pkg/common"

	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/cipkg/base/workflow"
	"go.chromium.org/luci/common/errors"
)

type Environment struct {
	Executable string
	CPython    generators.Generator
	Virtualenv generators.Generator
}

func CPython3FromCIPD(version string) generators.Generator {
	return &generators.CIPDExport{
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

func VirtualenvFromCIPD(version string) generators.Generator {
	return &generators.CIPDExport{
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

//go:embed bootstrap.py pep425tags.py
var bootstrapEmbed embed.FS
var bootstrapGen = generators.InitEmbeddedFS("bootstrap", bootstrapEmbed)

func (e *Environment) Pep425Tags() generators.Generator {
	// Generate an empty virtual environment to probe the pep425tags
	empty := &workflow.Generator{
		Name: "python_venv",
		Args: []string{
			common.Python("{{.cpython}}", e.Executable),
			filepath.Join("{{.bootstrap}}", "bootstrap.py"),
		},
		Dependencies: []generators.Dependency{
			{Type: generators.DepsHostTarget, Generator: e.CPython, Runtime: true},
			{Type: generators.DepsHostTarget, Generator: e.Virtualenv},
			{Type: generators.DepsHostTarget, Generator: bootstrapGen},
		},
	}
	return &workflow.Generator{
		Name: "python_pep425tags",
		Args: []string{
			common.PythonVENV("{{.python_venv}}", e.Executable),
			filepath.Join("{{.bootstrap}}", "pep425tags.py"),
		},
		Dependencies: []generators.Dependency{
			{Type: generators.DepsHostTarget, Generator: empty},
			{Type: generators.DepsHostTarget, Generator: bootstrapGen},
		},
	}
}

func (e *Environment) WithWheels(wheels generators.Generator) generators.Generator {
	return &workflow.Generator{
		Name: "python_venv",
		Args: []string{
			common.Python("{{.cpython}}", e.Executable),
			"-BssE",
			filepath.Join("{{.bootstrap}}", "bootstrap.py"),
		},
		Dependencies: []generators.Dependency{
			{Type: generators.DepsHostTarget, Generator: e.CPython, Runtime: true},
			{Type: generators.DepsHostTarget, Generator: e.Virtualenv},
			{Type: generators.DepsHostTarget, Generator: wheels},
			{Type: generators.DepsHostTarget, Generator: bootstrapGen},
		},
	}
}

func CPythonFromPath(dir, cipdName string) (generators.Generator, error) {
	cpythonDir := dir
	if !filepath.IsAbs(dir) {
		path, err := os.Executable()
		if err != nil {
			return nil, errors.Annotate(err, "failed to get executable").Err()
		}
		if runtime.GOOS != "windows" {
			if path, err = filepath.EvalSymlinks(path); err != nil {
				return nil, errors.Annotate(err, "failed to eval symlink to executable").Err()
			}
		}
		cpythonDir = filepath.Join(filepath.Dir(path), dir)
	}

	v, err := os.Open(filepath.Join(cpythonDir, ".versions", fmt.Sprintf("%s.cipd_version", cipdName)))
	if err != nil {
		return nil, errors.Annotate(err, "Bundled Python %s not found. Use VPYTHON_BYPASS if prebuilt cpython not available on this platform", dir).Err()
	}
	defer v.Close()
	version, err := io.ReadAll(v)
	if err != nil {
		return nil, errors.Annotate(err, "failed to read version file").Err()
	}
	return &generators.ImportTargets{
		Name: "cpython",
		Targets: map[string]generators.ImportTarget{
			".": {Source: cpythonDir, Version: string(version), Mode: fs.ModeDir, FollowSymlinks: true},
		},
	}, nil
}
