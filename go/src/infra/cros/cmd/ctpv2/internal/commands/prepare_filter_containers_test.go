// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"testing"
)

func TestGetBuildFromGCSPath(t *testing.T) {
	res := getBuildFromGCSPath("gs://chromeos-image-archive/eve-release/R120-15662.91.0")
	if getBuildFromGCSPath("gs://chromeos-image-archive/eve-release/R120-15662.91.0") != 15662 {
		t.Fatalf("Build incorrectly parsed. Expected 15662, got: %v", res)
	}
	res = getBuildFromGCSPath("gs://chromeos-image-archive/dedede-cq/R123-15771.0.0-94499-875671100350224633")
	if res != 15771 {
		t.Fatalf("Build incorrectly parsed. Expected 15771, got: %v", res)
	}
}
