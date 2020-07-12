// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirmeta

import (
	"io/ioutil"
	"os"
	"path"
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
//
// If expand is true, the r.Mapping has an entry for each directory in
// the tree, even if that directory does not define any metadata.
// If expand is false, then the r.Mapping is neither reduced nor
// expanded, but represents data from the files as is.
func (r *MappingReader) ReadAll(expand bool) error {
	r.Mapping = *NewMapping(0)
	return filepath.Walk(r.Root, func(dir string, info os.FileInfo, err error) error {
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
		}

		if expand {
			// Compute full metadata for the dir.
			// Note: filepath.Walk walks in lexical order, so by this time we have
			// expanded parent.
			meta := cloneMeta(r.Dirs[path.Dir(key)])
			Merge(meta, r.Dirs[key])
			r.Dirs[key] = meta
		}

		return nil
	})
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
