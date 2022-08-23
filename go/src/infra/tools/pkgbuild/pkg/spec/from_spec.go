// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package spec

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/pkgbuild/pkg/stdenv"

	"google.golang.org/protobuf/proto"
)

// TODO(fancl): Use all:from_spec/build-support after go 1.18.
//go:embed from_spec/*
var fromSpecSupport embed.FS

// Load 3pp Spec and convert it into a stdenv generator.
type SpecLoader struct {
	Directory string

	// Ideally we shouldn't have a separated Platform in the loader. But it's
	// difficult to translate spec into stdenv generator because we need the
	// Platform to construct the Spec.Create and call SpecLoader.FromSpec
	// recursively for dependencies.
	Platform string

	pkgs map[string]*stdenv.Generator
}

func NewSpecLoader(dir string, plat string) *SpecLoader {
	return &SpecLoader{
		Directory: dir,
		Platform:  plat,
		pkgs:      make(map[string]*stdenv.Generator),
	}
}

func (l *SpecLoader) LoadPackageDef(pkg string) (*PackageDef, error) {
	return LoadPackageDef(l.Directory, pkg)
}

func (l *SpecLoader) FromSpec(pkg string) (*stdenv.Generator, error) {
	if g, ok := l.pkgs[pkg]; ok {
		return g, nil
	}

	def, err := l.LoadPackageDef(pkg)
	if err != nil {
		return nil, err
	}

	// Copy files for building from spec
	defDrv := &builtins.CopyFiles{
		Name:  fmt.Sprintf("%s_from_spec_def", def.Name),
		Files: def.Dir,
	}
	fromSpecFS, err := fs.Sub(fromSpecSupport, "from_spec")
	if err != nil {
		return nil, err
	}
	fromSpecDrv := &builtins.CopyFiles{
		Name:  "from_spec_support",
		Files: fromSpecFS,
	}

	// Construct Create from Spec
	create := &Spec_Create{}

	for _, c := range def.Spec.GetCreate() {
		if c.GetPlatformRe() != "" {
			matched, err := regexp.MatchString(c.GetPlatformRe(), l.Platform)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
		}

		proto.Merge(create, c)
	}

	// Fetch source
	source := create.GetSource()

	var src stdenv.Source
	switch source.GetMethod().(type) {
	case *Spec_Create_Source_Git:
		// source.GetGit()
	case *Spec_Create_Source_Url:
		u := source.GetUrl()
		ext := u.GetExtension()
		if ext == "" {
			ext = ".tar.gz"
		}
		src = &stdenv.SourceURL{
			URL:           u.GetDownloadUrl(),
			Filename:      fmt.Sprintf("%s-%s%s", def.Name, u.GetVersion(), ext),
			HashAlgorithm: builtins.HashIgnore,
		}
	case *Spec_Create_Source_Cipd:
		// source.GetCipd()
	case *Spec_Create_Source_Script:
		// source.GetScript()
	}

	// Get patches
	var patches []string
	for _, pdir := range source.GetPatchDir() {
		dir, err := fs.ReadDir(def.Dir, pdir)
		if err != nil {
			return nil, err
		}

		prefix := fmt.Sprintf("{{.%s}}", defDrv.Name)
		for _, d := range dir {
			patches = append(patches, filepath.Join(prefix, pdir, d.Name()))
		}
	}

	// Get build commands
	build := create.GetBuild()
	installArgs := build.GetInstall()
	if len(installArgs) == 0 {
		installArgs = []string{"install.sh"}
	}
	installArgs[0] = filepath.Join(fmt.Sprintf("{{.%s}}", defDrv.Name), installArgs[0])
	fromSpecInstall, err := json.Marshal(installArgs)
	if err != nil {
		return nil, err
	}

	// Generate dependencies
	deps := []utilities.BaseDependency{
		{Type: cipkg.DepsBuildHost, Generator: defDrv},
		{Type: cipkg.DepsBuildHost, Generator: fromSpecDrv},
	}
	for _, dep := range build.GetTool() {
		g, err := l.FromSpec(dep)
		if err != nil {
			return nil, err
		}
		deps = append(deps, utilities.BaseDependency{
			Type:      cipkg.DepsBuildHost,
			Generator: g,
		})
	}
	for _, dep := range build.GetDep() {
		g, err := l.FromSpec(dep)
		if err != nil {
			return nil, err
		}
		deps = append(deps, utilities.BaseDependency{
			Type:      cipkg.DepsHostTarget,
			Generator: g,
		})
	}

	g := &stdenv.Generator{
		Name:         def.Name,
		Source:       src,
		Dependencies: deps,
		Env: []string{
			fmt.Sprintf("patches=%s", strings.Join(patches, string(os.PathListSeparator))),
			fmt.Sprintf("fromSpecInstall=%s", fromSpecInstall),
			fmt.Sprintf("_3PP_PLATFORM=%s", l.Platform),
		},
	}

	if strings.HasPrefix(l.Platform, "mac-") {
		// TODO(fancl): Set CROSS_TRIPLE and MACOSX_DEPLOYMENT_TARGET for Mac.
	}

	l.pkgs[pkg] = g
	return g, nil
}
