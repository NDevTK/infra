// Copyright 2023 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package buildbucket

import (
	"context"
	"strings"
	"testing"

	"infra/cros/internal/cmd"
)

// TestIsLUCIToolAuthed tests tryRunBase.IsLUCIToolAuthed() by mocking `bb auth-info` output.
func TestIsLUCIToolAuthed(t *testing.T) {
	c := &Client{}
	ctx := context.Background()
	for i, tc := range []struct {
		bbExitCode   int
		expectAuthed bool
		expectErr    bool
	}{
		{0, true, false},
		{1, false, false},
		{2, false, true},
	} {
		c.cmdRunner = FakeAuthInfoRunner("bb", tc.bbExitCode)
		authed, err := c.IsLUCIToolAuthed(ctx, "bb")
		if authed != tc.expectAuthed {
			t.Errorf("#%d: IsLUCIToolAuthed returned %v; want %v", i, authed, tc.expectAuthed)
		}
		if tc.expectErr && err == nil {
			t.Errorf("#%d: IsLUCIToolAuthed returned no error; want error", i)
		}
		if !tc.expectErr && err != nil {
			t.Errorf("#%d: IsLUCIToolAuthed returned an error; want no error. Returned error: %+v", i, err)
		}
	}
}

// TestEnsureLUCIToolAuthed tests tryRunBase.EnsureLUCIToolAuthed() by mocking `bb auth-info` output.
func TestEnsureLUCIToolAuthed(t *testing.T) {
	c := &Client{}
	ctx := context.Background()
	for i, tc := range []struct {
		bbExitCode  int
		expectError bool
	}{
		{0, false}, // User is logged in
		{1, true},  // User is logged out
		{2, true},  // Unexpected error
	} {
		c.cmdRunner = FakeAuthInfoRunner("bb", tc.bbExitCode)
		err := c.EnsureLUCIToolAuthed(ctx, "bb")
		if err == nil && tc.expectError {
			t.Errorf("#%d: EnsureLUCIToolAuthed returned no error; want error", i)
		} else if err != nil && !tc.expectError {
			t.Errorf("#%d: EnsureLUCIToolAuthed returned an error; want no error. Returned error: %+v", i, err)
		}
	}
}

// TestEnsureLUCIToolsAuthed tests tryRunBase.EnsureLUCIToolsAuthed() by mocking `bb auth-info` and `led auth-info` output.
func TestEnsureLUCIToolsAuthed(t *testing.T) {
	c := &Client{}
	ctx := context.Background()
	for i, tc := range []struct {
		bbExitCode  int
		ledExitCode int
		expectError bool
	}{
		{0, 0, false},
		{0, 1, true},
		{1, 1, true},
	} {
		c.cmdRunner = &cmd.FakeCommandRunnerMulti{CommandRunners: []cmd.FakeCommandRunner{
			FakeAuthInfoRunner("bb", tc.bbExitCode),
			FakeAuthInfoRunner("led", tc.ledExitCode),
		}}
		err := c.EnsureLUCIToolsAuthed(ctx, "bb", "led")
		if err == nil && tc.expectError {
			t.Errorf("#%d: EnsureLUCIToolsAuthed returned no error; want error", i)
		} else if err != nil && !tc.expectError {
			t.Errorf("#%d: EnsureLUCIToolsAuthed returned an error; want no error. Returned error: %+v", i, err)
		} else if err != nil {
			if tc.bbExitCode > 0 && !strings.Contains(err.Error(), "bb auth-login") {
				t.Errorf("#%d: EnsureLUCIToolsAuthed error didn't advise the user to run `bb auth-login`. Returned error: %+v", i, err)
			} else if tc.bbExitCode == 0 && strings.Contains(err.Error(), "bb auth-login") {
				t.Errorf("#%d: EnsureLUCIToolsAuthed error incorrectly advised the user to run `bb auth-login`. Returned error: %+v", i, err)
			}
			if tc.ledExitCode > 0 && !strings.Contains(err.Error(), "led auth-login") {
				t.Errorf("#%d: EnsureLUCIToolsAuthed error didn't advise the user to run `led auth-login`. Returned error: %+v", i, err)
			} else if tc.ledExitCode == 0 && strings.Contains(err.Error(), "led auth-login") {
				t.Errorf("#%d: EnsureLUCIToolsAuthed error incorrectly advised the user to run `led auth-login`. Returned error: %+v", i, err)
			}
		}
	}
}
