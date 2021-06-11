// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package recovery provides ability to run recovery tasks against on the target units.
package recovery

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/tlw"
)

// Run runs the recovery tasks against the provide unit.
// Process includes:
//   - Verification of input data.
//   - Collect DUTs info.
//   - Load execution plan for required task with verification.
// 	 - Send DUTs info to inventory.
func Run(ctx context.Context, in *Input) error {
	if err := verify(in); err != nil {
		return errors.Annotate(err, "run recovery").Err()
	}
	return errors.Reason("not implemented").Err()
}

// Input provides input arguments for recovery process.
type Input struct {
	Access tlw.Access
	// Unit represents some device setup against which running some tests or task in the system.
	// The unit can be represented as a single DUT or group of the DUTs registered in inventory as single unit.
	Unit string
}

func verify(in *Input) error {
	if in == nil {
		return errors.Reason("verify: input parameters is not provided").Err()
	}
	if in.Unit == "" {
		return errors.Reason("verify: unit is not provide in input parameters").Err()
	}
	if in.Access == nil {
		return errors.Reason("verify: tlw access point is not provided").Err()
	}
	return nil
}
