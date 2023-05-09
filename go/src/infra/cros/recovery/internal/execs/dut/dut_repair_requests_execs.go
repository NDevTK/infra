// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dut

import (
	"context"
	"strings"

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

// hasAnyRepairRequestsExec checks is any specified repair_request is present.
func hasAnyRepairRequestsExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetDut() == nil {
		return errors.Reason("has any repair-requirests: dut is not provided").Err()
	}
	args := info.GetActionArgs(ctx)
	requests := args.AsStringSlice(ctx, "requests", nil)
	if len(requests) == 0 {
		return nil
	}
	requestMap := make(map[tlw.RepairRequest]bool, len(requests))
	for _, r := range requests {
		if r == "" {
			continue
		}
		er := tlw.RepairRequest(strings.ToUpper(r))
		if er == tlw.RepairRequestUnknown {
			continue
		}
		requestMap[er] = true
	}
	for _, rr := range info.GetDut().RepairRequests {
		if _, ok := requestMap[rr]; ok {
			return errors.Reason("has any repair-requirests: found %v", rr).Err()
		}
	}
	return nil
}

// removeRepairRequestsExec removes provided repair_requests.
func removeRepairRequestsExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetDut() == nil {
		return errors.Reason("remove repair-requirests: dut is not provided").Err()
	}
	args := info.GetActionArgs(ctx)
	requests := args.AsStringSlice(ctx, "requests", nil)
	if len(requests) == 0 {
		// Nothing to remove.
		return nil
	}
	requestMap := make(map[tlw.RepairRequest]bool, len(requests))
	for _, r := range requests {
		if r == "" {
			continue
		}
		er := tlw.RepairRequest(strings.ToUpper(r))
		if er == tlw.RepairRequestUnknown {
			continue
		}
		requestMap[er] = true
	}
	var newRepairRequests []tlw.RepairRequest
	for _, rr := range info.GetDut().RepairRequests {
		if _, ok := requestMap[rr]; ok {
			log.Debugf(ctx, "Found and removed %q.", rr)
		} else {
			newRepairRequests = append(newRepairRequests, rr)
		}
	}
	info.GetDut().RepairRequests = newRepairRequests
	return nil
}

// resetRepairRequestsExec removes all present repair_requests.
func resetRepairRequestsExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetDut() == nil {
		return errors.Reason("reset repair-requirests: dut is not provided").Err()
	}
	info.GetDut().RepairRequests = nil
	return nil
}

// addRepairRequestsExec add provided repair_requests.
func addRepairRequestsExec(ctx context.Context, info *execs.ExecInfo) error {
	if info.GetDut() == nil {
		return errors.Reason("add repair-requirests: dut is not provided").Err()
	}
	args := info.GetActionArgs(ctx)
	requests := args.AsStringSlice(ctx, "requests", nil)
	if len(requests) == 0 {
		// Nothing to add.
		return nil
	}
	requestMap := make(map[tlw.RepairRequest]bool, len(info.GetDut().RepairRequests))
	for _, r := range info.GetDut().RepairRequests {
		requestMap[r] = true
	}
	for _, e := range requests {
		e = strings.ToUpper(strings.TrimSpace(e))
		if e == "" {
			// Skip empty values.
			continue
		}
		er := tlw.RepairRequest(e)
		switch er {
		case tlw.RepairRequestProvision, tlw.RepairRequestUpdateUSBKeyImage, tlw.RepairRequestReimageByUSBKey:
			if _, ok := requestMap[er]; ok {
				log.Debugf(ctx, "The repair-request %q already present", er)
			} else {
				info.GetDut().RepairRequests = append(info.GetDut().RepairRequests, er)
				requestMap[er] = true
			}
		default:
			return errors.Reason("add repair-requirests: unsupported request %q", er).Err()
		}
	}
	return nil
}

func init() {
	execs.Register("dut_has_no_repair_requests", hasNoRepairRequestsExec)
	execs.Register("dut_has_any_repair_requests", hasAnyRepairRequestsExec)
	execs.Register("dut_remove_repair_requests", removeRepairRequestsExec)
	execs.Register("dut_reset_repair_requests", resetRepairRequestsExec)
	execs.Register("dut_add_repair_requests", addRepairRequestsExec)
}
