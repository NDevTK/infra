// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"context"

	"golang.org/x/time/rate"

	"infra/cros/recovery/internal/log"
)

// SSHProvider provide access to SSH client manager.
//
// Provider gives option to use pool or create new client always.
type SSHProvider interface {
	Get(ctx context.Context, addr string) (SSHClient, error)
	Close() error
	Config() Config
	SetUser(username string)
	Clone() SSHProvider
}

// Implementation of SSHProvider.
type sshProviderImpl struct {
	username string // overrides SSH config username
	config   Config
	limiter  *rate.Limiter
}

// NewProvider creates new provider for use.
//
//	clientConfig: SSH configuration to configure the new clients.
//	tlsConfig: Optional TLS configuration to establish SSH connections over TLS channel.
func NewProvider(config Config, limiter *rate.Limiter) SSHProvider {
	return &sshProviderImpl{
		config:  config,
		limiter: limiter,
	}
}

// Get provides SSH client for requested host.
func (c *sshProviderImpl) Get(ctx context.Context, addr string) (SSHClient, error) {
	if c.limiter != nil && !c.limiter.Allow() {
		log.Debugf(ctx, "Connection rate is limited to 1 connection per %v ms", 1000/c.limiter.Limit())
		if err := c.limiter.Wait(ctx); err != nil {
			return nil, err
		}
	}
	return NewClient(ctx, addr, c.username, c.config)
}

// Close closing used resource of the provider.
func (c *sshProviderImpl) Close() error {
	return nil
}

// Config returns SSH provider configuration.
func (c *sshProviderImpl) Config() Config {
	return c.config
}

// SetUser sets an username to override SSH config user.
func (c *sshProviderImpl) SetUser(username string) {
	c.username = username
}

// Clone creates a new SSH provider with its own username property.
// Changes to the username in the clone won't affect the original provider.
func (c *sshProviderImpl) Clone() SSHProvider {
	return &sshProviderImpl{
		username: c.username,
		config:   c.config,
		limiter:  c.limiter,
	}
}
