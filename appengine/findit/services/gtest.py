# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""This module is for gtest-related operations.

It provides functions to:
  * normalize the test names
  * concatenate gtest logs
  * Remove platform from step name.
"""

import base64
from collections import defaultdict
import cStringIO

from services import constants

_PRE_TEST_PREFIX = 'PRE_'

# Invalid gtest result error codes.
# TODO(crbug.com/785463): Use enum for error codes.
RESULTS_INVALID = 10

_NON_FAILURE_STATUSES = ['SUCCESS', 'SKIPPED', 'UNKNOWN']


def IsTestResultsInExpectedFormat(test_results_log):
  """Checks if the log can be parsed by gtest.

  Args:
    test_results_log (dict): It should be in below format:
    {
        'all_tests': ['test1',
                      'test2',
                      ...],
        'per_iteration_data': [
            {
                'test1': [
                  {
                      'status': 'SUCCESS',
                      'output_snippet': 'output',
                      ...
                  }
                ],
                'test2': [
                    {},
                    {},
                    ...
                ]
            }
        ]
    }

  """
  return (isinstance(test_results_log, dict) and
          isinstance(test_results_log.get('all_tests'), list) and
          isinstance(test_results_log.get('per_iteration_data'), list) and all(
              isinstance(i, dict)
              for i in test_results_log.get('per_iteration_data')))


class GtestResults(object):

  def RemoveAllPrefixes(self, test):
    """Removes prefixes from test names.

    Args:
      test (str): A test's name, eg: 'suite1.PRE_test1'.

    Returns:
      base_test (str): A base test name, eg: 'suite1.test1'.
    """
    test_name_start = max(test.find('.'), 0)
    if test_name_start == 0:
      return test

    test_suite = test[:test_name_start]
    test_name = test[test_name_start + 1:]
    pre_position = test_name.find(_PRE_TEST_PREFIX)
    while pre_position == 0:
      test_name = test_name[len(_PRE_TEST_PREFIX):]
      pre_position = test_name.find(_PRE_TEST_PREFIX)
    base_test = '%s.%s' % (test_suite, test_name)
    return base_test

  # TODO(crbug/805732): Get rid of repeated decode/encode operations.
  def ConcatenateTestLog(self, string1, string2):
    """Concatenates the base64 encoded log.

    Tests if one string is a substring of another,
        if yes, returns the longer string,
        otherwise, returns the concatenation.

    Args:
      string1: base64-encoded string.
      string2: base64-encoded string.

    Returns:
      base64-encoded string.
    """
    str1 = base64.b64decode(string1)
    str2 = base64.b64decode(string2)
    if str2 in str1:
      return string1
    elif str1 in str2:
      return string2
    else:
      return base64.b64encode(str1 + str2)

  def GetConsistentTestFailureLog(self, gtest_result):
    """Analyzes the archived gtest json results and extract reliable failures.

    Args:
      gtest_result (dict): A JSON file for failed step log.

    Returns:
      A string contains the names of reliable test failures and related
      log content.
      If gtest_results_log in gtest json result is 'invalid', we will return
      'invalid' as the result.
      If we find out that all the test failures in this step are flaky, we will
      return 'flaky' as result.
    """
    sio = cStringIO.StringIO()
    for iteration in gtest_result['per_iteration_data']:
      for test_name in iteration.keys():
        is_reliable_failure = True

        for test_run in iteration[test_name]:
          # We will ignore the test if some of the attempts were success.
          if test_run['status'] == 'SUCCESS':
            is_reliable_failure = False
            break

        if is_reliable_failure:  # all attempts failed
          for test_run in iteration[test_name]:
            sio.write(base64.b64decode(test_run['output_snippet_base64']))

    failed_test_log = sio.getvalue()
    sio.close()

    if not failed_test_log:
      return constants.FLAKY_FAILURE_LOG

    return failed_test_log

  def DoesTestExist(self, gtest_result, test_name):
    """Determines whether test_name is in gtest_result's 'all_tests' field.

    Args:
      gtest_result (dict): A gtest's json output expected to be in the format:
          {
              'all_tests': [(str)],
              ...,
          }
      test_name (str): The name of the test to check.

    Returns:
      True if the test exists according to gtest_result, False otherwise.
    """
    return test_name in (gtest_result.get('all_tests') or [])

  def IsTestEnabled(self, test_name, gtest_result):
    """Returns True if the test is enabled, False otherwise."""
    if not gtest_result:
      return False

    all_tests = gtest_result.get('all_tests', [])
    disabled_tests = gtest_result.get('disabled_tests', [])

    # Checks if one test was enabled by checking the test results.
    # If the disabled tests array is empty, we assume the test is enabled.
    return test_name in all_tests and test_name not in disabled_tests

  def _MergeListsOfDicts(self, merged, shard):
    output = []
    for i in xrange(max(len(merged), len(shard))):
      merged_dict = merged[i] if i < len(merged) else {}
      shard_dict = shard[i] if i < len(shard) else {}
      output_dict = merged_dict.copy()
      output_dict.update(shard_dict)
      output.append(output_dict)
    return output

  def GetMergedTestResults(self, shard_results):
    """Merges the shards into one.

    Args:
      shard_results (list): A list of dicts with individual shard results.

    Returns:
      A dict with the following form:
      {
        'all_tests':[
          'AllForms/FormStructureBrowserTest.DataDrivenHeuristics/0',
          'AllForms/FormStructureBrowserTest.DataDrivenHeuristics/1',
          'AllForms/FormStructureBrowserTest.DataDrivenHeuristics/10',
          ...
        ]
        'per_iteration_data':[
          {
            'AllForms/FormStructureBrowserTest.DataDrivenHeuristics/109': [
              {
                'elapsed_time_ms': 4719,
                'losless_snippet': true,
                'output_snippet': '[ RUN      ] run outputs\\n',
                'output_snippet_base64': 'WyBSVU4gICAgICBdIEFsbEZvcm1zL0Zvcm1T'
                'status': 'SUCCESS'
              }
            ],
          },
          ...
        ]
      }
    """
    merged_results = {'all_tests': set(), 'per_iteration_data': []}
    for shard_result in shard_results:
      merged_results['all_tests'].update(shard_result.get('all_tests', []))
      merged_results['per_iteration_data'] = self._MergeListsOfDicts(
          merged_results['per_iteration_data'],
          shard_result.get('per_iteration_data', []))
    merged_results['all_tests'] = sorted(merged_results['all_tests'])
    return merged_results

  def GetFailedTestsInformation(self, test_results_log):
    """Parses the json data to get all the reliable failures' information."""
    failed_test_log = {}
    reliable_failed_tests = {}

    for iteration in (test_results_log.get('per_iteration_data') or []):
      for test_name in iteration.keys():

        if (any(test['status'] in _NON_FAILURE_STATUSES
                for test in iteration[test_name])):
          # Ignore the test if any of the attempts didn't fail.
          # If a test is skipped, that means it was not run at all.
          # Treats it as success since the status cannot be determined.
          continue

        # Stores the output to the step's log_data later.
        failed_test_log[test_name] = ''
        for test in iteration[test_name]:
          failed_test_log[test_name] = self.ConcatenateTestLog(
              failed_test_log[test_name], test.get('output_snippet_base64', ''))
        reliable_failed_tests[test_name] = self.RemoveAllPrefixes(test_name)

    return failed_test_log, reliable_failed_tests

  def IsTestResultUseful(self, test_results_log):
    """Checks if the log contains useful information."""
    # If this task doesn't have result, per_iteration_data will look like
    # [{}, {}, ...]
    return test_results_log and any(
        test_results_log.get('per_iteration_data') or [])

  def GetTestsRunStatuses(self, test_results_log):
    """Parses test results and gets accumulated test run statuses.

      Args:
      test_results_log (dict): A dict of all test results in the task.

    Returns:
      tests_statuses (dict): A dict of different statuses for each test.
    """
    tests_statuses = defaultdict(lambda: defaultdict(int))
    if test_results_log:
      for iteration in test_results_log.get('per_iteration_data'):
        for test_name, tests in iteration.iteritems():
          tests_statuses[test_name]['total_run'] += len(tests)
          for test in tests:
            tests_statuses[test_name][test['status']] += 1

    return tests_statuses
