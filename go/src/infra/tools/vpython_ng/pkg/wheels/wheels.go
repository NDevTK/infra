// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wheels

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"

	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/cipd/client/cipd/template"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/hardcoded/chromeinfra"
	"go.chromium.org/luci/vpython/api/vpython"
	"go.chromium.org/luci/vpython/spec"
	"go.chromium.org/luci/vpython/wheel"
	"google.golang.org/protobuf/encoding/prototext"
)

func FromSpec(spec *vpython.Spec, tags cipkg.Generator) (cipkg.Generator, error) {
	raw, err := prototext.Marshal(spec)
	if err != nil {
		return nil, errors.Annotate(err, "failed to marshal vpython spec").Err()
	}

	return &utilities.BaseGenerator{
		Name:    "wheels",
		Builder: "builtin:udf:ensureWheels",
		Args:    []string{"v1", string(raw)},
		Dependencies: []cipkg.Dependency{
			{Type: cipkg.DepsHostTarget, Generator: tags},
		},
	}, nil
}

func init() {
	builtins.RegisterUserDefinedFunction("ensureWheels", ensureWheels)
}

func ensureWheels(ctx context.Context, cmd *exec.Cmd) error {
	// cmd.Args = ["builtin:udf:generateWheelsDir", Version, Spec]

	// Parse spec file
	var s vpython.Spec
	if err := prototext.Unmarshal([]byte(cmd.Args[2]), &s); err != nil {
		return err
	}

	// Parse tags file
	var tags []*vpython.PEP425Tag
	tagsDir := builtins.GetEnv("python_pep425tags", cmd.Env)
	raw, err := os.Open(filepath.Join(tagsDir, "pep425tags.json"))
	if err != nil {
		return err
	}
	defer raw.Close()
	if err := json.NewDecoder(raw).Decode(&tags); err != nil {
		return err
	}

	// Remove unmatched wheels from spec
	if err := spec.NormalizeSpec(&s, tags); err != nil {
		return err
	}

	// Get vpython template from tags
	expander := template.DefaultExpander()
	if t := pep425TagSelector(tags); t != nil {
		if err := addPEP425CIPDTemplateForTag(expander, t); err != nil {
			return err
		}
	}

	// Translates packages' name in spec into a CIPD ensure file.
	pslice := make(ensure.PackageSlice, len(s.Wheel))
	for i, pkg := range s.Wheel {
		name, err := expander.Expand(pkg.Name)
		switch err {
		case template.ErrSkipTemplate:
			continue
		case nil:
		default:
			return errors.Annotate(err, "expanding %#v", pkg).Err()
		}
		pslice[i] = ensure.PackageDef{
			PackageTemplate:   name,
			UnresolvedVersion: pkg.Version,
		}
	}
	ef := ensure.File{
		ServiceURL:       chromeinfra.CIPDServiceURL,
		PackagesBySubdir: map[string]ensure.PackageSlice{"wheels": pslice},
	}
	var efs strings.Builder
	if err := ef.Serialize(&efs); err != nil {
		return err
	}

	// Construct CIPD command and execute
	// TODO: Replacing it with executing cipd binary directly
	cipd := exec.CommandContext(ctx, builtins.CIPDEnsureBuilder, efs.String())
	cipd.Env = cmd.Env
	cipd.Stdin = cmd.Stdin
	cipd.Stdout = cmd.Stdout
	cipd.Stderr = cmd.Stderr
	cipd.Dir = cmd.Dir

	if err := builtins.Execute(ctx, cipd); err != nil {
		return err
	}

	// Generate requirements.txt
	out := builtins.GetEnv("out", cmd.Env)

	wheels := filepath.Join(out, "wheels")
	ws, err := wheel.ScanDir(wheels)
	if err != nil {
		return errors.Annotate(err, "failed to scan wheels").Err()
	}
	if err := wheel.WriteRequirementsFile(filepath.Join(out, "requirements.txt"), ws); err != nil {
		return errors.Annotate(err, "failed to write requirements.txt").Err()
	}

	return nil
}
