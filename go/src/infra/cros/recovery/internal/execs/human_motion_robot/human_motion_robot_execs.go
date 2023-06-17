// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package human_motion_robot

import (
	"context"
	"strings"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"

	"go.chromium.org/luci/common/errors"
)

// setHMRStateExec sets the hmr state of the from the actionArgs argument.
//
// @actionArgs: the list of the string that contains the hmr state information.
// It should only contain one string in the format of: "state:x"
// x must be one of the keys from HumanMotionRobot_State_value
func setHMRStateExec(ctx context.Context, info *execs.ExecInfo) error {
	args := info.GetActionArgs(ctx)
	newState := strings.ToUpper(args.AsString(ctx, "state", ""))
	if newState == "" {
		return errors.Reason("set hmr state: state is not provided").Err()
	}
	if info.GetChromeos().GetHumanMotionRobot() == nil {
		return errors.Reason("set hmr state: hmr is not supported").Err()
	}
	log.Debugf(ctx, "Previous hmr state: %s", info.GetChromeos().GetHumanMotionRobot().GetState())
	if v, ok := tlw.HumanMotionRobot_State_value[newState]; ok {
		info.GetChromeos().GetHumanMotionRobot().State = tlw.HumanMotionRobot_State(v)
		log.Infof(ctx, "Set hmr state to be: %s", newState)
		return nil
	}
	return errors.Reason("set hmr state: state is %q not found", newState).Err()
}

func init() {
	execs.Register("set_hmr_state", setHMRStateExec)
}
