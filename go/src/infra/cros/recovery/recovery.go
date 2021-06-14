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
//   - Send DUTs info to inventory.
func Run(ctx context.Context, in *Input) error {
	if err := in.verify(); err != nil {
		return errors.Annotate(err, "run recovery: verify input").Err()
	}
	return errors.Reason("not implemented").Err()
}

// Input provides input arguments for recovery process.
type Input struct {
	Access tlw.Access
	// UnitName represents some device setup against which running some tests or task in the system.
	// The unit can be represented as a single DUT or group of the DUTs registered in inventory as single unit.
	UnitName string
}

func (in *Input) verify() error {
	if in == nil {
		return errors.Reason("input is empty").Err()
	} else if in.UnitName == "" {
		return errors.Reason("unit name is not provided").Err()
	} else if in.Access == nil {
		return errors.Reason("access point is not provided").Err()
	}
	return nil
}
