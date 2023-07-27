// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
)

// DockerCp represents `docker cp`.
type DockerCp struct {
	Source      string // source
	Destination string // destination
}

func (c *DockerCp) Execute(ctx context.Context) (string, string, error) {
	args := []string{"cp", c.Source, c.Destination}
	return execute(ctx, dockerCmd, args)
}
