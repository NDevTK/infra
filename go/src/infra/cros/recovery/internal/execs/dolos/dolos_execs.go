// Copyright 2024 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dolos

import (
	"context"
	"strings"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// setDolosStateExec sets the dolos state of the from the actionArgs argument.
//
// @actionArgs: the list of the string that contains the dolos state information.
// It should only contain one string in the format of: "state:x"
// x must be one of the keys from Dolos_State_value
func setDolosStateExec(ctx context.Context, info *execs.ExecInfo) error {
	args := info.GetActionArgs(ctx)
	newState := strings.ToUpper(args.AsString(ctx, "state", ""))
	if newState == "" {
		return errors.Reason("set dolos state: state is not provided").Err()
	}
	if dolos := info.GetChromeos().GetDolos(); dolos == nil || dolos.GetHostname() == "" {
		return errors.Reason("set dolos state: dolos is not supported").Err()
	}
	log.Debugf(ctx, "Previous dolos state: %s", info.GetChromeos().GetDolos().GetState())
	if v, ok := tlw.Dolos_State_value[newState]; ok {
		info.GetChromeos().GetDolos().State = tlw.Dolos_State(v)
		log.Infof(ctx, "Set dolos state to be: %s", newState)
		return nil
	}
	return errors.Reason("set dolos state: state is %q not found", newState).Err()
}

func init() {
	execs.Register("set_dolos_state", setDolosStateExec)
}
