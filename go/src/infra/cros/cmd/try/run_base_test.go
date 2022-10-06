// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"testing"

	"infra/cros/internal/assert"
)

func TestValidate_tryRunBase(t *testing.T) {
	r := tryRunBase{
		branch:     "main",
		production: true,
		patches:    []string{"crrev.com/c/1234567"},
	}
	err := r.validate()
	assert.ErrorContains(t, err, "only supported for staging")

	r = tryRunBase{
		branch:     "release-R106.15054.B",
		production: false,
		patches:    []string{"crrev.com/foo/1234567"},
	}
	err = r.validate()
	assert.ErrorContains(t, err, "invalid patch")

	r = tryRunBase{
		branch:     "release-R106.15054.B",
		production: false,
		buildspec:  "/not/a/gs/path",
	}
	err = r.validate()
	assert.ErrorContains(t, err, "gs://")

	r = tryRunBase{
		branch:     "release-R106.15054.B",
		production: true,
		buildspec:  "gs://chromeos-manifest-versions/foo.xml",
	}
	err = r.validate()
	assert.ErrorContains(t, err, "only supported for staging")
}
