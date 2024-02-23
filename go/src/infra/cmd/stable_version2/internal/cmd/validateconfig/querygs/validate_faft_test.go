// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package querygs

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var validateFaftUseCases = []struct {
	name        string
	in          string
	expectError bool
	want        string
}{
	{
		"happy path",
		"a-firmware/R10-11.12.13-aaaaaa",
		false,
		"gs://chromeos-image-archive/a-firmware/R10-11.12.13-aaaaaa/firmware_from_source.tar.bz2",
	},
	{
		"has gs in the path",
		"gs://chromeos-image-archive/a-firmware/R10-11.12.13-aaaaaa",
		true,
		"",
	},
}

func TestValidateFaft(t *testing.T) {
	t.Parallel()
	for _, uc := range validateFaftUseCases {
		ctx := context.Background()
		uc := uc
		var r Reader
		r.dld = fakeDownloader
		r.exst = fakeExistenceChecker
		got, err := r.validateFaft(uc.in)
		if uc.expectError && err == nil {
			t.Errorf("Expected error but got nil")
		}
		if diff := cmp.Diff(uc.want, got); diff != "" {
			t.Errorf("unexpected diff (-want +got): %s", diff)
		}
	}
}
