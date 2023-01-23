package commands

import (
	"infra/cros/cmd/cros_test_runner/internal/interfaces"
)

// All supported command types.
const (
	BuildInputValidationCmdType interfaces.CommandType = "BuildInputValidation"
	ParseEnvInfoCmdType         interfaces.CommandType = "ParseEnvInfoCmd"

	// For testing purposes only
	UnSupportedCmdType interfaces.CommandType = "UnSupportedCmd"
)
