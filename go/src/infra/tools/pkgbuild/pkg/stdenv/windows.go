// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package stdenv

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
)

func importWindows(cfg *Config, bins ...string) (gs []cipkg.Generator, err error) {
	// Import posix utilities from MinGW
	// We use bash.exe to locate where MinGW is installed.
	path, err := cfg.FindBinary("bash.exe")
	if err != nil {
		return nil, fmt.Errorf("failed to find MinGW: %w", err)
	}
	path = filepath.Dir(path)

	g := &builtins.Import{Name: "posix_import"}
	for _, bin := range bins {
		g.Targets = append(g.Targets, builtins.ImportTarget{
			Source:      filepath.Join(path, bin+".exe"),
			Destination: "bin",
			Type:        builtins.ImportExecutable,
		})
	}
	gs = append(gs, g)

	// Copy windows sdk
	winSDK := cfg.WinSDK
	if winSDK == nil {
		vsDir := os.Getenv("VSINSTALLDIR")
		if vsDir == "" {
			return nil, fmt.Errorf("failed to find visual studio: VSINSTALLDIR not set")
		}
		vsFs := os.DirFS(vsDir)
		mf, err := vsFs.Open("win_sdk/SDKManifest.xml")
		if err != nil {
			return nil, fmt.Errorf("failed to open sdk manifest: %w", err)
		}
		defer mf.Close()
		ver, err := io.ReadAll(mf)
		if err != nil {
			return nil, fmt.Errorf("failed to read sdk manifest: %w", err)
		}

		winSDK = &builtins.CopyFiles{
			Name:    "winsdk_files",
			Files:   vsFs,
			Version: string(ver),
		}
	}
	gs = append(gs, winSDK)

	// Import platform-specific tools
	g, err = builtins.FromPathBatch("windows_import", cfg.FindBinary,
		"attrib",
		"cmd",
		"where",
	)
	gs = append(gs, g)

	return
}

func (g *Generator) generateWindows(ctx *cipkg.BuildContext, tmpl *utilities.BaseGenerator) error {
	proc_arch := ctx.Platforms.Build.Arch()
	if proc_arch == "386" {
		proc_arch = "x86"
	}

	sdk_arch := map[string]string{
		"386":   "x86",
		"amd64": "x64",
		"arm64": "arm64",
	}[ctx.Platforms.Host.Arch()]
	if sdk_arch == "" {
		return fmt.Errorf("host architecture not supported yet: %s", ctx.Platforms.Host)
	}

	tmpl.Env = append(tmpl.Env,
		fmt.Sprintf("PROCESSOR_ARCHITECTURE=%s", proc_arch),
		"winsdk_root={{.winsdk_files}}",
		fmt.Sprintf("sdk_arch=%s", sdk_arch),
	)
	return nil
}
