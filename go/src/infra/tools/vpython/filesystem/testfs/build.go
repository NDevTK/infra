// Copyright 2017 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package testfs

import (
	"io/ioutil"
	"path/filepath"
	"sort"

	"infra/tools/vpython/filesystem"
)

// BuildDir is a sentinel value that can be provided to Build as content to
// instruct it to make a directory instead of a file.
var BuildDir = "\x00"

// Build constructs a filesystem hierarchy given a layout.
//
// The layouts keys should be ToSlash-style file paths. Its values should be the
// content that is written at those paths. Intermediate directories will be
// automatically created.
func Build(base string, layout map[string]string) error {
	keys := make([]string, 0, len(layout))
	for k := range layout {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, path := range keys {
		content := layout[path]
		path = filepath.Join(base, filepath.FromSlash(path))

		if content == BuildDir {
			// Make a directory.
			if err := filesystem.MakeDirs(path); err != nil {
				return err
			}
		} else {
			// Make a file.
			if err := filesystem.MakeDirs(filepath.Dir(path)); err != nil {
				return err
			}
			if err := ioutil.WriteFile(path, []byte(content), 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
