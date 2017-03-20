# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import defaultdict
import datetime
import logging
import time

from gae_libs.http.http_client_appengine import HttpClientAppengine
from common.pipeline_wrapper import BasePipeline
from model import analysis_status
from waterfall import swarming_util
from waterfall import waterfall_config
from waterfall.trigger_base_swarming_task_pipeline import NO_TASK


class ProcessBaseSwarmingTaskResultPipeline(BasePipeline):
  """A pipeline for monitoring swarming task and processing task result.

  This pipeline waits for result for a swarming task and processes the result to
  generate a dict for statuses for each test run.
  """

  HTTP_CLIENT = HttpClientAppengine()
  # Making this pipeline asynchronous by setting this class variable.
  async = True

  def __init__(self, *args, **kwargs):
    super(ProcessBaseSwarmingTaskResultPipeline, self).__init__(*args, **kwargs)
    # This attribute is meant for use by the unittest only.
    self.last_params = {}

  def _CheckTestsRunStatuses(self, output_json, *_):
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

  def _GetSwarmingTask(self):
    # Get the appropriate kind of Swarming Task (Wf or Flake).
    # Should be overwritten by subclass.
    raise NotImplementedError(
        '_GetSwarmingTask should be implemented in the child class')

  def _GetArgs(self):
    # Return list of arguments to call _CheckTestsRunStatuses with - output_json
    # Should be overwritten by subclass.
    raise NotImplementedError(
        '_GetArgs should be implemented in the child class')

  def _ConvertDateTime(self, time_string):
    """Convert UTC time string to datetime.datetime."""
    if not time_string:
      return None
    for fmt in ('%Y-%m-%dT%H:%M:%S.%f', '%Y-%m-%dT%H:%M:%S'):
      # When microseconds are 0, the '.123456' suffix is elided.
      try:
        return datetime.datetime.strptime(time_string, fmt)
      except ValueError:
        pass
    raise ValueError('Failed to parse %s' % time_string)  # pragma: no cover

  def delay_callback(self, **kwargs):  # pragma: no cover
    self.last_params = kwargs
    countdown = kwargs.get('server_query_interval_seconds', 60)
    task = self.get_callback_task(countdown=countdown, params=kwargs)
    task.add(self.queue_name)

  def _GetPipelineResult(self, step_name, step_name_no_platform, task):
    # The sub-classes may use properties of the task as part of the result.
    _ = task
    return step_name, step_name_no_platform

  # Arguments number differs from overridden method - pylint: disable=W0221
  def callback(self, task_id, step_name, call_args, deadline,
               server_query_interval_seconds, task_started, task_completed,
               step_name_no_platform, pipeline_id=None):
    """Monitors the swarming task and waits for it to complete."""
    _ = pipeline_id  # We don't do anything with this id.
    assert task_id
    task = self._GetSwarmingTask(*call_args)

    def check_task_completion():
      if task_completed and data is not None:
        task.created_time = (task.created_time or
                             self._ConvertDateTime(data.get('created_ts')))
        task.started_time = (task.started_time or
                             self._ConvertDateTime(data.get('started_ts')))
        task.completed_time = (task.completed_time or
                               self._ConvertDateTime(data.get('completed_ts')))
        task.put()
        pipeline_result = self._GetPipelineResult(
            step_name, step_name_no_platform, task)
        self.complete(pipeline_result)
      elif time.time() > deadline:  # pragma: no cover
        # Timeout.
        # Updates status as ERROR.
        task.status = analysis_status.ERROR
        task.error = {
            'code': swarming_util.TIMED_OUT,
            'message': 'Process swarming task result timed out'
        }
        task.put()
        timeout_hours = waterfall_config.GetSwarmingSettings().get(
            'task_timeout_hours')
        logging.error('Swarming task timed out after %d hours.' % timeout_hours)
        pipeline_result = self._GetPipelineResult(
            step_name, step_name_no_platform, task)
        self.complete(pipeline_result)
      else:
        self.last_params = {
            'task_id': task_id,
            'step_name': step_name,
            'call_args': call_args,
            'deadline': deadline,
            'server_query_interval_seconds': server_query_interval_seconds,
            'task_started': task_started,
            'task_completed': task_completed,
            'step_name_no_platform': step_name_no_platform,
        }
      # Update the stored callback url with possibly modified params.
        new_callback_url = self.get_callback_url(**self.last_params)
        if task.callback_url != new_callback_url:  # pragma: no cover
          task.callback_url = new_callback_url
          task.put()
        # TODO(robertocn): Remove this line when the system reliably receives
        # notifications from swarming via pubsub.
        self.delay_callback(**self.last_params)

    data, error = swarming_util.GetSwarmingTaskResultById(
        task_id, self.HTTP_CLIENT)

    if error:
      # An error occurred at some point when trying to retrieve data from
      # the swarming server, even if eventually successful.
      task.error = error
      task.put()

      if not data:
        # Even after retry, no data was recieved.
        task.status = analysis_status.ERROR
        task.put()
        check_task_completion()
        return

    task_state = data['state']
    exit_code = (data.get('exit_code') if
                 task_state == swarming_util.STATE_COMPLETED else None)
    step_name_no_platform = (
        step_name_no_platform or swarming_util.GetTagValue(
            data.get('tags', {}), 'ref_name'))

    if task_state not in swarming_util.STATES_RUNNING:
      task_completed = True

      if (task_state == swarming_util.STATE_COMPLETED and
          int(exit_code) != swarming_util.TASK_FAILED):
        outputs_ref = data.get('outputs_ref')

        # If swarming task aborted because of errors in request arguments,
        # it's possible that there is no outputs_ref.
        if not outputs_ref:
          task.status = analysis_status.ERROR
          task.error = {
              'code': swarming_util.NO_TASK_OUTPUTS,
              'message': 'outputs_ref is None'
          }
          task.put()
          check_task_completion()
          return

        output_json, error = swarming_util.GetSwarmingTaskFailureLog(
            outputs_ref, self.HTTP_CLIENT)

        task.status = analysis_status.COMPLETED

        if error:
          task.error = error

          if not output_json:
            # Retry was ultimately unsuccessful.
            task.status = analysis_status.ERROR

        tests_statuses = self._CheckTestsRunStatuses(output_json, *call_args)
        task.tests_statuses = tests_statuses
        task.put()
      else:
        if exit_code is not None:
          # Swarming task completed, but the task failed.
          code = int(exit_code)
          message = swarming_util.EXIT_CODE_DESCRIPTIONS[code]
        else:
          # The swarming task did not complete.
          code = swarming_util.STATES_NOT_RUNNING_TO_ERROR_CODES[task_state]
          message = task_state

        task.status = analysis_status.ERROR
        task.error = {
            'code': code,
            'message': message
        }
        task.put()

        logging_str = 'Swarming task stopped with status: %s' % task_state
        if exit_code:  # pragma: no cover
          logging_str += ' and exit_code: %s - %s' % (
              exit_code, swarming_util.EXIT_CODE_DESCRIPTIONS[code])
        logging.error(logging_str)

      tags = data.get('tags', {})
      priority_str = swarming_util.GetTagValue(tags, 'priority')
      if priority_str:
        task.parameters['priority'] = int(priority_str)
      task.put()
    else:  # pragma: no cover
      if task_state == 'RUNNING' and not task_started:
        # swarming task just starts, update status.
        task_started = True
        task.status = analysis_status.RUNNING
        task.put()
    check_task_completion()

  # Arguments number differs from overridden method - pylint: disable=W0221
  def run(self, master_name, builder_name, build_number, step_name,
          task_id=None, *args):
    """Monitors a swarming task.

    Args:
      master_name (str): The master name.
      builder_name (str): The builder name.
      build_number (str): The build number.
      step_name (str): The failed test step name.
      task_id (str): The task id to query the swarming server on the progresss
        of a swarming task.
    """
    call_args = self._GetArgs(master_name, builder_name, build_number,
                              step_name, *args)
    task = self._GetSwarmingTask(*call_args)
    if not task_id:
      task_id = task.task_id

    # Check to make this method idempotent.
    if task.callback_url and self.pipeline_id in task.callback_url:
      return

    timeout_hours = waterfall_config.GetSwarmingSettings().get(
        'task_timeout_hours')
    deadline = time.time() + timeout_hours * 60 * 60
    server_query_interval_seconds = waterfall_config.GetSwarmingSettings().get(
        'server_query_interval_seconds')
    task_started = False
    task_completed = False
    step_name_no_platform = None

    if task_id.lower() == NO_TASK:  # pragma: no branch
      # This situation happens in flake analysis: if the step with flaky test
      # didn't exist in checked build, we should skip the build.
      task = self._GetSwarmingTask(*call_args)
      task.task_id = None
      task.status = analysis_status.SKIPPED
      task.put()
      self._UpdateMasterFlakeAnalysis(
          *call_args, pass_rate=-1, flake_swarming_task=task)
      self.complete(self._GetPipelineResult(
            step_name, step_name_no_platform, task))
      return

    self.last_params = {
        'task_id': task_id,
        'step_name': step_name,
        'call_args': call_args,
        'deadline': deadline,
        'server_query_interval_seconds': server_query_interval_seconds,
        'task_started': task_started,
        'task_completed': task_completed,
        'step_name_no_platform': step_name_no_platform,
    }

    task.callback_url = self.get_callback_url(**self.last_params)
    task.put()

    self.callback(**self.last_params)
