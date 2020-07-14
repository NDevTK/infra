// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirmeta

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/encoding/prototext"

	"go.chromium.org/luci/common/errors"

	dirmetapb "infra/tools/dirmeta/proto"
)

// Filename is the standard name of the metadata file.
const Filename = "DIR_METADATA"

// ReadMetadata reads metadata from a single directory.
// See also MappingReader.
//
// Returns (nil, nil) if the metadata is not defined.
func ReadMetadata(dir string) (*dirmetapb.Metadata, error) {
	fullPath := filepath.Join(dir, Filename)
	contents, err := ioutil.ReadFile(fullPath)
	switch {
	case os.IsNotExist(err):
		// Try the legacy file.
		return readOwners(dir)

	case err != nil:
		return nil, err
	}

	var ret dirmetapb.Metadata
	if err := prototext.Unmarshal(contents, &ret); err != nil {
		return nil, errors.Annotate(err, "failed to parse %q", fullPath).Err()
	}
	return &ret, nil
}

// MappingReader reads Mapping from the file system.
type MappingReader struct {
	// Root is a path to the root directory.
	Root string
	// Mapping is the result of reading.
	Mapping
}

// ReadAll reads metadata from the entire directory tree, overwriting
// r.Mapping.
func (r *MappingReader) ReadAll(form dirmetapb.MappingForm) error {
	r.Mapping = *NewMapping(0)
	err := filepath.Walk(r.Root, func(dir string, info os.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case !info.IsDir():
			return nil
		}

		key := r.mustDirKey(dir)

		switch meta, err := ReadMetadata(dir); {
		case err != nil:
			return errors.Annotate(err, "failed to read metadata of %q", dir).Err()

		case meta != nil:
			r.Dirs[key] = meta

		case form == dirmetapb.MappingForm_FULL:
			// Put an empty metadata, so that ComputeAll() populates it below.
			r.Dirs[key] = &dirmetapb.Metadata{}
		}
		return nil
	})
	if err != nil {
		return err
	}

	switch form {
	case dirmetapb.MappingForm_REDUCED:
		r.Mapping.Reduce()
	case dirmetapb.MappingForm_COMPUTED, dirmetapb.MappingForm_FULL:
		r.Mapping.ComputeAll()
	}

	return nil
}

// ReadTowards reads metadata of directories on the node path from r.Root to
// target. It skips directories for which it already has metadata.
func (r *MappingReader) ReadTowards(target string) error {
	root := filepath.Clean(r.Root)
	target = filepath.Clean(target)

	for {
		switch key, err := r.DirKey(target); {
		case err != nil:
			return err
		case r.Dirs[key] == nil:
			switch meta, err := ReadMetadata(target); {
			case err != nil:
				return errors.Annotate(err, "failed to read metadata of %q", target).Err()

			case meta != nil:
				if r.Dirs == nil {
					r.Dirs = map[string]*dirmetapb.Metadata{}
				}
				r.Dirs[key] = meta
			}
		}

		if target == root {
			return nil
		}

		// Go up.
		parent := filepath.Dir(target)
		if parent == target {
			// We have reached the root of the file system, but not `root`.
			// This is impossible because DirKey would have failed.
			panic("impossible")
		}
		target = parent
	}
}

// DirKey returns a r.Dirs key for the given dir on the file system.
// The path must be a part of the tree under r.Root.
func (r *MappingReader) DirKey(dir string) (string, error) {
	key, err := filepath.Rel(r.Root, dir)
	if err != nil {
		return "", err
	}

	// Dir keys use forward slashes.
	key = filepath.ToSlash(key)

	if strings.HasPrefix(key, "../") {
		return "", errors.Reason("the path %q must be under the root %q", dir, r.Root).Err()
	}

	return key, nil
}

// mustDirKey is like DirKey, but panics on failure.
func (r *MappingReader) mustDirKey(dir string) string {
	key, err := r.DirKey(dir)
	if err != nil {
		panic(err)
	}
	return key
}
