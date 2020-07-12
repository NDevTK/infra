// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirmeta

import (
	"path"

	"google.golang.org/protobuf/proto"

	dirmetapb "infra/tools/dirmeta/proto"
)

// Mapping is a mapping from a directory to its metadata.
//
// It wraps the corresponding protobuf message and adds utility functions.
type Mapping dirmetapb.Mapping

// NewMapping initializes an empty mapping.
func NewMapping(size int) *Mapping {
	return &Mapping{
		Dirs: make(map[string]*dirmetapb.Metadata, size),
	}
}

// Compute computes metadata for the given directory key.
func (m *Mapping) Compute(key string) *dirmetapb.Metadata {
	parent := path.Dir(key)
	if parent == key {
		return cloneMeta(m.Dirs[key])
	}

	ret := m.Compute(parent)
	Merge(ret, m.Dirs[key])
	return ret
}

// Proto converts m back to the protobuf message.
func (m *Mapping) Proto() *dirmetapb.Mapping {
	return (*dirmetapb.Mapping)(m)
}

// Merge merges metadata from src to dst, where dst contains inherited metadata
// and src contains directory-specific metadata.
// Does nothing is src is nil.
//
// The current implementation is just proto.Merge, but it may change in the
// future.
func Merge(dst, src *dirmetapb.Metadata) {
	if src != nil {
		proto.Merge(dst, src)
	}
}

// cloneMeta clones meta.
// If meta is nil, returns a new message.
func cloneMeta(meta *dirmetapb.Metadata) *dirmetapb.Metadata {
	if meta == nil {
		return &dirmetapb.Metadata{}
	}
	return proto.Clone(meta).(*dirmetapb.Metadata)
}
