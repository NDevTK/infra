# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""This module is for test-try-job-related operations.

It provides functions to:
  * Decide if a new test try job is needed.
  * Get reliable failures based on swarming rerun results.
  * Get parameters for starting a new test try job.
"""

from collections import defaultdict
import copy
import logging

from google.appengine.ext import ndb

from common.waterfall import failure_type
from libs import analysis_status
from libs import time_util
from model import analysis_approach_type
from model import result_status
from model.wf_analysis import WfAnalysis
from model.wf_swarming_task import WfSwarmingTask
from model.wf_try_job import WfTryJob
from model.wf_try_job_data import WfTryJobData
from services import try_job as try_job_service
from services.test_failure import ci_test_failure
from waterfall import build_util
from waterfall import suspected_cl_util
from waterfall import swarming_util
from waterfall import waterfall_config


def _GetStepsAndTests(failed_steps):
  """Extracts failed steps and tests from failed_steps data structure.

  Args:
    failed_steps: Failed steps and test, plus extra information. Example:
    {
        'step_a': {
            'last_pass': 4,
            'tests': {
                'test1': {
                    'last_pass': 4,
                    'current_failure': 6,
                    'first_failure': 5
                },
                'test2': {
                    'last_pass': 4,
                    'current_failure': 6,
                    'first_failure': 5
                }
            },
            'current_failure': 6,
            'first_failure': 5,
            'list_isolated_data': [
                {
                    'isolatedserver': 'https://isolateserver.appspot.com',
                    'namespace': 'default-gzip',
                    'digest': 'abcd'
                }
            ]
        },
        'step_b': {
            'current_failure': 3,
            'first_failure': 2,
            'last_pass': 1
        }
    }

  Returns:
    failed_steps_and_tests: Sorted list of lists of step and test names.
    Example:
    [
        ['step_a', 'test1'],
        ['step_a', 'test2'],
        ['step_b', None]
    ]
  """

  failed_steps_and_tests = []

  if not failed_steps:
    return failed_steps_and_tests

  for step_name, step in failed_steps.iteritems():
    for test_name in step.get('tests', [None]):
      failed_steps_and_tests.append([step_name, test_name])

  return sorted(failed_steps_and_tests)


def _GetMatchingTestFailureGroups(failed_steps_and_tests):
  groups = try_job_service.GetMatchingFailureGroups(failure_type.TEST)
  return [
      group for group in groups
      if group.failed_steps_and_tests == failed_steps_and_tests
  ]


def _IsTestFailureUniqueAcrossPlatforms(master_name, builder_name, build_number,
                                        build_failure_type, blame_list,
                                        failed_steps, heuristic_result):

  if build_failure_type != failure_type.TEST:
    logging.info('Expected test failure but get %s failure.' %
                 failure_type.GetDescriptionForFailureType(build_failure_type))
    return True

  failed_steps_and_tests = _GetStepsAndTests(failed_steps)
  if not failed_steps_and_tests:
    return True
  groups = _GetMatchingTestFailureGroups(failed_steps_and_tests)

  return try_job_service.IsBuildFailureUniqueAcrossPlatforms(
      master_name,
      builder_name,
      build_number,
      build_failure_type,
      blame_list,
      heuristic_result,
      groups,
      failed_steps_and_tests=failed_steps_and_tests)


def _HasBuildKeyForBuildInfoInFailureResultMap(master_name, builder_name,
                                               build_number):
  """Checks if there is any first failed test."""
  analysis = WfAnalysis.Get(master_name, builder_name, build_number)
  failure_result_map = analysis.failure_result_map
  current_build_key = build_util.CreateBuildId(master_name, builder_name,
                                               build_number)
  for step_keys in failure_result_map.itervalues():
    for test_key in step_keys.itervalues():
      if test_key == current_build_key:
        return True
  return False


def _NeedANewTestTryJob(master_name, builder_name, build_number, force_try_job):

  if (not force_try_job and
      waterfall_config.ShouldSkipTestTryJobs(master_name, builder_name)):
    logging.info('Test try jobs on %s, %s are not supported yet.', master_name,
                 builder_name)
    return False

  return _HasBuildKeyForBuildInfoInFailureResultMap(master_name, builder_name,
                                                    build_number)


def NeedANewTestTryJob(master_name,
                       builder_name,
                       build_number,
                       failure_info,
                       heuristic_result,
                       force_try_job=False):
  """Decides if a new test try job is needed.

  A new test try job is needed if:
  1. It passed preliminary checks in try_job_service.NeedANewWaterfallTryJob,
  2. It's for a test failure,
  3. It contains some first failed steps/tests,
  4. There is no other running or completed try job.

  Returns:
    A bool to indicate if a new try job is needed.
    A key to the entity of the try job.
  """
  need_new_try_job = try_job_service.NeedANewWaterfallTryJob(
      master_name, builder_name, build_number, force_try_job)

  if not need_new_try_job:
    return False, None

  try_job_type = failure_info['failure_type']
  if try_job_type != failure_type.TEST:
    logging.error('Checking for a test try job but got a %s failure.',
                  failure_type.GetDescriptionForFailureType(try_job_type))
    return False, None

  need_new_try_job = _NeedANewTestTryJob(master_name, builder_name,
                                         build_number, force_try_job)

  # TODO(chanli): enable the feature to trigger single try job for a group
  # when notification is ready.
  # We still call _IsBuildFailureUniqueAcrossPlatforms just so we have data for
  # failure groups.

  # TODO(chanli): Add checking for culprits of the group when enabling
  # single try job: add current build to suspected_cl.builds if the try job for
  # this group has already completed.
  if need_new_try_job:
    _IsTestFailureUniqueAcrossPlatforms(
        master_name, builder_name, build_number, try_job_type,
        failure_info['builds'][str(build_number)]['blame_list'],
        failure_info['failed_steps'], heuristic_result)

  try_job_was_created, try_job_key = try_job_service.ReviveOrCreateTryJobEntity(
      master_name, builder_name, build_number, force_try_job)
  need_new_try_job = need_new_try_job and try_job_was_created
  return need_new_try_job, try_job_key


def _GetLastPassTest(build_number, failed_steps):
  for step_failure in failed_steps.itervalues():
    for test_failure in step_failure.get('tests', {}).itervalues():
      if (test_failure['first_failure'] == build_number and
          test_failure.get('last_pass') is not None):
        return test_failure['last_pass']
  return None


def _GetGoodRevisionTest(master_name, builder_name, build_number, failure_info):
  last_pass = _GetLastPassTest(build_number, failure_info['failed_steps'])
  if last_pass is None:
    logging.warning('Couldn"t start try job for build %s, %s, %d because'
                    ' last_pass is not found.', master_name, builder_name,
                    build_number)
    return None

  return failure_info['builds'][str(last_pass)]['chromium_revision']


def GetParametersToScheduleTestTryJob(master_name, builder_name, build_number,
                                      failure_info, heuristic_result):
  parameters = {}
  parameters['bad_revision'] = failure_info['builds'][str(build_number)][
      'chromium_revision']
  parameters[
      'suspected_revisions'] = try_job_service.GetSuspectsFromHeuristicResult(
          heuristic_result)
  parameters['good_revision'] = _GetGoodRevisionTest(master_name, builder_name,
                                                     build_number, failure_info)

  parameters['task_results'] = GetReliableTests(master_name, builder_name,
                                                build_number, failure_info)

  parent_mastername = failure_info.get('parent_mastername') or master_name
  parent_buildername = failure_info.get('parent_buildername') or builder_name
  parameters['dimensions'] = waterfall_config.GetTrybotDimensions(
      parent_mastername, parent_buildername)
  parameters['cache_name'] = swarming_util.GetCacheName(parent_mastername,
                                                        parent_buildername)
  return parameters


# TODO(chanli@): move this function to swarming task related module.
def GetReliableTests(master_name, builder_name, build_number, failure_info):
  task_results = {}
  for step_name, step_failure in failure_info['failed_steps'].iteritems():
    if not ci_test_failure.AnyTestHasFirstTimeFailure(
        step_failure.get('tests', {}), build_number):
      continue
    task = WfSwarmingTask.Get(master_name, builder_name, build_number,
                              step_name)

    if not task or not task.classified_tests:
      logging.error('No result for swarming task %s/%s/%s/%s' %
                    (master_name, builder_name, build_number, step_name))
      continue

    if not task.reliable_tests:
      continue

    task_results[task.canonical_step_name or step_name] = task.reliable_tests

  return task_results


def GetBuildProperties(master_name, builder_name, build_number, good_revision,
                       bad_revision, suspected_revisions):
  properties = try_job_service.GetBuildProperties(
      master_name, builder_name, build_number, good_revision, bad_revision,
      failure_type.TEST, suspected_revisions)
  properties['target_testername'] = builder_name

  return properties


def _GetResultAnalysisStatus(analysis, result, all_flaked=False):
  """Returns the analysis status based on existing status and try job result.

  Args:
    analysis: The WfAnalysis entity corresponding to this try job.
    result: A result dict containing the result of this try job.
    all_flaked: A flag indicates if all failures are flaky.

  Returns:
    A result_status code.
  """
  if all_flaked:
    return result_status.FLAKY

  return try_job_service.GetResultAnalysisStatus(analysis, result)


def _GetTestFailureCausedByCL(result):
  if not result:
    return None

  failures = {}
  for step_name, step_result in result.iteritems():
    if step_result['status'] == 'failed':
      failures[step_name] = step_result['failures']

  return failures


def _GetUpdatedSuspectedCLs(analysis, result, culprits):
  """Returns a list of suspected CLs.

  Args:
    analysis: The WfAnalysis entity corresponding to this try job.
    result: A result dict containing the result of this try job.
    culprits: A list of suspected CLs found by the try job.

  Returns:
    A combined list of suspected CLs from those already in analysis and those
    found by this try job.
  """
  suspected_cls = analysis.suspected_cls[:] if analysis.suspected_cls else []
  suspected_cl_revisions = [cl['revision'] for cl in suspected_cls]

  for revision, try_job_suspected_cl in culprits.iteritems():
    suspected_cl_copy = copy.deepcopy(try_job_suspected_cl)
    if revision not in suspected_cl_revisions:
      suspected_cl_revisions.append(revision)
      failures = _GetTestFailureCausedByCL(
          result.get('report', {}).get('result', {}).get(revision))
      suspected_cl_copy['failures'] = failures
      suspected_cl_copy['top_score'] = None
      suspected_cls.append(suspected_cl_copy)

  return suspected_cls


def _GetUpdatedAnalysisResult(analysis, flaky_failures):
  if not analysis or not analysis.result or not analysis.result.get('failures'):
    return [], False

  analysis_result = copy.deepcopy(analysis.result)
  all_flaky = swarming_util.UpdateAnalysisResult(analysis_result,
                                                 flaky_failures)

  return analysis_result, all_flaky


def FindCulpritForEachTestFailure(result):
  culprit_map = defaultdict(dict)
  failed_revisions = set()

  # Recipe should return culprits with the format as:
  # 'culprits': {
  #     'step1': {
  #         'test1': 'rev1',
  #         'test2': 'rev2',
  #         ...
  #     },
  #     ...
  # }
  if result['report'].get('culprits'):
    for step_name, tests in result['report']['culprits'].iteritems():
      culprit_map[step_name]['tests'] = {}
      for test_name, revision in tests.iteritems():
        culprit_map[step_name]['tests'][test_name] = {'revision': revision}
        failed_revisions.add(revision)
  return culprit_map, list(failed_revisions)


def UpdateCulpritMapWithCulpritInfo(culprit_map, culprits):
  """Fills in commit_position and review url for each failed rev in map."""
  for step_culprit in culprit_map.values():
    for test_culprit in step_culprit.get('tests', {}).values():
      test_revision = test_culprit['revision']
      test_culprit.update(culprits[test_revision])


def GetCulpritDataForTest(culprit_map):
  """Gets culprit revision for each failure for try job metadata."""
  culprit_data = {}
  for step, step_culprit in culprit_map.iteritems():
    culprit_data[step] = {}
    for test, test_culprit in step_culprit['tests'].iteritems():
      culprit_data[step][test] = test_culprit['revision']
  return culprit_data


@ndb.transactional
def UpdateTryJobResult(master_name, builder_name, build_number, result,
                       try_job_id, culprits):
  try_job = WfTryJob.Get(master_name, builder_name, build_number)
  try_job_service.UpdateTryJobResultWithCulprit(try_job.test_results, result,
                                                try_job_id, culprits)
  try_job.status = analysis_status.COMPLETED
  try_job.put()


@ndb.transactional
def UpdateWfAnalysisWithTryJobResult(master_name, builder_name, build_number,
                                     result, culprits, flaky_failures):
  if not culprits and not flaky_failures:
    return

  analysis = WfAnalysis.Get(master_name, builder_name, build_number)
  # Update analysis result and suspected CLs with results of this try job if
  # culprits were found or failures are flaky.
  updated_result, all_flaked = _GetUpdatedAnalysisResult(
      analysis, flaky_failures)
  updated_result_status = _GetResultAnalysisStatus(analysis, result, all_flaked)
  updated_suspected_cls = _GetUpdatedSuspectedCLs(analysis, result, culprits)
  try_job_service.UpdateWfAnalysisWithTryJobResult(
      analysis, updated_result_status, updated_suspected_cls, updated_result)


def UpdateSuspectedCLs(master_name, builder_name, build_number, culprits,
                       result):
  if not culprits:
    return

  # Creates or updates each suspected_cl.
  for culprit in culprits.values():
    revision = culprit['revision']
    failures = _GetTestFailureCausedByCL(
        result.get('report', {}).get('result', {}).get(revision))

    suspected_cl_util.UpdateSuspectedCL(culprit['repo_name'], revision,
                                        culprit.get('commit_position'),
                                        analysis_approach_type.TRY_JOB,
                                        master_name, builder_name, build_number,
                                        failure_type.TEST, failures, None)
