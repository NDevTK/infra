// Copyright 2023 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"context"
	"log"

	"go.chromium.org/luci/common/errors"
	"golang.org/x/crypto/ssh"
)

// SSHClient provides base API to work with SSH client.
type SSHClient interface {
	NewSession() (*ssh.Session, error)
	IsAlive() bool
	Close() error
	Client() *ssh.Client
}

// Implementation of SSHClient.
type sshClientImpl struct {
	client *ssh.Client
}

// Close closing the native client.
func (c *sshClientImpl) Close() error {
	err := c.client.Close()
	return errors.Annotate(err, "close ssh client").Err()
}

// NewSession creates new SSH session to execute commands.
func (c *sshClientImpl) NewSession() (*ssh.Session, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return nil, errors.Annotate(err, "new session").Err()
	}
	return session, nil
}

// IsAlive checks if client is alive or not.
func (c *sshClientImpl) IsAlive() bool {
	_, _, err := c.client.SendRequest("keepalive@openssh.org", true, nil)
	return err == nil
}

// Client provide access to native client.
// TODO: Remove as this created only to support current state and any manipulation need to be wrapped to special functions.
func (c *sshClientImpl) Client() *ssh.Client {
	return c.client
}

// NewClient connects to SSH client to flesh connection.
func NewClient(ctx context.Context, addr string, config *ssh.ClientConfig) (SSHClient, error) {
	log.Printf("New Client Starting... with addr:%q\n", addr)
	ssh, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Printf("New Client created with error: %s\n", err)
		return nil, errors.Annotate(err, "new SSH client").Err()
	}
	log.Println("New Client created!")
	return &sshClientImpl{ssh}, nil
}
