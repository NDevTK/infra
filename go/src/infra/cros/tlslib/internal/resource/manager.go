// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package resource helps to manage AIP resources (https://google.aip.dev/121).
package resource

import (
	"fmt"
	"log"
	"sync"
)

// Resource represents a resource recommended by go/aip.
type Resource interface {
	// Close closes the resource.
	// It may be used in the Delete method of the resource.
	// This function may not be concurrency safe.
	// This function may be called multiple times.
	Close() error
}

// Manager tracks resources to support resource methods like Create, Delete,
// etc.
type Manager struct {
	resources sync.Map
}

// NewManager returns a new instance of Manager.
func NewManager() *Manager {
	return &Manager{
		resources: sync.Map{},
	}
}

// Add adds a entry in the manager for a resource.
func (m *Manager) Add(name string, r Resource) error {
	if _, loaded := m.resources.LoadOrStore(name, r); loaded {
		return fmt.Errorf("existing resource name %q", name)
	}
	return nil
}

// Remove deletes a resource by name and return it to the caller.
func (m *Manager) Remove(name string) (Resource, error) {
	r, ok := m.resources.LoadAndDelete(name)
	if !ok {
		return nil, fmt.Errorf("unknown name %q", name)
	}
	return r.(Resource), nil
}

// Close closes the manager and all resources it tracks.
func (m *Manager) Close() {
	m.resources.Range(func(key, value interface{}) bool {
		if err := value.(Resource).Close(); err != nil {
			log.Printf("Resource manager: close resource %q error: %s", key.(string), err)
		}
		return true
	})
}
