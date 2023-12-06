// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"context"
	"crypto/tls"
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
func (c *sshClientImpl) ForwardLocalToRemote(localAddr, remoteAddr string, errFunc func(error)) (*Forwarder, error) {
	connFunc := func() (net.Conn, error) {
		return c.Client().Dial(supportNetwork, remoteAddr)
	}
	l, err := net.Listen(supportNetwork, localAddr)
	if err != nil {
		return nil, err
	}
	return newForwarder(l, connFunc, errFunc)
}

// NewProxyClient establishes an authenticated SSH connection to target host
// using TLS channel as the underlying transport.
func NewProxyClient(ctx context.Context, addr string, config *ssh.ClientConfig, tlsConfig *tls.Config) (SSHClient, error) {
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		log.Printf("Error creating a new TLS connection: %s\n", err)
		return nil, errors.Annotate(err, "new proxy client").Err()
	}
	var c ssh.Conn
	var chans <-chan ssh.NewChannel
	var reqs <-chan *ssh.Request
	done := make(chan bool)
	go func() {
		c, chans, reqs, err = ssh.NewClientConn(conn, addr, config)
		done <- true
	}()
	select {
	case <-ctx.Done():
		conn.Close()
		return nil, errors.Annotate(ctx.Err(), "new proxy client").Err()
	case <-done:
	}
	if err != nil {
		return nil, errors.Annotate(err, "new proxy client").Err()
	}
	return &sshClientImpl{ssh.NewClient(c, chans, reqs)}, nil
}

// NewClient connects to SSH client to flesh connection.
func NewClient(ctx context.Context, addr string, config *ssh.ClientConfig) (SSHClient, error) {
	var c *ssh.Client
	var err error
	done := make(chan bool)
	go func() {
		c, err = ssh.Dial("tcp", addr, config)
		done <- true
	}()
	select {
	case <-ctx.Done():
		return nil, errors.Annotate(ctx.Err(), "new SSH client").Err()
	case <-done:
	}
	if err != nil {
		log.Printf("Error creating a new SSH client: %s\n", err)
		return nil, errors.Annotate(err, "new SSH client").Err()
	}
	return &sshClientImpl{c}, nil
}
