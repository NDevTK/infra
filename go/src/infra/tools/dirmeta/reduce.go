// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirmeta

import (
	"path"
	"sort"

	"google.golang.org/protobuf/reflect/protoreflect"

	dirmetapb "infra/tools/dirmeta/proto"
)

// Reduce returns a new mapping with all redundancies removed.
func (m *Mapping) Reduce() *Mapping {
	// Compute the dir keys in the shortest-path to longest-path order,
	// such that when processing a node, we have a guarantee that its nearest
	// ancestor is already processed.
	orderedKeys := m.keysByLength()

	// First, for each existing node, compute expanded metadata.
	expanded := NewMapping(len(m.Dirs))
	for _, dir := range orderedKeys {
		if ancestor := expanded.nearestAncestor(dir); ancestor == nil {
			expanded.Dirs[dir] = m.Dirs[dir]
		} else {
			meta := cloneMeta(ancestor)
			Merge(meta, m.Dirs[dir])
			expanded.Dirs[dir] = meta
		}
	}

	// Then compute a mapping without redundant information.
	ret := NewMapping(len(m.Dirs))
	for _, dir := range orderedKeys {
		meta := cloneMeta(m.Dirs[dir])
		if ancestor := expanded.nearestAncestor(dir); ancestor != nil {
			excludeSame(meta.ProtoReflect(), ancestor.ProtoReflect())
		}
		if !isEmpty(meta.ProtoReflect()) {
			ret.Dirs[dir] = meta
		}
	}

	return ret
}

// excludeSame mutates m in-place to clear fields that have same values as ones
// in exclude.
func excludeSame(m, exclude protoreflect.Message) {
	m.Range(func(f protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case !exclude.Has(f):
			// It cannot be the same.
			return true

		case f.Kind() == protoreflect.MessageKind:
			// Recurse.
			excludeSame(v.Message(), exclude.Get(f).Message())
			// Clear the field if it became empty.
			if isEmpty(v.Message()) {
				m.Clear(f)
			}

		case f.Cardinality() == protoreflect.Repeated:
			panic("Reduce() is not implemented for repeated fields. We don't have them as of writing")

		case scalarValuesEqual(v, exclude.Get(f), f.Kind()):
			m.Clear(f)
		}
		return true
	})
}

// scalarValuesEqual returns true if a and b are determined to be equal.
// May return false negatives.
func scalarValuesEqual(a, b protoreflect.Value, kind protoreflect.Kind) bool {
	switch kind {
	case protoreflect.BoolKind:
		return a.Bool() == b.Bool()
	case protoreflect.EnumKind:
		return a.Enum() == b.Enum()
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		return a.Int() == b.Int()
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return a.Float() == b.Float()
	case protoreflect.StringKind:
		return a.String() == b.String()
	default:
		return false
	}
}

// nearestAncestor returns metadata of the nearest ancestor.
func (m *Mapping) nearestAncestor(dir string) *dirmetapb.Metadata {
	for {
		parent := path.Dir(dir)
		if parent == dir {
			// We have reached the root.
			return nil
		}
		dir = parent

		if meta, ok := m.Dirs[dir]; ok {
			return meta
		}
	}
}

// keysByLength returns keys sorted by length.
// Key "." is treated as shortest of all.
func (m *Mapping) keysByLength() []string {
	ret := make([]string, 0, len(m.Dirs))
	for k := range m.Dirs {
		ret = append(ret, k)
	}

	sortKey := func(dirKey string) int {
		// "." is considered shortest of all.
		if dirKey == "." {
			return -1
		}
		return len(dirKey)
	}
	sort.Slice(ret, func(i, j int) bool {
		return sortKey(ret[i]) < sortKey(ret[j])
	})
	return ret
}

// isEmpty returns true if m has no populated fields.
func isEmpty(m protoreflect.Message) bool {
	found := false
	m.Range(func(f protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		found = true
		return false
	})
	return !found
}
