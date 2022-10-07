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

type ImportTarget struct {
	Source      string
	Destination string
	Version     string
	Type        int
}

type Import struct {
	Name    string
	Targets []ImportTarget
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

	for _, t := range i.Targets {
		subdir := filepath.Join(out, t.Destination)
		if err := os.MkdirAll(subdir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory: %#v: %w", subdir, err)
		}
		var newname string
		switch t.Type {
		case ImportNormalFile:
			newname = filepath.Join(subdir, filepath.Base(t.Source))
		case ImportExecutable:
			// TODO: For windows we need to create a bat file instead for dll searching
			newname = filepath.Join(subdir, filepath.Base(t.Source))
		case ImportDirectory:
			if err := os.Remove(subdir); err != nil {
				return fmt.Errorf("failed to remove output dir: %w", err)
			}
			newname = subdir
		}

		if err := os.Symlink(t.Source, newname); err != nil {
			return fmt.Errorf("failed to symlink import: %#v: %w", i, err)
		}
	}

	if err := os.MkdirAll(filepath.Join(out, "build-support"), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create build-support: %w", err)
	}
	f, err := os.Create(filepath.Join(out, "build-support", "builtin_import.stamp"))
	if err != nil {
		return fmt.Errorf("failed to touch import stamp: %w", err)
	}
	f.Close()

	return nil
}

var importFromPathMap = make(map[string]struct {
	target *ImportTarget
	err    error
})

// FindBinaryFunc should return a path for the provided binary name.
// e.g. exec.LookPath searches the binary in the PATH.
type FindBinaryFunc func(bin string) (path string, err error)

// FromPathBatch(...) is a wrapper for builtins.Import generator. It finds
// binaries using finder func and caches the result based on the name. if
// finder is nil, binaries will be searched from the PATH environment.
func FromPathBatch(name string, finder FindBinaryFunc, bins ...string) (cipkg.Generator, error) {
	if finder == nil {
		finder = exec.LookPath
	}

	i := &Import{Name: name}
	for _, bin := range bins {
		ret, ok := importFromPathMap[bin]
		if !ok {
			ret.target, ret.err = func() (*ImportTarget, error) {
				path, err := finder(bin)
				if err != nil {
					return nil, fmt.Errorf("failed to find binary: %s: %w", bin, err)
				}
				return &ImportTarget{
					Source:      path,
					Destination: "bin",
					Type:        ImportExecutable,
				}, nil
			}()

			importFromPathMap[bin] = ret
		}

		if ret.err != nil {
			return nil, ret.err
		}
		i.Targets = append(i.Targets, *ret.target)
	}
	return i, nil
}
