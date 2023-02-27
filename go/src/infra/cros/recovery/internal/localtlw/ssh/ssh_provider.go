// Copyright 2023 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"context"

	"go.chromium.org/luci/common/errors"
	"golang.org/x/crypto/ssh"

	"infra/libs/sshpool"
)

// SSHProvider provide access to SSH client manager.
//
// Provider gives option to use pool or create new client always.
type SSHProvider interface {
	GetContext(ctx context.Context, addr string) (SSHClient, error)
	Put(addr string, sc SSHClient)
	Close() error
}

// Implementation of SSHProvider.
type sshProviderImpl struct {
	config     *ssh.ClientConfig
	useSSHPool bool
	pool       *sshpool.Pool
}

// NewProvider creates new provider for use.
func NewProvider(config *ssh.ClientConfig) SSHProvider {
	p := &sshProviderImpl{
		config: config,
		// Use pool by default.
		// Following changes will made it an optional.
		useSSHPool: true,
	}
	if p.useSSHPool {
		p.pool = sshpool.New(p.config)
	}
	return p
}

// GetContext provides SSH client for requested host.
func (c *sshProviderImpl) GetContext(ctx context.Context, addr string) (SSHClient, error) {
	if c.useSSHPool {
		s, err := c.pool.GetContext(ctx, addr)
		if err != nil {
			return nil, errors.Annotate(err, "get contex from provier").Err()
		}
		return &sshClientImpl{s}, nil
	}
	return NewClient(ctx, addr, c.config)
}

// Put wrapper method to work with pool.
func (c *sshProviderImpl) Put(addr string, sc SSHClient) {
	if c.useSSHPool {
		c.pool.Put(addr, sc.Client())
	}
	// Do nothing. themethod only for pool option.
}

// Close closing used resource of the provider.
func (c *sshProviderImpl) Close() error {
	if c.pool != nil {
		if err := c.pool.Close(); err != nil {
			return errors.Annotate(err, "close provider").Err()
		}
	}
	// If we do not use pool then provider does not track created clients.
	return nil
}

// IsUseSSHPool reports if provider works with pool or not.
func (c *sshProviderImpl) IsUseSSHPool() bool {
	return c.useSSHPool
}
