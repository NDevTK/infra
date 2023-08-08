// Copyright 2019 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"testing"
)

// Running go test runs the smoke test.

func Test(t *testing.T) {
	t.Skip("requires python2")

	// TODO(maruel): Redirect child task output to log.
	if err := mainImpl(context.Background()); err != nil {
		t.Fatal(err)
	}
}
