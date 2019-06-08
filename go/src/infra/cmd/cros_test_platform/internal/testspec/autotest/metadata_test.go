// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package autotest

import (
	"io"
	"strings"
	"testing"
)

func TestGetPopulatesSuiteTests(t *testing.T) {

}

func TestGetReturnsPartialResults(t *testing.T) {

}

func newFakeControlFilesLoader(tests, suites map[string]string) *fakeLoader {
	return &fakeLoader{
		tests:  stringReadersForValues(tests),
		suites: stringReadersForValues(suites),
	}
}

func stringReadersForValues(m map[string]string) map[string]io.Reader {
	var ret map[string]io.Reader
	for k, v := range m {
		ret[k] = strings.NewReader(v)
	}
	return ret
}

type fakeLoader struct {
	tests  map[string]io.Reader
	suites map[string]io.Reader
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
