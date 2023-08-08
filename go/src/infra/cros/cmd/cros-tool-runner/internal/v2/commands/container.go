// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"context"
	"strings"
)

// ContainerStop represents `docker container stop`.
type ContainerStop struct {
	Names []string // names (or ids) of containers to be removed
}

func (c *ContainerStop) Execute(ctx context.Context) (string, string, error) {
	args := []string{"container", "stop"}
	args = append(args, c.Names...)
	return execute(ctx, dockerCmd, args)
}

// ContainerInspect represents `docker container inspect`.
type ContainerInspect struct {
	Names  []string // names (or ids) of containers to be inspected
	Format string   // value for the --format option. e.g. {{.Id}}
}

func (c *ContainerInspect) Execute(ctx context.Context) (string, string, error) {
	args := []string{"container", "inspect"}
	if strings.TrimSpace(c.Format) != "" {
		args = append(args, "-f", c.Format)
	}
	args = append(args, c.Names...)
	return execute(ctx, dockerCmd, args)
}

// ContainerPort represents `docker container port`.
type ContainerPort struct {
	Name string // name (or id) of the container
}

func (c *ContainerPort) Execute(ctx context.Context) (string, string, error) {
	args := []string{"container", "port", c.Name}
	return execute(ctx, dockerCmd, args)
}

// HostIpAddresses represents `hostname --all-ip-addresses`. Output is all IP
// addresses separated by a whitespace.
type HostIpAddresses struct{}

func (c *HostIpAddresses) Execute(ctx context.Context) (string, string, error) {
	args := []string{"--all-ip-addresses"}
	return execute(ctx, "hostname", args)
}
