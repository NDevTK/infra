// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_commands

import (
	"infra/cros/cmd/common_lib/interfaces"
)

// All supported common command types.
const (
	// Ctr service related commands
	CtrServiceStartAsyncCmdType interfaces.CommandType = "CtrServiceStartAsync"
	CtrServiceStopCmdType       interfaces.CommandType = "CtrServiceStop"
	GcloudAuthCmdType           interfaces.CommandType = "GcloudAuth"

	// For testing purposes only
	UnSupportedCmdType interfaces.CommandType = "UnSupportedCmd"
)
