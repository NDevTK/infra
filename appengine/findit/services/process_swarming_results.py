# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import logging
from collections import defaultdict
from waterfall import swarming_util


def CheckTestsRunStatuses(output_json, *_):
  """Checks result status for each test run and saves the numbers accordingly.

  Args:
    output_json (dict): A dict of all test results in the swarming task.

  Returns:
    tests_statuses (dict): A dict of different statuses for each test.

  Currently for each test, we are saving number of total runs,
  number of succeeded runs and number of failed runs.
  """
  tests_statuses = defaultdict(lambda: defaultdict(int))
  if output_json:
    for iteration in output_json.get('per_iteration_data'):
      for test_name, tests in iteration.iteritems():
        tests_statuses[test_name]['total_run'] += len(tests)
        for test in tests:
          tests_statuses[test_name][test['status']] += 1

  return tests_statuses


def GetSwarmingTaskError(task_id, result_data, result_error, log_data,
                         log_error):
  """Checks for error, and returns the error if any."""
  if result_error:
    return result_error

  if not result_data:
    logging.error(
        'No data was found for task_id %d even through task was completed.',
        task_id)
    return {
        'code': swarming_util.NO_OUTPUT_JSON,
        'message': 'No data was found even though task was completed.'
    }

  outputs_ref = result_data['outputs_ref']
  if not outputs_ref:
    logging.error('outputs_ref for task %s is None', task_id)
    return {
        'code': swarming_util.NO_TASK_OUTPUTS,
        'message': 'outputs_ref is None'
    }

  if log_error or not log_data:
    logging.error('output_json for task %s is None', task_id)
    return log_error or {
        'code': swarming_util.NO_OUTPUT_JSON,
        'message': 'output_json is None',
    }

  if not log_data.get('per_iteration_data'):
    logging.error('outputs_ref.per_iteration_data for task %s is None', task_id)
    return {
        'code': swarming_util.NO_PER_ITERATION_DATA,
        'message': 'per_iteration_data is empty or missing'
    }

  return None


def IsTaskComplete(data):
  """Returns True if the task is done running on swarming."""
  task_state = data['state']
  return (task_state is swarming_util.STATE_COMPLETED or
          task_state in swarming_util.STATES_NOT_RUNNING)