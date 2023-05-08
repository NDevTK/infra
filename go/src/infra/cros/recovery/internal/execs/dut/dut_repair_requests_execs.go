// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"

	"go.chromium.org/luci/common/errors"

	"infra/cros/recovery/internal/execs"
	"infra/cros/recovery/internal/log"
	"infra/cros/recovery/tlw"
)

// hasNoRepairRequestsExec checks that DUT has no repair_requests.
func hasNoRepairRequestsExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetDut() == nil {
		return errors.Reason("has no repair-requirests: dut is not provided").Err()
	}
	if len(info.GetDut().RepairRequests) == 0 {
		log.Debugf(ctx, "Total 0 repair-requiests.")
		return nil
	}
	for _, rr := range info.GetDut().RepairRequests {
		if rr != tlw.RepairRequestUnknown {
			return errors.Reason("has no repair-requirests: found %q repair-request", rr).Err()
		}
	}
	return nil
}

func init() {
	execs.Register("dut_has_no_repair_requests", hasNoRepairRequestsExec)
}
