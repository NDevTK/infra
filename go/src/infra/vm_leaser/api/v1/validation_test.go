// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package vmleaserpb

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLeaseVMValidate(t *testing.T) {
	Convey("LeaseVMRequest Validate", t, func() {
		Convey("Valid request - successful path", func() {
			req := &LeaseVMRequest{
				HostReqs: &VMRequirements{
					GceImage:       "test-image",
					GceRegion:      "test-region",
					GceProject:     "test-project",
					GceMachineType: "test-machine-type",
					GceDiskSize:    100,
				},
			}
			err := req.Validate()
			So(err, ShouldBeNil)
		})
		Convey("Invalid request - missing host requirements", func() {
			req := &LeaseVMRequest{}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "VM requirements must be set.")
		})
		Convey("Invalid request - missing image", func() {
			req := &LeaseVMRequest{
				HostReqs: &VMRequirements{
					GceRegion:      "test-region",
					GceProject:     "test-project",
					GceMachineType: "test-machine-type",
					GceDiskSize:    100,
				},
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE image must be set.")
		})
		Convey("Invalid request - missing region", func() {
			req := &LeaseVMRequest{
				HostReqs: &VMRequirements{
					GceImage:       "test-image",
					GceProject:     "test-project",
					GceMachineType: "test-machine-type",
					GceDiskSize:    100,
				},
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE region (zone) must be set.")
		})
		Convey("Invalid request - missing project", func() {
			req := &LeaseVMRequest{
				HostReqs: &VMRequirements{
					GceImage:       "test-image",
					GceRegion:      "test-region",
					GceMachineType: "test-machine-type",
					GceDiskSize:    100,
				},
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE project must be set.")
		})
		Convey("Invalid request - missing machine type", func() {
			req := &LeaseVMRequest{
				HostReqs: &VMRequirements{
					GceImage:    "test-image",
					GceRegion:   "test-region",
					GceProject:  "test-project",
					GceDiskSize: 100,
				},
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE machine type must be set.")
		})
		Convey("Invalid request - missing disk size", func() {
			req := &LeaseVMRequest{
				HostReqs: &VMRequirements{
					GceImage:       "test-image",
					GceRegion:      "test-region",
					GceProject:     "test-project",
					GceMachineType: "test-machine-type",
				},
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE machine disk size must be set (in GB).")
		})
	})
}

func TestReleaseVMValidate(t *testing.T) {
	Convey("ReleaseVMRequest Validate", t, func() {
		Convey("Valid request - successful path", func() {
			req := &ReleaseVMRequest{
				LeaseId:    "test-lease-id",
				GceProject: "test-project",
				GceRegion:  "test-region",
			}
			err := req.Validate()
			So(err, ShouldBeNil)
		})
		Convey("Invalid request - missing lease id", func() {
			req := &ReleaseVMRequest{
				GceProject: "test-project",
				GceRegion:  "test-region",
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Lease ID must be set.")
		})
		Convey("Invalid request - missing project", func() {
			req := &ReleaseVMRequest{
				LeaseId:   "test-lease-id",
				GceRegion: "test-region",
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE project must be set.")
		})
		Convey("Invalid request - missing region", func() {
			req := &ReleaseVMRequest{
				LeaseId:    "test-lease-id",
				GceProject: "test-project",
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE region (zone) must be set.")
		})
	})
}

func TestVMRequirementsValidate(t *testing.T) {
	Convey("VMRequirements Validate", t, func() {
		Convey("Valid request - successful path", func() {
			req := &VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			err := req.Validate()
			So(err, ShouldBeNil)
		})
		Convey("Invalid request - missing image", func() {
			req := &VMRequirements{
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE image must be set.")
		})
		Convey("Invalid request - missing region", func() {
			req := &VMRequirements{
				GceImage:       "test-image",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE region (zone) must be set.")
		})
		Convey("Invalid request - missing project", func() {
			req := &VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceMachineType: "test-machine-type",
				GceDiskSize:    100,
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE project must be set.")
		})
		Convey("Invalid request - missing machine type", func() {
			req := &VMRequirements{
				GceImage:    "test-image",
				GceRegion:   "test-region",
				GceProject:  "test-project",
				GceDiskSize: 100,
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE machine type must be set.")
		})
		Convey("Invalid request - missing disk size", func() {
			req := &VMRequirements{
				GceImage:       "test-image",
				GceRegion:      "test-region",
				GceProject:     "test-project",
				GceMachineType: "test-machine-type",
			}
			err := req.Validate()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "GCE machine disk size must be set (in GB).")
		})
	})
}
