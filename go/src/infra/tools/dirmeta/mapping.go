// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirmeta

import (
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

// Proto converts m back to the protobuf message.
func (m *Mapping) Proto() *dirmetapb.Mapping {
	return (*dirmetapb.Mapping)(m)
}

// Merge merges metadata from src to dest, where dst contains inherited metadata
// and src contains directory-specific metadata.
//
// The current implementation is just proto.Merge, but it may change in the
// future.
func Merge(dst, src *dirmetapb.Metadata) {
	proto.Merge(dst, src)
}

// cloneMeta clones meta.
// If meta is nil, returns a new message.
func cloneMeta(meta *dirmetapb.Metadata) *dirmetapb.Metadata {
	if meta == nil {
		return &dirmetapb.Metadata{}
	}
	return proto.Clone(meta).(*dirmetapb.Metadata)
}
