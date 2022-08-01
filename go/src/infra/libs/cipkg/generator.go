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
	var plat Platform
	switch dep.Type {
	case DepsBuildBuild:
		plat = Platform{ctx.Platform.Build, ctx.Platform.Build, ctx.Platform.Build}
	case DepsBuildHost:
		plat = Platform{ctx.Platform.Build, ctx.Platform.Build, ctx.Platform.Host}
	case DepsBuildTarget:
		plat = Platform{ctx.Platform.Build, ctx.Platform.Build, ctx.Platform.Target}
	case DepsHostHost:
		plat = Platform{ctx.Platform.Build, ctx.Platform.Host, ctx.Platform.Host}
	case DepsHostTarget:
		plat = Platform{ctx.Platform.Build, ctx.Platform.Host, ctx.Platform.Target}
	case DepsTargetTarget:
		plat = Platform{ctx.Platform.Build, ctx.Platform.Target, ctx.Platform.Target}
	default:
		return nil, ErrUnknowDependencyType
	}
	drv, meta, err := dep.Generator.Generate(ctx.WithPlatform(plat))
	if err != nil {
		return nil, err
	}
	return ctx.Storage.Add(drv, meta), nil
}
