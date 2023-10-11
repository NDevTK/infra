// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package validation

import (
	"regexp"

	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// Lease parent resource in requests should compile to
	// `projects/${project}` or `projects/${project}/zones/${zone}`.
	// The parent resource for all leases should be a GCP project or a project
	// with a zone specified.
	ValidLeaseParent = regexp.MustCompile(`^projects\/(?P<project>[-|\w]+)\/?(?:zones\/(?P<zone>[-|\w]+))?\/?$`)
)

// ValidateLeaseVMRequest validates input requests of LeaseVMRequest.
func ValidateLeaseVMRequest(r *api.LeaseVMRequest) error {
	hostReqs := r.GetHostReqs()
	if hostReqs == nil {
		return status.Errorf(codes.InvalidArgument, "VM requirements must be set.")
	}
	if r.GetTestingClient() == api.VMTestingClient_VM_TESTING_CLIENT_CROSFLEET {
		l := r.GetLabels()
		if l == nil {
			return status.Errorf(codes.InvalidArgument, "Labels should not be nil")
		}

		client, ok := l["client"]
		if !ok || client != "crosfleet" {
			return status.Errorf(codes.InvalidArgument, "Labels should contain \"client\"=\"crosfleet\"")
		}
		_, ok = l["leased-by"]
		if !ok {
			return status.Errorf(codes.InvalidArgument, "Labels should contain \"leased-by\"={email}")
		}
	}
	if err := ValidateVMRequirements(hostReqs); err != nil {
		return err
	}
	return nil
}

// ValidateReleaseVMRequest validates input requests of ReleaseVMRequest.
func ValidateReleaseVMRequest(r *api.ReleaseVMRequest) error {
	if r.GetLeaseId() == "" {
		return status.Errorf(codes.InvalidArgument, "Lease ID must be set.")
	}
	if r.GetGceProject() == "" {
		return status.Errorf(codes.InvalidArgument, "GCE project must be set.")
	}
	if r.GetGceRegion() == "" {
		return status.Errorf(codes.InvalidArgument, "GCE region (zone) must be set.")
	}
	return nil
}

// ValidateVMRequirements validates the VMRequirements.
func ValidateVMRequirements(r *api.VMRequirements) error {
	if r.GetGceImage() == "" {
		return status.Errorf(codes.InvalidArgument, "GCE image must be set.")
	}
	if r.GetGceRegion() == "" {
		return status.Errorf(codes.InvalidArgument, "GCE region (zone) must be set.")
	}
	if r.GetGceProject() == "" {
		return status.Errorf(codes.InvalidArgument, "GCE project must be set.")
	}
	if r.GetGceMachineType() == "" {
		return status.Errorf(codes.InvalidArgument, "GCE machine type must be set.")
	}
	if r.GetGceDiskSize() == 0 {
		return status.Errorf(codes.InvalidArgument, "GCE machine disk size must be set (in GB).")
	}
	return nil
}

// ValidateLeaseParent validates the parent field to be `projects/${project}`.
func ValidateLeaseParent(parent string) error {
	if !ValidLeaseParent.MatchString(parent) {
		return status.Errorf(codes.InvalidArgument, "parent must be in the format `projects/${project}` or `projects/${project}/zones/${zone}`")
	}
	return nil
}
