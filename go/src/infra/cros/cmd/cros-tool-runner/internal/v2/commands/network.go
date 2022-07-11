// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
)

// NetworkCreate represents `docker network create`.
type NetworkCreate struct {
	Name string // name of network to be created
}

func (c *NetworkCreate) Execute(ctx context.Context) (string, string, error) {
	args := []string{"network", "create", c.Name}
	return execute(ctx, dockerCmd, args)
}

// NetworkRemove represents `docker network remove`.
type NetworkRemove struct {
	Names []string // names of network to be removed
}

func (c *NetworkRemove) Execute(ctx context.Context) (string, string, error) {
	args := []string{"network", "remove"}
	args = append(args, c.Names...)
	return execute(ctx, dockerCmd, args)
}

// NetworkInspect represents `docker network inspect`.
type NetworkInspect struct {
	Names  []string // names of network to be inspected
	Format string   // value for the --format option
}

func (c *NetworkInspect) Execute(ctx context.Context) (string, string, error) {
	args := []string{"network", "inspect"}
	if utils.trim(c.Format) != "" {
		args = append(args, "-f", c.Format)
	}
	args = append(args, c.Names...)
	return execute(ctx, dockerCmd, args)
}

// NetworkList represents `docker network ls`.
type NetworkList struct {
	Names  []string // names of network to be listed
	Format string   // value for the --format option. e.g. {{.ID}}
}

func (c *NetworkList) Execute(ctx context.Context) (string, string, error) {
	args := []string{"network", "ls"}
	if utils.trim(c.Format) != "" {
		args = append(args, "--format", c.Format)
	}
	for _, name := range c.Names {
		args = append(args, "--filter", "name="+name)
	}
	return execute(ctx, dockerCmd, args)
}
