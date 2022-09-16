// Copyright 2022 The ChromiumOS Authors.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package main

import (
	"context"
	"fmt"
	"os/exec"
	"testing"

	"infra/cros/internal/cmd"
)

// TestIsBBAuthed tests myjobRunBase.IsBBAuthed() by mocking `bb auth-info` output.
func TestIsBBAuthed(t *testing.T) {
	var m myjobRunBase
	ctx := context.Background()
	for _, testCase := range []struct {
		bbExitCode   int
		expectAuthed bool
		expectErr    bool
	}{
		{0, true, false},
		{1, false, false},
		{2, false, true},
	} {
		// As far as I can tell, Go doesn't have a great way to mock a command's exit code.
		// Instead, we can use exec to force errors with the desired exit status.
		cmdFailError := exec.Command("bash", "-c", fmt.Sprintf("exit %d", testCase.bbExitCode)).Run()
		m.cmdRunner = cmd.FakeCommandRunner{
			ExpectedCmd: []string{"bb", "auth-info"},
			FailCommand: testCase.bbExitCode != 0,
			FailError:   cmdFailError,
		}
		authed, err := m.IsBBAuthed(ctx)
		if authed != testCase.expectAuthed {
			t.Errorf("IsBBAuthed returned %v with `bb auth-info` exit code %d; want %v", authed, testCase.bbExitCode, testCase.expectAuthed)
		}
		if testCase.expectErr && err == nil {
			t.Errorf("IsBBAuthed returned no error with `bb auth-info` exit code %d; want error", testCase.bbExitCode)
		}
		if !testCase.expectErr && err != nil {
			t.Errorf("IsBBAuthed returned an error with `bb auth-info` exit code %d; want no error. Returned error: %+v", testCase.bbExitCode, err)
		}
	}
}
