// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package resource

import (
	"fmt"
	"sync"
)

// Resource represents a resource recommended by go/aip.
type Resource interface {
	Close() error
}

// Manager manage all resources.
type Manager struct {
	mu        sync.Mutex
	resources map[string]Resource
}

// NewManager returns a new instance of Manager.
func NewManager() *Manager {
	return &Manager{
		resources: make(map[string]Resource),
	}
}

// CreateResource creates a entry in the manager for a resource.
func (m *Manager) CreateResource(name string, r Resource) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.resources[name]; ok {
		return fmt.Errorf("existing resource name %q", name)
	}
	m.resources[name] = r
	return nil
}

// DeleteResource deletes a resource by name and return it to the caller.
func (m *Manager) DeleteResource(name string) (Resource, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.resources[name]
	if !ok {
		return nil, fmt.Errorf("unknown name %q", name)
	}
	delete(m.resources, name)
	return r, nil
}

// Close closes the manager and all resources it manages.
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, v := range m.resources {
		v.Close()
	}
}
