// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package tsmon

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocateFile_NotFound(t *testing.T) {
	candidates := []string{"/tmp/nonexistent-file-1", "/tmp/nonexistent-file-2"}

	_, err := locateFile(candidates)

	if err == nil {
		t.Errorf("Expect error but got nil")
	}
}

func TestLocateFile_Found(t *testing.T) {
	now := time.Now()
	fullPath := filepath.Join("/tmp/", now.Format("existent-file-2006-01-02T15:04:05.999999999Z07:00"))
	_, _ = os.Create(fullPath)
	t.Cleanup(func() { _ = os.Remove(fullPath) })
	candidates := []string{"/tmp/nonexistent-file-1", "/tmp/nonexistent-file-2", fullPath}

	found, err := locateFile(candidates)

	if err != nil {
		t.Errorf("Expect no error but got %v", err)
	}
	if found != fullPath {
		t.Errorf("Expect %s but got %v", fullPath, found)
	}
}

func TestInit_NotBroken(t *testing.T) {
	_ = Init()
}

func TestShutdown_NotBroken(t *testing.T) {
	Shutdown()
}
