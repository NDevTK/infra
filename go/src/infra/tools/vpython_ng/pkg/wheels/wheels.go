// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wheels

import (
	"context"
	"encoding/base64"
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
	"go.chromium.org/luci/vpython/api/vpython"
	"go.chromium.org/luci/vpython/spec"
	"go.chromium.org/luci/vpython/wheel"
	"google.golang.org/protobuf/proto"
)

func FromSpec(spec *vpython.Spec, tags cipkg.Generator) (cipkg.Generator, error) {
	raw, err := proto.Marshal(spec)
	if err != nil {
		return nil, errors.Annotate(err, "failed to marshal vpython spec").Err()
	}

	env := []string{
		"python_pep425tags={{.python_pep425tags}}",
	}

	return &utilities.BaseGenerator{
		Name:    "wheels",
		Builder: "builtin:udf:ensureWheels",
		Args:    []string{"v1", base64.RawStdEncoding.EncodeToString(raw)},
		Dependencies: []utilities.BaseDependency{
			{Type: cipkg.DepsHostTarget, Generator: tags},
		},
		Env: env,
	}, nil
}

func init() {
	builtins.RegisterUserDefinedFunction("ensureWheels", ensureWheels)
}

func ensureWheels(ctx context.Context, cmd *exec.Cmd) error {
	// cmd.Args = ["builtin:udf:ensureWheels", Version, Spec]

	// Parse spec file
	var s vpython.Spec
	rawSpec, err := base64.RawStdEncoding.DecodeString(cmd.Args[2])
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(rawSpec, &s); err != nil {
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
		p := PlatformForPEP425Tag(t)
		expander = p.Expander()
		if err := addPEP425CIPDTemplateForTag(expander, t); err != nil {
			return err
		}
	}

	// Translates packages' name in spec into a CIPD ensure file.
	ef, err := ensureFileFromWheels(expander, s.Wheel)
	if err != nil {
		return err
	}
	var efs strings.Builder
	if err := ef.Serialize(&efs); err != nil {
		return err
	}

	// Construct CIPD command and execute
	export := builtins.CIPDCommand("export", "--root", builtins.GetEnv("out", cmd.Env), "--ensure-file", "-")
	export.Env = os.Environ() // Pass host environment variables to cipd.
	export.Dir = cmd.Dir
	export.Stdin = strings.NewReader(efs.String())
	export.Stdout = cmd.Stdout
	export.Stderr = cmd.Stderr

	if err := export.Run(); err != nil {
		return errors.Annotate(err, "failed to export packages").Err()
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

func ensureFileFromWheels(expander template.Expander, wheels []*vpython.Spec_Package) (*ensure.File, error) {
	names := make(map[string]struct{})
	pslice := make(ensure.PackageSlice, len(wheels))
	for i, pkg := range wheels {
		name, err := expander.Expand(pkg.Name)
		if err != nil {
			if err == template.ErrSkipTemplate {
				continue
			}
			return nil, errors.Annotate(err, "expanding %v", pkg).Err()
		}
		if _, ok := names[name]; ok {
			return nil, errors.Reason("duplicated package: %v", pkg).Err()
		}
		names[name] = struct{}{}

		pslice[i] = ensure.PackageDef{
			PackageTemplate:   name,
			UnresolvedVersion: pkg.Version,
		}
	}
	return &ensure.File{
		PackagesBySubdir: map[string]ensure.PackageSlice{"wheels": pslice},
	}, nil
}

// Verify the spec for all VerifyPep425Tag listed in the spec. This will ensure
// all packages existed for these platforms.
//
// TODO: Maybe implement it inside a derivation after we executing cipd binary
// directly.
func Verify(spec *vpython.Spec) error {
	for _, t := range spec.VerifyPep425Tag {
		p := PlatformForPEP425Tag(t)
		e := p.Expander()
		if err := addPEP425CIPDTemplateForTag(e, t); err != nil {
			return err
		}
		ef, err := ensureFileFromWheels(e, spec.Wheel)
		if err != nil {
			return err
		}
		ef.VerifyPlatforms = []template.Platform{p}
		var efs strings.Builder
		if err := ef.Serialize(&efs); err != nil {
			return err
		}

		cmd := builtins.CIPDCommand("ensure-file-verify", "-ensure-file", "-")
		cmd.Stdin = strings.NewReader(efs.String())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
