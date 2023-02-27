// Copyright 2023 The Chromium OS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"log"
	"net"

	"go.chromium.org/luci/common/errors"
	"golang.org/x/crypto/ssh"
)

// SSHClient provides base API to work with SSH client.
type SSHClient interface {
	NewSession() (*ssh.Session, error)
	IsAlive() bool
	Close() error
	Client() *ssh.Client
	ForwardLocalToRemote(localAddr, remoteAddr string, errFunc func(error)) (*Forwarder, error)
}

const (
	supportNetwork = "tcp"
)

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

// ForwardLocalToRemote creates a new Forwarder that forwards connections from localAddr to remoteAddr using s.
// network is passed to net.Listen. Only TCP networks are supported.
// localAddr is passed to net.Listen and typically takes the form "host:port" or "ip:port".
// remoteAddr uses the same format but is resolved by the remote SSH server.
// If non-nil, errFunc will be invoked asynchronously on a goroutine with connection or forwarding errors.
func (s *sshClientImpl) ForwardLocalToRemote(localAddr, remoteAddr string, errFunc func(error)) (*Forwarder, error) {
	connFunc := func() (net.Conn, error) {
		return s.Client().Dial(supportNetwork, remoteAddr)
	}
	l, err := net.Listen(supportNetwork, localAddr)
	if err != nil {
		return nil, err
	}
	return newForwarder(l, connFunc, errFunc)
}

// NewClient connects to SSH client to flesh connection.
func NewClient(addr string, config *ssh.ClientConfig) (SSHClient, error) {
	ssh, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Printf("New Client created with error: %s\n", err)
		return nil, errors.Annotate(err, "new SSH client").Err()
	}
	return &sshClientImpl{ssh}, nil
}
