package testresults

import (
	"context"
	"time"

	"go.chromium.org/luci/server/span"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

// June 17th, 2022 is a Friday. The preceding 5 * 24 weekday hour
// are as follows:
//             Inclusive              - Exclusive
// Interval 0: (-1 day) Thursday 8am  - (now)    Friday 8am
// Interval 1: (-2 day) Wednesday 8am - (-1 day) Thursday 8am
// Interval 2: (-3 day) Tuesday 8am   - (-2 day) Wednesday 8am
// Interval 3: (-4 day) Monday 8am    - (-3 day) Tuesday 8am
// Interval 4: (-7 day) Friday 8am    - (-4 day) Monday 8am
var referenceTime = time.Date(2022, time.June, 17, 8, 0, 0, 0, time.UTC)

// CreateQueryFailureRateTestData creates test data in Spanner for testing
// QueryFailureRate.
func CreateQueryFailureRateTestData(ctx context.Context) error {
	var1 := pbutil.Variant("key1", "val1", "key2", "val1")
	var2 := pbutil.Variant("key1", "val2", "key2", "val1")
	var3 := pbutil.Variant("key1", "val2", "key2", "val2")

	_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
		insertTV := func(partitionTime time.Time, variant *pb.Variant, invId string, runStatuses []RunStatus, run *PresubmitRun, changeListNumber ...int64) {
			baseTestResult := NewTestResult().
				WithProject("project").
				WithTestID("test_id").
				WithVariantHash(pbutil.VariantHash(variant)).
				WithPartitionTime(partitionTime).
				WithIngestedInvocationID(invId).
				WithSubRealm("realm").
				WithStatus(pb.TestResultStatus_PASS).
				WithPresubmitRun(run)

			var changelists []Changelist
			for _, clNum := range changeListNumber {
				changelists = append(changelists, Changelist{
					Host:     "mygerrit",
					Change:   clNum,
					Patchset: 5,
				})
			}
			baseTestResult = baseTestResult.WithChangelists(changelists)

			trs := NewTestVerdict().
				WithBaseTestResult(baseTestResult.Build()).
				WithRunStatus(runStatuses...).
				Build()
			for _, tr := range trs {
				span.BufferWrite(ctx, tr.SaveUnverified())
			}
		}

		// pass, fail is shorthand here for expected and unexpected run,
		// where for the purposes of this RPC, a flaky run counts as
		// "expected" (as it has at least one expected result).
		passFail := []RunStatus{Flaky, Unexpected}
		failPass := []RunStatus{Unexpected, Flaky}
		pass := []RunStatus{Flaky}
		fail := []RunStatus{Unexpected}
		failFail := []RunStatus{Unexpected, Unexpected}

		day := 24 * time.Hour

		userRun := &PresubmitRun{Owner: "user"}
		automationRun := &PresubmitRun{Owner: "automation"}

		insertTV(referenceTime.Add(-6*day), var1, "inv1", failPass, userRun, 10)
		// duplicate-cl result should not be used, inv3 result should be
		// used instead (as only one verdict per changelist is used, and
		// inv3 is more recent).
		insertTV(referenceTime.Add(-4*day), var1, "duplicate-cl", failPass, userRun, 1)
		// duplicate-cl2 result should not be used, inv3 result should be used instead
		// (as only one verdict per changelist is used, and inv3 is flaky
		// and this is not).
		insertTV(referenceTime.Add(-1*time.Hour), var1, "duplicate-cl2", pass, userRun, 1)

		insertTV(referenceTime.Add(-4*day), var1, "inv2", pass, nil, 2)
		insertTV(referenceTime.Add(-2*time.Hour), var1, "inv3", failPass, userRun, 1)

		insertTV(referenceTime.Add(-3*day), var1, "inv4", failPass, automationRun)
		insertTV(referenceTime.Add(-3*day), var1, "inv5", passFail, nil, 3)
		insertTV(referenceTime.Add(-2*day), var1, "inv6", fail, userRun, 4)
		insertTV(referenceTime.Add(-3*day), var1, "inv7", failFail, nil)
		// should not be used, as tests multiple CLs, and too hard
		// to deduplicate the verdicts.
		insertTV(referenceTime.Add(-2*day), var1, "many-cl", failPass, userRun, 1, 3)

		// should not be used, as times  fall outside the queried intervals.
		insertTV(referenceTime.Add(-7*day-time.Microsecond), var1, "too-early", failPass, userRun, 5)
		insertTV(referenceTime, var1, "too-late", failPass, userRun, 6)

		insertTV(referenceTime.Add(-4*day), var2, "inv1", failPass, userRun, 1)
		insertTV(referenceTime.Add(-3*day), var2, "inv2", failPass, userRun, 2)

		insertTV(referenceTime.Add(-5*day), var3, "duplicate-cl1", passFail, userRun, 1)
		insertTV(referenceTime.Add(-3*day), var3, "duplicate-cl2", failPass, userRun, 1)
		insertTV(referenceTime.Add(-1*day), var3, "inv8", failPass, userRun, 1)

		return nil
	})
	return err
}

func QueryFailureRateSampleRequest() (project string, asAtTime time.Time, testVariants []*pb.TestVariantIdentifier) {
	var1 := pbutil.Variant("key1", "val1", "key2", "val1")
	var3 := pbutil.Variant("key1", "val2", "key2", "val2")
	testVariants = []*pb.TestVariantIdentifier{
		{
			TestId:  "test_id",
			Variant: var1,
		},
		{
			TestId:  "test_id",
			Variant: var3,
		},
	}
	asAtTime = time.Date(2022, time.June, 17, 8, 0, 0, 0, time.UTC)
	return "project", asAtTime, testVariants
}

// QueryFailureRateSampleResponse returns expected response data from QueryFailureRate
// after being invoked with QueryFailureRateSampleRequest.
// It is assumed test data was setup with CreateQueryFailureRateTestData.
func QueryFailureRateSampleResponse() *pb.QueryTestVariantFailureRateResponse {
	var1 := pbutil.Variant("key1", "val1", "key2", "val1")
	var3 := pbutil.Variant("key1", "val2", "key2", "val2")

	day := 24 * time.Hour

	intervals := []*pb.QueryTestVariantFailureRateResponse_Interval{
		{
			IntervalAge: 1,
			StartTime:   timestamppb.New(referenceTime.Add(-1 * day)),
			EndTime:     timestamppb.New(referenceTime),
		},
		{
			IntervalAge: 2,
			StartTime:   timestamppb.New(referenceTime.Add(-2 * day)),
			EndTime:     timestamppb.New(referenceTime.Add(-1 * day)),
		},
		{
			IntervalAge: 3,
			StartTime:   timestamppb.New(referenceTime.Add(-3 * day)),
			EndTime:     timestamppb.New(referenceTime.Add(-2 * day)),
		},
		{
			IntervalAge: 4,
			StartTime:   timestamppb.New(referenceTime.Add(-4 * day)),
			EndTime:     timestamppb.New(referenceTime.Add(-3 * day)),
		},
		{
			IntervalAge: 5,
			StartTime:   timestamppb.New(referenceTime.Add(-7 * day)),
			EndTime:     timestamppb.New(referenceTime.Add(-4 * day)),
		},
	}

	analysis := []*pb.TestVariantFailureRateAnalysis{
		{
			TestId:  "test_id",
			Variant: var1,
			IntervalStats: []*pb.TestVariantFailureRateAnalysis_IntervalStats{
				{
					IntervalAge:           1,
					TotalRunFlakyVerdicts: 1, // inv3.
				},
				{
					IntervalAge:                2,
					TotalRunUnexpectedVerdicts: 1, // inv6.
				},
				{
					IntervalAge:                3,
					TotalRunFlakyVerdicts:      2, // inv4, inv5.
					TotalRunUnexpectedVerdicts: 1, // inv7.
				},
				{
					IntervalAge:              4,
					TotalRunExpectedVerdicts: 1, // inv2.
				},
				{
					IntervalAge:           5,
					TotalRunFlakyVerdicts: 1, //inv1.

				},
			},
			RunFlakyVerdictExamples: []*pb.TestVariantFailureRateAnalysis_VerdictExample{
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					IngestedInvocationId: "inv3",
					Changelists:          expectedPBChangelist(1),
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-3 * day)),
					IngestedInvocationId: "inv4",
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-3 * day)),
					IngestedInvocationId: "inv5",
					Changelists:          expectedPBChangelist(3),
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-6 * day)),
					IngestedInvocationId: "inv1",
					Changelists:          expectedPBChangelist(10),
				},
			},
			// inv4 should not be included as it is a CL authored by automation.
			RecentVerdicts: []*pb.TestVariantFailureRateAnalysis_RecentVerdict{
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					IngestedInvocationId: "inv3",
					Changelists:          expectedPBChangelist(1),
					HasUnexpectedRuns:    true,
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-2 * day)),
					IngestedInvocationId: "inv6",
					Changelists:          expectedPBChangelist(4),
					HasUnexpectedRuns:    true,
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-3 * day)),
					IngestedInvocationId: "inv5",
					Changelists:          expectedPBChangelist(3),
					HasUnexpectedRuns:    true,
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-3 * day)),
					IngestedInvocationId: "inv7",
					HasUnexpectedRuns:    true,
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-4 * day)),
					IngestedInvocationId: "inv2",
					Changelists:          expectedPBChangelist(2),
					HasUnexpectedRuns:    false,
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-6 * day)),
					IngestedInvocationId: "inv1",
					Changelists:          expectedPBChangelist(10),
					HasUnexpectedRuns:    true,
				},
			},
		},
		{
			TestId:  "test_id",
			Variant: var3,
			IntervalStats: []*pb.TestVariantFailureRateAnalysis_IntervalStats{
				{
					IntervalAge:           1,
					TotalRunFlakyVerdicts: 1, // inv8.
				},
				{
					IntervalAge: 2,
				},
				{
					IntervalAge: 3,
				},
				{
					IntervalAge: 4,
				},
				{
					IntervalAge: 5,
				},
			},
			RunFlakyVerdictExamples: []*pb.TestVariantFailureRateAnalysis_VerdictExample{
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-1 * day)),
					IngestedInvocationId: "inv8",
					Changelists:          expectedPBChangelist(1),
				},
			},
			RecentVerdicts: []*pb.TestVariantFailureRateAnalysis_RecentVerdict{
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-1 * day)),
					IngestedInvocationId: "inv8",
					Changelists:          expectedPBChangelist(1),
					HasUnexpectedRuns:    true,
				},
			},
		},
	}

	return &pb.QueryTestVariantFailureRateResponse{
		Intervals:    intervals,
		TestVariants: analysis,
	}
}

func expectedPBChangelist(change int64) []*pb.Changelist {
	return []*pb.Changelist{
		{
			Host:     "mygerrit-review.googlesource.com",
			Change:   change,
			Patchset: 5,
		},
	}
}
