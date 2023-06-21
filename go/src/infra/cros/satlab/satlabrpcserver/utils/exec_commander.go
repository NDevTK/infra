// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package utils

import (
	"os/exec"
)

type IExecCommander interface {
	Exec(*exec.Cmd) ([]byte, error)

	Start(*exec.Cmd) error

	Wait(*exec.Cmd) error
}

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
