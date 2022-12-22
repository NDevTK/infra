// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmlab

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"infra/libs/vmlab/api"
	"infra/libs/vmlab/internal/instance/gcloud"
)

func TestNewInstanceApi_gcloud(t *testing.T) {
	want, _ := gcloud.New()
	ins, err := NewInstanceApi(api.ProviderId_GCLOUD)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cmp.Equal(ins, want) {
		t.Errorf("InstanceApi = %v, but want %v", ins, want)
	}
}

func TestNewInstanceApi_unimplemented(t *testing.T) {
	ins, err := NewInstanceApi(api.ProviderId_UNKNOWN)
	if ins != nil {
		t.Errorf("InstanceApi = %v, but want nil", ins)
	}
	if err == nil {
		t.Errorf("error should not be nil")
	}
}
