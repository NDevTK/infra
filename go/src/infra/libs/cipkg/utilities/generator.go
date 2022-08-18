// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utilities

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"infra/libs/cipkg"
)

type BaseDependency struct {
	Runtime   bool
	Type      cipkg.DependencyType
	Generator cipkg.Generator
}

type BaseGenerator struct {
	Name         string
	Builder      string
	Args         []string
	Env          []string
	Dependencies []BaseDependency
}

type BaseGeneratorResult struct {
	Derivation cipkg.Derivation
	Metadata   cipkg.PackageMetadata
	Packages   []cipkg.Package
}

func (g *BaseGenerator) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, cipkg.PackageMetadata, error) {
	// Generate dependencies' derivation
	var inputs, rDeps []string
	dirs := make(map[string]string)
	envDeps := make(map[string][]string)
	for _, dep := range g.Dependencies {
		d := &cipkg.Dependency{
			Type:      dep.Type,
			Generator: dep.Generator,
		}
		pkg, err := d.Generate(ctx)
		if err != nil {
			return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
		}

		drv := pkg.Derivation()
		inputs = append(inputs, drv.ID())
		dirs[drv.Name] = pkg.Directory()
		envDeps[d.Type.String()] = append(envDeps[d.Type.String()], pkg.Directory())
		if dep.Runtime {
			rDeps = append(rDeps, drv.ID())
		}
	}

	// Render templates for Builder, Args, Env
	tmpl := template.New(g.Name).Option("missingkey=error")
	builder, err := render(tmpl.New("builder"), g.Builder, dirs)
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
	}
	args, err := renderAll(tmpl, "arg", g.Args, dirs)
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
	}
	env, err := renderAll(tmpl, "env", g.Env, dirs)
	if err != nil {
		return cipkg.Derivation{}, cipkg.PackageMetadata{}, err
	}

	// Add dependencies' environment variables. Iterate through dependency types
	// to ensure a deterministic order.
	for i := cipkg.DepsUnknown; i < cipkg.DepsMaxNum; i++ {
		e := i.String()
		if deps, ok := envDeps[e]; ok {
			env = append(env, fmt.Sprintf("%s=%s", e, strings.Join(deps, string(os.PathListSeparator))))
		}
	}

	return cipkg.Derivation{
			Name:     g.Name,
			Platform: ctx.Platforms.Build.String(),
			Builder:  builder,
			Args:     args,
			Env:      env,
			Inputs:   inputs,
		}, cipkg.PackageMetadata{
			Dependencies: rDeps,
		}, nil
}

func renderAll(tmpl *template.Template, prefix string, raw []string, data interface{}) ([]string, error) {
	var ret []string
	for i, r := range raw {
		a, err := render(tmpl.New(fmt.Sprintf("%s_%d", prefix, i)), r, data)
		if err != nil {
			return nil, err
		}
		ret = append(ret, a)
	}
	return ret, nil
}

func render(tmpl *template.Template, raw string, data interface{}) (string, error) {
	t, err := tmpl.Parse(raw)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := t.Execute(&b, data); err != nil {
		return "", err
	}
	return b.String(), nil
}
