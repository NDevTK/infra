// Copyright 2021 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package env

import (
	"os"
	"testing"

	"infra/cros/internal/assert"
)

func TestRunningOnBot(t *testing.T) {
	os.Setenv("USER", "foo")
	assert.Assert(t, !RunningOnBot())

	os.Setenv("USER", "chrome-bot")
	assert.Assert(t, RunningOnBot())
}
