// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"context"
	"crypto/tls"

	"golang.org/x/crypto/ssh"
)

// SSHProvider provide access to SSH client manager.
//
// Provider gives option to use pool or create new client always.
type SSHProvider interface {
	Get(ctx context.Context, addr string) (SSHClient, error)
	Close() error
	Config() *ssh.ClientConfig

	// WithUser returns a new SSHProvider with an updated config.User as username.
	// This SSHProvider instance remains unchanged.
	WithUser(username string) SSHProvider
}

// Implementation of SSHProvider.
type sshProviderImpl struct {
	config    *ssh.ClientConfig
	tlsConfig *tls.Config
}

// NewProvider creates new provider for use.
//
//	clientConfig: SSH configuration to configure the new clients.
//	tlsConfig: Optional TLS configuration to establish SSH connections over TLS channel.
func NewProvider(clientConfig *ssh.ClientConfig, tlsConfig *tls.Config) SSHProvider {
	return &sshProviderImpl{
		config:    clientConfig,
		tlsConfig: tlsConfig,
	}
}

// Get provides SSH client for requested host.
func (c *sshProviderImpl) Get(ctx context.Context, addr string) (SSHClient, error) {
	if c.tlsConfig != nil {
		return NewProxyClient(ctx, addr, c.config, c.tlsConfig)
	}
	return NewClient(ctx, addr, c.config)
}

// Close closing used resource of the provider.
func (c *sshProviderImpl) Close() error {
	return nil
}

func (c *sshProviderImpl) Config() *ssh.ClientConfig {
	return c.config
}

// WithUser returns a new SSHProvider with an updated config.User as username.
// This SSHProvider instance remains unchanged.
func (c *sshProviderImpl) WithUser(username string) SSHProvider {
	newConfig := cloneSSHConfig(c.config)
	newConfig.User = username
	return NewProvider(newConfig, c.tlsConfig)
}
