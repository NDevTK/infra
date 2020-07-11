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

// ReadMetadata reads metadata from a given directory.
// See ReadMapping for a recursive version.
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
	Mapping
	// Root is a path to the root directory.
	Root string
}

// DirKey returns a r.Dirs key for the given path.
// The path must be a part of the tree under r.Root.
func (r *MappingReader) DirKey(path string) (string, error) {
	key, err := filepath.Rel(r.Root, path)
	if err != nil {
		return "", err
	}

	// Dir keys do not use backslashes.
	key = filepath.ToSlash(key)

	if strings.HasPrefix(key, "../") {
		return "", errors.Reason("the path %q must be under the root %q", path, r.Root).Err()
	}

	return key, nil
}

// Read reads metadata of the given directory.
func (r *MappingReader) Read(dir string) error {
	key, err := r.DirKey(dir)
	if err != nil {
		return err
	}

	switch meta, err := ReadMetadata(dir); {
	case err != nil:
		return err

	case meta != nil:
		if r.Dirs == nil {
			r.Dirs = map[string]*dirmetapb.Metadata{}
		}
		r.Dirs[key] = meta
	}
	return nil
}

// ReadFull reads metadata from the entire directory tree,
// overwriting r.Metadata.
//
// The resulting metadata is neither reduced nor expanded.
// It represents data from the files as is.
func (r *MappingReader) ReadFull() error {
	r.Dirs = map[string]*dirmetapb.Metadata{}
	return filepath.Walk(r.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		return errors.Annotate(r.Read(path), "failed to read metadata of %q", path).Err()
	})
}

// ReadTowards reads metadata of directories on the path from r.Root till target.
// It skips directories for which it already has metadata.
func (r *MappingReader) ReadTowards(target string) error {
	root := filepath.Clean(r.Root)
	target = filepath.Clean(target)

	for {
		switch key, err := r.DirKey(target); {
		case err != nil:
			return err
		case r.Dirs[key] == nil:
			if err := r.Read(target); err != nil {
				return errors.Annotate(r.Read(target), "failed to read metadata of %q", target).Err()
			}
		}

		if target == root {
			return nil
		}

		parent := filepath.Dir(target)
		if parent == target {
			// We have reached the root of the file system, but not `root`.
			// This is impossible because DirKey would have failed.
			panic("impossible")
		}
		target = parent
	}
}
