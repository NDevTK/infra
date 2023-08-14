// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package chromium

import (
	"os"
	"path/filepath"

	"go.chromium.org/luci/common/errors"
)

// prepareOutDir ensures that a dir exists and does not have files that match
// clearPattern glob, e.g. "*.jsonl.gz".
func PrepareOutDir(path string, append bool, clearPattern string) error {
	if err := os.MkdirAll(path, 0777); err != nil {
		return err
	}

	// If we're appending to existing data, don't delete existing files
	if append {
		return nil
	}

	// Remove existing files.
	existing, err := filepath.Glob(filepath.Join(path, clearPattern))
	if err != nil {
		return err
	}
	for _, f := range existing {
		if err := os.Remove(f); err != nil {
			return errors.Annotate(err, "failed to remove %q", f).Err()
		}
	}
	return nil
}
