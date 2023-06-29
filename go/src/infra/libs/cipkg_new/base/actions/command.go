// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package actions

import (
	"fmt"
	"strings"
	"text/template"

	"infra/libs/cipkg_new/core"
)

// ActionCommandTransformer is the default transformer for command action.
func ActionCommandTransformer(a *core.ActionCommand, deps []PackageDependency) (*core.Derivation, error) {
	drv := &core.Derivation{
		Args: a.Args,
		Env:  a.Env,
	}

	// Render templates
	dirs := make(map[string]string)
	for _, d := range deps {
		dirs[d.Package.Action.Metadata.Name] = d.Package.Handler.OutputDirectory()
	}
	if err := renderDerivation(dirs, drv); err != nil {
		return nil, err
	}

	return drv, nil
}

func renderDerivation(vals map[string]string, drv *core.Derivation) (err error) {
	tmpl := template.New("base").Option("missingkey=error")
	drv.Args, err = renderAll(tmpl, "arg", drv.Args, vals)
	if err != nil {
		return
	}
	drv.Env, err = renderAll(tmpl, "env", drv.Env, vals)
	if err != nil {
		return err
	}
	return nil
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
