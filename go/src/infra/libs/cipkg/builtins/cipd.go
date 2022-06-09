// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package builtins

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"infra/libs/cipkg"

	"go.chromium.org/luci/cipd/client/cipd"
	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/cipd/client/cipd/template"
)

// CIPDEnsure is used for downloading CIPD packages. It behaves similar to
// `cipd ensure` for the provided ensure file and use ${out} as the cipd
// root path.
const CIPDEnsureBuilder = BuiltinBuilderPrefix + "cipdEnsure"

type CIPDEnsure struct {
	Name     string
	Ensure   ensure.File
	Expander []string
}

func (c *CIPDEnsure) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	var w strings.Builder
	if err := c.Ensure.Serialize(&w); err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, fmt.Errorf("failed to encode ensure file: %v: %w", c.Ensure, err)
	}
	return cipkg.Derivation{
		Name:    c.Name,
		Builder: CIPDEnsureBuilder,
		Args:    []string{w.String()},
		Env:     c.Expander,
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

	opts := cipd.ClientOptions{
		Root:       out,
		ServiceURL: ef.ServiceURL,
		UserAgent:  fmt.Sprintf("cipkg, %s", cipd.UserAgent),
	}
	clt, err := cipd.NewClient(opts)
	if err != nil {
		return fmt.Errorf("failed to create cipd client: %w", err)
	}

	expander := template.DefaultExpander()
	for _, env := range cmd.Env {
		e := strings.SplitN(env, "=", 2)
		expander[e[0]] = e[1]
	}
	resolver := cipd.Resolver{Client: clt}
	resolved, err := resolver.Resolve(ctx, ef, expander)
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
				fmt.Fprintf(cmd.Stderr, "cipd root %q action %q for pin %q encountered error: %s", root, err.Action, err.Pin, err)
			}
		}
		if errorCount > 0 {
			return fmt.Errorf("cipd packages installation encountered %d error(s)", errorCount)
		}
	}
	return nil
}
