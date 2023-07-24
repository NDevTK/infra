// Copyright 2023 The Chromium Authors
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package common_test

import (
	"infra/cros/cmd/common_lib/common"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	_go "go.chromium.org/chromiumos/config/go"
	testapi "go.chromium.org/chromiumos/config/go/test/api"
	labapi "go.chromium.org/chromiumos/config/go/test/lab/api"
)

func TestDependencyInjectionBasic(t *testing.T) {
	Convey("basic injection", t, func() {
		original_proto := &testapi.CrosProvisionRequest{
			ProvisionState: &testapi.ProvisionState{
				Id: &testapi.ProvisionState_Id{
					Value: "Hello, World!",
				},
				PreventReboot: true,
			},
		}
		dut_address_proto := &labapi.IpEndpoint{
			Address: "localhost",
			Port:    4040,
		}
		injection_map := map[string]interface{}{}
		injection_map["dut_primary"] = dut_address_proto

		err := common.Inject(original_proto, "dutServer", injection_map, "dut_primary")

		So(err, ShouldBeNil)
		So(original_proto.DutServer, ShouldNotBeNil)
		So(original_proto.DutServer.Address, ShouldEqual, dut_address_proto.Address)
		So(original_proto.DutServer.Port, ShouldEqual, dut_address_proto.Port)
	})
}

func TestDependencyInjectionArray(t *testing.T) {
	Convey("array injection", t, func() {
		original_proto := &testapi.CrosProvisionRequest{
			ProvisionState: &testapi.ProvisionState{
				Id: &testapi.ProvisionState_Id{
					Value: "Hello, World!",
				},
				PreventReboot: true,
			},
		}
		dut_address_protos := []labapi.IpEndpoint{
			{
				Address: "not_expected",
				Port:    4040,
			},
			{
				Address: "expected",
				Port:    1234,
			},
		}

		injection_map := map[string]interface{}{}
		injection_map["duts"] = dut_address_protos

		err := common.Inject(original_proto, "dutServer", injection_map, "duts.1")

		So(err, ShouldBeNil)
		So(original_proto.DutServer, ShouldNotBeNil)
		So(original_proto.DutServer.Address, ShouldEqual, dut_address_protos[1].Address)
		So(original_proto.DutServer.Port, ShouldEqual, dut_address_protos[1].Port)
	})
}

func TestDependencyInjectionArrayAppend(t *testing.T) {
	Convey("array injection", t, func() {
		original_proto := &testapi.CrosProvisionRequest{
			ProvisionState: &testapi.ProvisionState{
				Id: &testapi.ProvisionState_Id{
					Value: "Hello, World!",
				},
				PreventReboot: true,
				Packages: []*testapi.ProvisionState_Package{
					{
						PackagePath: &_go.StoragePath{
							Path: "a",
						},
					},
					{
						PackagePath: &_go.StoragePath{
							Path: "b",
						},
					},
					{
						PackagePath: &_go.StoragePath{
							Path: "c",
						},
					},
				},
			},
		}
		new_package := &testapi.ProvisionState_Package{
			PackagePath: &_go.StoragePath{
				Path: "d",
			},
		}

		injection_map := map[string]interface{}{}
		injection_map["package"] = new_package

		err := common.Inject(original_proto, "provisionState.packages", injection_map, "package")

		So(err, ShouldBeNil)
		So(original_proto.ProvisionState.Packages, ShouldHaveLength, 4)
		So(original_proto.ProvisionState.Packages[3].PackagePath.Path, ShouldEqual, new_package.PackagePath.Path)
	})
}

func TestDependencyInjectionArrayOverride(t *testing.T) {
	Convey("array override injection", t, func() {
		original_proto := &testapi.CrosProvisionRequest{
			ProvisionState: &testapi.ProvisionState{
				Id: &testapi.ProvisionState_Id{
					Value: "Hello, World!",
				},
				PreventReboot: true,
				Packages:      []*testapi.ProvisionState_Package{},
			},
		}
		new_packages := []*testapi.ProvisionState_Package{
			{
				PackagePath: &_go.StoragePath{
					Path: "d",
				},
			},
			{
				PackagePath: &_go.StoragePath{
					Path: "e",
				},
			},
			{
				PackagePath: &_go.StoragePath{
					Path: "f",
				},
			},
		}

		injection_map := map[string]interface{}{}
		injection_map["packages"] = new_packages

		err := common.Inject(original_proto, "provisionState.packages", injection_map, "packages")

		So(err, ShouldBeNil)
		So(original_proto.ProvisionState.Packages, ShouldHaveLength, 3)
		So(original_proto.ProvisionState.Packages[0].PackagePath.Path, ShouldEqual, new_packages[0].PackagePath.Path)
		So(original_proto.ProvisionState.Packages[1].PackagePath.Path, ShouldEqual, new_packages[1].PackagePath.Path)
		So(original_proto.ProvisionState.Packages[2].PackagePath.Path, ShouldEqual, new_packages[2].PackagePath.Path)
	})
}
