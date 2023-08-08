// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"bytes"
	"os/exec"
	"path/filepath"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
)

func importDarwin(cfg *Config, bins ...string) (gs []cipkg.Generator, err error) {
	// Import posix utilities
	g, err := builtins.FromPathBatch("posix_import", cfg.FindBinary, bins...)
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

		xcode = &builtins.Import{
			Name: "xcode_import",
			Targets: []builtins.ImportTarget{
				{Source: path, Destination: "Developer", Version: ver, Type: builtins.ImportDirectory},
			},
		}
	}
	gs = append(gs, xcode)

	// Import platform-specific tools
	gs = append(gs, &builtins.Import{
		Name: "darwin_import",
		Targets: []builtins.ImportTarget{
			{Source: "/usr/bin/codesign", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/sw_vers", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/xcode-select", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/xcrun", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/hdiutil", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/pkgbuild", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/productbuild", Destination: "bin", Type: builtins.ImportExecutable},

			// Using compilers without wrappers require configuring Apple Framework properly, which isn't trivial.
			// See also: https://github.com/NixOS/nixpkgs/tree/master/pkgs/os-specific/darwin/apple-sdk
			{Source: "/usr/bin/cc", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/c++", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/clang", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/clang++", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/gcc", Destination: "bin", Type: builtins.ImportExecutable},
			{Source: "/usr/bin/g++", Destination: "bin", Type: builtins.ImportExecutable},
		},
	})

	return
}

func (g *Generator) generateDarwin(ctx *cipkg.BuildContext, tmpl *utilities.BaseGenerator) error {
	tmpl.Env = append(tmpl.Env,
		"osx_developer_root={{.darwin_import}}/Developer",

		// Env GREP added here to skip the configure testing them.
		// TODO(fancl): Update the specs to include gnu grep in the tools if
		// configure.ac expects gnu tools.
		"GREP={{.posix_import}}/bin/grep",
	)
	return nil
}
