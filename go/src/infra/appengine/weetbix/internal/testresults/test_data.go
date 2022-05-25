package testresults

import (
	"context"
	"time"

	"go.chromium.org/luci/server/span"
	"google.golang.org/protobuf/types/known/timestamppb"

	"infra/appengine/weetbix/pbutil"
	pb "infra/appengine/weetbix/proto/v1"
)

// CreateQueryFailureRateTestData creates test data in Spanner for testing
// QueryFailureRate.
func CreateQueryFailureRateTestData(ctx context.Context, referenceTime time.Time) error {
	var1 := pbutil.Variant("key1", "val1", "key2", "val1")
	var2 := pbutil.Variant("key1", "val2", "key2", "val1")
	var3 := pbutil.Variant("key1", "val2", "key2", "val2")

	_, err := span.ReadWriteTransaction(ctx, func(ctx context.Context) error {
		insertTV := func(partitionTime time.Time, variant *pb.Variant, invId string, runStatuses []RunStatus, changeListNumber ...int64) {
			baseTestResult := NewTestResult().
				WithProject("project").
				WithTestID("test_id").
				WithVariantHash(pbutil.VariantHash(variant)).
				WithPartitionTime(partitionTime).
				WithIngestedInvocationID(invId).
				WithSubRealm("realm").
				WithStatus(pb.TestResultStatus_PASS)

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

		passFail := []RunStatus{Flaky, Unexpected}
		failPass := []RunStatus{Unexpected, Flaky}
		pass := []RunStatus{Flaky}
		fail := []RunStatus{Unexpected}
		failFail := []RunStatus{Unexpected, Unexpected}

		insertTV(referenceTime.Add(-6*time.Hour), var1, "inv1", failPass, 10)
		// duplicate-cl result should not be used, inv3 result should be used instead
		// (as only one verdict per changelist is used).
		insertTV(referenceTime.Add(-4*time.Hour), var1, "duplicate-cl", pass, 1)
		insertTV(referenceTime.Add(-4*time.Hour), var1, "inv2", pass, 2)
		insertTV(referenceTime.Add(-2*time.Hour), var1, "inv3", failPass, 1)
		insertTV(referenceTime.Add(-3*time.Hour), var1, "inv4", failPass)
		insertTV(referenceTime.Add(-3*time.Hour), var1, "inv5", passFail, 3)
		insertTV(referenceTime.Add(-2*time.Hour), var1, "inv6", fail, 4)
		insertTV(referenceTime.Add(-3*time.Hour), var1, "inv7", failFail)
		// should not be used, as tests multiple CLs, and too hard
		// to deduplicate the verdicts.
		insertTV(referenceTime.Add(-2*time.Hour), var1, "many-cl", failPass, 1, 3)

		insertTV(referenceTime.Add(-4*time.Hour), var2, "inv1", failPass, 1)
		insertTV(referenceTime.Add(-3*time.Hour), var2, "inv2", failPass, 2)

		insertTV(referenceTime.Add(-5*time.Hour), var3, "duplicate-cl1", passFail, 1)
		insertTV(referenceTime.Add(-3*time.Hour), var3, "duplicate-cl2", failPass, 1)
		insertTV(referenceTime.Add(-1*time.Hour), var3, "inv8", failPass, 1)

		return nil
	})
	return err
}

func QueryFailureRateSampleRequest() (project string, testVariants []*pb.TestVariantIdentifier) {
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
	return "project", testVariants
}

// QueryFailureRateSampleResponse returns expected response data from QueryFailureRate
// after being invoked with QueryFailureRateSampleRequest (and an AfterPartitionTime of
// of referenceTime.Add(-6*time.Hour)).
// It is assumed test data was setup with CreateQueryFailureRateTestData.
func QueryFailureRateSampleResponse(referenceTime time.Time) []*pb.TestVariantFailureRateAnalysis {
	var1 := pbutil.Variant("key1", "val1", "key2", "val1")
	var3 := pbutil.Variant("key1", "val2", "key2", "val2")

	return []*pb.TestVariantFailureRateAnalysis{
		{
			TestId:  "test_id",
			Variant: var1,
			FailingRunRatio: &pb.FailingRunRatio{
				Numerator:   4, // inv3, inv4, inv6, inv7
				Denominator: 6, // as above, plus inv2, inv5
			},
			FailingRunExamples: []*pb.VerdictExample{
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					IngestedInvocationId: "inv3",
					Changelists:          expectedPBChangelist(1),
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					IngestedInvocationId: "inv6",
					Changelists:          expectedPBChangelist(4),
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-3 * time.Hour)),
					IngestedInvocationId: "inv4",
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-3 * time.Hour)),
					IngestedInvocationId: "inv7",
				},
			},
			FlakyVerdictRatio: &pb.FlakyVerdictRatio{
				Numerator:   2, // inv3, inv4
				Denominator: 4, // as above, plus inv2, inv5
			},
			FlakyVerdictExamples: []*pb.VerdictExample{
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-2 * time.Hour)),
					IngestedInvocationId: "inv3",
					Changelists:          expectedPBChangelist(1),
				},
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-3 * time.Hour)),
					IngestedInvocationId: "inv4",
				},
			},
			Sample: &pb.TestVariantFailureRateAnalysis_Sample{
				Verdicts:                 6,
				VerdictsPreDeduplication: 7,
			},
		},
		{
			TestId:  "test_id",
			Variant: var3,
			FailingRunRatio: &pb.FailingRunRatio{
				Numerator:   1, // inv8
				Denominator: 1, // inv8
			},
			FailingRunExamples: []*pb.VerdictExample{
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-1 * time.Hour)),
					IngestedInvocationId: "inv8",
					Changelists:          expectedPBChangelist(1),
				},
			},
			FlakyVerdictRatio: &pb.FlakyVerdictRatio{
				Numerator:   1, // inv8
				Denominator: 1, // inv8
			},
			FlakyVerdictExamples: []*pb.VerdictExample{
				{
					PartitionTime:        timestamppb.New(referenceTime.Add(-1 * time.Hour)),
					IngestedInvocationId: "inv8",
					Changelists:          expectedPBChangelist(1),
				},
			},
			Sample: &pb.TestVariantFailureRateAnalysis_Sample{
				Verdicts:                 1,
				VerdictsPreDeduplication: 3,
			},
		},
	}
}

func expectedPBChangelist(change int64) []*pb.Changelist {
	return []*pb.Changelist{
		{
			Host:     "mygerrit",
			Change:   change,
			Patchset: 5,
		},
	}
}
