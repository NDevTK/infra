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

	wheelFiles := &utilities.BaseGenerator{
		Name:    "wheel_files",
		Builder: "builtin:udf:ensureWheels",
		Args:    []string{"v1", string(raw)},
		Dependencies: []cipkg.Dependency{
			{Type: cipkg.DepsHostTarget, Generator: tags},
		},
	}

	return &utilities.BaseGenerator{
		Name:    "wheels",
		Builder: "builtin:udf:generateWheelsDir",
		Args:    []string{"v1", "{{.wheel_files}}"},
		Dependencies: []cipkg.Dependency{
			{Type: cipkg.DepsHostTarget, Generator: wheelFiles},
		},
	}, nil
}

func init() {
	builtins.RegisterUserDefinedFunction("ensureWheels", ensureWheels)
	builtins.RegisterUserDefinedFunction("generateWheelsDir", generateWheelsDir)
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

	// Translates the packages named in spec into a CIPD ensure file.
	pslice := make(ensure.PackageSlice, len(s.Wheel))
	for i, pkg := range s.Wheel {
		pslice[i] = ensure.PackageDef{
			PackageTemplate:   pkg.Name,
			UnresolvedVersion: pkg.Version,
		}
	}
	ef := ensure.File{
		ServiceURL:       chromeinfra.CIPDServiceURL,
		PackagesBySubdir: map[string]ensure.PackageSlice{"": pslice},
	}
	var efs strings.Builder
	if err := ef.Serialize(&efs); err != nil {
		return err
	}

	// Get vpython template from tags
	var env []string
	if t := pep425TagSelector(tags); t != nil {
		if env, err = getPEP425CIPDTemplateForTag(t); err != nil {
			return err
		}
	}

	// Construct CIPD command and execute
	cipd := exec.CommandContext(ctx, builtins.CIPDEnsureBuilder, efs.String())
	cipd.Env = append(env, cmd.Env...)
	cipd.Stdin = cmd.Stdin
	cipd.Stdout = cmd.Stdout
	cipd.Stderr = cmd.Stderr
	cipd.Dir = cmd.Dir

	return builtins.Execute(ctx, cipd)
}

func generateWheelsDir(ctx context.Context, cmd *exec.Cmd) error {
	// cmd.Args = ["builtin:udf:generateWheelsDir", Version, {{.wheel_files}}]
	files := cmd.Args[2]
	out := builtins.GetEnv("out", cmd.Env)
	ws, err := wheel.ScanDir(files)
	if err != nil {
		return errors.Annotate(err, "failed to scan wheels").Err()
	}
	if err := wheel.WriteRequirementsFile(filepath.Join(out, "requirements.txt"), ws); err != nil {
		return errors.Annotate(err, "failed to write requirements.txt").Err()
	}
	if err := os.Symlink(files, filepath.Join(out, "wheels")); err != nil {
		return errors.Annotate(err, "failed to symlink wheels").Err()
	}
	return nil
}
