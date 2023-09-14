// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"sync"

	"golang.org/x/crypto/ssh"

	"infra/libs/sshtunnel"
)

// tunnelManager keeps track of SSH tunnels. Any client using tunnelManager must
// call Close() to ensure all tunnels are cleaned up after use.
type tunnelManager struct {
	mu      sync.Mutex
	tunnels []*sshtunnel.Tunnel
}

// newTunnelManager creates a new tunnelManager which should be closed after
// use.
func newTunnelManager() *tunnelManager {
	return &tunnelManager{}
}

// NewTunnel creates a Tunnel and returns it.
func (m *tunnelManager) NewTunnel(localAddr string, remoteAddr string, c *ssh.Client) (*sshtunnel.Tunnel, error) {
	t, err := sshtunnel.NewTunnel(localAddr, remoteAddr, c)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tunnels = append(m.tunnels, t)
	return t, nil
}

// Close closes all tunnels in the list.
func (m *tunnelManager) Close() {
	var wg sync.WaitGroup
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tunnels {
		wg.Add(1)
		go func(t *sshtunnel.Tunnel) {
			defer wg.Done()
			t.Close()
		}(t)
	}
	wg.Wait()
	m.tunnels = nil
	log.Printf("sshtunnel manager: All tunnels closed")
}
