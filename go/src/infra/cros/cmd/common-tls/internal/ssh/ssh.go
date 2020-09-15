// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ssh

import (
	"sync"

	"golang.org/x/crypto/ssh"
)

// Client is used by ClientPool to help users close connections.
type Client struct {
	*ssh.Client
	// KnownGood is used in deciding if the Client can be Put back
	// into the pool.
	KnownGood bool
}

func (c *Client) close() {
	if c.Client == nil {
		return
	}
	c.Close()
	c.Client = nil
}

// ClientPool is a pool of SSH clients to reuse.
// Clients are pooled by the hostname they are connected to.
//
// Users should call Get, which returns a Client from the pool if available,
// or creates and returns a new Client.
// The returned Client is not guaranteed to be good,
// e.g., the connection may have broken while the Client was in the pool.
// The user should Put the Client back into the pool after use.
//
// The user should Put the Client back into the pool after use.
// If the user knows the Client is still usable, it should set Client.KnownGood
// to be true before the Client is Put back.
// The user should not close the Client as ClientPool will close it.
//
// The user should Close the pool after use, to free any SSH Clients
// in the pool.
type ClientPool struct {
	mu     sync.Mutex
	pool   map[string][]*Client
	config *ssh.ClientConfig
}

// NewClientPool returns a new ClientPool.
func NewClientPool(c *ssh.ClientConfig) *ClientPool {
	return &ClientPool{
		pool:   make(map[string][]*Client),
		config: c,
	}
}

// Get returns a Client with KnownGood as false.
// The user should:
//  1) defer a Put back of the Client.
//  2) set the Client.KnownGood to be true before the Client is Put back if the
//     Client is usable.
func (p *ClientPool) Get(host string) (*Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for n := len(p.pool[host]) - 1; n >= 0; n-- {
		c := p.pool[host][n]
		p.pool[host] = p.pool[host][:n]
		s, err := c.NewSession()
		if err != nil {
			// This Client is probably bad, so close and stop using it.
			go c.close()
			continue
		}
		s.Close()
		// KnownGood is set to false as the user is responsible for
		// returning a good Client into the pool.
		c.KnownGood = false
		return c, nil
	}
	c, err := ssh.Dial("tcp", host, p.config)
	return &Client{c, false}, err
}

// Put puts the Client back into the pool if Client.KnownGood is true.
// Otherwise, the Client is closed.
func (p *ClientPool) Put(host string, c *Client) {
	if c == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if c.KnownGood {
		p.pool[host] = append(p.pool[host], c)
	} else {
		c.close()
	}
}

// Close closes all Clients in the ClientPool.
func (p *ClientPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for hostname, cs := range p.pool {
		for _, c := range cs {
			go c.close()
		}
		delete(p.pool, hostname)
	}
	return nil
}
