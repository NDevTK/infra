// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package executor

import (
	"log"
	"os/exec"
	"time"
)

// IExecCommander is a interface to define which functions
// we can use in a commander.
type IExecCommander interface {
	Exec(*exec.Cmd) ([]byte, error)

	Start(*exec.Cmd) error

	Wait(*exec.Cmd) error
}

// ExecCommander is a wrapper uses to execute a command.
type ExecCommander struct{}

func (e *ExecCommander) Exec(cmd *exec.Cmd) ([]byte, error) {
	return cmd.Output()
}

func (e *ExecCommander) Start(cmd *exec.Cmd) error {
	return cmd.Start()
}

func (e *ExecCommander) Wait(cmd *exec.Cmd) error {
	return cmd.Wait()
}

// FakeCommander is used to fake a result when
// user write some test cases.
type FakeCommander struct {
	CmdOutput string
	Err       error
	FakeFn    func(*exec.Cmd) ([]byte, error)
}

func (f *FakeCommander) Exec(in *exec.Cmd) ([]byte, error) {
	if f.FakeFn != nil {
		return f.FakeFn(in)
	}
	return []byte(f.CmdOutput), f.Err
}

func (f *FakeCommander) Start(c *exec.Cmd) error {
	return nil
}

func (f *FakeCommander) Wait(c *exec.Cmd) error {
	go func() {
		if _, err := c.Stdout.Write([]byte(f.CmdOutput)); err != nil {
			log.Printf("can't write data to the pipe, got an error: %v", err)
		}
	}()

	time.Sleep(time.Millisecond)

	return nil
}
