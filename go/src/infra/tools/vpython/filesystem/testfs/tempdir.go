// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package testfs

import (
	"io/ioutil"

	"infra/tools/vpython/filesystem"
)

// WithTempDir creates a temporary directory and passes it to fn. After fn
// exits, the directory is cleaned up.
func WithTempDir(prefix string, fn func(string) error) error {
	tdir, err := ioutil.TempDir("", prefix)
	if err != nil {
		return err
	}
	defer func() {
		_ = filesystem.RemoveAll(tdir)
	}()
	return fn(tdir)
}

// MustWithTempDir calls WithTempDir and panics if any failures occur.
func MustWithTempDir(prefix string, fn func(string)) func() {
	return func() {
		err := WithTempDir(prefix, func(tdir string) error {
			fn(tdir)
			return nil
		})
		if err != nil {
			panic(err)
		}
	}
}
