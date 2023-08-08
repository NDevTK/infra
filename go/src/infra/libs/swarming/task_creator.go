// Copyright 2020 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package swarming

import (
	"context"
	"fmt"
	"io"
	"time"

	"infra/cmdsupport/cmdlib"
	sw "infra/libs/skylab/swarming"

	"github.com/google/uuid"
	"go.chromium.org/luci/auth"
	"go.chromium.org/luci/auth/client/authcli"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/errors"
)

const (
	// LuciProject is used to tag the chromeos tasks.
	LuciProject = "chromeos"

	// DefaultAdminTaskPriority is the priority used as default for the admin tasks
	// if other was not specified.
	DefaultAdminTaskPriority = 25

	// DUTIDDimensionKey is the dimension key for dut ID.
	DUTIDDimensionKey = "dut_id"

	// DUTNameDimensionKey is the dimension key for dut hostname.
	DUTNameDimensionKey = "dut_name"

	// IDDimensionKey is the dimension key for ID.
	IDDimensionKey = "id"

	// PoolDimensionKey is the dimension key for pool.
	PoolDimensionKey = "pool"
)

// TaskCreator creates Swarming tasks
type TaskCreator struct {
	// Client is Swarming API Client
	client *sw.Client
	// SwarmingService is a path to Swarming API
	swarmingService string
	// Session is an ID that is used to mark tasks and for tracking all of the tasks created in a logical session.
	session string
	// Authenticator is used to get user info
	authenticator *auth.Authenticator
	// LogdogService is the logdog service for the task logs
	LogdogService string
	// logdogTaskCode keeps unique code for each creating task. Please call GenerateLogdogTaskCode() for each task.
	logdogTaskCode string
	// SwarmingServiceAccount is the service account to be used.
	SwarmingServiceAccount string
	// LUCIProject is the name of the project used to create the task.
	LUCIProject string
}

// TaskInfo contains information of the created task.
type TaskInfo struct {
	// ID of the created task in the Swarming.
	ID string
	// TaskURL provides the URL to the created task in Swarming.
	TaskURL string
}

// GetID gets an ID and returns "" by default.
func (t *TaskInfo) GetID() string {
	if t == nil {
		return ""
	}
	return t.ID
}

// NewTaskCreator creates and initialize the TaskCreator.
func NewTaskCreator(ctx context.Context, authFlags *authcli.Flags, swarmingService string) (*TaskCreator, error) {
	a, err := cmdlib.NewAuthenticator(ctx, authFlags)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create TaskCreator. Authenticator error").Err()
	}
	h, err := a.Client()
	if err != nil {
		return nil, errors.Annotate(err, "failed to create TaskCreator. Cannot create http client").Err()
	}

	service, err := sw.NewClient(h, swarmingService)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create TaskCreator. Cannot create API client").Err()
	}

	tc := &TaskCreator{
		client:          service,
		swarmingService: swarmingService,
		session:         uuid.New().String(),
		LUCIProject:     LuciProject,
		authenticator:   a,
	}
	return tc, nil
}

// LogdogURL returns the logdog URL for task logs, empty string if logdog service not set.
//
// The logdogURL has to be unique for each task and to guaranty it please call GenerateLogdogTaskCode() before create new task.
func (tc *TaskCreator) LogdogURL() string {
	if tc.LogdogService != "" {
		return fmt.Sprintf("logdog://%s/%s/%s/+/annotations", tc.LogdogService, tc.LUCIProject, tc.logdogTaskCode)
	}
	return ""
}

// GenerateLogdogTaskCode generate new unique code for each task used in logdog URL.
func (tc *TaskCreator) GenerateLogdogTaskCode() {
	tc.logdogTaskCode = uuid.New().String()
}

// MapToSwarmingDimensions converts from a go map to SwarmingRpcsStringPair
func MapToSwarmingDimensions(dims map[string]string) []*swarming.SwarmingRpcsStringPair {
	var dimensions []*swarming.SwarmingRpcsStringPair
	for dimKey, dimValue := range dims {
		dimensions = append(dimensions, &swarming.SwarmingRpcsStringPair{
			Key:   dimKey,
			Value: dimValue,
		})
	}
	return dimensions
}

// Schedule registers task in the swarming
func (tc *TaskCreator) schedule(ctx context.Context, req *swarming.SwarmingRpcsNewTaskRequest) (*TaskInfo, error) {
	ctx, cf := context.WithTimeout(ctx, 60*time.Second)
	defer cf()
	resp, err := tc.client.CreateTask(ctx, req)
	if err != nil {
		return nil, errors.Annotate(err, "failed to create task").Err()
	}
	return &TaskInfo{
		ID:      resp.TaskId,
		TaskURL: tc.taskURL(resp.TaskId),
	}, nil
}

// taskURL generates URL to the task in swarming.
func (tc *TaskCreator) taskURL(id string) string {
	return sw.TaskURL(tc.swarmingService, id)
}

// sessionTag return admin session tag for swarming.
func (tc *TaskCreator) sessionTag() string {
	return fmt.Sprintf("admin_session:%s", tc.session)
}

// SessionTasksURL gets URL to see all created tasks belong to the session.
func (tc *TaskCreator) SessionTasksURL() string {
	tags := []string{
		tc.sessionTag(),
	}
	return sw.TaskListURLForTags(tc.swarmingService, tags)
}

// changeDUTStateCommand creates command to initiate state change.
// After set state we wait 180 seconds to apply state changes to the dut.
func changeDUTStateCommand(task string) []string {
	return []string{
		"/bin/sh",
		"-c",
		fmt.Sprintf("/opt/infra-tools/skylab_swarming_worker -task-name %s; echo Zzz...; do sleep 180", task),
	}
}

func (tc *TaskCreator) combineTags(toolName, logDogURL string, customTags []string) []string {
	tags := []string{
		fmt.Sprintf("skylab-tool:%s", toolName),
		fmt.Sprintf("luci_project:%s", tc.LUCIProject),
		fmt.Sprintf("pool:%s", sw.SkylabPool),
		tc.sessionTag(),
	}
	if logDogURL != "" {
		// log_location is required to see the logs in the swarming
		tags = append(tags, fmt.Sprintf("log_location:%s", logDogURL))
	}
	return append(tags, customTags...)
}

// PrintResults prints results of the task creation.
func (tc *TaskCreator) PrintResults(wr io.Writer, successMap map[string]*TaskInfo, errorMap map[string]error, batchLink bool) {
	if len(errorMap) > 0 {
		fmt.Fprintln(wr, "\n### Failed to create ###")
		for host, err := range errorMap {
			fmt.Fprintf(wr, "%s: %s\n", host, err.Error())
		}
	}
	if len(successMap) > 0 {
		// We say "successfully created tasks" instead of "successfully created swarming tasks" because
		// some tasks are buildbucket tasks.
		fmt.Fprintf(wr, "\n### Successfully created tasks - %d ###\n", len(successMap))
		for host, task := range successMap {
			fmt.Fprintf(wr, "%s: %s\n", host, task.TaskURL)
		}
		if batchLink && len(successMap) > 1 {
			fmt.Fprintln(wr, "\n### Batch tasks URL ###")
			fmt.Fprintln(wr, tc.SessionTasksURL())
		}
	}
}
