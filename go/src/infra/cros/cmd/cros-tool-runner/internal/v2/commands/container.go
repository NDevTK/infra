// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
)

// ContainerStop represents `docker container stop`.
type ContainerStop struct {
	Names []string // names of containers to be removed
}

func (c *ContainerStop) Execute(ctx context.Context) (string, string, error) {
	args := []string{"container", "stop"}
	args = append(args, c.Names...)
	return execute(ctx, "docker", args)
}
