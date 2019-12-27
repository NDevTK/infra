// Copyright 2019 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package request provides a library to create swarming requests based on
// skylab test or task parameters.
package request

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"go.chromium.org/chromiumos/infra/proto/go/test_platform/skylab_test_runner"
	buildbucket_pb "go.chromium.org/luci/buildbucket/proto"
	swarming "go.chromium.org/luci/common/api/swarming/swarming/v1"
	"go.chromium.org/luci/common/data/strpair"
	"go.chromium.org/luci/common/errors"

	"infra/libs/skylab/inventory"
	swarming_inventory "infra/libs/skylab/inventory/swarming"
	"infra/libs/skylab/worker"
)

// Args defines the set of arguments for creating a request.
type Args struct {
	// Cmd specifies the payload command to run for the request.
	Cmd          worker.Command
	SwarmingTags []string
	// ProvisionableDimensions specifies the provisionable dimensions in raw
	// string form; e.g. {"provisionable-cros-version:foo-cq-R75-1.2.3.4"}
	ProvisionableDimensions []string
	// Dimensions specifies swarming dimensions in raw string form.
	//
	// It is preferable to specify dimensions via the SchedulableLabels
	// argument. This argument should only be used for user-supplied freeform
	// dimensions; e.g. {"label-power:battery"}
	//
	// TODO(akeshet): This feature is needed to support `skylab create-test`
	// which allows arbitrary user-specified dimensions. If and when that
	// feature is dropped, then this feature can be dropped as well.
	Dimensions []string
	// SchedulableLabels specifies schedulable label requirements that will
	// be translated to dimensions.
	SchedulableLabels inventory.SchedulableLabels
	Timeout           time.Duration
	Priority          int64
	ParentTaskID      string
	//Pubsub Topic for status updates on the tests run for the request
	StatusTopic string
	// BuilderID identifies the builder that will run the test task.
	BuilderID *buildbucket_pb.BuilderID
	// Test describes the test to be run.
	Test *skylab_test_runner.Request_Test
}

// BuildbucketNewBuildRequest returns the Buildbucket request to create the
// test_runner build with these arguments.
func (a *Args) BuildbucketNewBuildRequest() (*buildbucket_pb.ScheduleBuildRequest, error) {
	dims, err := a.getBBDimensions()
	if err != nil {
		return nil, errors.Annotate(err, "create bb request").Err()
	}

	provisionableLabels, err := provisionDimensionsToLabelDict(a.ProvisionableDimensions)
	if err != nil {
		return nil, errors.Annotate(err, "create bb request").Err()
	}

	// TODO(crbug.com/1036559#c1): Add timeouts.
	req, err := requestToStructPB(&skylab_test_runner.Request{
		Prejob: &skylab_test_runner.Request_Prejob{
			ProvisionableLabels: provisionableLabels,
		},
		Test: a.Test,
	})
	if err != nil {
		return nil, errors.Annotate(err, "create bb request").Err()
	}

	props := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"request": req,
		},
	}

	tags, err := splitTagPairs(a.SwarmingTags)
	if err != nil {
		return nil, errors.Annotate(err, "create bb request").Err()
	}

	return &buildbucket_pb.ScheduleBuildRequest{
		Builder:    a.BuilderID,
		Properties: props,
		Tags:       tags,
		Dimensions: dims,
		Priority:   int32(a.Priority),
		Swarming: &buildbucket_pb.ScheduleBuildRequest_Swarming{
			ParentRunId: a.ParentTaskID,
		},
		Notify: getNotificationConfig(a.StatusTopic),
	}, nil
}

// getBBDimensions returns both required and optional dimensions that will be
// used to match this request with a Swarming bot.
func (a *Args) getBBDimensions() ([]*buildbucket_pb.RequestedDimension, error) {
	dims := schedulableLabelsToRequestedDimensions(&a.SchedulableLabels)

	optionalDimExpiration := 30 * time.Second
	pd, err := stringsToRequestedDimensions(a.ProvisionableDimensions, &optionalDimExpiration)
	if err != nil {
		return nil, errors.Annotate(err, "get BB dimensions").Err()
	}
	dims = append(dims, pd...)

	// TODO(zamorzaev): move the dut_state dimension to the builder config.
	dims = append(dims, &buildbucket_pb.RequestedDimension{
		Key:   "dut_state",
		Value: "ready",
	})

	extraDims, err := stringsToRequestedDimensions(a.Dimensions, nil)
	if err != nil {
		return nil, errors.Annotate(err, "get BB dimensions").Err()
	}
	dims = append(dims, extraDims...)
	return dims, nil
}

// schedulableLabelsToRequestedDimensions converts a slice of strings in foo:bar form to a slice of BB
// rpc requested dimensions.
func schedulableLabelsToRequestedDimensions(inv *inventory.SchedulableLabels) []*buildbucket_pb.RequestedDimension {
	var rd []*buildbucket_pb.RequestedDimension
	id := swarming_inventory.Convert(inv)
	for key, values := range id {
		for _, value := range values {
			rd = append(rd, &buildbucket_pb.RequestedDimension{
				Key:   key,
				Value: value,
			})
		}
	}
	return rd
}

// stringsToRequestedDimensions converts a slice of strings in foo:bar form to
// a slice of BB rpc requested dimensions.
func stringsToRequestedDimensions(dims []string, expiration *time.Duration) ([]*buildbucket_pb.RequestedDimension, error) {
	var expirationPb *duration.Duration
	if expiration != nil {
		expirationPb = ptypes.DurationProto(*expiration)
	}

	ret := make([]*buildbucket_pb.RequestedDimension, len(dims))
	for i, d := range dims {
		k, v := strpair.Parse(d)
		if v == "" {
			return nil, fmt.Errorf("malformed dimension with key '%s' has no value", k)
		}
		ret[i] = &buildbucket_pb.RequestedDimension{
			Key:        k,
			Value:      v,
			Expiration: expirationPb,
		}
	}
	return ret, nil
}

// provisionDimensionsToLabelDict converts provisionable dimensions to labels.
func provisionDimensionsToLabelDict(dims []string) (map[string]string, error) {
	labels := make(map[string]string)
	for _, d := range dims {
		k, v := strpair.Parse(d)
		if v == "" {
			return nil, fmt.Errorf("malformed provisionable dimension with key '%s' has no value", k)
		}
		k = strings.TrimPrefix(k, "provisionable-")
		labels[k] = v
	}
	return labels, nil
}

// splitTagPairs converts a slice of strings in foo:bar form to a slice of BB
// rpc string pairs.
func splitTagPairs(tags []string) ([]*buildbucket_pb.StringPair, error) {
	ret := make([]*buildbucket_pb.StringPair, len(tags))
	for i, t := range tags {
		k, v := strpair.Parse(t)
		if v == "" {
			return nil, fmt.Errorf("malformed tag with key '%s' has no value", k)
		}
		ret[i] = &buildbucket_pb.StringPair{
			Key:   k,
			Value: v,
		}
	}
	return ret, nil
}

// requestToStructPB converts a skylab_test_runner.Request into a Struct
// with the same JSON presentation.
func requestToStructPB(from *skylab_test_runner.Request) (*structpb.Value, error) {
	m := jsonpb.Marshaler{}
	jsonStr, err := m.MarshalToString(from)
	if err != nil {
		return nil, err
	}
	reqStruct := &structpb.Struct{}
	if err := jsonpb.UnmarshalString(jsonStr, reqStruct); err != nil {
		return nil, err
	}
	return &structpb.Value{
		Kind: &structpb.Value_StructValue{StructValue: reqStruct},
	}, nil
}

// getNotificationConfig constructs a valid NotificationConfig.
func getNotificationConfig(topic string) *buildbucket_pb.NotificationConfig {
	if topic == "" {
		// BB will crash if it encounters a non-nil NotificationConfig with an
		// empty PubsubTopic.
		return nil
	}
	return &buildbucket_pb.NotificationConfig{
		PubsubTopic: topic,
	}
}

// SwarmingNewTaskRequest returns the Swarming request to create the Skylab
// task with these arguments.
func (a *Args) SwarmingNewTaskRequest() (*swarming.SwarmingRpcsNewTaskRequest, error) {
	dims, err := a.StaticDimensions()
	if err != nil {
		return nil, errors.Annotate(err, "create request").Err()
	}
	slices, err := getSlices(a.Cmd, dims, a.ProvisionableDimensions, a.Timeout)
	if err != nil {
		return nil, errors.Annotate(err, "create request").Err()
	}

	req := &swarming.SwarmingRpcsNewTaskRequest{
		Name:         a.Cmd.TaskName,
		Tags:         a.SwarmingTags,
		TaskSlices:   slices,
		Priority:     a.Priority,
		ParentTaskId: a.ParentTaskID,
		PubsubTopic:  a.StatusTopic,
	}
	return req, nil
}

// StaticDimensions returns the dimensions required on a Swarming bot that can
// service this request.
//
// StaticDimensions() do not include dimensions used to optimize task
// scheduling.
func (a *Args) StaticDimensions() ([]*swarming.SwarmingRpcsStringPair, error) {
	ret := schedulableLabelsToPairs(a.SchedulableLabels)
	d, err := stringToPairs(a.Dimensions...)
	if err != nil {
		return nil, errors.Annotate(err, "get static dimensions").Err()
	}
	ret = append(ret, d...)
	ret = append(ret, &swarming.SwarmingRpcsStringPair{
		Key:   "pool",
		Value: "ChromeOSSkylab",
	})
	return ret, nil
}

// getSlices generates and returns the set of swarming task slices for the given test task.
func getSlices(cmd worker.Command, staticDimensions []*swarming.SwarmingRpcsStringPair, provisionableDimensions []string, timeout time.Duration) ([]*swarming.SwarmingRpcsTaskSlice, error) {
	slices := make([]*swarming.SwarmingRpcsTaskSlice, 1, 2)

	dims, _ := stringToPairs("dut_state:ready")
	dims = append(dims, staticDimensions...)

	provisionablePairs, err := stringToPairs(provisionableDimensions...)
	if err != nil {
		return nil, errors.Annotate(err, "create slices").Err()
	}

	s0Dims := append(dims, provisionablePairs...)
	slices[0] = taskSlice(cmd.Args(), s0Dims, timeout)

	if len(provisionableDimensions) != 0 {
		cmd.ProvisionLabels = provisionDimensionsToLabels(provisionableDimensions)
		s1Dims := dims
		slices = append(slices, taskSlice(cmd.Args(), s1Dims, timeout))
	}

	finalSlice := slices[len(slices)-1]
	finalSlice.ExpirationSecs = int64(timeout.Seconds())

	return slices, nil
}

func taskSlice(command []string, dimensions []*swarming.SwarmingRpcsStringPair, timeout time.Duration) *swarming.SwarmingRpcsTaskSlice {
	return &swarming.SwarmingRpcsTaskSlice{
		// We want all slices to wait, at least a little while, for bots with
		// metching dimensions.
		// For slice 0: This allows the task to try to re-use provisionable
		// labels that get set by previous tasks with the same label that are
		// about to finish.
		// For slice 1: This allows the task to wait for devices to get
		// repaired, if there are no devices with dut_state:ready.
		WaitForCapacity: true,
		// Slice 0 should have a fairly short expiration time, to reduce
		// overhead for tasks that are the first ones enqueue with a particular
		// provisionable label. This value will be overwritten for the final
		// slice of a task.
		ExpirationSecs: 30,
		Properties: &swarming.SwarmingRpcsTaskProperties{
			Command:              command,
			Dimensions:           dimensions,
			ExecutionTimeoutSecs: int64(timeout.Seconds()),
		},
	}
}

// provisionDimensionsToLabels converts provisionable dimensions to labels.
func provisionDimensionsToLabels(dims []string) []string {
	labels := make([]string, len(dims))
	for i, l := range dims {
		labels[i] = strings.TrimPrefix(l, "provisionable-")
	}
	return labels
}

// stringToPairs converts a slice of strings in foo:bar form to a slice of swarming
// rpc string pairs.
func stringToPairs(dimensions ...string) ([]*swarming.SwarmingRpcsStringPair, error) {
	pairs := make([]*swarming.SwarmingRpcsStringPair, len(dimensions))
	for i, d := range dimensions {
		k, v := strpair.Parse(d)
		if v == "" {
			return nil, fmt.Errorf("malformed dimension with key '%s' has no value", k)
		}
		pairs[i] = &swarming.SwarmingRpcsStringPair{Key: k, Value: v}
	}
	return pairs, nil
}

func schedulableLabelsToPairs(inv inventory.SchedulableLabels) []*swarming.SwarmingRpcsStringPair {
	dimensions := swarming_inventory.Convert(&inv)
	pairs := make([]*swarming.SwarmingRpcsStringPair, 0, len(dimensions))
	for key, values := range dimensions {
		for _, value := range values {
			pairs = append(pairs, &swarming.SwarmingRpcsStringPair{Key: key, Value: value})
		}
	}
	return pairs
}
