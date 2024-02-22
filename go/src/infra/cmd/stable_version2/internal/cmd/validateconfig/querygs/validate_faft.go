// Copyright 2024 The Chromium Authors
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
// See b:286114085 for detail information
func (r *Reader) validateFaft(faftVersion string) (string, error) {
	if strings.HasPrefix(faftVersion, "gs://") {
		return "", fmt.Errorf("validate faft: path is not expected to have gs:// %q", faftVersion)
	}

	path := fmt.Sprintf("gs://chromeos-image-archive/%s/firmware_from_source.tar.bz2", faftVersion)
	err := r.exst(gs.Path(path))
	if err == nil {
		return path, nil
	}
	fmt.Printf("The firmware file is not present on the path %q and failed with %v\n", path, err)
	return "", fmt.Errorf("validate faft: %w", err)
}
