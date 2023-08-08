// Copyright 2023 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cache

import (
	"testing"

	"go.chromium.org/luci/common/errors"

	"infra/cros/internal/assert"
	"infra/cros/recovery/internal/execs"
)

var httpResponseCodeTestCases = map[string]int{
	"curl: (22) The requested URL returned error: 500":                      500,
	"curl: (22) The requested URL returned error: 500, returned error: 500": 0,
	"curl: (22) The requested URL returned error: 501":                      501,
	"curl: (22) The requested URL returned error: 502":                      502,
	"curl: (22) The requested URL returned error: 404":                      404,
	"returned error: 404": 404,
	"":                    0,
}

func TestExtractHttpResponseCode(t *testing.T) {
	for k, v := range httpResponseCodeTestCases {
		errAnnotator := errors.Reason("http code extractor test")
		errAnnotator.Tag(errors.TagValue{Key: execs.StdErrTag, Value: k})
		assert.IntsEqual(t, ExtractHttpResponseCode(errAnnotator.Err()), v)
	}
}
