// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package managers

import (
	"context"
	"sync"

	"go.chromium.org/chromiumos/config/go/test/api"
)

type CTRDummy struct {
}

func NewCTRDummy() *CTRDummy {
	return &CTRDummy{}
}

func (ex *CTRDummy) StartContainer(ctx context.Context, req *api.StartContainerRequest) (*api.StartContainerResponse, error) {
	return nil, nil
}
func (ex *CTRDummy) StopContainer(ctx context.Context, foo string) error {
	return nil
}
func (ex *CTRDummy) GetContainer(ctx context.Context, foo string) (*api.GetContainerResponse, error) {
	return nil, nil
}

// Mmight not be required.
func (ex *CTRDummy) StartManager(ctx context.Context, foo string) error {
	return nil
}
func (ex *CTRDummy) StopManager(ctx context.Context, foo string) error {
	return nil
}
func (ex *CTRDummy) Initialize(ctx context.Context) error {
	return nil
}

type CloudDummy struct {
	CtrClient         string
	EnvVarsToPreserve []string

	wg              *sync.WaitGroup
	isServerRunning bool
}

func NewCloudDummy() *CloudDummy {
	return &CloudDummy{}
}

func (ex *CloudDummy) StartContainer(ctx context.Context, req *api.StartContainerRequest) (*api.StartContainerResponse, error) {
	return nil, nil
}
func (ex *CloudDummy) StopContainer(ctx context.Context, foo string) error {
	return nil
}
func (ex *CloudDummy) GetContainer(ctx context.Context, foo string) (*api.GetContainerResponse, error) {
	return nil, nil
}

// Mmight not be required.
func (ex *CloudDummy) StartManager(ctx context.Context, foo string) error {
	return nil
}

// Mmight not be required.
func (ex *CloudDummy) StopManager(ctx context.Context, foo string) error {
	return nil
}
func (ex *CloudDummy) Initialize(ctx context.Context) error {
	return nil
}
