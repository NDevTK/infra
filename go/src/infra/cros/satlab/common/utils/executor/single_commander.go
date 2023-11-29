// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package executor

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

// IExecCommander is a interface to define which functions
// we can use in a commander.
type IExecCommander interface {
	// CombinedOutput runs the command and returns its combined standard
	// output and standard error. It wrpas the error inside `fmt.Errorf`.
	//
	// if we want to parse *ExitError, we can use `errors.Unwrap` to achieve.
	CombinedOutput(*exec.Cmd) ([]byte, error)

	// Output runs the command and returns its standard output.
	// Any returned error will usually be of type *ExitError.
	// If c.Stderr was nil, Output populates ExitError.Stderr.
	Output(*exec.Cmd) ([]byte, error)

	// Start starts the specified command but does not wait for it to complete.
	//
	// If Start returns successfully, the c.Process field will be set.
	//
	// After a successful call to Start the Wait method must be called in
	Start(*exec.Cmd) error

	// Wait waits for the command to exit and waits for any copying to
	// stdin or copying from stdout or stderr to complete.
	//
	// The command must have been started by Start.
	//
	// The returned error is nil if the command runs, has no problems
	// copying stdin, stdout, and stderr, and exits with a zero exit
	// status.
	//
	// If the command fails to run or doesn't complete successfully, the
	// error is of type *ExitError. Other error types may be
	// returned for I/O problems.
	//
	// If any of c.Stdin, c.Stdout or c.Stderr are not an *os.File, Wait also waits
	// for the respective I/O loop copying to or from the process to complete.
	//
	Wait(*exec.Cmd) error
}

// ExecCommander is a wrapper uses to execute a command.
type ExecCommander struct{}

func (e *ExecCommander) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	s, err := cmd.CombinedOutput()
	if err != nil {
		return s, fmt.Errorf("%w: %s", err, string(s[:]))
	}
	return s, err
}

func (e *ExecCommander) Output(cmd *exec.Cmd) ([]byte, error) {
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

func (f *FakeCommander) CombinedOutput(in *exec.Cmd) ([]byte, error) {
	if f.FakeFn != nil {
		return f.FakeFn(in)
	}
	return []byte(f.CmdOutput), f.Err
}

func (f *FakeCommander) Output(in *exec.Cmd) ([]byte, error) {
	if f.FakeFn != nil {
		return f.FakeFn(in)
	}
	return []byte(f.CmdOutput), f.Err
}

func (f *FakeCommander) Start(in *exec.Cmd) error {
	if f.FakeFn != nil {
		_, err := f.FakeFn(in)
		return err
	}

	return f.Err
}

func (f *FakeCommander) Wait(c *exec.Cmd) error {
	go func() {
		if _, err := c.Stdout.Write([]byte(f.CmdOutput)); err != nil {
			log.Printf("can't write data to the pipe, got an error: %v", err)
		}
	}()

	time.Sleep(time.Second * 2)

	return nil
}
