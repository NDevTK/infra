// Copyright 2020 The Chromium OS Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package controller

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"

	ufspb "infra/unifiedfleet/api/v1/proto"
	"infra/unifiedfleet/app/model/configuration"
	. "infra/unifiedfleet/app/model/datastore"
	"infra/unifiedfleet/app/model/registration"
)

func mockKVM(id string) *ufspb.KVM {
	return &ufspb.KVM{
		Name: id,
	}
}

func TestCreateKVM(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	rack1 := &ufspb.Rack{
		Name: "rack-1",
		Rack: &ufspb.Rack_ChromeBrowserRack{
			ChromeBrowserRack: &ufspb.ChromeBrowserRack{},
		},
	}
	registration.CreateRack(ctx, rack1)
	Convey("CreateKVM", t, func() {
		Convey("Create new kvm with already existing kvm - error", func() {
			kvm1 := &ufspb.KVM{
				Name: "kvm-1",
			}
			_, err := registration.CreateKVM(ctx, kvm1)

			resp, err := CreateKVM(ctx, kvm1, "rack-5")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "KVM kvm-1 already exists in the system")
		})

		Convey("Create new kvm with non existing chromePlatform", func() {
			kvm2 := &ufspb.KVM{
				Name:           "kvm-2",
				ChromePlatform: "chromePlatform-1",
			}
			resp, err := CreateKVM(ctx, kvm2, "rack-1")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "There is no ChromePlatform with ChromePlatformID chromePlatform-1 in the system")
		})

		Convey("Create new kvm with existing resources", func() {
			chromePlatform2 := &ufspb.ChromePlatform{
				Name: "chromePlatform-2",
			}
			_, err := configuration.CreateChromePlatform(ctx, chromePlatform2)
			So(err, ShouldBeNil)

			kvm2 := &ufspb.KVM{
				Name:           "kvm-2",
				ChromePlatform: "chromePlatform-2",
			}
			resp, err := CreateKVM(ctx, kvm2, "rack-1")
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, kvm2)
		})

		Convey("Create new kvm with existing rack with kvms", func() {
			rack := &ufspb.Rack{
				Name: "rack-10",
				Rack: &ufspb.Rack_ChromeBrowserRack{
					ChromeBrowserRack: &ufspb.ChromeBrowserRack{
						Kvms: []string{"kvm-5"},
					},
				},
			}
			_, err := registration.CreateRack(ctx, rack)
			So(err, ShouldBeNil)

			kvm1 := &ufspb.KVM{
				Name: "kvm-20",
			}
			resp, err := CreateKVM(ctx, kvm1, "rack-10")
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, kvm1)

			mresp, err := registration.GetRack(ctx, "rack-10")
			So(err, ShouldBeNil)
			So(mresp.GetChromeBrowserRack().GetKvms(), ShouldResemble, []string{"kvm-5", "kvm-20"})
		})
	})
}

func TestUpdateKVM(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("UpdateKVM", t, func() {
		Convey("Update kvm with non-existing kvm", func() {
			rack1 := &ufspb.Rack{
				Name: "rack-1",
			}
			_, err := registration.CreateRack(ctx, rack1)
			So(err, ShouldBeNil)

			kvm1 := &ufspb.KVM{
				Name: "kvm-1",
			}
			resp, err := UpdateKVM(ctx, kvm1, "rack-1")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "There is no KVM with KVMID kvm-1 in the system")
		})

		Convey("Update kvm with new rack", func() {
			rack3 := &ufspb.Rack{
				Name: "rack-3",
				Rack: &ufspb.Rack_ChromeBrowserRack{
					ChromeBrowserRack: &ufspb.ChromeBrowserRack{
						Kvms: []string{"kvm-3"},
					},
				},
			}
			_, err := registration.CreateRack(ctx, rack3)
			So(err, ShouldBeNil)

			rack4 := &ufspb.Rack{
				Name: "rack-4",
				Rack: &ufspb.Rack_ChromeBrowserRack{
					ChromeBrowserRack: &ufspb.ChromeBrowserRack{
						Kvms: []string{"kvm-4"},
					},
				},
			}
			_, err = registration.CreateRack(ctx, rack4)
			So(err, ShouldBeNil)

			kvm3 := &ufspb.KVM{
				Name: "kvm-3",
			}
			_, err = registration.CreateKVM(ctx, kvm3)
			So(err, ShouldBeNil)

			resp, err := UpdateKVM(ctx, kvm3, "rack-4")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, kvm3)

			mresp, err := registration.GetRack(ctx, "rack-3")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(mresp.GetChromeBrowserRack().GetKvms(), ShouldBeNil)

			mresp, err = registration.GetRack(ctx, "rack-4")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(mresp.GetChromeBrowserRack().GetKvms(), ShouldResemble, []string{"kvm-4", "kvm-3"})
		})

		Convey("Update kvm with same rack", func() {
			rack := &ufspb.Rack{
				Name: "rack-5",
				Rack: &ufspb.Rack_ChromeBrowserRack{
					ChromeBrowserRack: &ufspb.ChromeBrowserRack{
						Kvms: []string{"kvm-5"},
					},
				},
			}
			_, err := registration.CreateRack(ctx, rack)
			So(err, ShouldBeNil)

			kvm1 := &ufspb.KVM{
				Name: "kvm-5",
			}
			_, err = registration.CreateKVM(ctx, kvm1)
			So(err, ShouldBeNil)

			resp, err := UpdateKVM(ctx, kvm1, "rack-5")
			So(err, ShouldBeNil)
			So(resp, ShouldNotBeNil)
			So(resp, ShouldResembleProto, kvm1)
		})

		Convey("Update kvm with non existing rack", func() {
			kvm1 := &ufspb.KVM{
				Name: "kvm-6",
			}
			_, err := registration.CreateKVM(ctx, kvm1)
			So(err, ShouldBeNil)

			resp, err := UpdateKVM(ctx, kvm1, "rack-6")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "There is no Rack with RackID rack-6 in the system.")
		})

	})
}

func TestDeleteKVM(t *testing.T) {
	t.Parallel()
	ctx := testingContext()
	Convey("DeleteKVM", t, func() {
		Convey("Delete KVM by existing ID with machine reference", func() {
			KVM1 := &ufspb.KVM{
				Name: "KVM-1",
			}
			_, err := registration.CreateKVM(ctx, KVM1)
			So(err, ShouldBeNil)

			chromeBrowserMachine1 := &ufspb.Machine{
				Name: "machine-1",
				Device: &ufspb.Machine_ChromeBrowserMachine{
					ChromeBrowserMachine: &ufspb.ChromeBrowserMachine{
						KvmInterface: &ufspb.KVMInterface{
							Kvm: "KVM-1",
						},
					},
				},
			}
			_, err = registration.CreateMachine(ctx, chromeBrowserMachine1)
			So(err, ShouldBeNil)

			err = DeleteKVM(ctx, "KVM-1")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, CannotDelete)

			resp, err := registration.GetKVM(ctx, "KVM-1")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, KVM1)
		})
		Convey("Delete KVM by existing ID with rack reference", func() {
			KVM2 := &ufspb.KVM{
				Name: "KVM-2",
			}
			_, err := registration.CreateKVM(ctx, KVM2)
			So(err, ShouldBeNil)

			chromeBrowserRack1 := &ufspb.Rack{
				Name: "rack-1",
				Rack: &ufspb.Rack_ChromeBrowserRack{
					ChromeBrowserRack: &ufspb.ChromeBrowserRack{
						Kvms: []string{"KVM-2", "KVM-5"},
					},
				},
			}
			_, err = registration.CreateRack(ctx, chromeBrowserRack1)
			So(err, ShouldBeNil)

			err = DeleteKVM(ctx, "KVM-2")
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, CannotDelete)

			resp, err := registration.GetKVM(ctx, "KVM-2")
			So(resp, ShouldNotBeNil)
			So(err, ShouldBeNil)
			So(resp, ShouldResembleProto, KVM2)
		})

		Convey("Delete KVM successfully by existing ID without references", func() {
			KVM4 := &ufspb.KVM{
				Name: "KVM-4",
			}
			_, err := registration.CreateKVM(ctx, KVM4)
			So(err, ShouldBeNil)

			err = DeleteKVM(ctx, "KVM-4")
			So(err, ShouldBeNil)

			resp, err := registration.GetKVM(ctx, "KVM-4")
			So(resp, ShouldBeNil)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, NotFound)
		})
	})
}
