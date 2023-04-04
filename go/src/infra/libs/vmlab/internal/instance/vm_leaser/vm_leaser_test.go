// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmleaser

import (
	"testing"

	"infra/libs/vmlab/api"
)

func TestCreate(t *testing.T) {
	vmLeaser, err := New()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	ins, err := vmLeaser.Create(&api.CreateVmInstanceRequest{})
	if ins != nil {
		t.Errorf("VmInstance = %v, but want nil", ins)
	}
	if err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestDelete(t *testing.T) {
	vmLeaser, err := New()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	err = vmLeaser.Delete(&api.VmInstance{})
	if err == nil {
		t.Errorf("error should not be nil")
	}
}

func TestList(t *testing.T) {
	vmLeaser, err := New()
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	_, err = vmLeaser.List(&api.ListVmInstancesRequest{})
	if err == nil {
		t.Errorf("error should not be nil")
	}
}
