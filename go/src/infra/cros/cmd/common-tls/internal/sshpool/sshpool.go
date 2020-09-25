// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package sshpool helps manage a pool of SSH clients.
package sshpool

import (
	"sync"

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
	mu     sync.Mutex
	pool   map[string][]*ssh.Client
	config *ssh.ClientConfig
}

// New returns a new Pool.
func New(c *ssh.ClientConfig) *Pool {
	return &Pool{
		pool:   make(map[string][]*ssh.Client),
		config: c,
	}
}

// Get returns a good SSH client
func (p *Pool) Get(host string) (*ssh.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for n := len(p.pool[host]) - 1; n >= 0; n-- {
		c := p.pool[host][n]
		p.pool[host] = p.pool[host][:n]
		s, err := c.NewSession()
		if err != nil {
			// This SSH client is probably bad, so close and stop using it.
			go c.Close()
			continue
		}
		s.Close()
		return c, nil
	}
	c, err := ssh.Dial("tcp", host, p.config)
	return c, err
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
		go c.Close()
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
			go c.Close()
		}
		delete(p.pool, hostname)
	}
	return nil
}
