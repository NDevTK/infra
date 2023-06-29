// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package base

import (
	"context"
	"fmt"
	"os"
	"strings"

	"infra/libs/cipkg_new/base/generators"
	"infra/libs/cipkg_new/core"

	"go.chromium.org/luci/common/system/environ"
)

// General purpose generator which generates dependencies and computes
// cross-compile tuples recursively.
type Generator struct {
	Metadata     *core.Action_Metadata
	Args         []string
	Env          environ.Env
	Dependencies []generators.Dependency
}

func (g *Generator) Generate(ctx context.Context, plats generators.Platforms) (*core.Action, error) {
	var deps []*core.Action_Dependency
	envDeps := make(map[string][]string)
	for _, d := range g.Dependencies {
		a, err := d.Generate(ctx, plats)
		if err != nil {
			return nil, err
		}
		deps = append(deps, &core.Action_Dependency{
			Action:  a,
			Runtime: d.Runtime,
		})
		envDeps[d.Type.String()] = append(envDeps[d.Type.String()], fmt.Sprintf("{{.%s}}", a.Metadata.Name))
	}

	env := g.Env.Clone()

	// Add dependencies' environment variables. Iterate through dependency types
	// to ensure a deterministic order.
	for i := generators.DepsUnknown; i < generators.DepsMaxNum; i++ {
		if deps, ok := envDeps[i.String()]; ok {
			env.Set(i.String(), strings.Join(deps, string(os.PathListSeparator)))
		}
	}

	return &core.Action{
		Metadata: g.Metadata,
		Deps:     deps,
		Spec: &core.Action_Command{
			Command: &core.ActionCommand{
				Args: g.Args,
				Env:  env.Sorted(),
			},
		},
	}, nil
}
