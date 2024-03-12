// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package querygs

import (
	"context"
	"fmt"
	"strings"

	"go.chromium.org/luci/common/gcloud/gs"
	"go.chromium.org/luci/common/logging"
)

// validateFaft validates the remote FAFT entry to make sure
// that a firmware bundle actually exists there.
//
// See b:241150358 for an example of this issue happening.
// See b:241155320 for more information about this change itself.
// See b:286114085 for detail information
func (r *Reader) validateFaft(ctx context.Context, faftPath string) (string, error) {
	if strings.HasPrefix(faftPath, "gs://") {
		return "", fmt.Errorf("validate faft: path is not expected to have gs:// %q", faftPath)
	}

	//check if the fw image path has .tar.bz2 extension
	if !strings.HasSuffix(faftPath, ".tar.bz2") {
		faftPath += "/firmware_from_source.tar.bz2"
	}

	gsPath := fmt.Sprintf("gs://chromeos-image-archive/%s", faftPath)
	if err := r.exst(gs.Path(gsPath)); err != nil {
		logging.Errorf(ctx, "The firmware file is not present on the path %q and failed with %v\n", gsPath, err)
		return "", fmt.Errorf("validate faft: %w", err)
	}
	return gsPath, nil
}
