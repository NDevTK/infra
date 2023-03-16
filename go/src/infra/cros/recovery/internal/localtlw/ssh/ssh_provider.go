// Copyright 2023 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"golang.org/x/crypto/ssh"
)

// SSHProvider provide access to SSH client manager.
//
// Provider gives option to use pool or create new client always.
type SSHProvider interface {
	Get(addr string) (SSHClient, error)
	Close() error
}

// Implementation of SSHProvider.
type sshProviderImpl struct {
	config *ssh.ClientConfig
}

// NewProvider creates new provider for use.
func NewProvider(config *ssh.ClientConfig) SSHProvider {
	return &sshProviderImpl{
		config: config,
	}
}

// Get provides SSH client for requested host.
func (c *sshProviderImpl) Get(addr string) (SSHClient, error) {
	return NewClient(addr, c.config)
}

// Close closing used resource of the provider.
func (c *sshProviderImpl) Close() error {
	return nil
}
