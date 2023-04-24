// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package api

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/protobuf/types/known/durationpb"

	vmleaserpb "infra/vm_leaser/api/v1"
)

func TestValidateVmLeaserBackend(t *testing.T) {
	Convey("CreateVmInstanceRequest Validate", t, func() {
		Convey("Valid request - successful path", func() {
			d, err := time.ParseDuration("60s")
			So(err, ShouldBeNil)

			req := &CreateVmInstanceRequest{
				Config: &Config{
					Backend: &Config_VmLeaserBackend_{
						VmLeaserBackend: &Config_VmLeaserBackend{
							Env: Config_VmLeaserBackend_ENV_LOCAL,
							VmRequirements: &vmleaserpb.VMRequirements{
								GceImage:       "test-image",
								GceRegion:      "test-region",
								GceProject:     "test-project",
								GceMachineType: "test-machine-type",
								GceDiskSize:    100,
							},
							LeaseDuration: durationpb.New(d),
						},
					},
				},
			}
			err = req.ValidateVmLeaserBackend()
			So(err, ShouldBeNil)
		})
		Convey("Invalid request - empty request", func() {
			req := &CreateVmInstanceRequest{}
			err := req.ValidateVmLeaserBackend()
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "invalid argument: no config found")
		})
		Convey("Invalid request - empty VmLeaserBackend", func() {
			req := &CreateVmInstanceRequest{
				Config: &Config{},
			}
			err := req.ValidateVmLeaserBackend()
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "invalid argument: bad backend: want vmleaser")
		})
		Convey("Invalid request - wrong backend", func() {
			req := &CreateVmInstanceRequest{
				Config: &Config{
					Backend: &Config_GcloudBackend{
						GcloudBackend: &Config_GCloudBackend{},
					},
				},
			}
			err := req.ValidateVmLeaserBackend()
			So(err, ShouldNotBeNil)
			So(err, ShouldErrLike, "invalid argument: bad backend: want vmleaser")
		})
	})
}
