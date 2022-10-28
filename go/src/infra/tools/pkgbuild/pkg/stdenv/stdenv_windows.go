// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"embed"

	"infra/libs/cipkg"
)

var (
	//go:embed resources/windows
	setupFiles embed.FS
	setup      cipkg.Generator
)

func (g *Generator) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	panic("not implemented")
}
