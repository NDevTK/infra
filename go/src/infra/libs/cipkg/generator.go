// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cipkg

import (
	"errors"
)

type Generator interface {
	Generate(ctx *BuildContext) (Derivation, PackageMetadata, error)
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
	Type      DependencyType
	Generator Generator
}

func (dep *Dependency) Generate(ctx *BuildContext) (Package, error) {
	var plats Platforms
	switch dep.Type {
	case DepsBuildBuild:
		plats = Platforms{ctx.Platforms.Build, ctx.Platforms.Build, ctx.Platforms.Build}
	case DepsBuildHost:
		plats = Platforms{ctx.Platforms.Build, ctx.Platforms.Build, ctx.Platforms.Host}
	case DepsBuildTarget:
		plats = Platforms{ctx.Platforms.Build, ctx.Platforms.Build, ctx.Platforms.Target}
	case DepsHostHost:
		plats = Platforms{ctx.Platforms.Build, ctx.Platforms.Host, ctx.Platforms.Host}
	case DepsHostTarget:
		plats = Platforms{ctx.Platforms.Build, ctx.Platforms.Host, ctx.Platforms.Target}
	case DepsTargetTarget:
		plats = Platforms{ctx.Platforms.Build, ctx.Platforms.Target, ctx.Platforms.Target}
	default:
		return nil, ErrUnknowDependencyType
	}
	drv, meta, err := dep.Generator.Generate(ctx.WithPlatform(plats))
	if err != nil {
		return nil, err
	}
	return ctx.Packages.Add(drv, meta), nil
}
