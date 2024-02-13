// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package gitignore implements .gitignore check predicate.
//
// Uses only checked in .gitignore files, ignoring any globals.
package gitignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"

	"go.chromium.org/luci/common/errors"

	"infra/cmd/cloudbuildhelper/fileset"
)

const gitDir = ".git"

// NewExcluder returns a predicate that checks whether the given absolute path
// under given `dir` is excluded by some `ignoreFile` (usually ".gitignore")
// file which is active in that directory.
func NewExcluder(dir, ignoreFile string) (fileset.Excluder, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	// Find a directory with ".git".
	repoRoot, err := findRepoRoot(dir)
	if err != nil {
		return nil, err
	}

	// Find possible ".gitignore" files in parent directories and *all*
	// ".gitignore" files recursively under `dir`.
	paths := scanUp(dir, repoRoot, ignoreFile)
	if paths, err = scanDown(paths, dir, ignoreFile); err != nil {
		return nil, err
	}

	// Load and parse them.
	var pats []gitignore.Pattern
	for _, path := range paths {
		parsed, err := readIgnoreFile(path)
		if err != nil {
			return nil, errors.Annotate(err, "when parsing %q", path).Err()
		}
		pats = append(pats, parsed...)
	}

	// Build a fileset.Excluder out of parsed patterns.
	matcher := gitignore.NewMatcher(pats)
	return func(absPath string, isDir bool) bool {
		return matcher.Match(splitPath(absPath), isDir)
	}, nil
}

// NewPatternExcluder returns a predicate that returns true if the path is
// excluded by any of the given .gitignore patterns.
func NewPatternExcluder(patterns []string) fileset.Excluder {
	var pats []gitignore.Pattern
	for _, pat := range patterns {
		pats = append(pats, gitignore.ParsePattern(pat, nil))
	}
	matcher := gitignore.NewMatcher(pats)
	return func(absPath string, isDir bool) bool {
		return matcher.Match(splitPath(absPath), isDir)
	}
}

// findRepoRoot searches for a parent directory that has ".git" child directory.
//
// `start` itself is also considered during the search. Returns `start` as well
// if no better parent can be found.
func findRepoRoot(start string) (string, error) {
	cur := start
	for {
		switch stat, err := os.Stat(filepath.Join(cur, gitDir)); {
		case err == nil && stat.IsDir():
			return cur, nil
		case err != nil && !os.IsNotExist(err):
			return "", errors.Annotate(err, "when searching for repo root of %q", start).Err()
		}
		par := filepath.Dir(cur)
		if par == cur {
			return start, nil // reached the file system root
		}
		cur = par
	}
}

// scanUp returns potential .gitignore paths in parents of `start` up to `root`.
//
// It is purely lexicographical computation, the file system is not touched.
// Files closer to `root` come first.
func scanUp(start, root, ignoreFile string) (paths []string) {
	cur := start
	for {
		par := filepath.Dir(cur)
		paths = append(paths, filepath.Join(par, ignoreFile))
		if par == root || par == cur {
			break
		}
		cur = par
	}
	// We need the result in reversed order (`root` to `start`).
	for l, r := 0, len(paths)-1; l < r; l, r = l+1, r-1 {
		paths[l], paths[r] = paths[r], paths[l]
	}
	return
}

// scanDown recursively searches for ".gitignore" files under `start`.
//
// Adds them to `paths` slice, returning it in the end.
func scanDown(paths []string, start, ignoreFile string) ([]string, error) {
	err := filepath.Walk(start, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Base(path) == ignoreFile {
			paths = append(paths, path)
		}
		return err
	})
	if err != nil {
		return nil, errors.Annotate(err, "when scanning for .gitignore in %q", start).Err()
	}
	return paths, nil
}

// splitPath splits a path into components, as weird go-git.v4 API wants it.
func splitPath(p string) []string {
	return strings.Split(filepath.Clean(p), string(filepath.Separator))
}

// readIgnoreFile reads a single git ignore file.
//
// Returns (nil, nil) if it doesn't exist.
func readIgnoreFile(path string) (pat []gitignore.Pattern, err error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}
	defer f.Close()

	// Weird go-git.v4 API wants paths split into components.
	domain := splitPath(filepath.Dir(path))

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		switch line := strings.TrimSpace(scanner.Text()); {
		case strings.HasPrefix(line, "#!include:"):
			// "#!include:<path>" is a syntax used by .gcloudignore files.
			inc := filepath.Join(filepath.Dir(path), strings.TrimPrefix(line, "#!include:"))
			included, err := readIgnoreFile(inc)
			if err != nil {
				return nil, err
			}
			pat = append(pat, included...)
		case len(line) > 0 && !strings.HasPrefix(line, "#"):
			pat = append(pat, gitignore.ParsePattern(line, domain))
		}
	}

	err = scanner.Err()
	return
}
