// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package gcloud

import (
	"errors"

	"infra/libs/vmlab/api"
)

// TODO(b/250961857): implement gcloud provider
// gcloudInstanceApi implements api.InstanceApi. The struct itself doesn't need
// to be public.
type gcloudInstanceApi struct{}

// New constructs a new api.InstanceApi with gcloud backend.
func New() (api.InstanceApi, error) {
	return &gcloudInstanceApi{}, nil
}

func (g *gcloudInstanceApi) Create(req *api.CreateVmInstanceRequest) (*api.VmInstance, error) {
	return nil, errors.New("not implemented")
}

func (g *gcloudInstanceApi) Delete(ins *api.VmInstance) error {
	return errors.New("not implemented")
}

func (g *gcloudInstanceApi) Cleanup(req *api.CleanupVmInstancesRequest) error {
	return errors.New("not implemented")
}
