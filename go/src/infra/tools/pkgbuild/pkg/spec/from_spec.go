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
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/pkgbuild/pkg/stdenv"

	"go.chromium.org/luci/cipd/client/cipd/platform"
	"google.golang.org/protobuf/proto"
)

// TODO(fancl): Use all:from_spec/build-support after go 1.18.
//go:embed from_spec/*
var fromSpecSupport embed.FS

// Load 3pp Spec and convert it into a stdenv generator.
type SpecLoader struct {
	Directory fs.FS

	pkgs map[string]*stdenv.Generator
}

func NewSpecLoader(dir fs.FS) *SpecLoader {
	return &SpecLoader{
		Directory: dir,
		pkgs:      make(map[string]*stdenv.Generator),
	}
}

func (l *SpecLoader) LoadPackageDef(pkg string) (*PackageDef, error) {
	return LoadPackageDef(l.Directory, pkg)
}

// FromSpec convert the 3pp spec to stdenv generator.
// Ideally we should use the Host Platform in BuildContext during the
// generation. But it's much easier to construct the Spec.Create before
// generate and call SpecLoader.FromSpec recursively for dependencies.
func (l *SpecLoader) FromSpec(pkg, host string) (*stdenv.Generator, error) {
	if g, ok := l.pkgs[pkg]; ok {
		if g == nil {
			return nil, fmt.Errorf("circular dependency detected: %s %s", pkg, host)
		}
		return g, nil
	}

	// Mark the package visited to prevent circular dependency.
	// Remove the mark if we end up with not updating the result.
	l.pkgs[pkg] = nil
	defer func() {
		if l.pkgs[pkg] == nil {
			delete(l.pkgs, pkg)
		}
	}()

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
			matched, err := regexp.MatchString(c.GetPlatformRe(), host)
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
		s := source.GetGit()
		ref, err := resolveGitRef(s)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve git ref: %w", err)
		}
		src = &stdenv.SourceGit{
			URL: s.GetRepo(),
			Ref: ref,
		}
	case *Spec_Create_Source_Url:
		s := source.GetUrl()
		ext := s.GetExtension()
		if ext == "" {
			ext = ".tar.gz"
		}
		src = &stdenv.SourceURL{
			URL:           s.GetDownloadUrl(),
			Filename:      fmt.Sprintf("%s-%s%s", def.Name, s.GetVersion(), ext),
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
		g, err := l.FromSpec(dep, platform.CurrentPlatform())
		if err != nil {
			return nil, err
		}
		deps = append(deps, utilities.BaseDependency{
			Type:      cipkg.DepsBuildHost,
			Generator: g,
		})
	}
	for _, dep := range build.GetDep() {
		g, err := l.FromSpec(dep, host)
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
			fmt.Sprintf("_3PP_PLATFORM=%s", host),
		},
	}

	if strings.HasPrefix(host, "mac-") {
		// TODO(fancl): Set CROSS_TRIPLE and MACOSX_DEPLOYMENT_TARGET for Mac.
	}

	l.pkgs[pkg] = g
	return g, nil
}

//go:embed resolve_git.py
var resolveGitScript string

type tagInfo struct {
	// Regulated semantic versioning tag e.g. 1.2.3
	// This may not be the corresponding git tag.
	Tag string

	// Git commit for the tag.
	Commit string
}

// resolveGitTag require python3 and git in the PATH.
func resolveGitRef(git *GitSource) (string, error) {
	cmd := exec.Command("python3", "-c", resolveGitScript)
	cmd.Stderr = os.Stderr

	in, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	out, err := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		return "", err
	}

	if err := json.NewEncoder(in).Encode(git); err != nil {
		return "", err
	}
	in.Close()

	var info tagInfo
	if err := json.NewDecoder(out).Decode(&info); err != nil {
		return "", err
	}
	out.Close()

	if err := cmd.Wait(); err != nil {
		return "", err
	}

	return info.Commit, nil
}
