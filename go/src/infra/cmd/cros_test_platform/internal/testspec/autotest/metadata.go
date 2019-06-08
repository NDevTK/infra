// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package autotest

import (
	"io"

	"go.chromium.org/chromiumos/infra/proto/go/chromite/api"
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
	return nil, nil
}
