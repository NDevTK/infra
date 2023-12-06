// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package sshpool helps manage a pool of SSH clients.
package sshpool

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Pool is a pool of SSH clients to reuse.
//
// Clients are pooled by the hostname they are connected to.
//
// Users should call Get, which returns a Client from the pool if available,
// or creates and returns a new Client.
// The returned Client is not guaranteed to be good,
// e.g., the connection may have broken while the Client was in the pool.
//
// The user should put the SSH client back into the pool after use.
// The user should not close the Client as Pool will close it if bad.
//
// The user should Close the pool after use, to free any SSH Clients in the pool.
type Pool struct {
	mu        sync.Mutex
	pool      map[string][]*ssh.Client
	config    *ssh.ClientConfig
	tlsConfig *tls.Config
	wg        sync.WaitGroup
}

// New returns a new Pool. The provided ssh config is used for new SSH
// connections if pool has none to reuse.
//
//	config: SSH configuration to configure the new clients.
//	tlsConfig: Optional TLS configuration to establish SSH connections over TLS channel.
func New(config *ssh.ClientConfig, tlsConfig *tls.Config) *Pool {
	return &Pool{
		pool:      make(map[string][]*ssh.Client),
		config:    config,
		tlsConfig: tlsConfig,
	}
}

// Get returns a good SSH client.
func (p *Pool) Get(host string) (*ssh.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for n := len(p.pool[host]) - 1; n >= 0; n-- {
		c := p.pool[host][n]
		p.pool[host] = p.pool[host][:n]
		if !verifyClientIsAlive(c) {
			log.Printf("sshpool Get: SSH client for %q is bad, closing it now!\n", host)
			p.closeClient(c)
			continue
		}
		return c, nil
	}
	log.Printf("sshpool Get: dial new SSH client for %q\n", host)
	if p.tlsConfig == nil {
		return ssh.Dial("tcp", host, p.config)
	}
	return p.getProxyClient(host)
}

// verifyClientIsAlive verifies if the client is alive and can continue to use.
func verifyClientIsAlive(c *ssh.Client) bool {
	// Verify by request.
	if _, _, err := c.SendRequest("keepalive@openssh.org", true, nil); err != nil {
		return false
	}
	// verify by ability work with sessions.
	if s, err := c.NewSession(); err != nil {
		return false
	} else {
		s.Close()
	}
	// All checks passed. The client should be good!
	return true
}

// getProxyClient returns an active SSH client established over TLS connection.
func (p *Pool) getProxyClient(host string) (*ssh.Client, error) {
	conn, err := tls.Dial("tcp", host, p.tlsConfig)
	if err != nil {
		log.Printf("sshpool getProxyClient: error creating a new TLS connection: %s\n", err)
		return nil, err
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, host, p.config)
	if err != nil {
		log.Printf("sshpool getProxyClient: error creating a new SSH connection over TLS channel: %s\n", err)
		conn.Close()
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

// GetContext returns a good SSH client within the context timeout.
func (p *Pool) GetContext(ctx context.Context, host string) (*ssh.Client, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("sshpool GetContext: timeout when trying to connect to %s", host)
		default:
			if c, err := p.Get(host); err == nil {
				return c, err
			}
			log.Printf("sshpool GetContext: retrying connection to %s", host)
			// Add a slight delay to not hammer the host with SSH connections.
			time.Sleep(2 * time.Second)
		}
	}
}

// Put puts the client back in the pool if it is good.
// Otherwise, the Client is closed.
func (p *Pool) Put(host string, c *ssh.Client) {
	if c == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	s, err := c.NewSession()
	if err != nil {
		// This SSH client is probably bad, so close and don't put into the pool.
		p.closeClient(c)
		return
	}
	s.Close()
	p.pool[host] = append(p.pool[host], c)
}

// Close closes all SSH clients in the Pool.
func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for hostname, cs := range p.pool {
		for _, c := range cs {
			p.closeClient(c)
		}
		delete(p.pool, hostname)
	}
	p.wg.Wait()
	return nil
}

// closeClient closes the supplied ssh.Client.
// Safe to pass in an already closed ssh.Client.
func (p *Pool) closeClient(c *ssh.Client) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		// Ignore the error returned in case the client is already closed.
		// Which could happen if the DUT was rebooted, but the ssh.Client
		// is being put back into the pool.
		log.Printf("sshpool closeClient: started waiting")
		_ = c.Close()
		log.Printf("sshpool closeClient: client closed")
	}()
}
