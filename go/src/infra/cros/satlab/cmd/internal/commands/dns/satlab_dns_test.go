// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dns

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// TestMakeNewContent tests updating lines in a DNS file
func TestMakeNewContent(t *testing.T) {
	t.Parallel()

	type input struct {
		content        string
		newRecords     map[string]string
		deletedRecords map[string]bool
	}
	type test struct {
		name   string
		input  input
		output string
	}

	tests := []test{{
		name: "test new add to end",
		input: input{
			content: strings.Join([]string{
				tabify("addr1 host1"),
				tabify("addr2 host2"),
			}, "\n"),
			newRecords: map[string]string{
				"host1": "addr1-UPDATE",
				"host4": "addr4-NEW",
				"host3": "addr3-NEW",
			},
		},
		output: strings.Join([]string{
			tabify("addr1-UPDATE host1"),
			tabify("addr2 host2"),
			tabify("addr3-NEW host3"),
			tabify("addr4-NEW host4"),
		}, "\n") + "\n"}, {
		name: "test update records end unchanged",
		input: input{
			content: strings.Join([]string{
				tabify("addr1 host1"),
				tabify("addr2 host2"),
			}, "\n"),
			newRecords: map[string]string{
				"host1": "addr1-UPDATE",
			},
		},
		output: strings.Join([]string{
			tabify("addr1-UPDATE host1"),
			tabify("addr2 host2"),
		}, "\n") + "\n"}, {
		name: "test deleted records",
		input: input{
			content: strings.Join([]string{
				tabify("addr1 host1"),
				tabify("addr2 host2"),
			}, "\n"),
			newRecords: map[string]string{},
			deletedRecords: map[string]bool{
				"host2": true,
			},
		},
		output: strings.Join([]string{
			tabify("addr1 host1"),
		}, "\n") + "\n"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := makeNewContent(tc.input.content, tc.input.newRecords, tc.input.deletedRecords)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if !containIdenticalLines(tc.output, actual) {
				t.Errorf("unexpected diff. got: %s,\n expected: %s,\ninput: %+v", actual, tc.output, tc.input)
			}
		})
	}

}

// Tabify replaces arbitrary whitespace with tabs.
func tabify(s string) string {
	return strings.Join(strings.Fields(s), "\t")
}

// containIdenticalLines returns whether or not two strings have identical LINES regardless of order of these lines
func containIdenticalLines(x string, y string) bool {
	xArr := strings.SplitAfter(x, "\n")
	yArr := strings.SplitAfter(y, "\n")

	diff := cmp.Diff(xArr, yArr, cmpopts.SortSlices(func(a, b string) bool { return a < b }))
	return diff == ""
}

// makeFakeHostsfilesReader provides ability to echo any string desired.
func makeFakeHostsfileReader(content string) hostsfileReaderFunc {
	return func() (string, error) {
		return content, nil
	}
}

// ensureRecordCall is bag of items that are params to ensureRecord.
type ensureRecordCall struct {
	content        string
	newRecords     map[string]string
	deletedRecords map[string]bool
}

// makeFakeRecordEnsurer provides a no-op implementation that tracks all function calls.
func makeFakeRecordEnsurer(calls *[]ensureRecordCall) recordEnsurer {
	return func(content string, newRecords map[string]string, deletedRecords map[string]bool) error {
		*calls = append(*calls, ensureRecordCall{content: content, newRecords: newRecords, deletedRecords: deletedRecords})
		return nil
	}
}

// TestDeleteRecord verifies we are passing along correct args to ensureRecords.
func TestDeleteRecord(t *testing.T) {
	t.Parallel()

	recordEnsurerCalls := make([]ensureRecordCall, 0)
	fakeRecordEnsurer := makeFakeRecordEnsurer(&recordEnsurerCalls)
	fakeHostsfileReader := makeFakeHostsfileReader("content")
	toDelete := "test"

	expectedReturn := "content"
	expectedRecordEnsurerCalls := []ensureRecordCall{{"content", map[string]string{}, map[string]bool{"test": true}}}

	result, err := DeleteRecord(fakeRecordEnsurer, fakeHostsfileReader, toDelete)
	if result != expectedReturn {
		t.Errorf("diff in return val, expected: %s, got: %s", "content", result)
	}
	if err != nil {
		t.Errorf("unexpected err %+v", err)
	}
	// cmp.Diff needs us to export a way to compare maps, so we stick with reflect here
	if !reflect.DeepEqual(expectedRecordEnsurerCalls, recordEnsurerCalls) {
		t.Errorf("unexpected diff in calls to ensure record. got: %+v, diff: %+v", recordEnsurerCalls[0], expectedRecordEnsurerCalls)
	}
}
