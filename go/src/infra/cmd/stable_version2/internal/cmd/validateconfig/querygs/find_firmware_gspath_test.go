// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package querygs

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"go.chromium.org/luci/common/gcloud/gs"
)

func TestMilestonesInOrder(t *testing.T) {
	expected := []int{5, 6, 7, 8, 9, 10, 4, 3, 2, 1}
	actual := milestonesInOrder(5)
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected diff (-want, +got):\n%s", diff)
	}
}

func TestFindFirmwarePath(t *testing.T) {
	t.Parallel()
	var r Reader
	r.dld = fakeDownloader
	r.exst = fakeExistenceChecker
	_, err := r.FindFirmwarePath("a", 10, 10, 10, "10")
	if err != nil {
		t.Error(err)
	}
}

// fakeDownloader successfully produces an empty byte slice.
func fakeDownloader(gsPath gs.Path) ([]byte, error) {
	return []byte(""), nil
}

// fakeExistenceChecker always concludes that its argument exists
func fakeExistenceChecker(gsPath gs.Path) error {
	return nil
}
