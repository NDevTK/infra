// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package executor

import (
	"bytes"
	"io"
	"log"
	"os/exec"

	"infra/cros/satlab/satlabrpcserver/utils"
)

// MultipleCommandsExecutor a executor for executing multiple commands.
//
// Support to execute multiple command like this `ls -a | grep -o | sort -r`
// This requires `bash` support like this `exec.Command("bash", "-c", "ls -a | grep -o | sort -r")`
//
// This `struct` is created by combining multiple commands by `io.Pipe`. The
// concept is to pipe the previous command to next command.
type MultipleCommandsExecutor struct {
	// cmds the commands that we want to combine to execute
	cmds []*exec.Cmd
	// outs the writers that used for command's communication
	outs []*io.PipeWriter
	// res final result buffer
	res *bytes.Buffer
	// executor that used for executing the command
	executor utils.IExecCommander
}

// New the function used for creating an `MultipleCommandsExecutor`
//
// For example:
// `ls -a | grep -o | sort -r`
// multipleCmdsExecutor := New(
//
//	exec.Command("ls", "-a"),
//	exec.Command("grep", "-o"),
//	exec.Command("sort", "-r"),
//
// )
func New(cmds ...*exec.Cmd) *MultipleCommandsExecutor {
	var outBuffer bytes.Buffer
	lastIdx := len(cmds) - 1
	outs := make([]*io.PipeWriter, lastIdx)

	// use `io.Pipe` to make a connection to all commands
	// For example:
	// cmd1 -> pipe -> cmd2 -> pipe -> cmd3 -> ... -> out buffer
	for idx := 0; idx < lastIdx; idx++ {
		inPipe, outPipe := io.Pipe()
		cmds[idx].Stdout = outPipe
		cmds[idx+1].Stdin = inPipe
		outs[idx] = outPipe
	}
	cmds[lastIdx].Stdout = &outBuffer

	return &MultipleCommandsExecutor{
		cmds:     cmds,
		outs:     outs,
		res:      &outBuffer,
		executor: &utils.ExecCommander{},
	}
}

// Exec start executing the commands and waiting for the result.
// We will close all the `io.PipeWriter` after finished.
func (c *MultipleCommandsExecutor) Exec() (*bytes.Buffer, error) {
	// Start all commands
	for _, cmd := range c.cmds {
		if err := c.executor.Start(cmd); err != nil {
			return nil, err
		}
	}

	// Wait for the commands' result
	for _, cmd := range c.cmds {
		if err := c.executor.Wait(cmd); err != nil {
			return nil, err
		}
	}

	// Close all `io.PipeWriter`
	for idx := 0; idx < len(c.cmds)-1; idx++ {
		if err := c.outs[idx].Close(); err != nil {
			// if we face closing pipe problem.
			// As closing pipe doesn't affect the result.
			// We stop immediately, and then log the error message.
			log.Printf("Close the `io.PipeWriter` failed, got an error: {%v}", err)
			break
		}
	}

	return c.res, nil
}
