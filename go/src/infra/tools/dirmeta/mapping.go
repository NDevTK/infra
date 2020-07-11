// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirmeta

import (
	"path"
	"sort"

	"google.golang.org/protobuf/proto"

	dirmetapb "infra/tools/dirmeta/proto"
)

// Mapping is a mapping from a directory to its metadata.
//
// It wraps the corresponding protobuf message and adds utility functions.
type Mapping dirmetapb.Mapping

// Proto converts m back to the protobuf message.
func (m *Mapping) Proto() *dirmetapb.Mapping {
	return (*dirmetapb.Mapping)(m)
}

// Reduce returns a new mapping with all redundancies removed.
func (m *Mapping) Reduce() *Mapping {
	panic("not implemented")
}

// Expand returns a new mapping where each dir has attributes inherited
// from the parent dir.
func (m *Mapping) Expand() *Mapping {
	ret := &Mapping{
		Dirs: make(map[string]*dirmetapb.Metadata, len(m.Dirs)),
	}

	// nearestAncestor returns metadata of the nearest ancestor in ret.
	nearestAncestor := func(dir string) *dirmetapb.Metadata {
		for {
			parent := path.Dir(dir)
			if parent == dir {
				// We have reached the root.
				return nil
			}
			dir = parent

			if meta, ok := ret.Dirs[dir]; ok {
				return meta
			}
		}
	}

	// Process directories in the shorest-path to longest-path order,
	// such that, when computing the expanded metadata for a given directory,
	// we only need to check the nearest ancestor.
	for _, dir := range m.dirsSortedByLength() {
		if ancestor := nearestAncestor(dir); ancestor == nil {
			ret.Dirs[dir] = m.Dirs[dir]
		} else {
			meta := proto.Clone(ancestor).(*dirmetapb.Metadata)
			proto.Merge(meta, m.Dirs[dir])
			ret.Dirs[dir] = meta
		}
	}

	return ret
}

// dirsSortedByLength returns directory names sorted by length.
// Directory "." is treated as shortest of all.
func (m *Mapping) dirsSortedByLength() []string {
	ret := make([]string, 0, len(m.Dirs))
	for k := range m.Dirs {
		ret = append(ret, k)
	}
	sort.Slice(ret, func(i, j int) bool {
		// "." is considered shortest of all.
		if ret[i] == "." {
			return true
		}
		return len(ret[i]) < len(ret[j])
	})
	return ret
}

// Inherited returns inherited metadata of the directory.
func (m *Mapping) Inherited(dir string) *dirmetapb.Metadata {
	ret := &dirmetapb.Metadata{}
	for _, dir := range ancestorsAndSelf(dir) {
		if md := m.Dirs[dir]; md != nil {
			proto.Merge(ret, md)
		}
	}
	return ret
}

// ancestorsAndSelf returns all ancestors of the dir and the dir itself,
// e.g. for "a/b/c" it returns [".", "a", "a/b", "a/b/c"].
func ancestorsAndSelf(dir string) []string {
	var ret []string

	dir = path.Clean(dir)
	for {
		ret = append(ret, dir)

		parent := path.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Reverse, so that we start with the root.
	for i, n := 0, len(ret); i < n/2; i++ {
		ret[i], ret[n-1-i] = ret[n-1-i], ret[i]
	}
	return ret
}
