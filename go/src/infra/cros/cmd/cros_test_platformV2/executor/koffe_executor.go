// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package executor

import (
	managers "infra/cros/cmd/cros_test_platformV2/docker_managers"

	"go.chromium.org/chromiumos/config/go/test/api"
)

func NewKoffeeExecutor(ctr managers.ContainerManager, resp *api.TestSuite, req *api.Filter) *FilterExecutor {
	return &FilterExecutor{Ctr: ctr, resp: resp, binaryName: req.Container.ServiceName, containerPath: req.Container.ContainerPath}
}
