// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"context"
	"crypto/tls"
	"math/rand"
	"net"
	"time"

	"go.chromium.org/luci/common/errors"
	"golang.org/x/crypto/ssh"

	"infra/cros/recovery/internal/log"
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
	supportNetwork             = "tcp"
	numberOfConnectionAttempts = 5
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

func connectToProxy(ctx context.Context, network string, proxy *proxyConfig) (net.Conn, error) {
	log.Debugf(ctx, "Establishing TLS connection to %s", proxy.GetAddr())
	conn, err := tls.Dial(network, proxy.GetAddr(), proxy.GetConfig())
	if err != nil {
		log.Warningf(ctx, "Failed to connect to proxy: %v", err)
		return nil, errors.Annotate(err, "connect to proxy").Err()
	}
	return conn, nil
}

func connectToProxyWithRetry(ctx context.Context, network string, proxy *proxyConfig) (net.Conn, error) {
	waitBetweenRetriesInMillis := rand.Intn(1000)
	timeToSleep := time.Duration(waitBetweenRetriesInMillis) * time.Millisecond
	var err error
	var conn net.Conn
	done := make(chan bool)
	for i := 0; i < numberOfConnectionAttempts; i++ {
		if i > 0 {
			log.Debugf(ctx, "Retrying TLS connection")
		}
		go func() {
			conn, err = connectToProxy(ctx, network, proxy)
			done <- true
		}()
		select {
		case <-ctx.Done():
			return nil, errors.Annotate(ctx.Err(), "connect to proxy with retry").Err()
		case <-done:
		}
		if err == nil {
			break
		}
		if i < numberOfConnectionAttempts-1 {
			time.Sleep(timeToSleep)
		}
	}
	return conn, err
}

// newProxyClient establishes an authenticated SSH connection to the target host
// using TLS channel as the underlying transport.
func newProxyClient(ctx context.Context, sshConfig *ssh.ClientConfig, proxy *proxyConfig) (SSHClient, error) {
	var conn net.Conn
	var err error
	log.Debugf(ctx, "Proxy config: %+v", *proxy)
	tlsCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	conn, err = connectToProxyWithRetry(tlsCtx, supportNetwork, proxy)
	if err != nil {
		log.Errorf(ctx, "Error creating a new TLS connection: %s", err)
		return nil, errors.Annotate(err, "new proxy client").Err()
	}
	log.Debugf(ctx, "Connected to proxy %s", proxy.GetAddr())
	var c ssh.Conn
	var chans <-chan ssh.NewChannel
	var reqs <-chan *ssh.Request
	done := make(chan bool)
	go func() {
		log.Debugf(ctx, "Establishing ssh over TLS: %s", proxy.GetAddr())
		c, chans, reqs, err = ssh.NewClientConn(conn, proxy.GetAddr(), sshConfig)
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

// NewClient starts a client connection to the given SSH host.
func NewClient(ctx context.Context, addr, username string, config Config) (SSHClient, error) {
	sshConfig := config.GetSSHConfig(addr)
	if username != "" {
		sshConfig.User = username
	}
	log.Debugf(ctx, "SSH config: %+v", *sshConfig)
	if proxy := config.GetProxy(addr); proxy != nil && proxy.GetConfig() != nil {
		return newProxyClient(ctx, sshConfig, proxy)
	}
	var c *ssh.Client
	var err error
	done := make(chan bool)
	go func() {
		c, err = ssh.Dial("tcp", addr, sshConfig)
		done <- true
	}()
	select {
	case <-ctx.Done():
		return nil, errors.Annotate(ctx.Err(), "new SSH client").Err()
	case <-done:
	}
	if err != nil {
		log.Errorf(ctx, "Error creating a new SSH client: %s", err)
		return nil, errors.Annotate(err, "new SSH client").Err()
	}
	return &sshClientImpl{c}, nil
}
