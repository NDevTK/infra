package analysis

const clusterPresubmitAnalysis = `
  WITH clustered_failures_latest AS (
	SELECT
	  cluster_algorithm,
	  cluster_id,
	  test_result_system,
	  test_result_id,
	  DATE(partition_time) as partition_time,
	  ARRAY_AGG(cf ORDER BY last_updated DESC LIMIT 1)[OFFSET(0)] as r
	FROM clustered_failures cf
	WHERE partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 DAY)
	GROUP BY cluster_algorithm, cluster_id, test_result_system, test_result_id, DATE(partition_time)
  ),
  clustered_failures_extended AS (
	SELECT
	  cluster_algorithm,
	  cluster_id,
	  r.is_included,
	  r.partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 12 HOUR) as is_12h,
	  r.partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 DAY) as is_1d,
	  -- The identity of the first changelist that was tested, assuming the
	  -- result was part of a presubmit run, and the owner of the presubmit
	  -- run was a user and not automation.
	  IF(ARRAY_LENGTH(r.changelists)>0 AND r.presubmit_run_owner='user',
		  CONCAT(r.changelists[OFFSET(0)].host, r.changelists[OFFSET(0)].change),
		  NULL) as presubmit_run_user_cl_id,
	  r.is_test_run_blocked as is_test_run_fail,
	FROM clustered_failures_latest
  )
  SELECT
    STRUCT(cluster_algorithm AS Algorithm, cluster_id as ID) as ClusterID,
	COUNT(DISTINCT IF(is_12h AND is_test_run_fail, presubmit_run_user_cl_id, NULL)) as DistinctUserClTestRunsFailed12h,
	COUNT(DISTINCT IF(is_1d AND is_test_run_fail, presubmit_run_user_cl_id, NULL)) as DistinctUserClTestRunsFailed1d,
  FROM clustered_failures_extended
  WHERE STRUCT(cluster_algorithm AS Algorithm, cluster_id as ID) IN UNNEST(@clusterIDs)
    AND is_included
  GROUP BY cluster_algorithm, cluster_id
`

const clusterSummariesAnalysis = `
  WITH clustered_failures_latest AS (
	SELECT
	  cluster_algorithm,
	  cluster_id,
	  test_result_system,
	  test_result_id,
	  DATE(partition_time) as partition_time,
	  ARRAY_AGG(cf ORDER BY last_updated DESC LIMIT 1)[OFFSET(0)] as r
	FROM clustered_failures cf
	WHERE partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
	GROUP BY cluster_algorithm, cluster_id, test_result_system, test_result_id, DATE(partition_time)
  ),
  clustered_failures_extended AS (
	SELECT
	  cluster_algorithm,
	  cluster_id,
	  r.is_included,
	  r.is_included_with_high_priority,
	  (ARRAY_LENGTH(r.exonerations) > 0) as is_exonerated,
	  r.build_status = 'FAILURE' as build_failed,
	  -- Presubmit run and tryjob is critical, and
	  (r.build_critical AND
		-- Exonerated for a reason other than NOT_CRITICAL or UNEXPECTED_PASS.
		-- Passes are not ingested by Weetbix, but if a test has both an unexpected pass
		-- and an unexpected failure, it will be exonerated for the unexpected pass.
		(EXISTS
		  (SELECT TRUE FROM UNNEST(r.exonerations) e
		  WHERE e.Reason = 'OCCURS_ON_MAINLINE' OR e.Reason = 'OCCURS_ON_OTHER_CLS'))) as is_critical_and_exonerated,
	  r.test_id,
	  r.failure_reason,
	  r.bug_tracking_component,
	  r.test_run_id,
	  r.partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 12 HOUR) as is_12h,
	  r.partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 1 DAY) as is_1d,
	  r.partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 3 DAY) as is_3d,
	  r.partition_time >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY) as is_7d,
	  -- The identity of the first changelist that was tested, assuming the
	  -- result was part of a presubmit run, and the owner of the presubmit
	  -- run was a user and not automation.
	  IF(ARRAY_LENGTH(r.changelists)>0 AND r.presubmit_run_owner='user',
		  CONCAT(r.changelists[OFFSET(0)].host, r.changelists[OFFSET(0)].change),
		  NULL) as presubmit_run_user_cl_id,
	  (r.is_ingested_invocation_blocked AND r.build_critical AND
		r.presubmit_run_mode = 'FULL_RUN') as is_presubmit_reject,
	  r.is_test_run_blocked as is_test_run_fail,
	FROM clustered_failures_latest
  )

  SELECT
	  cluster_algorithm,
	  cluster_id,

	  -- 1 day metrics.
	  COUNT(DISTINCT IF(is_1d AND is_presubmit_reject AND NOT is_exonerated AND build_failed, presubmit_run_user_cl_id, NULL)) as presubmit_rejects_1d,
	  COUNT(DISTINCT IF(is_1d AND is_presubmit_reject AND is_included_with_high_priority AND NOT is_exonerated AND build_failed, presubmit_run_user_cl_id, NULL)) as presubmit_rejects_residual_1d,
	  COUNT(DISTINCT IF(is_1d AND is_test_run_fail, test_run_id, NULL)) as test_run_fails_1d,
	  COUNT(DISTINCT IF(is_1d AND is_test_run_fail AND is_included_with_high_priority, test_run_id, NULL)) as test_run_fails_residual_1d,
	  COUNTIF(is_1d) as failures_1d,
	  COUNTIF(is_1d AND is_included_with_high_priority) as failures_residual_1d,
	  COUNTIF(is_1d AND is_critical_and_exonerated) as critical_failures_exonerated_1d,
	  COUNTIF(is_1d AND is_critical_and_exonerated AND is_included_with_high_priority) as critical_failures_exonerated_residual_1d,

	  -- 3 day metrics.
	  COUNT(DISTINCT IF(is_3d AND is_presubmit_reject AND NOT is_exonerated AND build_failed, presubmit_run_user_cl_id, NULL)) as presubmit_rejects_3d,
	  COUNT(DISTINCT IF(is_3d AND is_presubmit_reject AND is_included_with_high_priority AND NOT is_exonerated AND build_failed, presubmit_run_user_cl_id, NULL)) as presubmit_rejects_residual_3d,
	  COUNT(DISTINCT IF(is_3d AND is_test_run_fail, test_run_id, NULL)) as test_run_fails_3d,
	  COUNT(DISTINCT IF(is_3d AND is_test_run_fail AND is_included_with_high_priority, test_run_id, NULL)) as test_run_fails_residual_3d,
	  COUNTIF(is_3d) as failures_3d,
	  COUNTIF(is_3d AND is_included_with_high_priority) as failures_residual_3d,
	  COUNTIF(is_3d AND is_critical_and_exonerated) as critical_failures_exonerated_3d,
	  COUNTIF(is_3d AND is_critical_and_exonerated AND is_included_with_high_priority) as critical_failures_exonerated_residual_3d,

	  -- 7 day metrics.
	  COUNT(DISTINCT IF(is_7d AND is_presubmit_reject AND NOT is_exonerated AND build_failed, presubmit_run_user_cl_id, NULL)) as presubmit_rejects_7d,
	  COUNT(DISTINCT IF(is_7d AND is_presubmit_reject AND is_included_with_high_priority AND NOT is_exonerated AND build_failed, presubmit_run_user_cl_id, NULL)) as presubmit_rejects_residual_7d,
	  COUNT(DISTINCT IF(is_7d AND is_test_run_fail, test_run_id, NULL)) as test_run_fails_7d,
	  COUNT(DISTINCT IF(is_7d AND is_test_run_fail AND is_included_with_high_priority, test_run_id, NULL)) as test_run_fails_residual_7d,
	  COUNTIF(is_7d) as failures_7d,
	  COUNTIF(is_7d AND is_included_with_high_priority) as failures_residual_7d,
	  COUNTIF(is_7d AND is_critical_and_exonerated) as critical_failures_exonerated_7d,
	  COUNTIF(is_7d AND is_critical_and_exonerated AND is_included_with_high_priority) as critical_failures_exonerated_residual_7d,

	  -- Other analysis.
	  ANY_VALUE(failure_reason) as example_failure_reason,
	  APPROX_TOP_COUNT(test_id, 5) as top_test_ids,
	  APPROX_TOP_COUNT(IF(bug_tracking_component.system = 'monorail', bug_tracking_component.component, NULL), 5) as top_monorail_components,
  FROM clustered_failures_extended
  WHERE is_included
  GROUP BY cluster_algorithm, cluster_id`
