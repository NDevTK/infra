// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// All supported command types.
const (
	// Build/env related commands
	BuildInputValidationCmdType interfaces.CommandType = "BuildInputValidation"
	ParseEnvInfoCmdType         interfaces.CommandType = "ParseEnvInfoCmd"

	// Inventory service related commands
	InvServiceStartCmdType interfaces.CommandType = "InvServiceStart"
	InvServiceStopCmdType  interfaces.CommandType = "InvServiceStop"
	LoadDutTopologyCmdType interfaces.CommandType = "LoadDutTopology"

	// Ctr service related commands
	CtrServiceStartAsyncCmdType interfaces.CommandType = "CtrServiceStartAsync"
	CtrServiceStopCmdType       interfaces.CommandType = "CtrServiceStop"
	GcloudAuthCmdType           interfaces.CommandType = "GcloudAuth"

	// Dut service related commands
	DutServiceStartCmdType interfaces.CommandType = "DutServiceStart"

	// Provision service related commands
	ProvisionServiceStartCmdType interfaces.CommandType = "ProvisionServiceStart"
	ProvisonInstallCmdType       interfaces.CommandType = "ProvisonInstall"

	// For testing purposes only
	UnSupportedCmdType interfaces.CommandType = "UnSupportedCmd"
)
