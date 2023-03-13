// Copyright 2022 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"context"
	"fmt"
	"io"

	"go.chromium.org/luci/common/errors"

	"infra/cmd/shivas/site"
	"infra/libs/skylab/buildbucket"
	"infra/libs/skylab/swarming"
	ufsUtil "infra/unifiedfleet/app/util"
)

// ScheduleDeployTask schedules a deploy task by Buildbucket for PARIS.
func ScheduleDeployTask(ctx context.Context, bc buildbucket.Client, e site.Environment, unit, sessionTag string, useLatestVersion bool) error {
	if unit == "" {
		return errors.Reason("schedule deploy task: unit name is empty").Err()
	}
	v := buildbucket.CIPDProd
	if useLatestVersion {
		v = buildbucket.CIPDLatest
	}
	adminServicePath := e.AdminService
	contextNamespace := ReadContextNamespace(ctx, ufsUtil.OSNamespace)
	if contextNamespace == ufsUtil.OSPartnerNamespace {
		// Partner do not have options with stable version.
		adminServicePath = ""
	}
	p := &buildbucket.Params{
		BuilderName:    "deploy",
		UnitName:       unit,
		TaskName:       string(buildbucket.Deploy),
		EnableRecovery: true,
		AdminService:   adminServicePath,
		// NOTE: We use the UFS service, not the Inventory service here.
		InventoryService:   e.UnifiedFleetService,
		InventoryNamespace: contextNamespace,
		UpdateInventory:    true,
		ExtraTags: []string{
			sessionTag,
			"task:deploy",
			"client:shivas",
			fmt.Sprintf("inventory_namespace:%s", contextNamespace),
			fmt.Sprintf("version:%s", v),
		},
	}
	url, _, err := buildbucket.ScheduleTask(ctx, bc, v, p)
	if err != nil {
		return errors.Annotate(err, "schedule deploy task").Err()
	}
	fmt.Printf("Triggered Deploy task %s. Follow the deploy job at %s\n", p.UnitName, url)
	return nil
}

// PrintTasksBatchLink prints batch link for scheduled tasks.
func PrintTasksBatchLink(wr io.Writer, swarmingService, commonTag string) {
	fmt.Fprintf(wr, "### Batch tasks URL ###\n")
	fmt.Fprintf(wr, "Created tasks: %s\n", TasksBatchLink(swarmingService, commonTag))
}

// TasksBatchLink created batch link to swarming for scheduled tasks.
func TasksBatchLink(swarmingService, commonTag string) string {
	return swarming.TaskListURLForTags(swarmingService, []string{commonTag})
}
