// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmlab

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"infra/libs/vmlab/api"
	"infra/libs/vmlab/internal/image/cloudsdk"
)

func TestNewImageApi_cloudsdk(t *testing.T) {
	want, _ := cloudsdk.New()
	imageApi, err := NewImageApi(api.ProviderId_CLOUDSDK)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cmp.Equal(imageApi, want) {
		t.Errorf("ImageApi = %v, but want %v", imageApi, want)
	}
}

func TestNewImageApi_unimplemented(t *testing.T) {
	imageApi, err := NewImageApi(api.ProviderId_GCLOUD)
	if imageApi != nil {
		t.Errorf("ImageApi = %v, but want nil", imageApi)
	}
	if err == nil {
		t.Errorf("error should not be nil")
	}
}
