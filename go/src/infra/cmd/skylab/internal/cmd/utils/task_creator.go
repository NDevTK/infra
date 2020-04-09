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

// HostMatcher is an interface for search criteria that can be used
// to locate DUTs to lease.
type HostMatcher interface {
	Dimensions() []*swarming_api.SwarmingRpcsStringPair
	Tag() string
}

// Hostname is the class that produces swarming Dimensions to look for
// a DUT to lease by hostname.
type Hostname struct {
	Data string
}

// Model is the class that produces swarming DImensions to look for
// a DUT by model.
type Model struct {
	Data string
}

// Dimensions converts a hostname into the swarming dimensions required to
// match exactly that hostname.
func (h *Hostname) Dimensions() []*swarming_api.SwarmingRpcsStringPair {
	return []*swarming_api.SwarmingRpcsStringPair{
		{Key: "pool", Value: "ChromeOSSkylab"},
		{Key: "dut_name", Value: h.Data},
	}
}

// Dimensions converts a model into the swarming dimensions required to
// match the repair_failed DUTs of that model in DUT_POOL_QUOTA.
func (m *Model) Dimensions() []*swarming_api.SwarmingRpcsStringPair {
	return []*swarming_api.SwarmingRpcsStringPair{
		{Key: "pool", Value: "ChromeOSSkylab"},
		{Key: "label-model", Value: m.Data},
		// We definitely do not want to steal DUTs from DUT_POOL_CTS
		{Key: "label-pool", Value: "DUT_POOL_QUOTA"},
		// Until we can implement real per-model caps, only allow people
		// to steal repair_failed DUTs for leasing.
		{Key: "dut_state", Value: "repair_failed"},
	}
}

// Tag returns the tag that is conventionally applied to swarming tasks
// matching a hostname.
func (h *Hostname) Tag() string {
	return fmt.Sprintf("dut-name:%s", h.Data)
}

// Tag returns the tag that is conventionally applied to lease swarming tasks
// that target a specific model.
func (m *Model) Tag() string {
	return fmt.Sprintf("model-name:%s", m.Data)
}

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
func (tc *TaskCreator) LeaseTask(ctx context.Context, m HostMatcher, durationSec int, reason string) (taskID string, err error) {
	c := []string{"/bin/sh", "-c", `while true; do sleep 60; echo Zzz...; done`}
	slices := []*swarming_api.SwarmingRpcsTaskSlice{{
		ExpirationSecs: int64(10 * 60),
		Properties: &swarming_api.SwarmingRpcsTaskProperties{
			Command:              c,
			Dimensions:           m.Dimensions(),
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
			m.Tag(),
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
