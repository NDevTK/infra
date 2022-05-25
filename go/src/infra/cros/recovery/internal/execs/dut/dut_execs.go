// Copyright 2021 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"regexp"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/dutstate"
	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	ufsProto "infra/unifiedfleet/api/v1/models"
)

// hasDutNameActionExec verifies that DUT provides name.
func hasDutNameActionExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.RunArgs.DUT != nil && info.RunArgs.DUT.Name != "" {
		log.Debugf(ctx, "DUT name: %q", info.RunArgs.DUT.Name)
		return nil
	}
	return errors.Reason("dut name is empty").Err()
}

// regexNameMatchExec checks if name match to provided regex.
func regexNameMatchExec(ctx context.Context, info *execs.ExecInfo) error {
	actionMap := info.GetActionArgs(ctx)
	d := info.RunArgs.DUT
	if d == nil {
		return errors.Reason("regex name match: DUT not found").Err()
	}
	regex := actionMap.AsString(ctx, "regex", "")
	if regex == "" {
		return errors.Reason("regex name match: regex is empty").Err()
	}
	m, err := regexp.MatchString(regex, d.Name)
	if err != nil {
		return errors.Annotate(err, "regex name match").Err()
	}
	if !m {
		return errors.Reason("regex name match: not match").Err()
	}
	return nil
}

// setDutStateExec sets the state of the DUT to the value passed in
// the action arguments.
func setDutStateExec(ctx context.Context, info *execs.ExecInfo) error {
	args := info.GetActionArgs(ctx)
	newState := strings.ToLower(args.AsString(ctx, "state", ""))
	if newState == "" {
		return errors.Reason("set dut state: state is not provided").Err()
	}
	state := dutstate.State(newState)
	if dutstate.ConvertToUFSState(state) == ufsProto.State_STATE_UNSPECIFIED {
		return errors.Reason("set dut state: unsupported state %q", newState).Err()
	}
	log.Debugf(ctx, "Old DUT state: %s", info.RunArgs.DUT.State)
	info.RunArgs.DUT.State = state
	log.Infof(ctx, "New DUT state: %s", newState)
	return nil
}

func init() {
	execs.Register("dut_has_name", hasDutNameActionExec)
	execs.Register("dut_regex_name_match", regexNameMatchExec)
	execs.Register("dut_set_state", setDutStateExec)
}
