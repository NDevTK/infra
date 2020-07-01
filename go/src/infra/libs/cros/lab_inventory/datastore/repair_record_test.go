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
		Hostname: hostname,
		AssetTag: assetTag,
		RepairTargetType: 0,
		RepairState: 0,
		BuganizerBugUrl: "https://b/12345678",
		ChromiumBugUrl: "https://crbug.com/12345678",
		DutRepairFailureDescription: "Mock DUT repair failure description.",
		DutVerifierFailureDescription: "Mock DUT verifier failure description.",
		ServoRepairFailureDescription: "Mock Servo repair failure description.",
		ServoVerifierFailureDescription: "Mock Servo verifier failure description.",
		Diagnosis: "Mock diagnosis.",
		RepairProcedure: "Mock repair procedure.",
		ManualRepairActions: []inv.DeviceManualRepairRecord_ManualRepairAction{1,2,3,6},
		TimeTaken: 15,
		CreatedTime: &timestamp.Timestamp{Seconds: 111, Nanos: 0},
		UpdatedTime: &timestamp.Timestamp{Seconds: 222, Nanos: 0},
		CompletedTime: &timestamp.Timestamp{Seconds: 222, Nanos: 0},
	}
}

func TestAddRecord(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	record1 := mockDeviceManualRepairRecord("chromeosxx-rowxx-rackxx-hostxx", "xxxxxxxx")
	record2 := mockDeviceManualRepairRecord("chromeosyy-rowyy-rackyy-hostyy", "yyyyyyyy")
	record3 := mockDeviceManualRepairRecord("chromeoszz-rowzz-rackzz-hostzz", "zzzzzzzz")
	record4 := mockDeviceManualRepairRecord("chromeosaa-rowaa-rackaa-hostaa", "aaaaaaaa")
	// record4 := mockDeviceManualRepairRecord("", "")
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

			// Retrieve records from datastore
			reqGet := []string{record1.GetHostname(), record2.GetHostname()}
			res = GetDeviceManualRepairRecords(ctx, reqGet)
			So(res, ShouldHaveLength, 2)
			for i, r := range records {
				So(res[i].Err, ShouldBeNil)
				So(res[i].Entity.Hostname, ShouldEqual, r.GetHostname())
			}
		})
		Convey("Add existing record to datastore", func() {
			req := []*inv.DeviceManualRepairRecord{record3}
			res, err := AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldBeNil)

			// Verify adding existing record
			req2 := []*inv.DeviceManualRepairRecord{record3, record4}
			res2, err := AddDeviceManualRepairRecords(ctx, req2)
			So(err, ShouldBeNil)
			So(res2, ShouldNotBeNil)
			So(res2, ShouldHaveLength, 2)
			So(res2[0].Err, ShouldNotBeNil)
			So(res2[0].Err.Error(), ShouldContainSubstring, "Record exists in the datastore")
			So(res2[1].Err, ShouldBeNil)
			So(res2[1].Entity.Hostname, ShouldEqual, record3.GetHostname())
		})
	})
}
