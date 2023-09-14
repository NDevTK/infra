package stdenv

import (
	"crypto"
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"infra/libs/cipkg"
	"infra/libs/cipkg/builtins"
	"infra/libs/cipkg/utilities"
)

//go:embed git_archive.py
var gitSource embed.FS

func (g *Generator) fetchSource() (cipkg.Generator, string, error) {
	// The name of the source derivation. It's also used in environment variable
	// srcs to pointing to the location of source file(s), which will be expanded
	// to absolute path by utilities.BaseGenerator.
	name := fmt.Sprintf("%s_source", g.Name)
	srcPath := fmt.Sprintf("{{.%s}}", name)
	switch s := g.Source.(type) {
	case *SourceGit:
		return &utilities.BaseGenerator{
			Name:    name,
			Builder: filepath.Join("{{.stdenv_python3}}", "bin", "python3"),
			Args:    []string{"-I", "-B", filepath.Join("{{.git_source}}", "git_archive.py"), s.URL, s.Ref},
			Env:     []string{"PATH=" + filepath.Join("{{.stdenv_git}}", "bin")},
			Dependencies: append([]utilities.BaseDependency{
				{Type: cipkg.DepsBuildHost, Generator: git},
				{Type: cipkg.DepsBuildHost, Generator: cpython},
				{Type: cipkg.DepsBuildHost, Generator: &builtins.CopyFiles{
					Name:  "git_source",
					Files: gitSource,
				}},
			}),
			Version:  s.Version,
			CacheKey: s.CacheKey,
		}, "srcs=" + filepath.Join(srcPath, "src.tar"), nil
	case *SourceURLs:
		urls := builtins.FetchURLs{
			Name: name,
		}
		var srcs []string
		for _, u := range s.URLs {
			urls.URLs = append(urls.URLs, builtins.FetchURL{
				Name:          name,
				URL:           u.URL,
				Filename:      u.Filename,
				HashAlgorithm: u.HashAlgorithm,
				HashString:    u.HashString,
			})
			srcs = append(srcs, filepath.Join(srcPath, u.Filename))
		}
		return &utilities.WithMetadata{
			Generator: &urls,
			Metadata: cipkg.PackageMetadata{
				Version:  s.Version,
				CacheKey: s.CacheKey,
			},
		}, fmt.Sprintf("srcs=%s", strings.Join(srcs, string(filepath.ListSeparator))), nil
	default:
		return nil, "", fmt.Errorf("unknown source type %#v:", s)
	}
}

type Source interface {
	isSourceMethod()
}

type SourceGit struct {
	// The url to the git repository. Support any protocol used by the git
	// command line interfaces.
	URL string

	// The reference for the git repo. Can be anything supported by git checkout.
	// Typical use cases:
	// - refs/tags/xxx
	// - 8e8722e14772727b0e1cd5bd925a0f089611a60b
	Ref string

	CacheKey string
	Version  string
}

func (s *SourceGit) isSourceMethod() {}

type SourceURL struct {
	URL           string
	Filename      string
	HashAlgorithm crypto.Hash
	HashString    string
}

type SourceURLs struct {
	URLs []SourceURL

	CacheKey string
	Version  string
}

func (s *SourceURLs) isSourceMethod() {}
