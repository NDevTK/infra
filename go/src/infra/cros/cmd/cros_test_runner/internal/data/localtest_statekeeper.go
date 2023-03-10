// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package data

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"

	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
)

// LocalTestStateKeeper represents all the data local test execution flow requires.
type LocalTestStateKeeper struct {
	interfaces.StateKeeper
	HwTestStateKeeper

	// CLI inputs
	Args *LocalArgs

	// Replace TestRequest inputs
	Tests       []string
	Tags        []string
	TagsExclude []string

	// IpEndpoints
	DutSshAddress         *labapi.IpEndpoint
	DutCacheServerAddress *labapi.IpEndpoint
	CacheServerAddress    *labapi.IpEndpoint

	// Replacement values for CftTestRequest
	ImagePath string
}

type LocalArgs struct {
	BuildBoard                      string
	BuildBucket                     string
	BuildNumber                     string
	HostName                        string
	Tests                           string
	Tags                            string
	TagsExclude                     string
	ContainerKeysRequestedForUpdate string
	Chroot                          string

	// Optional replacements for values that would have been updated from a skipped step
	DutAddress      string
	DutCacheAddress string
	CacheAddress    string

	// Flow control args. Should match LocalStepConfig proto
	SkipBuildDutTopology bool
	SkipDutServer        bool
	SkipProvision        bool
	SkipTestFinder       bool
	SkipTest             bool
	SkipCacheServer      bool
	SkipSshTunnel        bool
	SkipSshReverseTunnel bool
}
