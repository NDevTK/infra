// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package external

import (
	"context"
	"testing"
)

// TestValidUFSHostname ensures we return err when hostname given for UFS client
// is nil
func TestValidUFSHostname(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, err := NewUFSClient(ctx, "")
	if err == nil {
		t.Errorf("Expected an error when invalid host passed")
	}
}
