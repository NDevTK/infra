// Copyright 2022 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package commands

import (
	"context"
	"strings"
	"testing"
	"time"
)

type genericCommand struct {
	name string
	args []string
}

func (c *genericCommand) Execute(ctx context.Context) (string, string, error) {
	return execute(ctx, c.name, c.args)
}

func TestExecute_stdout(t *testing.T) {
	ctx := context.Background()
	expect := "CTRv2"
	cmd := &genericCommand{name: "echo", args: []string{expect}}
	stdout, stderr, err := cmd.Execute(ctx)
	trimmed := strings.TrimSpace(stdout)
	if trimmed != expect {
		t.Fatalf("stdout %s doesn't match expected %s", trimmed, expect)
	}
	if stderr != "" {
		t.Fatalf("stderr is not empty: %s", stderr)
	}
	if err != nil {
		t.Fatalf("err is not nil: %v", err)
	}
}

func TestExecute_stderr(t *testing.T) {
	ctx := context.Background()
	cmd := &genericCommand{name: "cat", args: []string{"/tmp"}}
	stdout, stderr, err := cmd.Execute(ctx)

	if stdout != "" {
		t.Fatalf("stdout is not empty: %s", stdout)
	}
	if stderr == "" {
		t.Fatalf("stderr should not be empty")
	}
	if err == nil {
		t.Fatalf("err should not be nil")
	}
}

func TestExecute_errOnly(t *testing.T) {
	ctx := context.Background()
	cmd := &genericCommand{name: "bad_command"}
	stdout, stderr, err := cmd.Execute(ctx)

	if stdout != "" {
		t.Fatalf("stdout is not empty: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr is not empty: %s", stderr)
	}
	if err == nil {
		t.Fatalf("err should not be nil")
	}
}

func TestExecute_deadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	cmd := &genericCommand{name: "ls", args: []string{"/tmp"}}
	_, _, err := cmd.Execute(ctx)

	if err.Error() != "context deadline exceeded" {
		t.Fatalf("err is not correct: %v", err)
	}
	if ctx.Err() != err {
		t.Fatalf("ctx.Err() is not correct: %v", ctx.Err())
	}
}

func TestExecute_cancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	cancel()
	cmd := &genericCommand{name: "ls", args: []string{"/tmp"}}
	_, _, err := cmd.Execute(ctx)

	if err.Error() != "context canceled" {
		t.Fatalf("err is not correct: %v", err)
	}
	if ctx.Err() != err {
		t.Fatalf("ctx.Err() is not correct: %v", ctx.Err())
	}
}
