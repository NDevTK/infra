// Copyright 2022 The ChromiumOS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package querygs

import (
	"fmt"
	"strings"

	"go.chromium.org/luci/common/gcloud/gs"
)

// validateFaft validates the remote FAFT entry to make sure
// that a firmware bundle actually exists there.
//
// See b:241150358 for an example of this issue happening.
// See b:241155320 for more information about this change itself.
func (r *Reader) validateFaft(faftVersion string) (string, error) {
	if strings.HasPrefix(faftVersion, "gs://") {
		return "", fmt.Errorf("faft version is not expected to be complete gs path %q", faftVersion)
	}

	paths := []string{
		// Empirically, most firmware versions contain a 'metadata.json' file.
		fmt.Sprintf("gs://chromeos-image-archive/%s/metadata.json", faftVersion),
		// However, some firmware versions do not contain a metadata.json file and a JSON proto instead.
		fmt.Sprintf("gs://chromeos-image-archive/%s/firmware_metadata.jsonpb", faftVersion),
	}

	err := fmt.Errorf("validate faft: internal error no paths selected")
	for _, path := range paths {
		// If we encounter an error here, do NOT log it. It leads to confusing log messages that look
		// like errors that are entirely routine.
		err = (r.exst)(gs.Path(path))
		if err == nil {
			return path, nil
		}
	}

	// Heuristically guess that the first path is what we should report as the "real" path if we did not
	// successfully find any of the paths.
	return paths[0], err
}
