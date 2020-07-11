// Copyright 2020 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package dirmeta

import (
	"io/ioutil"
	"os"
	"path/filepath"

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

// ReadMapping reads metadata from the given directory tree.
//
// The returned metadata is neither reduced, not expanded. It represents
// data from the files as is.
func ReadMapping(root string) (*Mapping, error) {
	ret := &Mapping{
		Dirs: map[string]*dirmetapb.Metadata{},
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		// The key in ret.Dirs must be relative to the root.
		key, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		// The key must be in canonical form.
		key = filepath.ToSlash(key)

		switch meta, err := ReadMetadata(path); {
		case err != nil:
			return errors.Annotate(err, "failed to read metadata of %q", path).Err()

		case meta != nil:
			ret.Dirs[key] = meta
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// ReadInherited reads metadata inherited by the target.
// The inheritance starts from the root.
//
// Reads only metadata from root towards target, as opposed to entire tree of
// the root.
func ReadInherited(root, target string) (*dirmetapb.Metadata, error) {
	root = filepath.Clean(root)
	target = filepath.Clean(target)

	var toRead []string
	for {
		toRead = append(toRead, target)

		if target == root {
			break
		}

		parent := filepath.Dir(target)
		if parent == target {
			// We have reached the root of the file system, but did not reach
			// the `root`.
			return nil, errors.Reason("target must be same as root or a sub-directory of root").Err()
		}
		target = parent
	}

	// Read the metadata in the opposite order, starting from the root.
	ret := &dirmetapb.Metadata{}
	for i := len(toRead) - 1; i >= 0; i-- {
		switch meta, err := ReadMetadata(toRead[i]); {
		case err != nil:
			return nil, errors.Annotate(err, "failed to read metadata of %q", toRead[i]).Err()
		case meta != nil:
			Merge(ret, meta)
		}
	}
	return ret, nil
}
