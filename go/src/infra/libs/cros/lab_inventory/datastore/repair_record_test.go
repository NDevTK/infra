package datastore

import (
	"testing"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/luci/appengine/gaetesting"

	inv "infra/appengine/cros/lab_inventory/api/v1"
)

func mockDeviceManualRepairRecord(hostname string, assetTag string) *inv.DeviceManualRepairRecord {
	return &inv.DeviceManualRepairRecord{
		Hostname:                        hostname,
		AssetTag:                        assetTag,
		RepairTargetType:                0,
		RepairState:                     0,
		BuganizerBugUrl:                 "https://b/12345678",
		ChromiumBugUrl:                  "https://crbug.com/12345678",
		DutRepairFailureDescription:     "Mock DUT repair failure description.",
		DutVerifierFailureDescription:   "Mock DUT verifier failure description.",
		ServoRepairFailureDescription:   "Mock Servo repair failure description.",
		ServoVerifierFailureDescription: "Mock Servo verifier failure description.",
		Diagnosis:                       "Mock diagnosis.",
		RepairProcedure:                 "Mock repair procedure.",
		ManualRepairActions:             []inv.DeviceManualRepairRecord_ManualRepairAction{1, 2, 3, 6},
		TimeTaken:                       15,
		CreatedTime:                     &timestamp.Timestamp{Seconds: 111, Nanos: 0},
		UpdatedTime:                     &timestamp.Timestamp{Seconds: 222, Nanos: 0},
		CompletedTime:                   &timestamp.Timestamp{Seconds: 222, Nanos: 0},
	}
}

func mockUpdatedRecord(hostname string, assetTag string) *inv.DeviceManualRepairRecord {
	return &inv.DeviceManualRepairRecord{
		Hostname:                        hostname,
		AssetTag:                        assetTag,
		RepairTargetType:                0,
		RepairState:                     1,
		BuganizerBugUrl:                 "https://b/12345678",
		ChromiumBugUrl:                  "https://crbug.com/12345678",
		DutRepairFailureDescription:     "Mock DUT repair failure description.",
		DutVerifierFailureDescription:   "Mock DUT verifier failure description.",
		ServoRepairFailureDescription:   "Mock Servo repair failure description.",
		ServoVerifierFailureDescription: "Mock Servo verifier failure description.",
		Diagnosis:                       "Mock diagnosis.",
		RepairProcedure:                 "Mock repair procedure.",
		ManualRepairActions:             []inv.DeviceManualRepairRecord_ManualRepairAction{1, 2, 3, 6},
		TimeTaken:                       30,
		CreatedTime:                     &timestamp.Timestamp{Seconds: 111, Nanos: 0},
		UpdatedTime:                     &timestamp.Timestamp{Seconds: 222, Nanos: 0},
		CompletedTime:                   &timestamp.Timestamp{Seconds: 222, Nanos: 0},
	}
}

func TestAddRecord(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	record1 := mockDeviceManualRepairRecord("chromeosxx-rowxx-rackxx-hostxx", "xxxxxxxx")
	record2 := mockDeviceManualRepairRecord("chromeosyy-rowyy-rackyy-hostyy", "yyyyyyyy")
	record3 := mockDeviceManualRepairRecord("chromeoszz-rowzz-rackzz-hostzz", "zzzzzzzz")
	record4 := mockDeviceManualRepairRecord("chromeosaa-rowaa-rackaa-hostaa", "aaaaaaaa")
	record5 := mockDeviceManualRepairRecord("", "")
	Convey("Add device manual repair record to datastore", t, func() {
		Convey("Add multiple device manual repair records to datastore", func() {
			records := []*inv.DeviceManualRepairRecord{record1, record2}
			res, err := AddDeviceManualRepairRecords(ctx, records)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 2)
			for i, r := range records {
				So(res[i].Err, ShouldBeNil)
				So(res[i].Entity.Hostname, ShouldEqual, r.GetHostname())
				So(res[i].Entity.AssetTag, ShouldEqual, r.GetAssetTag())
				So(res[i].Entity.RepairState, ShouldEqual, r.GetRepairState().String())
			}
		})
		Convey("Add existing record to datastore", func() {
			req := []*inv.DeviceManualRepairRecord{record3}
			res, err := AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldBeNil)

			// Verify adding existing record
			req = []*inv.DeviceManualRepairRecord{record3, record4}
			res, err = AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res, ShouldHaveLength, 2)
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "Record exists in the datastore")
			So(res[1].Err, ShouldBeNil)
			So(res[1].Entity.Hostname, ShouldEqual, record4.GetHostname())
		})
		Convey("Add record without hostname to datastore", func() {
			req := []*inv.DeviceManualRepairRecord{record5}
			res, err := AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "Missing hostname")
		})
	})
}

func TestGetRecord(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	record1 := mockDeviceManualRepairRecord("chromeosxx-rowxx-rackxx-hostxx", "xxxxxxxx")
	record2 := mockDeviceManualRepairRecord("chromeosyy-rowyy-rackyy-hostyy", "yyyyyyyy")
	ids1 := []string{
		record1.Hostname + "-" + record1.AssetTag,
		record2.Hostname + "-" + record2.AssetTag,
	}

	record3 := mockDeviceManualRepairRecord("chromeoszz-rowzz-rackzz-hostzz", "12345678")
	ids2 := []string{
		record2.Hostname + "-" + record2.AssetTag,
		record3.Hostname + "-" + record3.AssetTag,
	}
	Convey("Get device manual repair record from datastore", t, func() {
		Convey("Get multiple device manual repair records from datastore", func() {
			records := []*inv.DeviceManualRepairRecord{record1, record2}
			res, err := AddDeviceManualRepairRecords(ctx, records)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 2)
			for i, r := range records {
				So(res[i].Err, ShouldBeNil)
				So(res[i].Entity.ID, ShouldEqual, r.GetHostname()+"-"+r.GetAssetTag())
			}

			res = GetDeviceManualRepairRecords(ctx, ids1)
			So(res, ShouldHaveLength, 2)
			So(res[0].Err, ShouldBeNil)
			So(res[1].Err, ShouldBeNil)
		})
		Convey("Get non-existent device manual repair record from datastore", func() {
			res := GetDeviceManualRepairRecords(ctx, ids2)
			So(res, ShouldHaveLength, 2)
			So(res[0].Err, ShouldBeNil)
			So(res[1].Err, ShouldNotBeNil)
		})
		Convey("Get record with empty id", func() {
			res := GetDeviceManualRepairRecords(ctx, []string{""})
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldNotBeNil)
		})
	})
}

func TestUpdateRecord(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	record1 := mockDeviceManualRepairRecord("chromeosxx-rowxx-rackxx-hostxx", "xxxxxxxx")
	record1Update := mockUpdatedRecord("chromeosxx-rowxx-rackxx-hostxx", "xxxxxxxx")

	record2 := mockDeviceManualRepairRecord("chromeosyy-rowyy-rackyy-hostyy", "yyyyyyyy")

	record3 := mockDeviceManualRepairRecord("chromeoszz-rowzz-rackzz-hostzz", "zzzzzzzz")
	record3Update := mockUpdatedRecord("chromeoszz-rowzz-rackzz-hostzz", "zzzzzzzz")
	record4 := mockDeviceManualRepairRecord("chromeosaa-rowaa-rackaa-hostaa", "aaaaaaaa")

	record5 := mockDeviceManualRepairRecord("", "")
	Convey("Update record in datastore", t, func() {
		Convey("Update existing record to datastore", func() {
			req := []*inv.DeviceManualRepairRecord{record1}
			res, err := AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldBeNil)
			So(res[0].Record.GetRepairState(), ShouldEqual, record1.GetRepairState())

			reqUpdate := map[string]*inv.DeviceManualRepairRecord{
				record1.Hostname + "-" + record1.AssetTag: record1Update,
			}
			res2, err := UpdateDeviceManualRepairRecords(ctx, reqUpdate)
			So(err, ShouldBeNil)
			So(res2, ShouldHaveLength, 1)
			So(res2[0].Err, ShouldBeNil)
			So(res2[0].Record.GetRepairState(), ShouldEqual, record1Update.GetRepairState())
		})
		Convey("Update non-existent record in datastore", func() {
			reqUpdate := map[string]*inv.DeviceManualRepairRecord{
				record2.Hostname + "-" + record2.AssetTag: record2,
			}
			res, err := UpdateDeviceManualRepairRecords(ctx, reqUpdate)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "datastore: no such entity")
		})
		Convey("Update multiple records to datastore", func() {
			req := []*inv.DeviceManualRepairRecord{record3, record4}
			res, err := AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res, ShouldHaveLength, 2)
			So(res[0].Err, ShouldBeNil)
			So(res[0].Record.GetRepairState(), ShouldEqual, record3.GetRepairState())
			So(res[1].Err, ShouldBeNil)
			So(res[1].Record.GetRepairState(), ShouldEqual, record4.GetRepairState())

			reqUpdate := map[string]*inv.DeviceManualRepairRecord{
				record3.Hostname + "-" + record3.AssetTag: record3Update,
				record4.Hostname + "-" + record4.AssetTag: record4,
			}
			res2, err := UpdateDeviceManualRepairRecords(ctx, reqUpdate)
			So(err, ShouldBeNil)
			So(res2, ShouldHaveLength, 2)
			So(res2[0].Err, ShouldBeNil)
			So(res2[0].Record.GetRepairState(), ShouldEqual, record3Update.GetRepairState())
			So(res2[1].Err, ShouldBeNil)
			So(res2[1].Record.GetRepairState(), ShouldEqual, record4.GetRepairState())
		})
		Convey("Update record without ID to datastore", func() {
			reqUpdate := map[string]*inv.DeviceManualRepairRecord{
				record5.Hostname + "-" + record5.AssetTag: record5,
			}
			res, err := UpdateDeviceManualRepairRecords(ctx, reqUpdate)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "datastore: no such entity")
		})
	})
}
