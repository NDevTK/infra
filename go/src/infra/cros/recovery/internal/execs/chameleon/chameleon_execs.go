// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package chameleon

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/tlw"
)

// setStateBrokenExec sets state as BROKEN.
func setStateBrokenExec(ctx context.Context, info *execs.ExecInfo) error {
	info.GetChromeos().GetChameleon().State = tlw.Chameleon_BROKEN
	return nil
}

// setStateWorkingExec sets state as WORKING.
func setStateWorkingExec(ctx context.Context, info *execs.ExecInfo) error {
	info.GetChromeos().GetChameleon().State = tlw.Chameleon_WORKING
	return nil
}

// setStateNotApplicableExec sets state as NOT_APPLICABLE.
func setStateNotApplicableExec(ctx context.Context, info *execs.ExecInfo) error {
	info.GetChromeos().GetChameleon().State = tlw.Chameleon_NOT_APPLICABLE
	return nil
}

// chameleonNotPresentExec check if chameleon is absent.
// return error if chameleon exists.
func chameleonNotPresentExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetChromeos().GetChameleon().GetName() != "" {
		return errors.Reason("chameleon not present: chameleon hostname exist").Err()
	}
	return nil
}

func init() {
	execs.Register("chameleon_state_broken", setStateBrokenExec)
	execs.Register("chameleon_state_working", setStateWorkingExec)
	execs.Register("chameleon_state_not_applicable", setStateNotApplicableExec)
	execs.Register("chameleon_not_present", chameleonNotPresentExec)
}
