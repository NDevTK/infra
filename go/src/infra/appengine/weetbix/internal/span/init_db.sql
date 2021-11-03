-- Copyright 2021 The Chromium Authors. All rights reserved.
--
-- Use of this source code is governed by a BSD-style license that can be
-- found in the LICENSE file.

--------------------------------------------------------------------------------
-- This script initializes a Weetbix Spanner database.

-- Stores a test variant.
-- The test variant should be:
-- * currently flaky
-- * suspected of flakiness that needs to be verified
-- * flaky before but has been fixed, broken, disabled or removed
CREATE TABLE AnalyzedTestVariants (
  -- Security realm this test variant belongs to.
  Realm STRING(64) NOT NULL,

  -- Builder that the test variant runs on.
  -- It must have the same value as the builder variant.
  Builder STRING(MAX),

  -- Unique identifier of the test,
  -- see also luci.resultdb.v1.TestResult.test_id.
  TestId STRING(MAX) NOT NULL,

  -- key:value pairs to specify the way of running the test.
  -- See also luci.resultdb.v1.TestResult.variant.
  Variant ARRAY<STRING(MAX)>,

  -- A hex-encoded sha256 of concatenated "<key>:<value>\n" variant pairs.
  -- Combination of Realm, TestId and VariantHash can identify a test variant.
  VariantHash STRING(64) NOT NULL,

  -- Timestamp when the row of a test variant was created.
  CreateTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),

  -- Status of the flaky test variant, see AnalyzedTestVariantStatus.
  Status INT64 NOT NULL,
  -- Timestamp when the status field was last updated.
  StatusUpdateTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  -- Timestamp the next UpdateTestVariant task is enqueued.
  -- This timestamp is used as a token to validate an UpdateTestVariant is
  -- expected. A task with unmatched token will be silently ignored.
  NextUpdateTaskEnqueueTime TIMESTAMP,

  -- Compressed metadata for the test case.
  -- For example, the original test name, test location, etc.
  -- See TestResult.test_metadata for details.
  -- Test location is helpful for dashboards to get aggregated data by directories.
  TestMetadata BYTES(MAX),

  -- key:value pairs for the metadata of the test variant.
  -- For example the monorail component and team email.
  Tags ARRAY<STRING(MAX)>,

  -- Flake statistics, including flake rate, failure rate and counts.
  -- See FlakeStatistics proto.
  FlakeStatistics BYTES(MAX),
  -- Timestamp when the most recent flake statistics were computed.
  FlakeStatisticUpdateTime TIMESTAMP,
) PRIMARY KEY (Realm, TestId, VariantHash);

-- Used by finding test variants with FLAKY status on a builder in
-- CollectFlakeResults task.
CREATE NULL_FILTERED INDEX AnalyzedTestVariantsPerBuilderAndStatus
ON AnalyzedTestVariants (Realm, Builder, Status);

-- Stores results of a test variant in one invocation.
CREATE TABLE Verdicts (
  -- Primary Key of the parent AnalyzedTestVariants.
  -- Security realm this test variant belongs to.
  Realm STRING(64) NOT NULL,
  -- Unique identifier of the test,
  -- see also luci.resultdb.v1.TestResult.test_id.
  TestId STRING(MAX) NOT NULL,
  -- A hex-encoded sha256 of concatenated "<key>:<value>\n" variant pairs.
  -- Combination of Realm, TestId and VariantHash can identify a test variant.
  VariantHash STRING(64) NOT NULL,

  -- Id of the build invocation the results belong to.
  InvocationId STRING(MAX) NOT NULL,

  -- Flag indicates if the verdict belongs to a try build.
  IsPreSubmit BOOL,

  -- Flag indicates if the try build the verdict belongs to contributes to
  -- a CL's submission.
  -- Verdicts with HasContributedToClSubmission as False will be filtered out
  -- for deciding the test variant's status because they could be noises.
  -- This field is only meaningful for PreSubmit verdicts.
  HasContributedToClSubmission BOOL,

  -- If the unexpected results in the verdict are exonerated.
  Exonerated BOOL,

  -- Status of the results for the parent test variant in this verdict,
  -- See VerdictStatus.
  Status INT64 NOT NULL,

  -- Result counts in the verdict.
  -- Note that SKIP results are ignored in either of the counts.
  UnexpectedResultCount INT64,
  TotalResultCount INT64,

  --Creation time of the invocation containing this verdict.
  InvocationCreationTime TIMESTAMP NOT NULL,

  -- List of colon-separated key-value pairs, where key is the cluster algorithm
  -- and value is the cluster id.
  -- key can be repeated.
  -- The clusters the first test result of the verdict is in.
  -- Once the test result reaches its retention period in the clustering
  -- system, this will cease to be updated.
  Clusters ARRAY<STRING(MAX)>,

) PRIMARY KEY (Realm, TestId, VariantHash, InvocationId),
INTERLEAVE IN PARENT AnalyzedTestVariants ON DELETE CASCADE;

-- Used by finding most recent verdicts to calculate flakiness statistics.
CREATE NULL_FILTERED INDEX VerdictsByTInvocationCreationTime
 ON Verdicts (Realm, TestId, VariantHash, InvocationCreationTime DESC);

-- BugClusters contains the bugs tracked by Weetbix, and the failure clusters
-- they are associated with.
CREATE TABLE BugClusters (
  -- The LUCI Project this bug belongs to.
  Project STRING(40) NOT NULL,
  -- The bug. For monorail, the scheme is monorail/{project}/{numeric id}.
  Bug STRING(255) NOT NULL,
  -- The associated failure cluster. In future, the intent is to replace
  -- this in favour of a failure association rule.
  AssociatedClusterId STRING(MAX) NOT NULL,
  -- Whether the bug must still be updated by Weetbix. The only allowed
  -- values are true or NULL (to indicate false). Only if the bug has
  -- been closed and no failures have been observed for a while should
  -- this be NULL. This makes it easy to retrofit a NULL_FILTERED index
  -- in future, if it is needed for performance.
  IsActive BOOL,
) PRIMARY KEY (Project, Bug);

-- FailureAssociationRules associate failures with bugs. When a rule
-- is used to match incoming test failures, the resultant cluster is
-- known as a 'bug cluster' because the failures in it are associated
-- with a bug (via the failure association rule).
-- The ID of a bug cluster corresponding to a rule is
-- (RuleBasedClusteringAlgorithm, RuleID), where
-- RuleBasedClusteringAlgorithm is the algorithm name of the algorithm
-- that clusters failures based on failure association rules (e.g.
-- 'rules-v1'), and RuleId is the ID of the rule.
CREATE TABLE FailureAssociationRules (
  -- The LUCI Project this bug belongs to.
  Project STRING(40) NOT NULL,
  -- The unique identifier for the rule. This is a randomly generated
  -- 128-bit ID, encoded as 32 lowercase hexadecimal characters.
  RuleId STRING(32) NOT NULL,
  -- The rule predicate, defining which failures are being associated.
  RuleDefinition STRING(4096) NOT NULL,
  -- The time the rule was created.
  CreationTime TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  -- The last time the rule was updated.
  LastUpdated TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  -- The bug the failures are associated to. For monorail, the scheme
  -- is monorail/{project}/{numeric id}.
  Bug STRING(255) NOT NULL,
  -- Whether the bug must still be updated by Weetbix, and whether failures
  -- should still be matched against this rule. The only allowed
  -- values are true or NULL (to indicate false). Only if the bug has
  -- been closed and no failures have been observed for a while should
  -- this be NULL. This makes it easy to retrofit a NULL_FILTERED index
  -- in future, if it is needed for performance.
  IsActive BOOL,
  -- The clustering algorithm of the suggested cluster that this failure
  -- association rule was created from (if any).
  -- Until re-clustering is complete (and the residual impact of the source
  -- cluster has reduced to zero), SourceClusterAlgorithm and SourceClusterId
  -- tell bug filing to ignore the source suggested cluster when
  -- determining whether new bugs need to be filed.
  SourceClusterAlgorithm STRING(32) NOT NULL,
  -- The algorithm-defined Cluster ID of the suggested cluster that this
  -- bug cluster was created from (if any).
  SourceClusterId STRING(32) NOT NULL,
) PRIMARY KEY (Project, RuleId);

-- The failure association rule associated with a bug. This also enforces
-- the constraint that each rule must have a unique bug, even if the rules
-- are in different LUCI Projects.
-- This is required to ensure that automatic bug filing does not attempt to
-- take conflicting actions (e.g. simultaneously increase and decrease priority)
-- on the same bug, because of differing priorities in different projects.
CREATE UNIQUE INDEX FailureAssociationRuleByBug ON FailureAssociationRules(Bug);

-- Clustering state records the clustering state of failed test results, organised
-- by chunk.
CREATE TABLE ClusteringState (
  -- The LUCI Project the test results belong to.
  Project STRING(40) NOT NULL,
  -- The identity of the chunk of test results. 32 lowercase hexadecimal
  -- characters assigned by the ingestion process.
  ChunkId STRING(32) NOT NULL,
  -- The start of the retention period of the test results in the chunk.
  PartitionTime TIMESTAMP NOT NULL,
  -- The identity of the blob storing the chunk's test results.
  ObjectId STRING(32) NOT NULL,
  -- The version of clustering algorithms used to cluster test results in this
  -- chunk. (This is a version over the set of algorithms, distinct from the
  -- versions of a single algorithm, e.g.:
  -- v1 -> {failurereason-v1}, v2 -> {failurereason-v1, testname-v1},
  -- v3 -> {failurereason-v2, testname-v1}.)
  AlgorithmsVersion INT64 NOT NULL,
  -- The version of the set of failure association rules used to match test
  -- results in this chunk. This is the "Last Updated" time of the most
  -- recently updated failure association rule in the snapshot of failure
  -- association rules used to match the test results.
  RuleVersion TIMESTAMP NOT NULL,
  -- Serialized ChunkClusters proto containing which test result is in which
  -- cluster.
  Clusters BYTES(MAX) NOT NULL,
  -- The Spanner commit timestamp of when the row was last updated.
  LastUpdated TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (Project, ChunkId);

-- Stores transactional tasks reminders.
-- See https://go.chromium.org/luci/server/tq. Scanned by tq-sweeper-spanner.
CREATE TABLE TQReminders (
                             ID STRING(MAX) NOT NULL,
                             FreshUntil TIMESTAMP NOT NULL,
                             Payload BYTES(102400) NOT NULL,
) PRIMARY KEY (ID ASC);
