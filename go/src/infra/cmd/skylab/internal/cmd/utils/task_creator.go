// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package utils

import (
	"context"
	"fmt"
	"time"

	"infra/cmd/skylab/internal/site"
	"infra/libs/skylab/swarming"
	"infra/libs/skylab/worker"

	swarming_api "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/errors"
)

// TaskCreator creates Swarming tasks
type TaskCreator struct {
	Client      *swarming.Client
	Environment site.Environment
}

// RepairTask creates admin_repair task for particular DUT
func (tc *TaskCreator) RepairTask(ctx context.Context, host string, customTags []string, expirationSec int) (taskID string, err error) {
	c := worker.Command{
		TaskName: "admin_repair",
	}
	c.Config(tc.Environment.Wrapped())
	slices := []*swarming_api.SwarmingRpcsTaskSlice{{
		ExpirationSecs: int64(expirationSec),
		Properties: &swarming_api.SwarmingRpcsTaskProperties{
			Command: c.Args(),
			Dimensions: []*swarming_api.SwarmingRpcsStringPair{
				{Key: "pool", Value: "ChromeOSSkylab"},
				{Key: "dut_name", Value: host},
			},
			ExecutionTimeoutSecs: 5400,
		},
		WaitForCapacity: true,
	}}
	tags := []string{
		fmt.Sprintf("log_location:%s", c.LogDogAnnotationURL),
		fmt.Sprintf("luci_project:%s", tc.Environment.LUCIProject),
		"pool:ChromeOSSkylab",
		"skylab-tool:repair",
	}
	tags = append(tags, customTags...)
	r := &swarming_api.SwarmingRpcsNewTaskRequest{
		Name:           "admin_repair",
		Tags:           tags,
		TaskSlices:     slices,
		Priority:       25,
		ServiceAccount: tc.Environment.ServiceAccount,
	}
	ctx, cf := context.WithTimeout(ctx, 60*time.Second)
	defer cf()
	resp, err := tc.Client.CreateTask(ctx, r)
	if err != nil {
		return "", errors.Annotate(err, "failed to create task").Err()
	}
	return resp.TaskId, nil
}

// LeaseTask creates lease_task for particular DUT
func (tc *TaskCreator) LeaseTask(ctx context.Context, host string, model string, durationSec int, reason string) (taskID string, err error) {
	if err := validateLeaseTaskArgs(host, model); err != nil {
		return "", err
	}
	c := []string{"/bin/sh", "-c", `while true; do sleep 60; echo Zzz...; done`}
	slices := []*swarming_api.SwarmingRpcsTaskSlice{{
		ExpirationSecs: int64(10 * 60),
		Properties: &swarming_api.SwarmingRpcsTaskProperties{
			Command:              c,
			Dimensions:           createSwarmingDimensions(host, model),
			ExecutionTimeoutSecs: int64(durationSec),
		},
	}}
	r := &swarming_api.SwarmingRpcsNewTaskRequest{
		Name: "lease task",
		Tags: []string{
			"pool:ChromeOSSkylab",
			"skylab-tool:lease",
			// This quota account specifier is only relevant for DUTs that are
			// in the prod skylab DUT_POOL_QUOTA pool; it is irrelevant and
			// harmless otherwise.
			"qs_account:leases",
			fmt.Sprintf("dut-name:%s", host),
			fmt.Sprintf("lease-reason:%s", reason),
		},
		TaskSlices:     slices,
		Priority:       15,
		ServiceAccount: tc.Environment.ServiceAccount,
	}
	ctx, cf := context.WithTimeout(ctx, 60*time.Second)
	defer cf()
	resp, err := tc.Client.CreateTask(ctx, r)
	if err != nil {
		return "", errors.Annotate(err, "failed to create task").Err()
	}
	return resp.TaskId, nil
}

// validateLeaseTaskArgs performs local validation of LeaseTask before swarming gets a chance to see it.
func validateLeaseTaskArgs(host string, model string) error {
	hasHostname := host != ""
	hasModel := model != ""
	if hasHostname && hasModel {
		return fmt.Errorf("cannot have both hostname and model")
	}
	if !hasHostname && !hasModel {
		return fmt.Errorf("must have at least one of hostname or model")
	}
	return nil
}

// createSwarmingDimensions creates the swarming dimensions for a request targeting a model or hostname
// createSwarmingDimensions assumes that exactly one of host or model is non-empty.
func createSwarmingDimensions(host string, model string) []*swarming_api.SwarmingRpcsStringPair {
	if host != "" {
		return []*swarming_api.SwarmingRpcsStringPair{
			{Key: "pool", Value: "ChromeOSSkylab"},
			{Key: "dut_name", Value: host},
		}
	}
	return []*swarming_api.SwarmingRpcsStringPair{
		{Key: "pool", Value: "ChromeOSSkylab"},
		{Key: "label-model", Value: model},
		// We definitely do not want to steal DUTs from DUT_POOL_CTS
		{Key: "label-pool", Value: "DUT_POOL_QUOTA"},
		// Until we can implement real per-model caps, only allow people
		// to steal repair_failed DUTs for leasing.
		{Key: "dut_state", Value: "repair_failed"},
	}
}
