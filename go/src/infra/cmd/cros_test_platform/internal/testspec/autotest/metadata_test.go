// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package autotest

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"go.chromium.org/chromiumos/infra/proto/go/chromite/api"
	"go.chromium.org/luci/common/errors"
)

func TestGetMapsSuitesToTests(t *testing.T) {
	// TODO: split into separate test cases.
	fl := newFakeLoader()
	fl.AddTests([]string{"test_in_no_suite", "test_in_one_suite", "test_in_two_suites"})
	fl.AddSuites([]string{"suite_with_no_test", "suite_with_one_test", "suite_with_two_test"})
	ft := newFakeParseTestControlFn(map[string]*testMetadata{
		"test_in_no_suite": testWithNameAndSuites("test_in_no_suite", []string{}),
		"test_in_one_suite": testWithNameAndSuites(
			"test_in_one_suite",
			[]string{"suite_with_one_test"},
		),
		"test_in_two_suites": testWithNameAndSuites(
			"test_in_two_suites",
			[]string{"suite_with_one_test", "suite_with_two_test"},
		),
	})
	fs := newFakeParseSuiteControlFn(map[string]*api.AutotestSuite{
		"suite_with_no_test":   &api.AutotestSuite{Name: "suite_with_no_test"},
		"suite_with_one_test":  &api.AutotestSuite{Name: "suite_with_one_test"},
		"suite_with_two_tests": &api.AutotestSuite{Name: "suite_with_two_tests"},
	})
	g := getter{
		controlFileLoader:   fl,
		parseTestControlFn:  ft,
		parseSuiteControlFn: fs,
	}
	resp, err := g.Get("root")
	if err != nil {
		t.Fatalf("getter.Get(): %s", err)
	}

	want := map[string][]string{
		"suite_with_no_test":   []string{},
		"suite_with_one_test":  []string{"test_in_one_suite"},
		"suite_with_two_tests": []string{"test_in_one_suite", "test_in_two_suites"},
	}
	got := extractSuiteTests(resp.GetAutotest().GetSuites())
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("Suite.Tests differ, -want +got, %s", diff)
	}
}

func TestGetReturnsPartialResults(t *testing.T) {

}

// newFakeParseTestControlFn returns a fake parseTestControlFn that returns
// canned parse results.
//
// canned must map *contents of the control file* to their parse results. The
// returned parseTestControlFn returns error for any control file not in canned.
func newFakeParseTestControlFn(canned map[string]*testMetadata) parseTestControlFn {
	return func(text string) (*testMetadata, error) {
		tm, ok := canned[text]
		if !ok {
			return nil, errors.Reason("uncanned control file: %s", text).Err()
		}
		return tm, nil
	}
}

func testWithNameAndSuites(name string, suites []string) *testMetadata {
	return &testMetadata{
		AutotestTest: api.AutotestTest{
			Name: name,
		},
		Suites: suites,
	}
}

// newFakeParseSuiteControlFn returns a fake parseSuiteControlFn that returns
// canned parse results.
//
// canned must map *contents of the control file* to their parse results. The
// returned parseSuiteControlFn returns error for any control file not in
// canned.
func newFakeParseSuiteControlFn(canned map[string]*api.AutotestSuite) parseSuiteControlFn {
	return func(text string) (*api.AutotestSuite, error) {
		as, ok := canned[text]
		if !ok {
			return nil, errors.Reason("uncanned control file: %s", text).Err()
		}
		return as, nil
	}
}

func newFakeLoader() *fakeLoader {
	return &fakeLoader{
		tests:  make(map[string]io.Reader),
		suites: make(map[string]io.Reader),
	}
}

type fakeLoader struct {
	tests      map[string]io.Reader
	suites     map[string]io.Reader
	pathSuffix int
}

// AddTests adds the given texts as a test new control files at  arbitrary
// paths.
func (d *fakeLoader) AddTests(texts []string) {
	for _, t := range texts {
		d.tests[fmt.Sprintf("test%d", d.pathSuffix)] = strings.NewReader(t)
		d.pathSuffix++
	}
}

// RegisterSuite adds the given texts as a new suite control files at arbitrary
// paths.
func (d *fakeLoader) AddSuites(texts []string) {
	for _, t := range texts {
		d.suites[fmt.Sprintf("test%d", d.pathSuffix)] = strings.NewReader(t)
		d.pathSuffix++
	}
}

func (d *fakeLoader) Discover(string) error {
	return nil
}

func (d *fakeLoader) Tests() map[string]io.Reader {
	return d.tests
}

func (d *fakeLoader) Suites() map[string]io.Reader {
	return d.suites
}

func extractTestNames(tests []*api.AutotestTest) []string {
	m := make([]string, 0, len(tests))
	for _, t := range tests {
		m = append(m, t.Name)
	}
	return m
}

func extractSuiteNames(suites []*api.AutotestSuite) []string {
	m := make([]string, 0, len(suites))
	for _, s := range suites {
		m = append(m, s.Name)
	}
	return m
}

func extractSuiteTests(suites []*api.AutotestSuite) map[string][]string {
	m := make(map[string][]string)
	for _, s := range suites {
		ts := []string{}
		for _, t := range s.GetTests() {
			ts = append(ts, t.Name)
		}
		m[s.Name] = ts
	}
	return m
}
