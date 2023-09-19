// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
package platform

import (
	"infra/cros/satlab/common/utils/executor"
)

type Debian struct {
	execCommander executor.IExecCommander
}

func NewDebianPlatform() IPlatform {
	return &Debian{
		execCommander: &executor.ExecCommander{},
	}
}
