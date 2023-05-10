// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmleaserpb

import (
	"go.chromium.org/chromiumos/config/go/test/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValidateLeaseVMRequest validates input requests of LeaseVMRequest.
func ValidateLeaseVMRequest(r *api.LeaseVMRequest) error {
	hostReqs := r.GetHostReqs()
	if hostReqs == nil {
		return status.Errorf(codes.InvalidArgument, "VM requirements must be set.")
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
