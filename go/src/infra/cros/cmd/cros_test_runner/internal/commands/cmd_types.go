// Copyright 2023 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// All supported command types.
const (
	// Server related commands
	CommandsServerCmdType interfaces.CommandType = "CommandsServer"

	// Build/env related commands
	BuildInputValidationCmdType interfaces.CommandType = "BuildInputValidation"
	ParseEnvInfoCmdType         interfaces.CommandType = "ParseEnvInfoCmd"
	ProcessResultsCmdType       interfaces.CommandType = "ProcessResultsCmd"
	ParseArgsCmdType            interfaces.CommandType = "ParseArgs"

	// Container related commands
	UpdateContainerImagesLocallyCmdType interfaces.CommandType = "UpdateContainerImagesLocally"
	FetchContainerMetadataCmdType       interfaces.CommandType = "FetchContainerMetadata"

	// Inventory service related commands
	InvServiceStartCmdType  interfaces.CommandType = "InvServiceStart"
	InvServiceStopCmdType   interfaces.CommandType = "InvServiceStop"
	LoadDutTopologyCmdType  interfaces.CommandType = "LoadDutTopology"
	BuildDutTopologyCmdType interfaces.CommandType = "BuildDutTopology"

	// Ctr service related commands
	CtrServiceStartAsyncCmdType interfaces.CommandType = "CtrServiceStartAsync"
	CtrServiceStopCmdType       interfaces.CommandType = "CtrServiceStop"
	GcloudAuthCmdType           interfaces.CommandType = "GcloudAuth"

	// SSH service related commands
	SshStartTunnelCmdType        interfaces.CommandType = "SshTunnelStart"
	SshStartReverseTunnelCmdType interfaces.CommandType = "SshReverseTunnelStart"
	SshStopTunnelsCmdType        interfaces.CommandType = "SshTunnelsStop"

	// Cache server related commands
	CacheServerStartCmdType interfaces.CommandType = "CacheServerStart"

	// Dut service related commands
	DutServiceStartCmdType interfaces.CommandType = "DutServiceStart"

	// Dut VM test related commands
	DutVmGetImageCmdType interfaces.CommandType = "DutVmGetImage"
	DutVmLeaseCmdType    interfaces.CommandType = "DutVmLease"
	DutVmReleaseCmdType  interfaces.CommandType = "DutVmRelease"

	// Provision service related commands
	ProvisionServiceStartCmdType interfaces.CommandType = "ProvisionServiceStart"
	ProvisonInstallCmdType       interfaces.CommandType = "ProvisonInstall"

	// Test Finder service related commands
	TestFinderServiceStartCmdType interfaces.CommandType = "TestFinderServiceStart"
	TestFinderExecutionCmdType    interfaces.CommandType = "TestFinderExecution"

	// Test service related commands
	TestServiceStartCmdType interfaces.CommandType = "TestServiceStart"
	TestsExecutionCmdType   interfaces.CommandType = "TestsExecution"

	// Publish service related commands
	GcsPublishStartCmdType  interfaces.CommandType = "GcsPublishStart"
	GcsPublishUploadCmdType interfaces.CommandType = "GcsPublishUpload"

	TkoPublishStartCmdType  interfaces.CommandType = "TkoPublishStart"
	TkoPublishUploadCmdType interfaces.CommandType = "TkoPublishUpload"
	TkoDirectUploadCmdType  interfaces.CommandType = "TkoDirectUpload"

	RdbPublishStartCmdType  interfaces.CommandType = "RdbPublishStart"
	RdbPublishUploadCmdType interfaces.CommandType = "RdbPublishUpload"

	// Ufs related commands
	UpdateDutStateCmdType interfaces.CommandType = "UpdateDutState"

	// For testing purposes only
	UnSupportedCmdType interfaces.CommandType = "UnSupportedCmd"
)
