// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmleaserpb

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Validate validates input requests of LeaseVMRequest.
func (r *LeaseVMRequest) Validate() error {
	hostReqs := r.GetHostReqs()
	if hostReqs == nil {
		return status.Errorf(codes.InvalidArgument, "VM requirements must be set.")
	}
	if err := hostReqs.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates input requests of ReleaseVMRequest.
func (r *ReleaseVMRequest) Validate() error {
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

// Validate validates the VMRequirements.
func (r *VMRequirements) Validate() error {
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
