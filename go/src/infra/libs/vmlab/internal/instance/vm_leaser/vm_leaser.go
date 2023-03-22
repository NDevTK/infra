// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmleaser

import (
	"errors"

	"infra/libs/vmlab/api"
)

// vmLeaserInstanceApi implements api.InstanceApi
//
// The struct itself doesn't need to be public.
type vmLeaserInstanceApi struct{}

// New constructs a new api.InstanceApi with VM Leaser Service backend.
func New() (api.InstanceApi, error) {
	return &vmLeaserInstanceApi{}, nil
}

func (g *vmLeaserInstanceApi) Create(req *api.CreateVmInstanceRequest) (*api.VmInstance, error) {
	return nil, errors.New("not implemented")
}

func (g *vmLeaserInstanceApi) Delete(ins *api.VmInstance) error {
	return errors.New("not implemented")
}

func (g *vmLeaserInstanceApi) Cleanup(req *api.CleanupVmInstancesRequest) error {
	return errors.New("not implemented")
}
