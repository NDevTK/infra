// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/common_lib/interfaces"
)

// All supported command types.
const (
	FilterStartCmdType     interfaces.CommandType = "FilterStart"
	FilterStopCmdType      interfaces.CommandType = "FilterStop"
	FilterExecutionCmdType interfaces.CommandType = "FilterExecution"
	TranslateRequestType   interfaces.CommandType = "TranslateRequest"
)
