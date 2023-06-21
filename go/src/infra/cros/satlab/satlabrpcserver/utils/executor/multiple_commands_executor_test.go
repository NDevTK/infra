// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package executor

import (
	"os/exec"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"infra/cros/satlab/satlabrpcserver/fake"
)

func TestExecuteMultipleCommandShouldSuccess(t *testing.T) {
	commandsExecutor := New(
		exec.Command("cmd1"),
		exec.Command("cmd2"),
		exec.Command("cmd3"),
		exec.Command("cmd4"),
	)
	commandsExecutor.executor = &fake.FakeCommander{}

	// As `io.Pipe`.Write() is a blocking api
	// we put it to the background to execute.
	go func() {
		// Sent the first command output to simulate the final result
		// the result should pass through.
		if _, err := commandsExecutor.outs[0].Write([]byte("cmd1 out")); err != nil {
			t.Errorf("write some data to pipe writer failed, got an {%v}", err)
		}
	}()

	time.Sleep(time.Millisecond * 200)

	result, err := commandsExecutor.Exec()

	if err != nil {
		t.Errorf("execute multiple commands should success, got an {%v}", err)
		return
	}

	res := string(result.Bytes())
	if diff := cmp.Diff(res, "cmd1 out"); diff != "" {
		t.Errorf("Expect cmd1 out, got {%v}", res)
	}
}
