// Copyright 2023 The ChromiumOS Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"log"

	"infra/cros/support/internal/cli"
	"infra/cros/support/internal/manifest"
)

type Input struct {
	FromPath string `json:"from_path"`
	ToPath   string `json:"to_path"`
}

type Output struct {
	manifest.ManifestDiff
}

func main() {
	cli.Init()

	var input Input
	cli.MustUnmarshalInput(&input)

	// Load from and to manifests
	var fromManifest, toManifest manifest.Manifest

	if err := fromManifest.LoadFromXmlFile(input.FromPath); err != nil {
		log.Fatal(err)
	}

	if err := toManifest.LoadFromXmlFile(input.ToPath); err != nil {
		log.Fatal(err)
	}

	cli.MustMarshalOutput(fromManifest.Diff(toManifest))
}
