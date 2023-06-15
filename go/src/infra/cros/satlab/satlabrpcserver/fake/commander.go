// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package fake

import "os/exec"

type FakeCommander struct {
	CmdOutput string
	Err       error
}

func (f *FakeCommander) Exec(_ *exec.Cmd) ([]byte, error) {
	return []byte(f.CmdOutput), f.Err
}
