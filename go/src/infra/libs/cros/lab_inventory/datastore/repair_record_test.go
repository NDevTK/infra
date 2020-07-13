package datastore

import (
	"testing"

	"github.com/golang/protobuf/ptypes"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"

	. "github.com/smartystreets/goconvey/convey"
	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/appengine/gaetesting"

	inv "infra/appengine/cros/lab_inventory/api/v1"
)

func mockDeviceManualRepairRecord(hostname string, assetTag string, createdTime int64) *inv.DeviceManualRepairRecord {
	return &inv.DeviceManualRepairRecord{
		Hostname:                        hostname,
		AssetTag:                        assetTag,
		RepairTargetType:                inv.DeviceManualRepairRecord_TYPE_DUT,
		RepairState:                     inv.DeviceManualRepairRecord_STATE_NOT_STARTED,
		BuganizerBugUrl:                 "https://b/12345678",
		ChromiumBugUrl:                  "https://crbug.com/12345678",
		DutRepairFailureDescription:     "Mock DUT repair failure description.",
		DutVerifierFailureDescription:   "Mock DUT verifier failure description.",
		ServoRepairFailureDescription:   "Mock Servo repair failure description.",
		ServoVerifierFailureDescription: "Mock Servo verifier failure description.",
		Diagnosis:                       "Mock diagnosis.",
		RepairProcedure:                 "Mock repair procedure.",
		ManualRepairActions: []inv.DeviceManualRepairRecord_ManualRepairAction{
			inv.DeviceManualRepairRecord_ACTION_FIX_SERVO,
			inv.DeviceManualRepairRecord_ACTION_FIX_YOSHI_CABLE,
			inv.DeviceManualRepairRecord_ACTION_VISUAL_INSPECTION,
			inv.DeviceManualRepairRecord_ACTION_REIMAGE_DUT,
		},
		IssueFixed:    true,
		UserLdap:      "testing-account",
		TimeTaken:     15,
		CreatedTime:   &timestamp.Timestamp{Seconds: createdTime, Nanos: 0},
		UpdatedTime:   &timestamp.Timestamp{Seconds: 222, Nanos: 0},
		CompletedTime: &timestamp.Timestamp{Seconds: 222, Nanos: 0},
	}
}

func mockUpdatedRecord(hostname string, assetTag string, createdTime int64) *inv.DeviceManualRepairRecord {
	return &inv.DeviceManualRepairRecord{
		Hostname:                        hostname,
		AssetTag:                        assetTag,
		RepairTargetType:                inv.DeviceManualRepairRecord_TYPE_DUT,
		RepairState:                     inv.DeviceManualRepairRecord_STATE_COMPLETED,
		BuganizerBugUrl:                 "https://b/12345678",
		ChromiumBugUrl:                  "https://crbug.com/12345678",
		DutRepairFailureDescription:     "Mock DUT repair failure description.",
		DutVerifierFailureDescription:   "Mock DUT verifier failure description.",
		ServoRepairFailureDescription:   "Mock Servo repair failure description.",
		ServoVerifierFailureDescription: "Mock Servo verifier failure description.",
		Diagnosis:                       "Mock diagnosis.",
		RepairProcedure:                 "Mock repair procedure.",
		ManualRepairActions: []inv.DeviceManualRepairRecord_ManualRepairAction{
			inv.DeviceManualRepairRecord_ACTION_FIX_SERVO,
			inv.DeviceManualRepairRecord_ACTION_FIX_YOSHI_CABLE,
			inv.DeviceManualRepairRecord_ACTION_VISUAL_INSPECTION,
			inv.DeviceManualRepairRecord_ACTION_REIMAGE_DUT,
		},
		IssueFixed:    true,
		UserLdap:      "testing-account",
		TimeTaken:     30,
		CreatedTime:   &timestamp.Timestamp{Seconds: createdTime, Nanos: 0},
		UpdatedTime:   &timestamp.Timestamp{Seconds: 222, Nanos: 0},
		CompletedTime: &timestamp.Timestamp{Seconds: 222, Nanos: 0},
	}
}

func TestAddRecord(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	record1 := mockDeviceManualRepairRecord("chromeosxx-rowxx-rackxx-hostxx", "xxxxxxxx", 1)
	record2 := mockDeviceManualRepairRecord("chromeosyy-rowyy-rackyy-hostyy", "yyyyyyyy", 1)
	record3 := mockDeviceManualRepairRecord("chromeoszz-rowzz-rackzz-hostzz", "zzzzzzzz", 1)
	record4 := mockDeviceManualRepairRecord("chromeosaa-rowaa-rackaa-hostaa", "aaaaaaaa", 1)
	record5 := mockDeviceManualRepairRecord("", "", 1)

	rec1ID, _ := generateRepairRecordID(record1.Hostname, record1.AssetTag, ptypes.TimestampString(record1.CreatedTime))
	rec2ID, _ := generateRepairRecordID(record2.Hostname, record2.AssetTag, ptypes.TimestampString(record2.CreatedTime))
	ids1 := []string{rec1ID, rec2ID}

	rec3ID, _ := generateRepairRecordID(record3.Hostname, record3.AssetTag, ptypes.TimestampString(record3.CreatedTime))
	rec4ID, _ := generateRepairRecordID(record4.Hostname, record4.AssetTag, ptypes.TimestampString(record4.CreatedTime))
	ids2 := []string{rec3ID, rec4ID}
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

			res = GetDeviceManualRepairRecords(ctx, ids1)
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

			// Verify adding existing record.
			req = []*inv.DeviceManualRepairRecord{record3, record4}
			res, err = AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res, ShouldHaveLength, 2)
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "Record exists in the datastore")
			So(res[1].Err, ShouldBeNil)
			So(res[1].Entity.Hostname, ShouldEqual, record4.GetHostname())

			// Check both records are in datastore.
			res = GetDeviceManualRepairRecords(ctx, ids2)
			So(res, ShouldHaveLength, 2)
			for i, r := range req {
				So(res[i].Err, ShouldBeNil)
				So(res[i].Entity.Hostname, ShouldEqual, r.GetHostname())
				So(res[i].Entity.AssetTag, ShouldEqual, r.GetAssetTag())
				So(res[i].Entity.RepairState, ShouldEqual, r.GetRepairState().String())
			}
		})
		Convey("Add record without hostname to datastore", func() {
			req := []*inv.DeviceManualRepairRecord{record5}
			res, err := AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "Hostname cannot be empty")
		})
	})
}

func TestGetRecord(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	record1 := mockDeviceManualRepairRecord("chromeosyy-rowyy-rackyy-hostyy", "yyyyyyyy", 1)
	record2 := mockDeviceManualRepairRecord("chromeoszz-rowzz-rackzz-hostzz", "12345678", 1)
	rec1ID, _ := generateRepairRecordID(record1.Hostname, record1.AssetTag, ptypes.TimestampString(record1.CreatedTime))
	rec2ID, _ := generateRepairRecordID(record2.Hostname, record2.AssetTag, ptypes.TimestampString(record2.CreatedTime))
	ids1 := []string{rec1ID, rec2ID}
	Convey("Get device manual repair record from datastore", t, func() {
		Convey("Get non-existent device manual repair record from datastore", func() {
			records := []*inv.DeviceManualRepairRecord{record1}
			res, err := AddDeviceManualRepairRecords(ctx, records)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)

			res = GetDeviceManualRepairRecords(ctx, ids1)
			So(res, ShouldHaveLength, 2)
			So(res[0].Err, ShouldBeNil)
			So(res[1].Err, ShouldNotBeNil)
			So(res[1].Err.Error(), ShouldContainSubstring, "datastore: no such entity")
		})
		Convey("Get record with empty id", func() {
			res := GetDeviceManualRepairRecords(ctx, []string{""})
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "datastore: invalid key")
		})
	})
}

func TestGetRecordByPropertyName(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	datastore.GetTestable(ctx).Consistent(true)

	record1 := mockDeviceManualRepairRecord("chromeosyy-rowyy-rackyy-hostyy", "yyyyyyyy", 1)
	record2 := mockUpdatedRecord("chromeosyy-rowyy-rackyy-hostyy", "yyyyyyyy", 2)
	record3 := mockDeviceManualRepairRecord("chromeosyy-rowyy-rackyy-hostyy", "xxxxxxxx", 1)
	rec1ID, _ := generateRepairRecordID(record1.Hostname, record1.AssetTag, ptypes.TimestampString(record1.CreatedTime))
	rec2ID, _ := generateRepairRecordID(record2.Hostname, record2.AssetTag, ptypes.TimestampString(record2.CreatedTime))
	rec3ID, _ := generateRepairRecordID(record3.Hostname, record3.AssetTag, ptypes.TimestampString(record3.CreatedTime))
	ids := []string{rec1ID, rec2ID, rec3ID}
	records := []*inv.DeviceManualRepairRecord{record1, record2, record3}

	// Set up records in datastore and test
	resAdd, err := AddDeviceManualRepairRecords(ctx, records)

	Convey("Get device manual repair record from datastore by property name", t, func() {
		So(err, ShouldBeNil)
		So(resAdd, ShouldHaveLength, 3)
		for i, r := range records {
			So(resAdd[i].Err, ShouldBeNil)
			So(resAdd[i].Entity.Hostname, ShouldEqual, r.GetHostname())
			So(resAdd[i].Entity.AssetTag, ShouldEqual, r.GetAssetTag())
			So(resAdd[i].Entity.RepairState, ShouldEqual, r.GetRepairState().String())
		}
		So(resAdd[0].Entity.ID, ShouldNotEqual, resAdd[1].Entity.ID)
		So(resAdd[1].Entity.ID, ShouldNotEqual, resAdd[2].Entity.ID)

		resGet := GetDeviceManualRepairRecords(ctx, ids)
		So(resGet, ShouldHaveLength, 3)
		for i, r := range records {
			So(resGet[i].Err, ShouldBeNil)
			So(resGet[i].Entity.Hostname, ShouldEqual, r.GetHostname())
			So(resGet[i].Entity.AssetTag, ShouldEqual, r.GetAssetTag())
			So(resGet[i].Entity.RepairState, ShouldEqual, r.GetRepairState().String())
		}
		So(resGet[0].Entity.ID, ShouldNotEqual, resGet[1].Entity.ID)
		So(resGet[1].Entity.ID, ShouldNotEqual, resGet[2].Entity.ID)

		// Named property tests
		Convey("Get repair record by Hostname", func() {
			recs1 := []*inv.DeviceManualRepairRecord{record1, record2, record3}
			res, err := GetRepairRecordByPropertyName(ctx, "hostname", record1.GetHostname())
			So(res, ShouldHaveLength, 3)
			So(err, ShouldBeNil)
			for i, r := range recs1 {
				So(res[i].Err, ShouldBeNil)
				So(res[i].Entity.Hostname, ShouldEqual, r.GetHostname())
			}
			So(res[0].Entity.ID, ShouldNotEqual, res[1].Entity.ID)
			So(res[1].Entity.ID, ShouldNotEqual, res[2].Entity.ID)
			So(res[0].Entity.ID, ShouldNotEqual, res[2].Entity.ID)
		})
		Convey("Get repair record by AssetTag", func() {
			recs2 := []*inv.DeviceManualRepairRecord{record1, record2}
			res, err := GetRepairRecordByPropertyName(ctx, "asset_tag", record1.GetAssetTag())
			So(res, ShouldHaveLength, 2)
			So(err, ShouldBeNil)
			for i, r := range recs2 {
				So(res[i].Err, ShouldBeNil)
				So(res[i].Entity.AssetTag, ShouldEqual, r.GetAssetTag())
			}
			So(res[0].Entity.ID, ShouldNotEqual, res[1].Entity.ID)
		})
		Convey("Get repair record by RepairState", func() {
			res, err := GetRepairRecordByPropertyName(ctx, "repair_state", record2.GetRepairState().String())
			So(res, ShouldHaveLength, 1)
			So(err, ShouldBeNil)
			So(res[0].Err, ShouldBeNil)
			So(res[0].Entity.Hostname, ShouldEqual, record2.GetHostname())
			So(res[0].Entity.AssetTag, ShouldEqual, record2.GetAssetTag())
			So(res[0].Entity.RepairState, ShouldEqual, record2.GetRepairState().String())
		})
	})
}

func TestUpdateRecord(t *testing.T) {
	t.Parallel()
	ctx := gaetesting.TestingContextWithAppID("go-test")
	record1 := mockDeviceManualRepairRecord("chromeosxx-rowxx-rackxx-hostxx", "xxxxxxxx", 1)
	record1Update := mockUpdatedRecord("chromeosxx-rowxx-rackxx-hostxx", "xxxxxxxx", 1)

	record2 := mockDeviceManualRepairRecord("chromeosyy-rowyy-rackyy-hostyy", "yyyyyyyy", 1)

	record3 := mockDeviceManualRepairRecord("chromeoszz-rowzz-rackzz-hostzz", "zzzzzzzz", 1)
	record3Update := mockUpdatedRecord("chromeoszz-rowzz-rackzz-hostzz", "zzzzzzzz", 1)
	record4 := mockDeviceManualRepairRecord("chromeosaa-rowaa-rackaa-hostaa", "aaaaaaaa", 1)

	record5 := mockDeviceManualRepairRecord("", "", 1)
	Convey("Update record in datastore", t, func() {
		Convey("Update existing record to datastore", func() {
			rec1ID, _ := generateRepairRecordID(record1.Hostname, record1.AssetTag, ptypes.TimestampString(record1.CreatedTime))
			req := []*inv.DeviceManualRepairRecord{record1}
			res, err := AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldBeNil)

			res = GetDeviceManualRepairRecords(ctx, []string{rec1ID})
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldBeNil)
			So(res[0].Entity.RepairState, ShouldEqual, record1.GetRepairState().String())

			// Update and check
			reqUpdate := map[string]*inv.DeviceManualRepairRecord{rec1ID: record1Update}
			res2, err := UpdateDeviceManualRepairRecords(ctx, reqUpdate)
			So(err, ShouldBeNil)
			So(res2, ShouldHaveLength, 1)
			So(res2[0].Err, ShouldBeNil)

			res = GetDeviceManualRepairRecords(ctx, []string{rec1ID})
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldBeNil)
			So(res[0].Entity.RepairState, ShouldEqual, record1Update.GetRepairState().String())
		})
		Convey("Update non-existent record in datastore", func() {
			rec2ID, _ := generateRepairRecordID(record2.Hostname, record2.AssetTag, ptypes.TimestampString(record2.CreatedTime))
			reqUpdate := map[string]*inv.DeviceManualRepairRecord{rec2ID: record2}
			res, err := UpdateDeviceManualRepairRecords(ctx, reqUpdate)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "datastore: no such entity")

			res = GetDeviceManualRepairRecords(ctx, []string{rec2ID})
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldNotBeNil)
		})
		Convey("Update multiple records to datastore", func() {
			rec3ID, _ := generateRepairRecordID(record3.Hostname, record3.AssetTag, ptypes.TimestampString(record3.CreatedTime))
			rec4ID, _ := generateRepairRecordID(record4.Hostname, record4.AssetTag, ptypes.TimestampString(record4.CreatedTime))
			req := []*inv.DeviceManualRepairRecord{record3, record4}
			res, err := AddDeviceManualRepairRecords(ctx, req)
			So(err, ShouldBeNil)
			So(res, ShouldNotBeNil)
			So(res, ShouldHaveLength, 2)
			So(res[0].Err, ShouldBeNil)
			So(res[1].Err, ShouldBeNil)

			reqUpdate := map[string]*inv.DeviceManualRepairRecord{
				rec3ID: record3Update,
				rec4ID: record4,
			}
			res, err = UpdateDeviceManualRepairRecords(ctx, reqUpdate)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 2)
			So(res[0].Err, ShouldBeNil)
			So(res[1].Err, ShouldBeNil)

			res = GetDeviceManualRepairRecords(ctx, []string{rec3ID, rec4ID})
			So(res, ShouldHaveLength, 2)
			So(res[0].Err, ShouldBeNil)
			So(res[1].Err, ShouldBeNil)
			So(res[0].Entity.RepairState, ShouldEqual, record3Update.GetRepairState().String())
			So(res[1].Entity.RepairState, ShouldEqual, record4.GetRepairState().String())
		})
		Convey("Update record without ID to datastore", func() {
			rec5ID, _ := generateRepairRecordID(record5.Hostname, record5.AssetTag, ptypes.TimestampString(record5.CreatedTime))
			reqUpdate := map[string]*inv.DeviceManualRepairRecord{rec5ID: record5}
			res, err := UpdateDeviceManualRepairRecords(ctx, reqUpdate)
			So(err, ShouldBeNil)
			So(res, ShouldHaveLength, 1)

			// Error should occur when trying to get old entity from datastore
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "datastore: no such entity")

			res = GetDeviceManualRepairRecords(ctx, []string{rec5ID})
			So(res, ShouldHaveLength, 1)
			So(res[0].Err, ShouldNotBeNil)
			So(res[0].Err.Error(), ShouldContainSubstring, "datastore: no such entity")
		})
	})
}
