// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package cmd

import (
	"errors"
	"fmt"
	"testing"

	"infra/libs/vmlab/api"

	"infra/cmd/vmlab/internal/config"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

type mockInstanceApi struct {
	listInstancesFunc  func(req *api.ListVmInstancesRequest) ([]*api.VmInstance, error)
	deleteInstanceFunc func(ins *api.VmInstance) error
}

func (m *mockInstanceApi) List(req *api.ListVmInstancesRequest) ([]*api.VmInstance, error) {
	return m.listInstancesFunc(req)
}

func (m *mockInstanceApi) Delete(ins *api.VmInstance) error {
	return m.deleteInstanceFunc(ins)
}

func (m *mockInstanceApi) Create(req *api.CreateVmInstanceRequest) (*api.VmInstance, error) {
	return nil, errors.New("not supported")
}

// isEquivalentTo compares if two `cleanupInstancesResult` have the same number
// of total instances and the same items in deleted, failed instances. Ordering
// of items doesn't matter.
func (a *cleanupInstancesResult) isEquivalentTo(b *cleanupInstancesResult) bool {
	less := func(a, b string) bool { return a < b }
	if a.Total != b.Total ||
		!cmp.Equal(a.Deleted, b.Deleted, cmpopts.SortSlices(less)) ||
		!cmp.Equal(a.Failed, b.Failed, cmpopts.SortSlices(less)) {
		return false
	}
	return true
}

var INSTANCE_1 = &api.VmInstance{
	Name: "gcetest-instance-1",
}

var INSTANCE_2 = &api.VmInstance{
	Name: "gcetest-instance-2",
}

var CONFIG = &config.BuiltinConfig{
	ProviderId: api.ProviderId_GCLOUD,
	GcloudConfig: api.Config_GCloudBackend{
		Project:        "vmlab-project",
		Zone:           "us-west-2",
		MachineType:    "n2-standard-4",
		InstancePrefix: "gcetest-",
	},
}

var EXPECTED_REQUEST = &api.ListVmInstancesRequest{
	Config: &api.Config{
		Backend: &api.Config_GcloudBackend{
			GcloudBackend: &api.Config_GCloudBackend{
				Project:        "vmlab-project",
				Zone:           "us-west-2",
				MachineType:    "n2-standard-4",
				InstancePrefix: "gcetest-",
			},
		},
	},
	TagFilters: map[string]string{
		"swarming-bot-name": "test-bot",
	},
}

func TestCleanup(t *testing.T) {
	insApi := &mockInstanceApi{
		listInstancesFunc: func(req *api.ListVmInstancesRequest) ([]*api.VmInstance, error) {
			if diff := cmp.Diff(req, EXPECTED_REQUEST, protocmp.Transform()); diff != "" {
				return []*api.VmInstance{}, fmt.Errorf("bad request: %v, diff %v", req, diff)
			}
			return []*api.VmInstance{INSTANCE_1, INSTANCE_2}, nil
		},
		deleteInstanceFunc: func(ins *api.VmInstance) error {
			return nil
		},
	}

	result, err := cleanupInstances(insApi, CONFIG, "test-bot", 1000, false)

	if err != nil {
		t.Fatalf("cleanupInstances() returned error: %v", err)
	}

	expectedResult := cleanupInstancesResult{
		Total:   2,
		Deleted: []string{"gcetest-instance-1", "gcetest-instance-2"},
		Failed:  []string{},
	}

	if !expectedResult.isEquivalentTo(&result) {
		t.Errorf("Expected result to be %v, but is %v", expectedResult, result)
	}
}

func TestCleanupPartialSuccess(t *testing.T) {
	insApi := &mockInstanceApi{
		listInstancesFunc: func(req *api.ListVmInstancesRequest) ([]*api.VmInstance, error) {
			if diff := cmp.Diff(req, EXPECTED_REQUEST, protocmp.Transform()); diff != "" {
				return []*api.VmInstance{}, fmt.Errorf("bad request: %v, diff %v", req, diff)
			}
			return []*api.VmInstance{INSTANCE_1, INSTANCE_2}, nil
		},
		deleteInstanceFunc: func(ins *api.VmInstance) error {
			if ins.Name == "gcetest-instance-1" {
				return nil
			}
			return errors.New("failed to delete this instance.")
		},
	}

	result, err := cleanupInstances(insApi, CONFIG, "test-bot", 1000, false)

	if err != nil {
		t.Fatalf("cleanupInstances() returned error: %v", err)
	}

	expectedResult := cleanupInstancesResult{
		Total:   2,
		Deleted: []string{"gcetest-instance-1"},
		Failed:  []string{"gcetest-instance-2"},
	}

	if !expectedResult.isEquivalentTo(&result) {
		t.Errorf("Expected result to be %v, but is %v", expectedResult, result)
	}
}

func TestCleanupDryRun(t *testing.T) {
	insApi := &mockInstanceApi{
		listInstancesFunc: func(req *api.ListVmInstancesRequest) ([]*api.VmInstance, error) {
			if diff := cmp.Diff(req, EXPECTED_REQUEST, protocmp.Transform()); diff != "" {
				return []*api.VmInstance{}, fmt.Errorf("bad request: %v, diff %v", req, diff)
			}
			return []*api.VmInstance{INSTANCE_1, INSTANCE_2}, nil
		},
		deleteInstanceFunc: func(ins *api.VmInstance) error {
			return errors.New("delete should not be called for dryrun.")
		},
	}

	result, err := cleanupInstances(insApi, CONFIG, "test-bot", 1000, true)

	if err != nil {
		t.Fatalf("cleanupInstances() returned error: %v", err)
	}

	expectedResult := cleanupInstancesResult{
		Total:   2,
		Deleted: []string{"gcetest-instance-1", "gcetest-instance-2"},
		Failed:  []string{},
	}

	if !expectedResult.isEquivalentTo(&result) {
		t.Errorf("Expected result to be %v, but is %v", expectedResult, result)
	}
}
