// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package spec

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"

	"google.golang.org/protobuf/encoding/prototext"
)

// Run `protoc -I../recipes --go_out=src ../recipes/recipe_modules/support_3pp/spec.proto`
// from infra/go to generate code from 3pp spec proto.

type PackageDef struct {
	// package name is the raw package directory name. It shouldn't be used
	// directly since:
	// 1. It can be overridden by pkg_name_override in the spec
	// 2. We should always use a package's full name for referencing.
	packageName string

	Spec *Spec
	Dir  fs.FS
}

var validDerivationNameChar = regexp.MustCompile("^([0-9])|[^a-zA-Z0-9_]")

// DerivationName is a valid derivation name for using inside the pkgbuild.
func (p *PackageDef) DerivationName() string {
	// 1. Prepend the first character with _ if it's a number.
	// 2. Replace all characters other than numbers and alphabets with _.
	return validDerivationNameChar.ReplaceAllString(p.packageName, "_$1")
}

// FullName is the package's name constructed by <pkg_prefix>/<package_name>.
// This is used for referring other packages in the specs.
func (p *PackageDef) FullName() string {
	upload := p.Spec.GetUpload()
	if upload == nil {
		return p.packageName
	}
	return path.Join(upload.PkgPrefix, p.packageName)
}

// FullNameWithOverride is the package's name constructed by
// <pkg_prefix>/<package_name> with possible override from upload.
// This is used for uploading or retrieving packages from remote.
func (p *PackageDef) FullNameWithOverride() string {
	upload := p.Spec.GetUpload()
	if upload == nil {
		return p.packageName
	}
	name := upload.GetPkgNameOverride()
	if name == "" {
		name = p.packageName
	}
	return path.Join(upload.PkgPrefix, name)
}

func (p *PackageDef) CIPDPath(prefix, host string) string {
	u := path.Join(prefix, p.FullNameWithOverride())
	if !p.Spec.GetUpload().GetUniversal() {
		u = path.Join(u, host)
	}
	return u
}

func LoadPackageDef(name string, dir fs.FS) (*PackageDef, error) {
	f, err := dir.Open("3pp.pb")
	if err != nil {
		return nil, fmt.Errorf("failed to open 3pp spec: %w", err)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read 3pp spec: %w", err)
	}

	var spec Spec
	if err := prototext.Unmarshal(b, &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal 3pp spec: %w", err)
	}

	return &PackageDef{
		packageName: name,
		Spec:        &spec,
		Dir:         dir,
	}, nil
}

func FindPackageDefs(dir fs.FS) (defs []*PackageDef, err error) {
	err = fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.Name() != "3pp.pb" {
			return nil
		}

		// There are two common hierarchies:
		// - /path/to/pkg/3pp.pb
		// - /path/to/pkg/3pp/3pp.pb
		pkgPath := filepath.Dir(path)
		parent, name := filepath.Split(pkgPath)
		if name == "3pp" {
			name = filepath.Base(parent)
		}

		if name == "." || name == string(filepath.Separator) {
			return fmt.Errorf("invalid package: %s", path)
		}

		pkgDir, err := fs.Sub(dir, pkgPath)
		if err != nil {
			return err
		}
		def, err := LoadPackageDef(name, pkgDir)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", path, err)
		}

		defs = append(defs, def)
		return nil
	})
	return
}
