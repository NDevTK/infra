# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import defaultdict

from handlers import result_status
from model.base_build_model import BaseBuildModel
from model.wf_analysis import WfAnalysis
from model.wf_swarming_task import WfSwarmingTask
from waterfall import waterfall_config


def _GetResultAndFailureResultMap(master_name,
                                  builder_name,
                                  build_number,
                                  use_group_info=False):
  analysis = WfAnalysis.Get(master_name, builder_name, build_number)

  # If this analysis is part of a group, get the build analysis that opened the
  # group.
  if use_group_info and analysis and analysis.failure_group_key:
    analysis = WfAnalysis.Get(*analysis.failure_group_key)

  if not analysis:
    return None, None

  return analysis.result, analysis.failure_result_map


def _GetAllTestsForASwarmingTask(task_key, step_failure_result_map):
  all_tests = set()
  for test_name, test_task_key in step_failure_result_map.iteritems():
    if task_key == test_task_key:
      all_tests.add(test_name)
  return list(all_tests)


def _GenerateSwarmingTasksData(failure_result_map):
  """Collects info for all related swarming tasks.

  Returns: A dict as below:
      {
          'step1': {
              'swarming_tasks': {
                  'm/b/121': {
                      'task_info': {
                          'status': 'Completed',
                          'task_id': 'task1',
                          'task_url': ('https://chromium-swarm.appspot.com/user'
                                       '/task/task1')
                      },
                      'all_tests': ['test2', 'test3', 'test4'],
                      'reliable_tests': ['test2'],
                      'flaky_tests': ['test3', 'test4']
                  }
              }
          },
          'step2': {
              'swarming_tasks': {
                  'm/b/121': {
                      'task_info': {
                          'status': 'Pending'
                      },
                      'all_tests': ['test1']
                  }
              }
          },
          'step3': {
              'swarming_tasks': {
                  'm/b/121': {
                      'task_info': {
                          'status': 'No swarming rerun found'
                      },
                      'all_tests': ['test1']
                  }
              }
          }
      }
  """

  tasks_info = defaultdict(lambda: defaultdict(lambda: defaultdict(dict)))

  swarming_server = waterfall_config.GetSwarmingSettings()['server_host']

  for step_name, failure in failure_result_map.iteritems():
    step_tasks_info = tasks_info[step_name]['swarming_tasks']

    if isinstance(failure, dict):
      # Only swarming test failures have swarming re-runs.
      swarming_task_keys = set(failure.values())

      for key in swarming_task_keys:
        task_dict = step_tasks_info[key]
        referred_build_keys = BaseBuildModel.GetBuildInfoFromBuildKey(key)
        task = WfSwarmingTask.Get(*referred_build_keys, step_name=step_name)
        all_tests = _GetAllTestsForASwarmingTask(key, failure)
        task_dict['all_tests'] = all_tests
        if not task:  # In case task got manually removed from data store.
          task_info = {'status': result_status.NO_SWARMING_TASK_FOUND}
        else:
          task_info = {'status': task.status}

          # Get the step name without platform.
          # This value should have been saved in task.parameters;
          # in case of no such value saved, split the step_name.
          task_dict['ref_name'] = (
              step_name.split()[0]
              if not task.parameters or not task.parameters.get('ref_name') else
              task.parameters['ref_name'])

          if task.task_id:  # Swarming rerun has started.
            task_info['task_id'] = task.task_id
            task_info['task_url'] = 'https://%s/user/task/%s' % (
                swarming_server, task.task_id)
          if task.classified_tests:
            # Swarming rerun has completed.
            # Use its result to get reliable and flaky tests.
            # If task has not completed, there will be no try job yet,
            # the result will be grouped in unclassified failures temporarily.
            reliable_tests = task.classified_tests.get('reliable_tests', [])
            task_dict['reliable_tests'] = [
                test for test in reliable_tests if test in all_tests
            ]
            flaky_tests = task.classified_tests.get('flaky_tests', [])
            task_dict['flaky_tests'] = [
                test for test in flaky_tests if test in all_tests
            ]

        task_dict['task_info'] = task_info
    else:
      step_tasks_info[failure] = {
          'task_info': {
              'status': result_status.NON_SWARMING_NO_RERUN
          }
      }

  return tasks_info


def GetSwarmingTaskInfo(master_name, builder_name, build_number):
  _, failure_result_map = _GetResultAndFailureResultMap(
      master_name, builder_name, build_number)
  return (_GenerateSwarmingTasksData(failure_result_map)
          if failure_result_map else {})
