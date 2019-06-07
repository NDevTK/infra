// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package autotest

import (
	"io"

	"go.chromium.org/luci/common/errors"
)

func parseTestControl(r io.Reader) (*testMetadata, error) {
	return nil, errors.Reason("not impl").Err()
}
