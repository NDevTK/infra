// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import (
	"cloud.google.com/go/bigquery"

	"go.chromium.org/chromiumos/config/go/test/api"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform"
	"go.chromium.org/luci/luciexe/build"

	"infra/cros/cmd/common_lib/interfaces"
	"infra/cros/cmd/common_lib/tools/crostoolrunner"
)

// PrePostFilterStateKeeper represents all the data pre and post filter execution flow requires.
type PrePostFilterStateKeeper struct {
	interfaces.StateKeeper

	// Requests
	CtpV1Requests map[string]*test_platform.Request
	CtpV2Request  *api.CTPv2Request

	// Results
	AllTestResults map[string][]*TestResults

	// Tools and their related dependencies
	Ctr                   *crostoolrunner.CrosToolRunner
	DockerKeyFileLocation string

	// BQ Client for writing CTP level task info to.
	BQClient   *bigquery.Client
	BuildState *build.State
}
