// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package linux

import (
	"testing"

	"go.chromium.org/luci/common/errors"

	"infra/cros/internal/assert"
	"infra/cros/recovery/internal/components"
)

func Test_extractHTTPResponseCodeFromCurlErr(t *testing.T) {
	testCases := map[string]int{
		"curl: (22) The requested URL returned error: 500":                      500,
		"curl: (22) The requested URL returned error: 500, returned error: 500": 0,
		"curl: (22) The requested URL returned error: 501":                      501,
		"curl: (22) The requested URL returned error: 502":                      502,
		"curl: (22) The requested URL returned error: 404":                      404,
		"returned error: 404": 404,
		"":                    0,
	}
	for k, v := range testCases {
		errAnnotator := errors.Reason("http code extractor test")
		errAnnotator.Tag(errors.TagValue{Key: components.StdErrTag, Value: k})
		assert.IntsEqual(t, extractHTTPResponseCodeFromCurlErr(errAnnotator.Err()), v)
	}
}
