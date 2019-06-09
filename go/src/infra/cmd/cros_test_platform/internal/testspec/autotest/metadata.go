// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package autotest

import (
	"io"
	"io/ioutil"

	"go.chromium.org/chromiumos/infra/proto/go/chromite/api"
	"go.chromium.org/luci/common/errors"
)

// Get computes metadata for all test and suite control files
// found within the directory tree rooted at root.
//
// Get always returns a valid api.TestMetadataResponse. In case of
// errors, the returned metadata corredsponds to the successfully parsed
// control files.
func Get(root string) (*api.TestMetadataResponse, error) {
	g := getter{
		controlFileLoader:   &controlFilesLoaderImpl{},
		parseTestControlFn:  parseTestControl,
		parseSuiteControlFn: parseSuiteControl,
	}
	return g.Get(root)
}

type controlFileLoader interface {
	Discover(string) error
	Tests() map[string]io.Reader
	Suites() map[string]io.Reader
}

type testMetadata struct {
	api.AutotestTest
	Suites []string
}
type parseTestControlFn func(string) (*testMetadata, error)
type parseSuiteControlFn func(string) (*api.AutotestSuite, error)

type getter struct {
	controlFileLoader   controlFileLoader
	parseTestControlFn  parseTestControlFn
	parseSuiteControlFn parseSuiteControlFn
}

func (g *getter) Get(root string) (*api.TestMetadataResponse, error) {
	if err := g.controlFileLoader.Discover(root); err != nil {
		return nil, errors.Annotate(err, "get autotest metadata").Err()
	}
	tests, err := g.parseTests(g.controlFileLoader.Tests())
	if err != nil {
		return nil, errors.Annotate(err, "get autotest metadata").Err()
	}
	suites, err := g.parseSuites(g.controlFileLoader.Suites())
	if err != nil {
		return nil, errors.Annotate(err, "get autotest metadata").Err()
	}
	collectTestsInSuites(tests, suites)
	return &api.TestMetadataResponse{
		Autotest: &api.AutotestTestMetadata{
			Suites: suites,
			Tests:  extractAutotestTests(tests),
		},
	}, nil
}

func (g *getter) parseTests(controls map[string]io.Reader) ([]*testMetadata, error) {
	tests := make([]*testMetadata, 0, len(controls))
	for _, t := range controls {
		bt, err := ioutil.ReadAll(t)
		if err != nil {
			// TODO: FIXME (and test me)
			return nil, err
		}
		tm, err := g.parseTestControlFn(string(bt))
		if err != nil {
			// TODO: FIXME (and test me)
			return nil, err
		}
		tests = append(tests, tm)
	}
	return tests, nil
}

func (g *getter) parseSuites(controls map[string]io.Reader) ([]*api.AutotestSuite, error) {
	suites := make([]*api.AutotestSuite, 0, len(controls))
	for _, t := range controls {
		bt, err := ioutil.ReadAll(t)
		if err != nil {
			// TODO: FIXME (and test me)
			return nil, err
		}
		tm, err := g.parseSuiteControlFn(string(bt))
		if err != nil {
			// TODO: FIXME (and test me)
			return nil, err
		}
		suites = append(suites, tm)
	}
	return suites, nil
}

func collectTestsInSuites(tests []*testMetadata, suites []*api.AutotestSuite) {
	sm := make(map[string]*api.AutotestSuite)
	for _, s := range suites {
		sm[s.GetName()] = s
	}
	for _, t := range tests {
		for _, s := range t.Suites {
			appendTestToSuite(t, sm[s])
		}
	}
}

func appendTestToSuite(test *testMetadata, suite *api.AutotestSuite) {
	suite.Tests = append(suite.Tests, &api.AutotestSuite_TestReference{Name: test.GetName()})

}

func extractAutotestTests(tests []*testMetadata) []*api.AutotestTest {
	at := make([]*api.AutotestTest, 0, len(tests))
	for _, t := range tests {
		at = append(at, &t.AutotestTest)
	}
	return at
}
