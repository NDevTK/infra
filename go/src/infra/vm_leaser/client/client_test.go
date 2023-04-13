// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package client

import (
	"context"
	"testing"
)

// Test that instantiating a new client with an empty config fails.
func TestNewClient(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(ctx, nil)
	if client != nil {
		t.Error("client unexpectedly created")
	}
	if err == nil {
		t.Error("expected creating client to fail", err)
	}
}
