// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utilities

import (
	"fmt"
	"strings"
	"text/template"

	"infra/libs/cipkg"
)

type BaseGenerator struct {
	Name         string
	Builder      string
	Args         []string
	Env          []string
	Dependencies []cipkg.Dependency
}

func (g *BaseGenerator) Generate(ctx *cipkg.BuildContext) (cipkg.Derivation, error) {
	var inputs, envs []string
	for _, dep := range g.Dependencies {
		drv, err := dep.Generate(ctx)
		if err != nil {
			return cipkg.Derivation{}, err
		}
		inputs = append(inputs, drv.ID())
		pkg := ctx.Storage.Add(drv)
		envs = append(envs, fmt.Sprintf("%s=%s", drv.Name, pkg.Directory()))
	}
	envs = append(envs, g.Env...)

	envMap, err := envToMap(envs)
	if err != nil {
		return cipkg.Derivation{}, nil
	}

	tmpl := template.New(g.Name).Option("missingkey=error")
	builder, err := render(tmpl.New("builder"), g.Builder, envMap)
	if err != nil {
		return cipkg.Derivation{}, nil
	}
	var args []string
	for i, arg := range g.Args {
		a, err := render(tmpl.New(fmt.Sprintf("arg_%d", i)), arg, envMap)
		if err != nil {
			return cipkg.Derivation{}, nil
		}
		args = append(args, a)
	}

	return cipkg.Derivation{
		Name:     g.Name,
		Platform: ctx.Platform.Build,
		Builder:  builder,
		Args:     args,
		Env:      envs,
		Inputs:   inputs,
	}, nil
}

func envToMap(envs []string) (map[string]string, error) {
	m := make(map[string]string)
	for _, env := range envs {
		ss := strings.SplitN(env, "=", 2)
		if len(ss) != 2 {
			return nil, fmt.Errorf("invalid environment variable: %s", env)
		}
		m[ss[0]] = ss[1]
	}
	return m, nil
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
