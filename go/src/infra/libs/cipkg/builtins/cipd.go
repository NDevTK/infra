// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package builtins

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"infra/libs/cipkg"

	"go.chromium.org/luci/cipd/client/cipd"
	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/cipd/client/cipd/template"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/system/environ"
)

// CIPDEnsure is used for downloading CIPD packages. It behaves similar to
// `cipd ensure` for the provided ensure file and use ${out} as the cipd
// root path.
const CIPDEnsureBuilder = BuiltinBuilderPrefix + "cipdEnsure"

type CIPDEnsure struct {
	Name     string
	Ensure   ensure.File
	Expander template.Expander

	CacheDir          string
	MaxThreads        int
	ParallelDownloads int
	ServiceURL        string
}

func (c *CIPDEnsure) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	// Expand template based on ctx.Platform.Host before pass it to cipd client
	// for cross-compile
	expander := c.Expander
	if expander == nil {
		// TODO: Replace with proper parser for platform
		h := strings.Split(ctx.Platform.Host, "_")
		expander = template.Platform{OS: cipdOS(h[0]), Arch: cipdArch(h[1])}.Expander()
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
	addEnv := func(k, v string) {
		if v == "" {
			v = hostEnv.Get(k)
		}
		if v != "" {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	addEnv(cipd.EnvCacheDir, c.CacheDir)
	addEnv(cipd.EnvMaxThreads, strconv.Itoa(c.MaxThreads))
	addEnv(cipd.EnvParallelDownloads, strconv.Itoa(c.ParallelDownloads))
	addEnv(cipd.EnvCIPDServiceURL, c.ServiceURL)

	return cipkg.Derivation{
		Name:    c.Name,
		Builder: CIPDEnsureBuilder,
		Args:    []string{w.String()},
		Env:     env,
	}, cipkg.PackageMetadata{}, nil
}

func cipdEnsure(ctx context.Context, cmd *exec.Cmd) error {
	// cmd.Args = ["builtin:cipdEnsure", Ensure{...}]
	if len(cmd.Args) != 2 {
		return fmt.Errorf("invalid arguments: %v", cmd.Args)
	}
	out := GetEnv("out", cmd.Env)

	ef, err := ensure.ParseFile(strings.NewReader(cmd.Args[1]))
	if err != nil {
		return fmt.Errorf("failed to parse argument: %#v: %w", cmd.Args[1], err)
	}

	ctxWithEnv := environ.New(cmd.Env).SetInCtx(ctx)
	opts := cipd.ClientOptions{
		Root:       out,
		ServiceURL: ef.ServiceURL,
		UserAgent:  fmt.Sprintf("cipkg, %s", cipd.UserAgent),
	}
	clt, err := cipd.NewClientFromEnv(ctxWithEnv, opts)
	if err != nil {
		return fmt.Errorf("failed to create cipd client: %w", err)
	}
	defer clt.Close(ctx)

	resolver := cipd.Resolver{Client: clt}
	resolved, err := resolver.Resolve(ctx, ef, template.Expander{})
	if err != nil {
		return fmt.Errorf("failed to resolve CIPD package: %w", err)
	}

	actionMap, err := clt.EnsurePackages(ctx, resolved.PackagesBySubdir, nil)
	if err != nil {
		return fmt.Errorf("failed to install CIPD packages: %w", err)
	}
	if len(actionMap) > 0 {
		errorCount := 0
		for root, action := range actionMap {
			errorCount += len(action.Errors)
			for _, err := range action.Errors {
				fmt.Fprintf(cmd.Stderr, "cipd root %q action %q for pin %q encountered error: %s", root, err.Action, err.Pin, err.Error)
			}
		}
		if errorCount > 0 {
			return fmt.Errorf("cipd packages installation encountered %d error(s)", errorCount)
		}
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
