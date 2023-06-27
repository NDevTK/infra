// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executor

import (
	managers "infra/cros/cmd/cros_test_platformV2/docker_managers"

	buildapi "go.chromium.org/chromiumos/config/go/build/api"
	"go.chromium.org/chromiumos/config/go/test/api"
)

func NewKarbonExecutor(ctr managers.ContainerManager, req *api.Filter, containerMetadata map[string]*buildapi.ContainerImageInfo) (*FilterExecutor, error) {
	// TODO, Given the request, make the correct filter.

	return newFilterExecutor(ctr, req, containerMetadata)
}
