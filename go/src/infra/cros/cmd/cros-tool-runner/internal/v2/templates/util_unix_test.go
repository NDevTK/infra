// Copyright 2022 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

//go:build !windows
// +build !windows

package templates

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.chromium.org/chromiumos/config/go/test/lab/api"
)

func TestWriteToFile(t *testing.T) {
	now := time.Now()
	fullPath := filepath.Join("/tmp/", now.Format("write-file-test.20060102-150405.json"))
	err := TemplateUtils.writeToFile(fullPath, &api.IpEndpoint{
		Address: "xyz",
		Port:    123,
	})
	if err != nil {
		t.Fatalf("Unexpected error when writing to file")
	}
	t.Cleanup(func() { os.Remove(fullPath) })
	content, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Cannot open file")
	}
	expect := `{"address":"xyz","port":123}`
	actual := string(content)
	if actual != expect {
		t.Fatalf("File content doesn't match\nexpect: %s\nactual: %s", expect, actual)
	}
}
