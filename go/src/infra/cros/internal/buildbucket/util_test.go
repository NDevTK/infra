// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package buildbucket

import "testing"

// TestSeparateBucketFromBuilder tests separateBucketFromBuilder.
func TestSeparateBucketFromBuilder(t *testing.T) {
	t.Parallel()
	for i, tc := range []struct {
		fullBuilderName string
		expectedBucket  string
		expectedBuilder string
		expectError     bool
	}{
		{"chromeos/release/release-main-orchestrator", "chromeos/release", "release-main-orchestrator", false},
		{"too/short", "", "", true},
		{"too/long/by/far", "", "", true},
	} {
		bucket, builder, err := SeparateBucketFromBuilder(tc.fullBuilderName)
		if err != nil && !tc.expectError {
			t.Errorf("#%d: separateBucketFromBuilder(%s) returned an error; want no error. Returned error: %+v", i, tc.fullBuilderName, err)
		}
		if err == nil && tc.expectError {
			t.Errorf("#%d: separateBucketFromBuilder(%s) returned no error; want error", i, tc.fullBuilderName)
		}
		if bucket != tc.expectedBucket {
			t.Errorf("#%d: separateBucketFromBuilder(%s) returned unexpected bucket: got %s; want %s", i, tc.fullBuilderName, bucket, tc.expectedBucket)
		}
		if builder != tc.expectedBuilder {
			t.Errorf("#%d: separateBucketFromBuilder(%s) returned unexpected builder: got %s; want %s", i, tc.fullBuilderName, builder, tc.expectedBuilder)
		}
	}
}
