// Copyright 2021 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package ingestion

import (
	"encoding/hex"
	"fmt"
	"sort"
	"testing"
	"time"

	"infra/appengine/weetbix/internal/analysis"
	"infra/appengine/weetbix/internal/analysis/clusteredfailures"
	"infra/appengine/weetbix/internal/clustering/algorithms/testname"
	"infra/appengine/weetbix/internal/clustering/chunkstore"
	cpb "infra/appengine/weetbix/internal/clustering/proto"
	"infra/appengine/weetbix/internal/testutil"

	"cloud.google.com/go/bigquery"
	rdbpb "go.chromium.org/luci/resultdb/proto/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIngest(t *testing.T) {
	Convey(`With Ingestor`, t, func() {
		ctx := testutil.SpannerTestContext(t)
		chunkStore := chunkstore.NewFakeClient()
		clusteredFailures := clusteredfailures.NewFakeClient()
		analysis := analysis.NewClusteringHandler(clusteredFailures)
		ingestor := New(chunkStore, analysis)

		opts := Options{
			Project:          "chromium",
			PartitionTime:    time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			Realm:            "chromium:ci",
			RootInvocationID: "build-123456790123456",
			CQRunID:          "cq-run-123",
		}
		testIngestion := func(input []*rdbpb.TestVariant, expectedCFs []*clusteredfailures.Entry) {
			err := ingestor.Ingest(ctx, opts, input)
			So(err, ShouldBeNil)

			So(len(clusteredFailures.Insertions), ShouldEqual, len(expectedCFs))

			// Sort both actuals and expectations by key so that we compare corresponding rows.
			sortClusteredFailures(clusteredFailures.Insertions)
			sortClusteredFailures(expectedCFs)
			for i, exp := range expectedCFs {
				actual := clusteredFailures.Insertions[i]
				So(actual, ShouldNotBeNil)

				// Chunk ID and index is assigned by ingestion.
				copyExp := *exp
				So(actual.ChunkID, ShouldNotBeEmpty)
				So(actual.ChunkIndex, ShouldBeGreaterThanOrEqualTo, 0)
				copyExp.ChunkID = actual.ChunkID
				copyExp.ChunkIndex = actual.ChunkIndex

				// LastUpdated time is assigned by Spanner.
				So(actual.LastUpdated, ShouldNotBeZeroValue)
				copyExp.LastUpdated = actual.LastUpdated

				So(actual, ShouldResemble, &copyExp)
			}
		}

		Convey(`Ingest one failure`, func() {
			const uniqifier = 1
			const taskCount = 1
			const resultsPerTask = 1
			tv := newTestVariant(uniqifier, taskCount, resultsPerTask)
			tvs := []*rdbpb.TestVariant{tv}

			// Expect the test result to be clustered by both reason and test name.
			const taskNum = 0
			const resultNum = 0
			regexpCF := expectedClusteredFailure(uniqifier, taskCount, taskNum, resultsPerTask, resultNum)
			setRegexpClustered(regexpCF)
			testnameCF := expectedClusteredFailure(uniqifier, taskCount, taskNum, resultsPerTask, resultNum)
			setTestNameClustered(testnameCF)
			expectedCFs := []*clusteredfailures.Entry{regexpCF, testnameCF}

			Convey(`Unexpected failure`, func() {
				tv.Results[0].Result.Status = rdbpb.TestStatus_FAIL
				tv.Results[0].Result.Expected = false

				testIngestion(tvs, expectedCFs)
				So(len(chunkStore.Blobs), ShouldEqual, 1)
			})
			Convey(`Expected failure`, func() {
				tv.Results[0].Result.Status = rdbpb.TestStatus_FAIL
				tv.Results[0].Result.Expected = true

				// Expect no test results ingested for a passed test
				// (even if unexpected).
				expectedCFs = nil

				testIngestion(tvs, expectedCFs)
				So(len(chunkStore.Blobs), ShouldEqual, 0)
			})
			Convey(`Unexpected pass`, func() {
				tv.Results[0].Result.Status = rdbpb.TestStatus_PASS
				tv.Results[0].Result.Expected = false

				// Expect no test results ingested for a passed test
				// (even if unexpected).
				expectedCFs = nil
				testIngestion(tvs, expectedCFs)
				So(len(chunkStore.Blobs), ShouldEqual, 0)
			})
			Convey(`Unexpected skip`, func() {
				tv.Results[0].Result.Status = rdbpb.TestStatus_SKIP
				tv.Results[0].Result.Expected = false

				// Expect no test results ingested for a skipped test
				// (even if unexpected).
				expectedCFs = nil

				testIngestion(tvs, expectedCFs)
				So(len(chunkStore.Blobs), ShouldEqual, 0)
			})
			Convey(`Failure without variant`, func() {
				// Tests are allowed to have no variant.
				tv.Variant = nil
				tv.Results[0].Result.Variant = nil

				regexpCF.Variant = nil
				testnameCF.Variant = nil

				testIngestion(tvs, expectedCFs)
				So(len(chunkStore.Blobs), ShouldEqual, 1)
			})
			Convey(`Failure without failure reason`, func() {
				// Failures may not have a failure reason.
				tv.Results[0].Result.FailureReason = nil
				testnameCF.FailureReason = nil
				expectedCFs = []*clusteredfailures.Entry{testnameCF}

				testIngestion(tvs, expectedCFs)
				So(len(chunkStore.Blobs), ShouldEqual, 1)
			})
			Convey(`Failure without CQ Run`, func() {
				opts.CQRunID = ""
				regexpCF.CQRunID = bigquery.NullString{}
				testnameCF.CQRunID = bigquery.NullString{}

				testIngestion(tvs, expectedCFs)
				So(len(chunkStore.Blobs), ShouldEqual, 1)
			})
			Convey(`Failure with exoneration`, func() {
				tv.Exonerations = []*rdbpb.TestExoneration{
					{
						Name:            fmt.Sprintf("invocations/task-mytask/tests/test-name-%v/exonerations/exon-1", uniqifier),
						TestId:          tv.TestId,
						Variant:         proto.Clone(tv.Variant).(*rdbpb.Variant),
						VariantHash:     "hash",
						ExonerationId:   "exon-1",
						ExplanationHtml: "<p>Known flake affecting CQ</p>",
					},
				}
				testnameCF.IsExonerated = true
				regexpCF.IsExonerated = true

				testIngestion(tvs, expectedCFs)
				So(len(chunkStore.Blobs), ShouldEqual, 1)
			})
		})
		Convey(`Ingest multiple failures`, func() {
			const uniqifier = 1
			const tasksPerVariant = 2
			const resultsPerTask = 2
			tv := newTestVariant(uniqifier, tasksPerVariant, resultsPerTask)
			tvs := []*rdbpb.TestVariant{tv}

			var expectedCFs []*clusteredfailures.Entry
			var expectedCFsByTask [][]*clusteredfailures.Entry
			for t := 0; t < tasksPerVariant; t++ {
				var taskExp []*clusteredfailures.Entry
				for j := 0; j < resultsPerTask; j++ {
					regexpCF := expectedClusteredFailure(uniqifier, tasksPerVariant, t, resultsPerTask, j)
					setRegexpClustered(regexpCF)
					testnameCF := expectedClusteredFailure(uniqifier, tasksPerVariant, t, resultsPerTask, j)
					setTestNameClustered(testnameCF)
					taskExp = append(taskExp, regexpCF, testnameCF)
				}
				expectedCFsByTask = append(expectedCFsByTask, taskExp)
				expectedCFs = append(expectedCFs, taskExp...)
			}

			Convey(`Tasks and CQ run blocked`, func() {
				for _, exp := range expectedCFs {
					exp.IsRootInvocationBlocked = true
					exp.IsTaskBlocked = true
				}
				testIngestion(tvs, expectedCFs)
				So(len(chunkStore.Blobs), ShouldEqual, 1)
			})
			Convey(`Some tasks blocked and CQ Run not blocked`, func() {
				// Let the last retry of the last task pass.
				tv.Results[3].Result.Status = rdbpb.TestStatus_PASS
				// First task should be blocked.
				for _, exp := range expectedCFsByTask[0] {
					exp.IsRootInvocationBlocked = false
					exp.IsTaskBlocked = true
				}
				// Second task should not be blocked.
				for _, exp := range expectedCFsByTask[1] {
					exp.IsRootInvocationBlocked = false
					exp.IsTaskBlocked = false
				}
			})
		})
		Convey(`Ingest many failures`, func() {
			var tvs []*rdbpb.TestVariant
			var expectedCFs []*clusteredfailures.Entry

			const variantCount = 20
			const tasksPerVariant = 10
			const resultsPerTask = 10
			for uniqifier := 0; uniqifier < variantCount; uniqifier++ {
				tv := newTestVariant(uniqifier, tasksPerVariant, resultsPerTask)
				tvs = append(tvs, tv)
				for t := 0; t < tasksPerVariant; t++ {
					for j := 0; j < resultsPerTask; j++ {
						regexpCF := expectedClusteredFailure(uniqifier, tasksPerVariant, t, resultsPerTask, j)
						setRegexpClustered(regexpCF)
						testnameCF := expectedClusteredFailure(uniqifier, tasksPerVariant, t, resultsPerTask, j)
						setTestNameClustered(testnameCF)
						expectedCFs = append(expectedCFs, regexpCF, testnameCF)
					}
				}
			}
			// Verify more than one chunk is ingested.
			testIngestion(tvs, expectedCFs)
			So(len(chunkStore.Blobs), ShouldBeGreaterThan, 1)
		})
	})
}

func setTestNameClustered(e *clusteredfailures.Entry) {
	e.ClusterAlgorithm = "testname-v0.1"
	e.ClusterID = hex.EncodeToString((&testname.Algorithm{}).Cluster(&cpb.Failure{
		TestId: e.TestID,
	}))
}

func setRegexpClustered(e *clusteredfailures.Entry) {
	e.ClusterAlgorithm = "regexp-v0.1"
	e.ClusterID = "5b4886907ba205f9ee2d8815452cb6e7" // Cluster ID for "Failure reason."
}

func sortClusteredFailures(cfs []*clusteredfailures.Entry) {
	sort.Slice(cfs, func(i, j int) bool {
		return clusteredFailureKey(cfs[i]) < clusteredFailureKey(cfs[j])
	})
}

func clusteredFailureKey(cf *clusteredfailures.Entry) string {
	return fmt.Sprintf("%q/%q/%q/%q", cf.Project, cf.ClusterAlgorithm, hex.EncodeToString([]byte(cf.ClusterID)), cf.TestResultID)
}

func newTestVariant(uniqifier int, taskCount int, resultsPerTask int) *rdbpb.TestVariant {
	testID := fmt.Sprintf("ninja://test_name/%v", uniqifier)
	variant := &rdbpb.Variant{
		Def: map[string]string{
			"k1": "v1",
		},
	}
	tv := &rdbpb.TestVariant{
		TestId:       testID,
		Variant:      variant,
		VariantHash:  "hash",
		Status:       rdbpb.TestVariantStatus_UNEXPECTED,
		Exonerations: nil,
		TestMetadata: &rdbpb.TestMetadata{},
	}
	for i := 0; i < taskCount; i++ {
		for j := 0; j < resultsPerTask; j++ {
			tr := newTestResult(uniqifier, i, j)
			tr.TestId = testID
			tr.Variant = proto.Clone(variant).(*rdbpb.Variant)
			tr.VariantHash = "hash"
			tv.Results = append(tv.Results, &rdbpb.TestResultBundle{Result: tr})
		}
	}
	return tv
}

func newTestResult(uniqifier, taskNum, resultNum int) *rdbpb.TestResult {
	resultID := fmt.Sprintf("result-%v-%v", taskNum, resultNum)
	return &rdbpb.TestResult{
		Name:     fmt.Sprintf("invocations/task-%v/tests/test-name-%v/results/%s", taskNum, uniqifier, resultID),
		TestId:   fmt.Sprintf("ninja://test_name/%v", uniqifier),
		ResultId: resultID,
		Variant: &rdbpb.Variant{
			Def: map[string]string{
				"k1": "v1",
			},
		},
		Expected:    false,
		Status:      rdbpb.TestStatus_CRASH,
		SummaryHtml: "<p>Some SummaryHTML</p>",
		StartTime:   timestamppb.New(time.Date(2022, time.February, 12, 0, 0, 0, 0, time.UTC)),
		Duration:    durationpb.New(time.Second * 10),
		Tags: []*rdbpb.StringPair{
			{
				Key:   "monorail_component",
				Value: "Component>MyComponent",
			},
		},
		VariantHash:  "hash",
		TestMetadata: &rdbpb.TestMetadata{},
		FailureReason: &rdbpb.FailureReason{
			PrimaryErrorMessage: "Failure reason.",
		},
	}
}

func expectedClusteredFailure(uniqifier, taskCount, taskNum, resultsPerTask, resultNum int) *clusteredfailures.Entry {
	resultID := fmt.Sprintf("result-%v-%v", taskNum, resultNum)
	return &clusteredfailures.Entry{
		Project:          "chromium",
		ClusterAlgorithm: "", // Determined by clustering algorithm.
		ClusterID:        "", // Determined by clustering algorithm.
		TestResultID:     fmt.Sprintf("invocations/task-%v/tests/test-name-%v/results/%s", taskNum, uniqifier, resultID),
		LastUpdated:      time.Time{}, // Only known at runtime, Spanner commit timestamp.

		PartitionTime:              time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
		IsIncluded:                 true,
		IsIncludedWithHighPriority: true,

		ChunkID:    "",
		ChunkIndex: 0, // To be set by caller as needed.

		Realm:  "chromium:ci",
		TestID: fmt.Sprintf("ninja://test_name/%v", uniqifier),
		Variant: []*clusteredfailures.Variant{
			{
				Key:   "k1",
				Value: "v1",
			},
		},
		VariantHash:               "hash",
		FailureReason:             &clusteredfailures.FailureReason{PrimaryErrorMessage: "Failure reason."},
		Component:                 "Component>MyComponent",
		StartTime:                 time.Date(2022, time.February, 12, 0, 0, 0, 0, time.UTC),
		Duration:                  time.Second * 10,
		IsExonerated:              false,
		RootInvocationID:          "build-123456790123456",
		RootInvocationResultSeq:   int64(taskNum*resultsPerTask + resultNum + 1),
		RootInvocationResultCount: int64(taskCount * resultsPerTask),
		IsRootInvocationBlocked:   true,
		TaskID:                    fmt.Sprintf("task-%v", taskNum),
		TaskResultSeq:             int64(resultNum + 1),
		TaskResultCount:           int64(resultsPerTask),
		IsTaskBlocked:             true,
		CQRunID:                   bigquery.NullString{StringVal: "cq-run-123", Valid: true},
	}
}
