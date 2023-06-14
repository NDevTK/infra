// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package managers

import (
	"context"

	"go.chromium.org/chromiumos/config/go/test/api"
)

type ContainerManager interface {

	// Execute runs the exector
	// NOTE: these are for _once_ the manager is started/ready.
	StartContainer(context.Context, *api.StartTemplatedContainerRequest) (*api.StartContainerResponse, error)
	StopContainer(context.Context, string) error
	GetContainer(context.Context, string) (*api.GetContainerResponse, error)
	StartManager(context.Context, string) error
	StopManager(context.Context, string) error

	Initialize(context.Context) error

	// Response
}
