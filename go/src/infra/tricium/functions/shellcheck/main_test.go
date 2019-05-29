// Copyright 2018 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// +build !windows

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"infra/tricium/functions/shellcheck/runner"
)

const (
	testInputDir       = "testdata"
	triciumResultsPath = "tricium/data/results.json"
)

func TestRun(t *testing.T) {
	r := &runner.Runner{
		Path:   findShellCheckBin(),
		Dir:    testInputDir,
		Logger: runnerLogger,
	}
	version, err := r.Version()
	if err != nil {
		t.Skip("no valid shellcheck bin found; skipping test")
	}
	if !strings.HasPrefix(version, "0.6.") {
		t.Skipf("got shellcheck version %q want 0.6.x; skipping test", version)
	}

	outputDir, err := ioutil.TempDir("", "tricium-shellcheck-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(outputDir)

	run(r, testInputDir, outputDir, "*.other,*.sh")

	resultsData, err := ioutil.ReadFile(filepath.Join(outputDir, triciumResultsPath))
	if err != nil {
		panic(err)
	}

	var results map[string][]map[string]interface{}

	if err := json.Unmarshal(resultsData, &results); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	comments, ok := results["comments"]
	if !ok {
		t.Fatalf("results have no comments key")
	}

	if len(comments) != 2 {
		t.Fatalf("got %d comments want 2", len(comments))
	}

	comment := comments[0]
	assertMapKeyEqual(t, comment, "category", "ShellCheck/SC2034")
	assertMapKeyEqual(t, comment, "message",
		"warning: unused appears unused. Verify use (or export if used externally).")
	assertMapKeyEqual(t, comment, "url",
		"https://github.com/koalaman/shellcheck/wiki/SC2034")
	assertMapKeyEqual(t, comment, "path", "bad.sh")
	assertMapKeyEqual(t, comment, "startLine", float64(3))
	assertMapKeyEqual(t, comment, "endLine", float64(3))
	assertMapKeyEqual(t, comment, "startChar", float64(1))
	assertMapKeyEqual(t, comment, "endChar", float64(7))

	comment = comments[1]
	assertMapKeyEqual(t, comment, "category", "ShellCheck/SC1037")
	assertMapKeyEqual(t, comment, "message",
		"error: Braces are required for positionals over 9, e.g. ${10}.")
	assertMapKeyEqual(t, comment, "url",
		"https://github.com/koalaman/shellcheck/wiki/SC1037")
	assertMapKeyEqual(t, comment, "path", "bad.sh")
	assertMapKeyEqual(t, comment, "startLine", float64(4))
	assertMapKeyEqual(t, comment, "endLine", float64(4))
	assertMapKeyEqual(t, comment, "startChar", float64(6))
	assertMapKeyEqual(t, comment, "endChar", float64(6))
}

func assertMapKeyEqual(t *testing.T, m map[string]interface{}, k string, want interface{}) {
	t.Helper()
	got, _ := m[k]
	if !reflect.DeepEqual(got, want) {
		t.Errorf("key %q got %v want %v", k, got, want)
	}
}
