// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package executor

import (
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
}

func (f *FakeCommander) Exec(_ *exec.Cmd) ([]byte, error) {
	return []byte(f.CmdOutput), f.Err
}

func (f *FakeCommander) Start(c *exec.Cmd) error {
	return nil
}

func (f *FakeCommander) Wait(c *exec.Cmd) error {
	if c.Stdin != nil {
		l := 1024
		data := make([]byte, l)
		if _, err := c.Stdin.Read(data); err != nil {
			return err
		}

		go func() {
			// take the byte until meet the first \x0
			idx := 0
			for ; idx < l && data[idx] != 0; idx++ {
			}

			if _, err := c.Stdout.Write(data[:idx]); err != nil {
			}
		}()

		time.Sleep(time.Millisecond * 200)
	}
	return nil
}
