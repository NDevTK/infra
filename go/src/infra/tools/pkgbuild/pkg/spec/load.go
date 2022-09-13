// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package spec

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"google.golang.org/protobuf/encoding/prototext"
)

// Run `protoc -I../recipes --go_out=src ../recipes/recipe_modules/support_3pp/spec.proto`
// from infra/go to generate code from 3pp spec proto.

type PackageDef struct {
	Name string
	Spec *Spec
	Dir  fs.FS
}

func LoadPackageDef(dir fs.FS, pkg string) (*PackageDef, error) {
	// TODO(fancl): Is there a way we can verify the package version from spec?
	_, s := path.Split(pkg)              // toos/go117@1.17.10 => go117@1.17.10
	name := strings.SplitN(s, "@", 2)[0] // go117@1.17.10 => go117

	pkgDir, err := fs.Sub(dir, name)
	if err != nil {
		return nil, fmt.Errorf("failed to open 3pp package dir: %w", err)
	}
	f, err := pkgDir.Open("3pp.pb")
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
		Name: strings.ReplaceAll(name, "-", "_"),
		Spec: &spec,
		Dir:  pkgDir,
	}, nil
}
