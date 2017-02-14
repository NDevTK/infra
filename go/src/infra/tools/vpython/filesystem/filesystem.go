// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package filesystem

import (
	"os"
	"path/filepath"

	"github.com/luci/luci-go/common/errors"
)

// MakeDirs is a convenience wrapper around os.MkdirAll that applies a 0755
// mask to all created directories.
func MakeDirs(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.Annotate(err).Err()
	}
	return nil
}

// AbsPath is a convenience wrapper around filepath.Abs that accepts a string
// pointer, base, and updates it on successful resolution.
func AbsPath(base *string) error {
	v, err := filepath.Abs(*base)
	if err != nil {
		return errors.Annotate(err).Reason("unable to resolve absolute path").
			D("base", *base).
			Err()
	}
	*base = v
	return nil
}

// Touch creates a new, empty file at the specified path.
func Touch(path string, mode os.FileMode) error {
	fd, err := os.OpenFile(path, (os.O_CREATE | os.O_RDWR | os.O_TRUNC), mode)
	if err != nil {
		return errors.Annotate(err).Err()
	}
	return fd.Close()
}

// RemoveAll is a wrapper around os.RemoveAll which makes sure all files are
// writeable (recursively) prior to removing them.
func RemoveAll(path string) error {
	if err := MakeUserWritable(path); err != nil {
		return errors.Annotate(err).Reason("failed to mark user-writable").Err()
	}
	if err := os.RemoveAll(path); err != nil {
		return errors.Annotate(err).Err()
	}
	return nil
}

// MakeReadOnly recursively iterates through all of the files and directories
// starting at path and marks them read-only.
func MakeReadOnly(path string, filter func(string) bool) error {
	return recursiveChmod(path, filter, func(mode os.FileMode) os.FileMode {
		return mode & (^os.FileMode(0222))
	})
}

// MakeUserWritable recursively iterates through all of the files and
// directories starting at path and ensures that they are user-writable.
func MakeUserWritable(path string) error {
	return recursiveChmod(path, nil, func(mode os.FileMode) os.FileMode {
		return (mode | 0200)
	})
}

func recursiveChmod(path string, filter func(string) bool, chmod func(mode os.FileMode) os.FileMode) error {
	if filter == nil {
		filter = func(string) bool { return true }
	}

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Annotate(err).Err()
		}

		mode := info.Mode()
		if (mode.IsRegular() || mode.IsDir()) && filter(path) {
			if newMode := chmod(mode); newMode != mode {
				if err := os.Chmod(path, newMode); err != nil {
					return errors.Annotate(err).Reason("failed to set read-only: %(path)s").
						D("path", path).
						Err()
				}
			}
		}
		return nil
	})
	if err != nil {
		return errors.Annotate(err).Err()
	}
	return nil
}
