// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package resource

import (
	"testing"
)

type Foo string

func (f Foo) Close() error {
	return nil
}

func TestResourceManager(t *testing.T) {
	t.Parallel()
	m := NewManager()
	var r Foo = "resource"
	t.Run("Create a new resource", func(t *testing.T) {
		name := "name"
		err := m.CreateResource(name, &r)
		if err != nil {
			t.Errorf("CreateResource(%q, &%#v) failed: %s", name, r, err)
		}
	})
	t.Run("Create resource with conflict name", func(t *testing.T) {
		name := "name"
		err := m.CreateResource(name, &r)
		if err == nil {
			t.Errorf("CreateResource(%q, &%#v) succeeded for duplicate name, want error", name, r)
		}
	})
	t.Run("Delete an existing resource", func(t *testing.T) {
		name := "name"
		want := r
		got, err := m.DeleteResource(name)
		if err != nil {
			t.Errorf(`DeleteResource(%q) failed: %s`, name, err)
		}
		if *(got.(*Foo)) != want {
			t.Errorf("DeleteResource(%q) = %#v, want %#v", name, got, want)
		}
	})
	t.Run("Delete non existing resource", func(t *testing.T) {
		name := "name"
		_, err := m.DeleteResource(name)
		if err == nil {
			t.Errorf("DeleteResource(%q) succeeded for non existing resource, want error", name)
		}
	})
}
