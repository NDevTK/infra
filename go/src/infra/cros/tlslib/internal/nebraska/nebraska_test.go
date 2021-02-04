// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package nebraska

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

func newFileReader(cnt string, err error) func(string) ([]byte, error) {
	return func(string) ([]byte, error) { return []byte(cnt), err }
}

func TestNebraskaStartSuccess(t *testing.T) {
	t.Parallel()
	const port = 12345
	env := &Environment{
		fileReader: newFileReader(fmt.Sprintf("%d", port), nil),
		httpDownloader: func(string) (*http.Response, error) {
			return &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString("nebraska content"))}, nil
		},
		commander: func(c string, args ...string) *exec.Cmd {
			if c == "gsutil" {
				return exec.Command("true")
			}
			// A fake long running service.
			return exec.Command("tail", "-f", "/dev/null")
		},
	}
	n := NewServer(env)
	err := n.Start("gs://bucket/board/build", nil, "URL")
	if err != nil {
		t.Errorf("Start() failed: %s", err)
	}
	if n.Port != port {
		t.Errorf("Start() failed, got port %d, want %d", n.Port, port)
	}

	c := n.cmd
	done := make(chan error, 1)
	go func() { done <- c.Wait() }()
	n.Close()
	select {
	case <-done:
		break
	case <-time.After(time.Second):
		t.Errorf("Close() failed to kill Nebraska process")
	}
	if _, err := os.Stat(n.metadataDir); os.IsExist(err) {
		t.Errorf("Close() failed to remove metadata dir")
	}
	if _, err := os.Stat(n.runtimeRoot); os.IsExist(err) {
		t.Errorf("Close() failed to remove runtime root dir")
	}
	if n.cmd != nil {
		t.Errorf("Close() didn't reset cmd")
	}
}

func TestNebraskaStartTimeout(t *testing.T) {
	t.Parallel()
	env := &Environment{
		fileReader: newFileReader("", fmt.Errorf("")),
		httpDownloader: func(string) (*http.Response, error) {
			return &http.Response{Body: ioutil.NopCloser(bytes.NewBufferString("nebraska content"))}, nil
		},
		commander: func(c string, args ...string) *exec.Cmd { return exec.Command("true") },
	}
	n := NewServer(env)
	err := n.Start("gs://bucket/board/build", nil, "URL")
	if err == nil {
		t.Errorf("Start() succeeded when start Nebraska timeout, want error")
	}
}
