// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package metadata provides functions to lint Chrome OS integration test
// metadata.
package metadata_test

import (
	"flag"
	"fmt"
	"infra/cros/cmd/tclint/internal/metadata"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/kylelemons/godebug/pretty"
	metadataPB "go.chromium.org/chromiumos/config/go/api/test/metadata/v1"
)

// Intentionally uses verbose flag name to avoid collision with predefined flags
// in the testing package.
var update = flag.Bool("update-lint-golden-files", false, "Update the golden files for lint diff tests")

// Tests returned diagnostic messages by comparing against golden expectation
// files.
//
// Returned diagnostics are the public API for tclint tool. This test prevents
// unintended regressions in the messages. To avoid spurious failures due to
// changes in logic unrelated to the message creation, each test case must
// generate exactly one error message.
func TestErrorMessages(t *testing.T) {
	for _, tc := range discoverTestCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			var spec metadataPB.Specification
			loadFromJSON(t, tc.inputFile, &spec)
			r := metadata.Lint(&spec)
			got := r.Display()
			want := loadGoldenFile(t, tc)
			if diff := pretty.Compare(want, got); diff != "" {
				t.Errorf("lint errors expectations mismatch, -want +got: \n%s", diff)
				if *update {
					writeGoldenFile(t, tc.goldenFile, got)
				}
			}
			// Also fail if the number of errors in not 1.
			// This ensures that diff tests are not brittle to reordering.
			if len(r.Errors) != 1 {
				t.Errorf("diff tests must check exactly 1 error, found %d", len(r.Errors))
			}
		})
	}
}

func loadFromJSON(t *testing.T, path string, m proto.Message) {
	t.Helper()
	r, err := os.Open(path)
	if err != nil {
		t.Fatalf("load proto from %s: %s", path, err.Error())
	}
	if err := jsonpb.Unmarshal(r, m); err != nil {
		t.Fatalf("load proto from %s: %s", path, err.Error())
	}
}

func loadGoldenFile(t *testing.T, tc testCase) string {
	t.Helper()
	var want []byte
	if tc.goldenFileFound {
		var err error
		if want, err = ioutil.ReadFile(tc.goldenFile); err != nil {
			t.Fatalf("load golden file %s: %s", tc.goldenFile, err.Error())
		}
	}
	return string(want)
}

func writeGoldenFile(t *testing.T, f string, data string) {
	t.Helper()
	if err := ioutil.WriteFile(f, []byte(data), 0666); err != nil {
		t.Fatalf("write golden file %s: %s", f, err.Error())
	}
	t.Logf("Updated golden file %s", f)
}

type testCase struct {
	name            string
	inputFile       string
	goldenFile      string
	goldenFileFound bool
}

const (
	testDataDir   = "testdata"
	inputFileExt  = ".input"
	goldenFileExt = ".golden"
)

func discoverTestCases(t *testing.T) []testCase {
	t.Helper()
	inputFiles := map[string]string{}
	goldenFiles := map[string]string{}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("load test data: %s", err.Error())
	}
	dataDir := filepath.Join(wd, testDataDir)
	fs, err := ioutil.ReadDir(dataDir)
	if err != nil {
		t.Fatalf("load test data from %s: %s", dataDir, err.Error())
	}
	for _, f := range fs {
		n := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
		p := filepath.Join(testDataDir, f.Name())
		switch ext := filepath.Ext(f.Name()); ext {
		case inputFileExt:
			inputFiles[n] = p
		case goldenFileExt:
			goldenFiles[n] = p
		default:
			t.Fatalf("unhandled extension %s in testdata: %s", ext, p)
		}
	}

	if len(inputFiles) == 0 {
		t.Fatalf("no input files found in %s", dataDir)
	}

	td := []testCase{}
	for name, inputFile := range inputFiles {
		goldenFile, found := goldenFiles[name]
		if !found {
			goldenFile = filepath.Join(dataDir, fmt.Sprintf("%s%s", name, goldenFileExt))
			t.Errorf("no golden file for input file %s", inputFile)
		}
		td = append(td, testCase{
			name:            name,
			inputFile:       inputFile,
			goldenFile:      goldenFile,
			goldenFileFound: found,
		})
	}
	return td
}
