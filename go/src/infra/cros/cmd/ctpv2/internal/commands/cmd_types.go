// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package commands

import (
	"infra/cros/cmd/common_lib/interfaces"
)

// All supported command types.
const (
	PrepareFilterContainersCmdType interfaces.CommandType = "PrepareFilterContainers"
	FilterExecutionCmdType         interfaces.CommandType = "FilterExecution"
	FilterStopCmdType              interfaces.CommandType = "FilterStop"
	TranslateRequestType           interfaces.CommandType = "TranslateRequest"
)
