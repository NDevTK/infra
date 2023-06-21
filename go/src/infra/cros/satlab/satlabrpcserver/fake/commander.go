// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package fake

import (
	"log"
	"os/exec"
	"time"
)

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
				log.Println("Write data failed: ", err)
			}
		}()

		time.Sleep(time.Millisecond * 200)
	}
	return nil
}
