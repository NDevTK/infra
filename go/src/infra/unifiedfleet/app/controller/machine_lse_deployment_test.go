// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
	"google.golang.org/genproto/protobuf/field_mask"

	ufspb "infra/unifiedfleet/api/v1/models"
	"infra/unifiedfleet/app/model/history"
	"infra/unifiedfleet/app/model/inventory"
)

func mockMachineLSEDeployment(serialNumber string) *ufspb.MachineLSEDeployment {
	return &ufspb.MachineLSEDeployment{
		SerialNumber: serialNumber,
	}
}

func TestUpdateMachineLSEDeployment(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("UpdateMachineLSEDeployment", t, func() {
		Convey("Update MachineLSEDeployment for non-existing MachineLSEDeployment - happy path", func() {
			dr1 := mockMachineLSEDeployment("serial-1")
			resp, err := UpdateMachineLSEDeployment(ctx, dr1, nil)
			So(err, ShouldBeNil)
			So(resp.GetHostname(), ShouldEqual, "no-host-yet-serial-1")
			So(resp.GetSerialNumber(), ShouldEqual, "serial-1")

			resGet, err := inventory.GetMachineLSEDeployment(ctx, "serial-1")
			So(err, ShouldBeNil)
			So(resGet, ShouldResembleProto, resp)
		})

		Convey("Update MachineLSEDeployment for existing MachineLSEDeployment - happy path", func() {
			dr2 := mockMachineLSEDeployment("serial-2")
			_, err := UpdateMachineLSEDeployment(ctx, dr2, nil)
			So(err, ShouldBeNil)

			dr2.Hostname = "hostname-2"
			resp, err := UpdateMachineLSEDeployment(ctx, dr2, nil)
			So(err, ShouldBeNil)
			So(resp.GetHostname(), ShouldEqual, "hostname-2")
			So(resp.GetSerialNumber(), ShouldEqual, "serial-2")

			resGet, err := inventory.GetMachineLSEDeployment(ctx, "serial-2")
			So(err, ShouldBeNil)
			So(resGet, ShouldResembleProto, dr2)

			// Verify change events
			changes, err := history.QueryChangesByPropertyName(ctx, "name", "machineLSEDeployments/serial-2")
			So(err, ShouldBeNil)
			// Record the first-time registration
			So(changes, ShouldHaveLength, 2)
			So(changes[0].GetEventLabel(), ShouldEqual, "machine_lse_deployment")
			So(changes[0].GetOldValue(), ShouldEqual, LifeCycleRegistration)
			So(changes[0].GetNewValue(), ShouldEqual, LifeCycleRegistration)
			// Record the update
			So(changes[1].GetEventLabel(), ShouldEqual, "machine_lse_deployment.hostname")
			So(changes[1].GetOldValue(), ShouldEqual, "no-host-yet-serial-2")
			So(changes[1].GetNewValue(), ShouldEqual, "hostname-2")
			msgs, err := history.QuerySnapshotMsgByPropertyName(ctx, "resource_name", "machineLSEDeployments/serial-2")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 2)
			So(msgs[0].Delete, ShouldBeFalse)
			So(msgs[1].Delete, ShouldBeFalse)
		})

		Convey("Update MachineLSEDeployment for existing MachineLSEDeployment - partial update hostname & deployment env", func() {
			dr3 := mockMachineLSEDeployment("serial-3")
			_, err := UpdateMachineLSEDeployment(ctx, dr3, nil)
			So(err, ShouldBeNil)

			dr3.Hostname = "hostname-3"
			dr3.DeploymentIdentifier = "identifier-3"
			dr3.DeploymentEnv = ufspb.DeploymentEnv_AUTOPUSH
			resp, err := UpdateMachineLSEDeployment(ctx, dr3, &field_mask.FieldMask{Paths: []string{"hostname", "deployment_env"}})
			So(err, ShouldBeNil)
			So(resp.GetSerialNumber(), ShouldEqual, "serial-3")
			So(resp.GetHostname(), ShouldEqual, "hostname-3")
			So(resp.GetDeploymentIdentifier(), ShouldBeEmpty)

			resGet, err := inventory.GetMachineLSEDeployment(ctx, "serial-3")
			So(err, ShouldBeNil)
			dr3.DeploymentIdentifier = ""
			So(resGet.GetHostname(), ShouldEqual, "hostname-3")
			So(resGet.GetDeploymentIdentifier(), ShouldBeEmpty)
			So(resGet.GetDeploymentEnv(), ShouldEqual, ufspb.DeploymentEnv_AUTOPUSH)

			// Verify change events
			changes, err := history.QueryChangesByPropertyName(ctx, "name", "machineLSEDeployments/serial-3")
			So(err, ShouldBeNil)
			// Record the first-time registration
			So(changes, ShouldHaveLength, 3)
			So(changes[0].GetEventLabel(), ShouldEqual, "machine_lse_deployment")
			So(changes[0].GetOldValue(), ShouldEqual, LifeCycleRegistration)
			So(changes[0].GetNewValue(), ShouldEqual, LifeCycleRegistration)
			// Record the update
			So(changes[1].GetEventLabel(), ShouldEqual, "machine_lse_deployment.hostname")
			So(changes[1].GetOldValue(), ShouldEqual, "no-host-yet-serial-3")
			So(changes[1].GetNewValue(), ShouldEqual, "hostname-3")
			So(changes[2].GetEventLabel(), ShouldEqual, "machine_lse_deployment.deployment_env")
			So(changes[2].GetOldValue(), ShouldEqual, ufspb.DeploymentEnv_DEPLOYMENTENV_UNDEFINED.String())
			So(changes[2].GetNewValue(), ShouldEqual, ufspb.DeploymentEnv_AUTOPUSH.String())
			msgs, err := history.QuerySnapshotMsgByPropertyName(ctx, "resource_name", "machineLSEDeployments/serial-2")
			So(err, ShouldBeNil)
			So(msgs, ShouldHaveLength, 2)
			So(msgs[0].Delete, ShouldBeFalse)
			So(msgs[1].Delete, ShouldBeFalse)
		})
	})
}

func TestListMachineLSEDeployments(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	drs := make([]*ufspb.MachineLSEDeployment, 0, 4)
	for i := 0; i < 4; i++ {
		dr := mockMachineLSEDeployment(fmt.Sprintf("list-dr-%d", i))
		if i%2 == 0 {
			dr.Hostname = fmt.Sprintf("host-%d", i)
		}
		drs = append(drs, dr)
	}
	updatedDrs, _ := inventory.UpdateMachineLSEDeployments(ctx, drs)
	Convey("ListMachineLSEDeployments", t, func() {
		Convey("List MachineLSEDeployments - filter invalid - error", func() {
			_, _, err := ListMachineLSEDeployments(ctx, 5, "", "invalid=mx-1", false)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "Invalid field name invalid")
		})

		Convey("List MachineLSEDeployment - filter host - happy path", func() {
			resp, _, err := ListMachineLSEDeployments(ctx, 5, "", "host=host-0", false)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 1)
			So(resp[0], ShouldResembleProto, updatedDrs[0])

			resp, _, err = ListMachineLSEDeployments(ctx, 5, "", "host=host-2", false)
			So(err, ShouldBeNil)
			So(resp, ShouldHaveLength, 1)
			So(resp[0], ShouldResembleProto, updatedDrs[2])
		})

		Convey("List MachineLSEDeployment - Full listing - happy path", func() {
			resp, _, _ := ListMachineLSEDeployments(ctx, 5, "", "", false)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, updatedDrs)
		})
	})
}
