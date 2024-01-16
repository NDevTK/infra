// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"testing"
)

// TestCmd is a smoke test for the subcommand code. It just runs it to make sure
// it doesn't panic.
func TestCmd(t *testing.T) {
	for _, sub := range makeApp(nil).Commands {
		if sub.CommandRun == nil {
			continue
		}
		t.Run(sub.Name(), func(t *testing.T) {
			runApp(makeApp(nil), []string{sub.Name()})
		})
	}
}
