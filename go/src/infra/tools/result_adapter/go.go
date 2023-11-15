// Copyright 2021 The LUCI Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	resultpb "go.chromium.org/luci/resultdb/proto/v1"
	sinkpb "go.chromium.org/luci/resultdb/sink/proto/v1"
)

// ensureArgsValid checks the payload to ensure it looks like a `go test`
// invocation and that it has the `-json` flag.
func (r *goRun) ensureArgsValid(args []string) ([]string, error) {
	// Scan the arguments.
	var (
		testArgIndex = -1
		jsonFlagSeen = false
	)
	for i, t := range args {
		switch t {
		case "test":
			testArgIndex = i
		case "-json":
			jsonFlagSeen = true
		}
	}
	if testArgIndex != -1 && !jsonFlagSeen {
		// If there's a 'test' argument and no '-json' flag, automatically add one.
		args = append(args[:testArgIndex+1], append([]string{"-json"}, args[testArgIndex+1:]...)...)
		return args, nil
	}
	if cmd := strings.Join(args, " "); !strings.Contains(cmd, "test") && !strings.Contains(cmd, "json") {
		// Since the command line doesn't even mention "test" nor "json", probably safe enough to fail here.
		return nil, errors.Reason("Expected command to be an invocation of `go test -json` or equivalent: %q", args).Err()
	}
	// Otherwise it might be something that will emit JSON events in https://go.dev/cmd/test2json format,
	// such as 'GOROOT/src/run.bash -json' or 'sh -c "something && go test -json ./..."'.
	//
	// Use the arguments as is and let it fail at a higher level if the caller made a mistake.
	return args, nil
}

func (r *goRun) generateTestResults(ctx context.Context, data []byte) ([]*sinkpb.TestResult, error) {
	ordered, byID := goTestJsonToTestRecords(ctx, data, r.CopyTestOutput)
	return testRecordsToTestProtos(ctx, ordered, byID), nil
}

// goTestJsonToTestRecords parses one line at a time from the given output,
// which is expected to be the one produced by `go test -json <package>`.
// It converts each line to TestEvent and ingests it into a TestRecord.
// copyTestOutput optionally specifies where to write a copy of test output.
// The resulting TestRecord(s) are returned to the caller as a slice in the
// same order as they were initially seen, and in a map where the test's id
// maps to its TestRecord.
func goTestJsonToTestRecords(ctx context.Context, data []byte, copyTestOutput io.StringWriter) ([]*TestRecord, map[string]*TestRecord) {
	var ordered []*TestRecord
	var byID = make(map[string]*TestRecord)
	// Ensure that the scanner below returns the last line in the output.
	if !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, []byte("\n")...)
	}
	lines := bufio.NewScanner(bytes.NewReader(data))
	// Iterate over output, parsing an event from each line and making the
	// appropriate record ingest it.
	for lines.Scan() {
		l := lines.Bytes()
		if len(l) == 0 {
			continue
		}
		tEvt, err := parseRow(l)
		if err != nil {
			logging.Warningf(ctx, "cannot parse row %q, %s", string(l), err)
			continue
		}
		currentRecord := byID[tEvt.id()]
		if currentRecord == nil {
			currentRecord = &TestRecord{TestID: tEvt.id()}
			ordered = append(ordered, currentRecord)
			byID[currentRecord.TestID] = currentRecord
		}
		currentRecord.ingest(tEvt)
		if copyTestOutput != nil {
			copyTestOutput.WriteString(tEvt.Output)
		}
	}
	return ordered, byID
}

// testRecordsToTestProtos converts the TestRecords returned by the above into
// a list of TestResult protos suitable for sending to result sink.
func testRecordsToTestProtos(ctx context.Context, ordered []*TestRecord, byID map[string]*TestRecord) []*sinkpb.TestResult {
	ret := make([]*sinkpb.TestResult, 0, 8)
	for _, record := range ordered {
		if record.IsPackage {
			continue
		}
		tr := sinkpb.TestResult{}
		switch record.Result {
		case "pass", "bench":
			tr.Status = resultpb.TestStatus_PASS
			tr.Expected = true
		case "fail":
			tr.Status = resultpb.TestStatus_FAIL
		case "skip":
			tr.Status = resultpb.TestStatus_SKIP
			tr.Expected = true
		case "":
			// It has been observed that test2json may fail to parse the status
			// of a test when multiple tests run in parallel in the same package
			// and produce certain output (such as goconvey output). In those
			// cases it's okay to mark the tests as passing if the whole
			// package passed.
			if byID[record.PackageName].Result == "pass" {
				logging.Warningf(ctx,
					"Status for test %s is missing from the list of test events. Setting to `pass` because package passed.", record.TestID)
				tr.Status = resultpb.TestStatus_PASS
				tr.Expected = true
			} else {
				// A test interrupted by SIGTERM, SIGABORT, SIGKILL will usually
				// have its status unset.
				tr.Status = resultpb.TestStatus_ABORT
			}
		}
		if record.Output.Len() > 0 {
			a := sinkpb.Artifact{}
			a.Body = &sinkpb.Artifact_Contents{
				Contents: []byte(record.Output.String()),
			}
			tr.Artifacts = map[string]*sinkpb.Artifact{"output": &a}
			tr.SummaryHtml = `<p><text-artifact artifact-id="output"></p>`
		}
		tr.TestId = record.TestID
		tr.Duration = durationpb.New(time.Duration(int64(record.Elapsed * float64(time.Second))))
		tr.StartTime = timestamppb.New(record.Started)
		ret = append(ret, &tr)
	}
	return ret
}

func parseRow(s []byte) (*TestEvent, error) {
	new := &TestEvent{}
	return new, json.Unmarshal(s, new)
}

// TestEvent represents each json object produced by `go test -json`.
// Details at https://go.dev/cmd/test2json.
type TestEvent struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

// id identifies tests by their package name, e.g. `go.chromium.org/luci/resultdb/sink`
// joined by a period to the name of the test: e.g. `TestNewServer`, so
// the test id would be e.g. `go.chromium.org/luci/resultdb/sink.TestNewServer`.
// Events that apply to the whole package are identified only by the package name.
func (te *TestEvent) id() string {
	if te.Test == "" {
		return te.Package
	}
	// Test names in Go may contain Unicode printable runes, but
	// ResultDB currently only allows ASCII printable runes. See crbug.com/1446084.
	// Work around that by temporarily escaping non-ASCII test names to ASCII.
	// TODO(crbug.com/1446084): Drop maybeEscape after the ResultDB fix rolls out.
	return fmt.Sprintf("%s.%s", te.Package, maybeEscape(te.Test))
}

// maybeEscape returns s unmodified if it consists entirely of ASCII runes,
// or else with each non-ASCII rune replaced by its hex Unicode code point
// inside round brackets. For example, "TestNameIsASCII" is returned as is,
// but "TestSeeሴLater" is escaped to "TestSee(U+1234)Later".
func maybeEscape(s string) string {
	for i, r := range s {
		if r < utf8.RuneSelf {
			continue
		}
		// Rare case: at least one non-ASCII rune, so need to escape.
		var b strings.Builder
		b.WriteString(s[:i]) // Fast-forward to first non-ASCII rune.
		for _, r := range s[i:] {
			if r < utf8.RuneSelf {
				b.WriteByte(byte(r))
			} else {
				b.WriteString(fmt.Sprintf("(%U)", r))
			}
		}
		return b.String()
	}
	// Common case.
	return s
}

// TestRecord represents the results of a single test or package.
// Several test events will apply to each TestRecord.
type TestRecord struct {
	TestID      string // TestID is the ID of the test this TestRecord represents.
	IsPackage   bool
	PackageName string // Import path of the Go package that the test is a part of.
	Result      string // Out of a subset of the values for TestEvent.Action as applicable.
	Started     time.Time
	Elapsed     float64 //seconds
	Output      strings.Builder
}

// ingest updates the fields of the test record according to the contents of
// the given test event.
// Output events from a specific test will be associated with the corresponding
// test record.
// Tests running in parallel that choose not to use t.Log/t.Error and instead
// write to stdout/stderr directly may result in output being associated with
// the wrong test. See https://go.dev/issue/23036#issuecomment-355669573.
func (tr *TestRecord) ingest(te *TestEvent) {
	if te.Test == "" {
		tr.IsPackage = true
	}
	if tr.PackageName == "" {
		tr.PackageName = te.Package
	}
	switch te.Action {
	// Action string values from https://go.dev/cmd/test2json.
	case "pause":
	case "cont":
	case "run":
		tr.Started = te.Time
	case "output":
		tr.Output.WriteString(te.Output)
	default:
		tr.Result = te.Action
		if te.Elapsed > 0 {
			tr.Elapsed = te.Elapsed
		}
	}
}