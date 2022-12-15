// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package spec

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
	"infra/tools/pkgbuild/pkg/stdenv"

	"go.chromium.org/luci/cipd/client/cipd/platform"
)

// TODO(fancl): Use all:from_spec/build-support after go 1.18.
//
//go:embed from_spec/*
var fromSpecSupport embed.FS

// Load 3pp Spec and convert it into a stdenv generator.
type SpecLoader struct {
	cipdPackagePrefix     string
	cipdSourceCachePrefix string
	sourceResolver        SourceResolver

	embedSupportFilesDerivation cipkg.Generator

	// Mapping packages's full name to definition.
	specs map[string]*PackageDef
	pkgs  map[string]*stdenv.Generator
}

type SpecLoaderConfig struct {
	CIPDPackagePrefix     string
	CIPDSourceCachePrefix string
	SourceResolver        SourceResolver
}

func DefaultSpecLoaderConfig(vpythonSpecPath string) *SpecLoaderConfig {
	return &SpecLoaderConfig{
		CIPDPackagePrefix:     "",
		CIPDSourceCachePrefix: "sources",
		SourceResolver: &DefaultSourceResolver{
			VPythonSpecPath: vpythonSpecPath,
		},
	}
}

func NewSpecLoader(dir fs.FS, cfg *SpecLoaderConfig) (*SpecLoader, error) {
	// Copy embedded files
	fromSpecFS, err := fs.Sub(fromSpecSupport, "from_spec")
	if err != nil {
		return nil, err
	}
	embedSupportFilesDerivation := &builtins.CopyFiles{
		Name:  "from_spec_support",
		Files: fromSpecFS,
	}

	defs, err := FindPackageDefs(dir)
	if err != nil {
		return nil, err
	}

	specs := make(map[string]*PackageDef)
	for _, def := range defs {
		specs[def.FullName()] = def
	}

	return &SpecLoader{
		cipdPackagePrefix:     cfg.CIPDPackagePrefix,
		cipdSourceCachePrefix: cfg.CIPDSourceCachePrefix,
		sourceResolver:        cfg.SourceResolver,

		embedSupportFilesDerivation: embedSupportFilesDerivation,

		specs: specs,
		pkgs:  make(map[string]*stdenv.Generator),
	}, nil
}

// List all loaded specs' full names by alphabetical order.
func (l *SpecLoader) ListAllByFullName() (names []string) {
	for name := range l.specs {
		names = append(names, name)
	}
	sort.Strings(names)
	return
}

// FromSpec converts the 3pp spec to stdenv generator by its full name, which
// builds the package for running on the cipd host platform.
// Ideally we should use the Host Platform in BuildContext during the
// generation. But it's much easier to construct the Spec.Create before
// generate and call SpecLoader.FromSpec recursively for dependencies.
func (l *SpecLoader) FromSpec(fullName, hostCipdPlatform string) (*stdenv.Generator, error) {
	pkgCacheKey := fmt.Sprintf("%s@%s", fullName, hostCipdPlatform)
	if g, ok := l.pkgs[pkgCacheKey]; ok {
		if g == nil {
			return nil, fmt.Errorf("circular dependency detected: %s", pkgCacheKey)
		}
		return g, nil
	}

	// Mark the package visited to prevent circular dependency.
	// Remove the mark if we end up with not updating the result.
	l.pkgs[pkgCacheKey] = nil
	defer func() {
		if l.pkgs[pkgCacheKey] == nil {
			delete(l.pkgs, pkgCacheKey)
		}
	}()

	def := l.specs[fullName]
	if def == nil {
		return nil, fmt.Errorf("package spec not available: %s", fullName)
	}

	// Copy files for building from spec
	defDerivation := &builtins.CopyFiles{
		Name:  fmt.Sprintf("%s_from_spec_def", def.DerivationName()),
		Files: def.Dir,
	}

	// Parse create spec for host
	create, err := newCreateParser(hostCipdPlatform, def.Spec.GetCreate())
	if err != nil {
		return nil, err
	}
	if err := create.ParseSource(def, l.cipdPackagePrefix, l.cipdSourceCachePrefix, hostCipdPlatform, l.sourceResolver); err != nil {
		return nil, err
	}
	if err := create.FindPatches(defDerivation); err != nil {
		return nil, err
	}
	if err := create.ParseBuilder(defDerivation); err != nil {
		return nil, err
	}
	if err := create.LoadDependencies(l); err != nil {
		return nil, err
	}

	plat := utilities.PlatformFromCIPD(hostCipdPlatform)
	g := &stdenv.Generator{
		Name:   def.DerivationName(),
		Source: create.Source,
		Dependencies: append([]utilities.BaseDependency{
			{Type: cipkg.DepsBuildHost, Generator: defDerivation},
			{Type: cipkg.DepsBuildHost, Generator: l.embedSupportFilesDerivation},
		}, create.Dependencies...),
		Env: append([]string{
			fmt.Sprintf("patches=%s", strings.Join(create.Patches, string(os.PathListSeparator))),
			fmt.Sprintf("fromSpecInstall=%s", create.Installer),
			fmt.Sprintf("_3PP_PLATFORM=%s", hostCipdPlatform),
			fmt.Sprintf("_3PP_TOOL_PLATFORM=%s", platform.CurrentPlatform()),

			// TODO: These should be moved to go package
			fmt.Sprintf("GOOS=%s", plat.OS()),
			fmt.Sprintf("GOARCH=%s", plat.Arch()),
		}, create.Enviroments...),
		CacheKey: def.CIPDPath(l.cipdPackagePrefix, hostCipdPlatform),
		Version:  create.Version,
	}

	switch hostCipdPlatform {
	case "mac-amd64":
		g.Env = append(g.Env, "MACOSX_DEPLOYMENT_TARGET=10.10")
	case "mac-arm64":
		g.Env = append(g.Env, "MACOSX_DEPLOYMENT_TARGET=11.0")
		// TODO(fancl): set CROSS_TRIPLE for Mac?
	}

	l.pkgs[pkgCacheKey] = g
	return g, nil
}

// A parser for Spec_Create spec. It converts the merged create section in the
// spec to information we need for constructing a stdenv generator.
type createParser struct {
	Source       stdenv.Source
	Version      string
	Patches      []string
	Installer    string
	Dependencies []utilities.BaseDependency
	Enviroments  []string

	host   string
	create *Spec_Create
}

var (
	ErrPackageNotAvailable = errors.New("package not available on the target platform")
)

// Merge create specs for the host platform. Return a parser with the merged
// spec.
func newCreateParser(host string, creates []*Spec_Create) (*createParser, error) {
	p := &createParser{
		host: host,
	}

	for _, c := range creates {
		if c.GetPlatformRe() != "" {
			matched, err := regexp.MatchString(c.GetPlatformRe(), host)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
		}

		if c.GetUnsupported() == true {
			return nil, ErrPackageNotAvailable
		}

		if p.create == nil {
			p.create = &Spec_Create{}
		}
		protoMerge(p.create, c)
	}

	if p.create == nil {
		return nil, ErrPackageNotAvailable
	}

	// To make this create rule self-consistent instead of just having the last
	// platform_re to be applied.
	p.create.PlatformRe = ""

	return p, nil
}

// Extract the cache path from URL
func gitCachePath(url string) string {
	url = strings.TrimPrefix(url, "https://chromium.googlesource.com/external/")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	return path.Clean(url)
}

// Fetch the latest version and convert source section in create to source
// definition in stdenv. Versions are fetched during the parsing so the source
// definition can be deterministic.
// Source may be cached based on CacheKey.
func (p *createParser) ParseSource(def *PackageDef, packagePrefix, sourceCachePrefix, hostCipdPlatform string, resolver SourceResolver) error {
	source := p.create.GetSource()

	// Subdir is only used by go packages before go module and can be easily
	// replaced by a simple move after unpack stage.
	if source.GetSubdir() != "" {
		return fmt.Errorf("source.subdir not supported.")
	}

	if source.GetUnpackArchive() {
		p.Enviroments = append(p.Enviroments, "_3PP_UNPACK_ARCHIVE=1")
	}

	if source.GetNoArchivePrune() {
		p.Enviroments = append(p.Enviroments, "_3PP_NO_ARCHIVE_PRUNE=1")
	}

	s, v, err := func() (stdenv.Source, string, error) {
		switch source.GetMethod().(type) {
		case *Spec_Create_Source_Git:
			s := source.GetGit()
			info, err := resolver.ResolveGitSource(s)
			if err != nil {
				return nil, "", fmt.Errorf("failed to resolve git ref: %w", err)
			}
			return &stdenv.SourceGit{
				URL: s.GetRepo(),
				Ref: info.Commit,

				CacheKey: (&url.URL{
					Path: path.Join(packagePrefix, sourceCachePrefix, "git", gitCachePath(s.GetRepo())),
					RawQuery: url.Values{
						"subdir": {"src"},
						"tag":    {fmt.Sprintf("2@%s", info.Tag)},
					}.Encode(),
				}).String(),
			}, info.Tag, nil
		case *Spec_Create_Source_Url:
			s := source.GetUrl()
			ext := s.GetExtension()
			if ext == "" {
				ext = ".tar.gz"
			}
			return &stdenv.SourceURLs{
				URLs: []stdenv.SourceURL{
					{URL: s.GetDownloadUrl(), Filename: fmt.Sprintf("raw_source_0%s", ext), HashAlgorithm: builtins.HashIgnore},
				},
				CacheKey: (&url.URL{
					Path: path.Join(packagePrefix, sourceCachePrefix, "url", def.FullNameWithOverride(), p.host),
					RawQuery: url.Values{
						"tag": {fmt.Sprintf("2@%s", s.GetVersion())},
					}.Encode(),
				}).String(),
			}, s.GetVersion(), nil
		case *Spec_Create_Source_Cipd:
			// source.GetCipd()
		case *Spec_Create_Source_Script:
			s := source.GetScript()
			info, err := resolver.ResolveScriptSource(hostCipdPlatform, def.Dir, s)
			if err != nil {
				return nil, "", fmt.Errorf("failed to resolve latest: %w", err)
			}

			// info.Name is optional.
			names := info.Name
			if len(names) == 0 {
				for i := range info.URL {
					names = append(names, fmt.Sprintf("raw_source_%d%s", i, info.Ext))
				}
			}

			// Number of names must equal to urls.
			if len(names) != len(info.URL) {
				return nil, "", fmt.Errorf("failed to get download urls: number of urls should be equal to artifacts: %w", err)
			}

			var urls []stdenv.SourceURL
			for i, url := range info.URL {
				urls = append(urls, stdenv.SourceURL{
					URL:           url,
					Filename:      names[i],
					HashAlgorithm: builtins.HashIgnore,
				})
			}

			return &stdenv.SourceURLs{
				URLs: urls,
				CacheKey: (&url.URL{
					Path: path.Join(packagePrefix, sourceCachePrefix, "script", def.FullNameWithOverride(), p.host),
					RawQuery: url.Values{
						"tag": {fmt.Sprintf("2@%s", info.Version)},
					}.Encode(),
				}).String(),
			}, info.Version, nil
		}
		return nil, "", fmt.Errorf("unknown source type from spec")
	}()
	if err != nil {
		return err
	}

	p.Enviroments = append(p.Enviroments, fmt.Sprintf("_3PP_VERSION=%s", v))
	if pv := p.create.GetSource().GetPatchVersion(); pv != "" {
		p.Enviroments = append(p.Enviroments, fmt.Sprintf("_3PP_PATCH_VERSION=%s", pv))
		v = v + "." + pv
	}
	p.Version = v
	p.Source = s

	return nil
}

func (p *createParser) FindPatches(drv *builtins.CopyFiles) error {
	source := p.create.GetSource()

	prefix := fmt.Sprintf("{{.%s}}", drv.Name)
	for _, pdir := range source.GetPatchDir() {
		dir, err := fs.ReadDir(drv.Files, pdir)
		if err != nil {
			return err
		}

		for _, d := range dir {
			p.Patches = append(p.Patches, filepath.Join(prefix, pdir, d.Name()))
		}
	}

	return nil
}

func (p *createParser) ParseBuilder(drv *builtins.CopyFiles) error {
	build := p.create.GetBuild()

	installArgs := build.GetInstall()
	if len(installArgs) == 0 {
		installArgs = []string{"install.sh"}
	}
	installArgs[0] = filepath.Join(fmt.Sprintf("{{.%s}}", drv.Name), installArgs[0])

	installer, err := json.Marshal(installArgs)
	if err != nil {
		return err
	}
	p.Installer = string(installer)

	return nil
}

func (p *createParser) LoadDependencies(l *SpecLoader) error {
	build := p.create.GetBuild()
	if build == nil {
		p.Enviroments = append(p.Enviroments, "_3PP_NO_INSTALL=1")
		return nil
	}

	fromSpecByURI := func(dep, host string) (cipkg.Generator, error) {
		// tools/go117@1.17.10
		var name, ver string
		ss := strings.SplitN(dep, "@", 2)
		name = ss[0]
		if len(ss) == 2 {
			ver = ss[1]
		}

		g, err := l.FromSpec(name, host)
		if err != nil {
			return nil, fmt.Errorf("failed to load dependency %s on %s: %w", name, host, err)
		}
		if ver != "" && ver != g.Version {
			return nil, fmt.Errorf("dependency version mismatch: %s, require: %s, have: %s", dep, ver, g.Version)
		}

		return g, nil
	}

	buildPlat := platform.CurrentPlatform()
	for _, dep := range build.GetTool() {
		g, err := fromSpecByURI(dep, buildPlat)
		if err != nil {
			return err
		}
		p.Dependencies = append(p.Dependencies, utilities.BaseDependency{
			Type:      cipkg.DepsBuildHost,
			Generator: g,
		})
	}
	for _, dep := range build.GetDep() {
		g, err := fromSpecByURI(dep, p.host)
		if err != nil {
			return err
		}
		p.Dependencies = append(p.Dependencies, utilities.BaseDependency{
			Type:      cipkg.DepsHostTarget,
			Generator: g,
		})
	}

	return nil
}

// Convert CIPD platform to cipkg platform.
func ParseCIPDPlatform(plat string) (cipkg.Platform, error) {
	idx := strings.Index(plat, "-")
	if idx == -1 {
		return nil, fmt.Errorf("invalid cipd target platform: %s", plat)
	}
	os, arch := plat[:idx], plat[idx+1:]
	if os == "mac" {
		os = "darwin"
	}
	if arch == "armv6l" {
		arch = "arm"
	}
	return utilities.NewPlatform(os, arch), nil
}
