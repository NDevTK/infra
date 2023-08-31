// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dns

import (
	"testing"

	"infra/cros/satlab/common/utils/executor"

	"github.com/google/go-cmp/cmp"
)

func TestReadHostsToIPMapShouldSuccess(t *testing.T) {
	t.Parallel()

	// Create a mock data
	executor := executor.FakeCommander{}
	executor.CmdOutput = `
192.168.231.137	satlab-0wgtfqin1846803b-one
192.168.231.137	satlab-0wgtfqin1846803b-host5
192.168.231.222	satlab-0wgtfqin1846803b-host11
192.168.231.222	satlab-0wgtfqin1846803b-host12
  `

	// Act
	res, err := ReadHostsToIPMap(&executor)

	if err != nil {
		t.Errorf("got an error: %v\n", err)
	}

	// Asset - because there are two duplicated IPs
	expectedResult := map[string]string{}
	expectedResult["192.168.231.137"] = "satlab-0wgtfqin1846803b-host5"
	expectedResult["192.168.231.222"] = "satlab-0wgtfqin1846803b-host12"

	if diff := cmp.Diff(res, expectedResult); diff != "" {
		t.Errorf("Expected: %v, got: %v", expectedResult, res)
	}
}

func TestReadHostsToHostMapShouldSuccess(t *testing.T) {
	t.Parallel()

	// Create a mock data
	executor := executor.FakeCommander{}
	executor.CmdOutput = `
192.168.231.137	satlab-0wgtfqin1846803b-one
192.168.231.137	satlab-0wgtfqin1846803b-host5
192.168.231.222	satlab-0wgtfqin1846803b-host11
192.168.231.222	satlab-0wgtfqin1846803b-host12
  `

	// Act
	res, err := ReadHostsToHostMap(&executor)

	if err != nil {
		t.Errorf("got an error: %v\n", err)
	}

	// Asset
	expectedResult := map[string]string{}
	expectedResult["satlab-0wgtfqin1846803b-one"] = "192.168.231.137"
	expectedResult["satlab-0wgtfqin1846803b-host5"] = "192.168.231.137"
	expectedResult["satlab-0wgtfqin1846803b-host11"] = "192.168.231.222"
	expectedResult["satlab-0wgtfqin1846803b-host12"] = "192.168.231.222"

	if diff := cmp.Diff(res, expectedResult); diff != "" {
		t.Errorf("Expected: %v, got: %v", expectedResult, res)
	}
}
