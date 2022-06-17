// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package builtins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"infra/libs/cipkg"
)

// Import is used to import file/directory from host environment. The builder
// itself won't detect the change of the imported file/directory. A version
// string should be generated to indicate the change if it matters.
const ImportBuilder = BuiltinBuilderPrefix + "import"

const (
	ImportNormalFile = iota
	ImportExecutable
	ImportDirectory
)

type Import struct {
	Name    string
	Path    string
	Version string
	Target  string

	Type int
}

func (i *Import) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	s, err := json.Marshal(i)
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, fmt.Errorf("failed to encode import: %#v: %w", i, err)
	}
	return cipkg.Derivation{
		Name:    i.Name,
		Builder: ImportBuilder,
		Args:    []string{string(s)},
	}, cipkg.PackageMetadata{}, nil
}

func importFromHost(ctx context.Context, cmd *exec.Cmd) error {
	// cmd.Args = ["builtin:import", Import{...}]
	if len(cmd.Args) != 2 {
		return fmt.Errorf("invalid arguments: %v", cmd.Args)
	}
	out := GetEnv("out", cmd.Env)

	var i Import
	if err := json.Unmarshal([]byte(cmd.Args[1]), &i); err != nil {
		return fmt.Errorf("failed to decode import: %#v: %w", cmd.Args[1], err)
	}

	subdir := filepath.Join(out, i.Target)
	if err := os.MkdirAll(subdir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %#v: %w", subdir, err)
	}
	var newname string
	switch i.Type {
	case ImportNormalFile:
		newname = filepath.Join(subdir, filepath.Base(i.Path))
	case ImportExecutable:
		// TODO: For windows we need to create a bat file instead for dll searching
		newname = filepath.Join(subdir, filepath.Base(i.Path))
	case ImportDirectory:
		if err := os.Remove(subdir); err != nil {
			return fmt.Errorf("failed to remove output dir: %w", err)
		}
		newname = subdir
	}

	if err := os.Symlink(i.Path, newname); err != nil {
		return fmt.Errorf("failed to symlink import: %#v: %w", i, err)
	}
	return nil
}

var importFromPathMap = make(map[string]struct {
	g   cipkg.Generator
	err error
})

// FromHost(bin) is a wrapper for builtins.Import generator. It finds binaries
// in the PATH environment and caches the result.
func FromPath(bin string) (cipkg.Generator, error) {
	ret, ok := importFromPathMap[bin]
	if ok {
		return ret.g, ret.err
	}

	ret.g, ret.err = func() (cipkg.Generator, error) {
		path, err := exec.LookPath(bin)
		if err != nil {
			return nil, fmt.Errorf("failed to find binary: %s: %w", bin, err)
		}
		return &Import{
			Name: bin,
			Path: path,
			Type: ImportExecutable,
		}, nil
	}()

	importFromPathMap[bin] = ret
	return ret.g, ret.err
}
