// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package kbqpb

import (
	"testing"

	"cloud.google.com/go/bigquery"
)

// TestActionIsValueSaver tests that a pointer to an action implements the ValueSaver interface.
// This test is trivial at runtime but nontrivial at compile time.
func TestActionIsValueSaver(t *testing.T) {
	t.Parallel()
	var _ bigquery.ValueSaver = &Action{}
}

// TestActionSave tests saving an action. New subtests should be added here for new fields
// to make sure that they are exported correctly.
func TestActionSave(t *testing.T) {
	t.Parallel()
	t.Run("model", func(t *testing.T) {
		m, _, err := (&Action{Name: "aaaaa", Model: "hi"}).Save()
		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if m["model"] != "hi" {
			t.Errorf("unexpected model: %q", m["model"])
		}
	})
	t.Run("board", func(t *testing.T) {
		m, _, err := (&Action{Name: "aaaaa", Board: "hi"}).Save()
		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if m["board"] != "hi" {
			t.Errorf("unexpected board: %q", m["board"])
		}
	})
}
