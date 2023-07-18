// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package builtins

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"infra/libs/cipkg"

	"go.chromium.org/luci/cipd/client/cipd"
	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/cipd/client/cipd/template"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/system/environ"
)

// CIPDExport is used for downloading CIPD packages. It behaves similar to
// `cipd export` for the provided ensure file and use ${out} as the cipd
// root path.
const CIPDExportBuilder = BuiltinBuilderPrefix + "cipdExport"

type CIPDExport struct {
	Name     string
	Ensure   ensure.File
	Expander template.Expander

	ConfigFile          string
	CacheDir            string
	HTTPUserAgentPrefix string
	MaxThreads          int
	ParallelDownloads   int
	AdmissionPlugin     string
	ServiceURL          string
}

func (c *CIPDExport) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	// Expand template based on ctx.Platform.Host before pass it to cipd client
	// for cross-compile
	expander := c.Expander
	if expander == nil {
		expander = template.Platform{
			OS:   cipdOS(ctx.Platforms.Host.OS()),
			Arch: cipdArch(ctx.Platforms.Host.Arch()),
		}.Expander()
	}

	ef, err := expandEnsureFile(&c.Ensure, expander)
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, fmt.Errorf("failed to expand ensure file: %v: %w", c.Ensure, err)
	}

	var w strings.Builder
	if err := ef.Serialize(&w); err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, fmt.Errorf("failed to encode ensure file: %v: %w", ef, err)
	}

	// Construct environment variables for cipd client.
	// Use cipd related host environment as default value.
	var env []string
	hostEnv := environ.FromCtx(ctx.Context)
	addEnv := func(k string, v any) {
		if vv := reflect.ValueOf(v); !vv.IsValid() || vv.IsZero() {
			if hostV, ok := hostEnv.Lookup(k); !ok {
				return
			} else {
				v = hostV
			}
		}
		env = append(env, fmt.Sprintf("%s=%v", k, v))
	}

	addEnv(cipd.EnvConfigFile, c.ConfigFile)
	addEnv(cipd.EnvCacheDir, c.CacheDir)
	addEnv(cipd.EnvHTTPUserAgentPrefix, c.HTTPUserAgentPrefix)
	addEnv(cipd.EnvMaxThreads, c.MaxThreads)
	addEnv(cipd.EnvParallelDownloads, c.ParallelDownloads)
	addEnv(cipd.EnvAdmissionPlugin, c.AdmissionPlugin)
	addEnv(cipd.EnvCIPDServiceURL, c.ServiceURL)

	return cipkg.Derivation{
		Name:    c.Name,
		Builder: CIPDExportBuilder,
		Args:    []string{w.String()},
		Env:     env,
	}, cipkg.PackageMetadata{}, nil
}

func cipdExport(ctx context.Context, cmd *exec.Cmd) error {
	// cmd.Args = ["builtin:cipdEnsure", Ensure{...}]
	if len(cmd.Args) != 2 {
		return fmt.Errorf("invalid arguments: %v", cmd.Args)
	}
	out := GetEnv("out", cmd.Env)

	export := CIPDCommand("export", "--root", out, "--ensure-file", "-")
	export.Env = append(os.Environ(), cmd.Env...) // Workaround for crbug/1462669.
	export.Dir = cmd.Dir
	export.Stdin = strings.NewReader(cmd.Args[1])
	export.Stdout = cmd.Stdout
	export.Stderr = cmd.Stderr

	if err := export.Run(); err != nil {
		return fmt.Errorf("failed to export packages: %w", err)
	}

	return nil
}

func cloneEnsureFile(ef *ensure.File) (*ensure.File, error) {
	var s bytes.Buffer
	if err := ef.Serialize(&s); err != nil {
		return nil, err
	}
	return ensure.ParseFile(&s)
}

func expandEnsureFile(ef *ensure.File, expander template.Expander) (*ensure.File, error) {
	ef, err := cloneEnsureFile(ef)
	if err != nil {
		return nil, err
	}

	for dir, slice := range ef.PackagesBySubdir {
		var s ensure.PackageSlice
		for _, p := range slice {
			pkg, err := p.Expand(expander)
			switch err {
			case template.ErrSkipTemplate:
				continue
			case nil:
			default:
				return nil, errors.Annotate(err, "expanding %#v", pkg).Err()
			}
			p.PackageTemplate = pkg
			s = append(s, p)
		}
		ef.PackagesBySubdir[dir] = s
	}
	return ef, nil
}

func cipdOS(os string) string {
	if os == "darwin" {
		return "mac"
	}
	return os
}

func cipdArch(arch string) string {
	if arch == "arm" {
		return "armv6l"
	}
	return arch
}

// Create a exec.Cmd for cipd which lookup and expands 'cipd' to it's path.
// exec.Command already did that and store the path in Cmd.Path, but doesn't
// work properly for .bat script.
func CIPDCommand(arg ...string) *exec.Cmd {
	cipd := "cipd"
	if path, err := exec.LookPath("cipd"); err == nil {
		cipd = path
	}

	// Use cmd to execute batch file on windows.
	if filepath.Ext(cipd) == ".bat" {
		return exec.Command("cmd.exe", append([]string{"/C", cipd}, arg...)...)
	}

	return exec.Command(cipd, arg...)
}
