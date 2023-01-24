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

	// For testing purposes only
	UnSupportedCmdType interfaces.CommandType = "UnSupportedCmd"
)
