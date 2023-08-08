// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"fmt"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
)

// Return the dockcross image for the platform.
// TODO(fancl): build the container using pkgbuild.
func containers(plat cipkg.Platform) string {
	const prefix = "gcr.io/chromium-container-registry/infra-dockerbuild/"
	const version = ":v1.4.18"
	if plat.OS() != "linux" {
		return ""
	}
	switch plat.Arch() {
	case "amd64":
		return prefix + "manylinux-x64-py3" + version
	case "arm64":
		return prefix + "linux-arm64-py3" + version
	case "arm":
		return prefix + "linux-armv6-py3" + version
	default:
		return ""
	}
}

func importLinux(cfg *Config, bins ...string) (gs []cipkg.Generator, err error) {
	// Import posix utilities
	g, err := builtins.FromPathBatch("posix_import", cfg.FindBinary, bins...)
	if err != nil {
		return nil, err
	}
	gs = append(gs, g)

	// Import docker
	g, err = builtins.FromPathBatch("docker_import", cfg.FindBinary, "docker")
	if err != nil {
		return nil, err
	}
	gs = append(gs, g)

	return
}

func (g *Generator) generateLinux(ctx *cipkg.BuildContext, tmpl *utilities.BaseGenerator) error {
	containers := containers(ctx.Platforms.Host)
	if containers == "" {
		return fmt.Errorf("containers not available for %s", ctx.Platforms.Host)
	}

	tmpl.Env = append(tmpl.Env,
		fmt.Sprintf("dockerImage=%s", containers),
	)
	return nil
}
