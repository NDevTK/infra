// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"bytes"
	"os/exec"
	"path/filepath"

	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/cipkg/base/workflow"
)

func importDarwin(cfg *Config, bins ...string) (gs []generators.Generator, err error) {
	// Import posix utilities
	g, err := generators.FromPathBatch("posix_import", cfg.FindBinary, bins...)
	if err != nil {
		return nil, err
	}
	gs = append(gs, g)

	// Import xcode directory
	xcode := cfg.XcodeDeveloper
	if xcode == nil {
		cmd := exec.Command("xcode-select", "--print-path")
		out, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		path := string(bytes.TrimSpace(out))

		cmd = exec.Command(filepath.Join(path, "usr", "bin", "xcodebuild"), "-version")
		out, err = cmd.Output()
		if err != nil {
			return nil, err
		}
		ver := string(bytes.TrimSpace(out))

		xcode = &generators.ImportTargets{
			Name: "xcode_import",
			Targets: map[string]generators.ImportTarget{
				"Developer": {Source: path, Version: ver},
			},
		}
	}
	gs = append(gs, xcode)

	// Import platform-specific tools
	gs = append(gs, &generators.ImportTargets{
		Name: "darwin_import",
		Targets: map[string]generators.ImportTarget{
			"bin/codesign":     {Source: "/usr/bin/codesign"},
			"bin/xcode-select": {Source: "/usr/bin/xcode-select"},
			"bin/xcrun":        {Source: "/usr/bin/xcrun"},
			"bin/hdiutil":      {Source: "/usr/bin/hdiutil"},
			"bin/pkgbuild":     {Source: "/usr/bin/pkgbuild"},
			"bin/productbuild": {Source: "/usr/bin/productbuild"},

			// Using compilers without wrappers require configuring Apple Framework properly, which isn't trivial.
			// See also: https://github.com/NixOS/nixpkgs/tree/master/pkgs/os-specific/darwin/apple-sdk
			"bin/cc":      {Source: "/usr/bin/cc"},
			"bin/c++":     {Source: "/usr/bin/c++"},
			"bin/clang":   {Source: "/usr/bin/clang"},
			"bin/clang++": {Source: "/usr/bin/clang++"},
			"bin/gcc":     {Source: "/usr/bin/gcc"},
			"bin/g++":     {Source: "/usr/bin/g++"},
		},
	})

	return
}

func (g *Generator) generateDarwin(plats generators.Platforms, tmpl *workflow.Generator) error {
	tmpl.Env.Set("osx_developer_root", "{{.darwin_import}}/Developer")

	// Env GREP added here to skip the configure testing them.
	// TODO(fancl): Update the specs to include gnu grep in the tools if
	// configure.ac expects gnu tools.
	tmpl.Env.Set("GREP", "{{.posix_import}}/bin/grep")
	return nil
}
