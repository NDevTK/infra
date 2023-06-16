// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ufs

import (
	"context"

	"infra/cros/dutstate"

	"go.chromium.org/luci/auth/client/authcli"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"google.golang.org/genproto/protobuf/field_mask"

	ufsAPI "infra/unifiedfleet/api/v1/rpc"
)

// Allowlist of DUT states that are safe to overwrite.
var dutStatesSafeForOverwrite = map[dutstate.State]bool{
	dutstate.NeedsRepair: true,
	dutstate.Ready:       true,
	dutstate.Reserved:    true,
}

// SafeUpdateUFSDUTState attempts to safely update the DUT state to the
// given value in UFS. States other than Ready and NeedsRepair are
// ignored.
func SafeUpdateUFSDUTState(ctx context.Context, authFlags *authcli.Flags, dutName, dutState, ufsService string) error {
	c, err := NewClient(ctx, ufsService, authFlags)
	if err != nil {
		return errors.Annotate(err, "save update ufs state").Err()
	}
	info := dutstate.Read(ctx, c, dutName)
	logging.Infof(ctx, "Receive DUT state from UFS: %s", info.State)
	if info.DeviceId == "" {
		return errors.Reason("save update ufs state: deviceId not found").Err()
	}
	if dutStatesSafeForOverwrite[info.State] {
		req := &ufsAPI.UpdateTestDataRequest{
			DeviceId:      info.DeviceId,
			Hostname:      dutName,
			ResourceState: dutstate.ConvertToUFSState(dutstate.State(dutState)),
			UpdateMask:    &field_mask.FieldMask{Paths: []string{"dut.state"}},
		}
		_, err = c.UpdateTestData(ctx, req)
		return errors.Annotate(err, "save update ufs state").Err()
	}
	logging.Warningf(ctx, "Not saving requested DUT state %s, since current DUT state is %s, which should never be overwritten!", dutState, info.State)
	return nil
}
