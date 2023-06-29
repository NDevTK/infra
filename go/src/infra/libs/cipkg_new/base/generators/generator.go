// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package generators

import (
	"context"

	"infra/libs/cipkg_new/core"

	"go.chromium.org/luci/common/errors"
)

// Generator is the interface for generating actions.
type Generator interface {
	Generate(ctx context.Context, plats Platforms) (*core.Action, error)
}

var (
	ErrUnknowDependencyType = errors.New("unknown dependency type")
)

// Different dependency types are used to calculate dependency's cross-compile
// platform from the dependent's.
type DependencyType int

func (t DependencyType) String() string {
	switch t {
	case DepsBuildBuild:
		return "depsBuildBuild"
	case DepsBuildHost:
		return "depsBuildHost"
	case DepsBuildTarget:
		return "depsBuildTarget"
	case DepsHostHost:
		return "depsHostHost"
	case DepsHostTarget:
		return "depsHostTarget"
	case DepsTargetTarget:
		return "depsTargetTarget"
	default:
		return "depsUnknown"
	}
}

const (
	DepsUnknown DependencyType = iota
	DepsBuildBuild
	DepsBuildHost
	DepsBuildTarget
	DepsHostHost
	DepsHostTarget
	DepsTargetTarget
	DepsMaxNum
)

type Dependency struct {
	Generator Generator
	Type      DependencyType
	Runtime   bool
}

func (dep *Dependency) Generate(ctx context.Context, plats Platforms) (*core.Action, error) {
	var depPlats Platforms
	switch dep.Type {
	case DepsBuildBuild:
		depPlats = Platforms{plats.Build, plats.Build, plats.Build}
	case DepsBuildHost:
		depPlats = Platforms{plats.Build, plats.Build, plats.Host}
	case DepsBuildTarget:
		depPlats = Platforms{plats.Build, plats.Build, plats.Target}
	case DepsHostHost:
		depPlats = Platforms{plats.Build, plats.Host, plats.Host}
	case DepsHostTarget:
		depPlats = Platforms{plats.Build, plats.Host, plats.Target}
	case DepsTargetTarget:
		depPlats = Platforms{plats.Build, plats.Target, plats.Target}
	default:
		return nil, ErrUnknowDependencyType
	}
	return dep.Generator.Generate(ctx, depPlats)
}
