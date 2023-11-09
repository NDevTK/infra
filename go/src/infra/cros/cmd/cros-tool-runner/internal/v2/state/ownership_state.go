// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package state

import "log"

// OwnershipRecorder is the interface to manage ownership state in the server.
// If an entity (container/network) is started/created by the server, the server
// "owns" the entity and is responsible to recycle (stop/remove) at server exit.
// An ownership is identified by the entity name (supplied by client) and the ID
// (generated by docker). Ownership with the same name can be redeclared to a
// different ID; the old ID will be abandoned permanently.
// Redeclaration may happen due to local development, client bugs, and/or
// unexpected stop of containers. As long as an action is accepted by the
// underlying docker commands, the server follows suit and records the state.
type OwnershipRecorder interface {
	// RecordOwnership declares ownership of a name and the associated ID. The
	// same name can be redeclared to overwrite the existing ID.
	RecordOwnership(name string, id string)
	// HasOwnership checks current ownership.
	HasOwnership(name string, id string) bool
	// RemoveOwnership revokes ownership of a name.
	RemoveOwnership(name string)
	// GetIdsToClearOwnership returns current IDs in reverse order of
	// declarations. The IDs can be used to recycle entities.
	GetIdsToClearOwnership() []string
	// GetMapping returns a copy of name to ID mapping.
	GetMapping() map[string]string
	// Clear reset the state.
	Clear()
	// GetIdForOwner returns a copy of ID mapping to name.
	GetIdForOwner(name string) string
}

// ownershipState is the implementation of OwnershipRecorder. It uses a history
// array to record all declarations on names, and a map of name to ID to track
// the current state.
type ownershipState struct {
	history []string
	mapping map[string]string
}

// newOwnershipState returns an instance of ownershipState.
func newOwnershipState() OwnershipRecorder {
	return &ownershipState{history: make([]string, 0), mapping: make(map[string]string)}
}

func (o *ownershipState) RecordOwnership(name string, id string) {
	o.history = append(o.history, name)
	if val, ok := o.mapping[name]; ok {
		log.Printf("warning: updating name %s ownership id from %s to %s", name, val, id)
	}
	o.mapping[name] = id
}

func (o *ownershipState) HasOwnership(name string, id string) bool {
	if val, ok := o.mapping[name]; ok {
		return id == val
	}
	return false
}

func (o *ownershipState) RemoveOwnership(name string) {
	delete(o.mapping, name)
	log.Printf("warning: name %s ownership has been removed", name)
}

func (o *ownershipState) GetIdsToClearOwnership() []string {
	size := len(o.history)
	result := make([]string, len(o.mapping))
	cloneMapping := make(map[string]string, len(o.mapping))
	for k, v := range o.mapping {
		cloneMapping[k] = v
	}

	index := 0
	for i := range o.history {
		name := o.history[size-i-1]
		if val, ok := cloneMapping[name]; ok {
			result[index] = val
			index++
			delete(cloneMapping, name)
		}
	}
	return result
}

func (o *ownershipState) GetMapping() map[string]string {
	return o.mapping
}

func (o *ownershipState) Clear() {
	o.history = make([]string, 0)
	o.mapping = make(map[string]string)
}

func (o *ownershipState) GetIdForOwner(name string) string {
	if val, ok := o.mapping[name]; ok {
		return val
	}
	return ""
}