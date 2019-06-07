// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package autotest

import (
	"errors"
	"io"

	"go.chromium.org/chromiumos/infra/proto/go/chromite/api"
)

// GetTestMetadata computes metadata for all test and suite control files
// found within the directory tree rooted at root.
//
// GetTestMetadata always returns a valid api.TestMetadataResponse. In case of
// errors, the returned metadata corredsponds to the successfully parsed
// control files.
func GetTestMetadata(root string) (*api.TestMetadataResponse, error) {
	var d controlFileLoader = &controlFilesLoaderImpl{}
	_ = d
	return nil, errors.New("not implemented")
}

type controlFileLoader interface {
	Discover(string) error
	Tests() map[string]io.Reader
	Suites() map[string]io.Reader
}

type testMetadata struct {
	api.AutotestTest
	suites []string
}

type parseTestControlFn func(io.Reader) (*testMetadata, error)
type parseSuiteControlFn func(io.Reader) (*api.AutotestSuite, error)
