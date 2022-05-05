// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package pagination

import (
	"go.chromium.org/luci/common/errors"
)

type PageSizeLimiter struct {
	Max     int32
	Default int32
}

// Adjust the requested pageSize according to PageSizeLimiter.Max and
// PageSizeLimiter.Default as necessary.
func (psl *PageSizeLimiter) Adjust(pageSize int32) int32 {
	switch {
	case pageSize >= psl.Max:
		return psl.Max
	case pageSize > 0:
		return pageSize
	default:
		return psl.Default
	}
}

// ValidatePageSize returns a non-nil error if pageSize is invalid.
// Returns nil if pageSize is 0.
func ValidatePageSize(pageSize int32) error {
	if pageSize < 0 {
		return errors.Reason("negative").Err()
	}
	return nil
}
