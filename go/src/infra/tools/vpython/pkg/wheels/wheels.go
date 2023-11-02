// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package wheels

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/types/known/anypb"

	"go.chromium.org/luci/cipd/client/cipd"
	"go.chromium.org/luci/cipd/client/cipd/ensure"
	"go.chromium.org/luci/cipd/client/cipd/template"
	"go.chromium.org/luci/cipkg/base/actions"
	"go.chromium.org/luci/cipkg/base/generators"
	"go.chromium.org/luci/cipkg/core"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/system/environ"
	"go.chromium.org/luci/vpython/api/vpython"
	"go.chromium.org/luci/vpython/spec"
	"go.chromium.org/luci/vpython/wheel"

	"infra/tools/vpython/pkg/common"
)

type vpythonSpecGenerator struct {
	spec       *vpython.Spec
	pep425tags generators.Generator
}

func (g *vpythonSpecGenerator) Generate(ctx context.Context, plats generators.Platforms) (*core.Action, error) {
	p, err := g.pep425tags.Generate(ctx, plats)
	if err != nil {
		return nil, err
	}
	s, err := anypb.New(g.spec)
	if err != nil {
		return nil, err
	}
	return &core.Action{
		Name: "wheels",
		Deps: []*core.Action{p},
		Spec: &core.Action_Extension{Extension: s},
	}, nil
}

func FromSpec(spec *vpython.Spec, pep425tags generators.Generator) generators.Generator {
	return &vpythonSpecGenerator{spec: spec, pep425tags: pep425tags}
}

func MustSetTransformer(cipdCacheDir string, ap *actions.ActionProcessor) {
	v := &vpythonSpecTransformer{
		cipdCacheDir: cipdCacheDir,
	}
	actions.MustSetTransformer[*vpython.Spec](ap, v.Transform)
}

type vpythonSpecTransformer struct {
	cipdCacheDir string
}

func (v *vpythonSpecTransformer) Transform(spec *vpython.Spec, deps []actions.Package) (*core.Derivation, error) {
	drv, err := actions.ReexecDerivation(spec, true)
	if err != nil {
		return nil, err
	}
	env := environ.New(drv.Env)
	env.Set(cipd.EnvCacheDir, v.cipdCacheDir)
	for _, d := range deps {
		drv.FixedOutput += "+" + d.DerivationID
		env.Set(d.Action.Name, d.Handler.OutputDirectory())
	}
	drv.Env = env.Sorted()
	return drv, nil
}

func MustSetExecutor(reexec *actions.ReexecRegistry) {
	actions.MustSetExecutor[*vpython.Spec](reexec, actionVPythonSpecExecutor)
}

func actionVPythonSpecExecutor(ctx context.Context, a *vpython.Spec, out string) error {
	envs := environ.FromCtx(ctx)

	// Parse tags file
	var tags []*vpython.PEP425Tag
	tagsDir := envs.Get("python_pep425tags")
	raw, err := os.Open(filepath.Join(tagsDir, "pep425tags.json"))
	if err != nil {
		return err
	}
	defer raw.Close()
	if err := json.NewDecoder(raw).Decode(&tags); err != nil {
		return err
	}

	// Remove unmatched wheels from spec
	if err := spec.NormalizeSpec(a, tags); err != nil {
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
	ef, err := ensureFileFromWheels(expander, a.Wheel)
	if err != nil {
		return err
	}
	var efs strings.Builder
	if err := ef.Serialize(&efs); err != nil {
		return err
	}

	// Execute cipd export
	if err := actions.ActionCIPDExportExecutor(ctx, &core.ActionCIPDExport{
		EnsureFile: efs.String(),
		Env:        envs.Sorted(),
	}, out); err != nil {
		return err
	}

	// Generate requirements.txt
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

		cmd := common.CIPDCommand("ensure-file-verify", "-ensure-file", "-")
		cmd.Stdin = strings.NewReader(efs.String())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
