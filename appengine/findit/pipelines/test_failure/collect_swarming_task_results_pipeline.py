# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging

from common import exceptions
from dto.collect_swarming_task_results_inputs import (
    CollectSwarmingTaskResultsInputs)
from dto.collect_swarming_task_results_outputs import (
    CollectSwarmingTaskResultsOutputs)
from gae_libs.pipelines import AsynchronousPipeline
from services.test_failure import test_swarming

_COUNTDOWN_LIMIT = 5 * 60


def _GetCountDown(callback_count):
  """Gets how long should next task waits.

  The wait time increases linearly until hits the limit.
  """
  return callback_count * 60 if callback_count < 5 else _COUNTDOWN_LIMIT


class CollectSwarmingTaskResultsPipeline(AsynchronousPipeline):
  """A pipeline to collect results of all swarming reruns.
  """

  input_type = CollectSwarmingTaskResultsInputs
  output_type = CollectSwarmingTaskResultsOutputs

  def TimeoutSeconds(self):
    return 2 * 60 * 60  # 2 hours. This will enable a timeout callback.

  def OnTimeout(
      self, collect_consistent_failure_params, parameters):  # pragma: no cover.
    pass

  def RunImpl(self, collect_consistent_failure_params):
    if self.GetCallbackParameters().get('steps'):
      # For idempotent operation.
      logging.warning(
          'RunImpl invoked again for collecting swarming task results.')
      return

    steps = test_swarming.GetsStepsToCollectSwarmingTaskResults(
        collect_consistent_failure_params)

    if not steps:
      # Either build has not completed yet or no first time failure in build.
      # Bail out.
      return

    self.SaveCallbackParameters({'steps': steps, 'callback_count': 0})

    # Schedules a callback immediately.
    self.ScheduleCallbackTask(countdown=0)

  def CallbackImpl(self, collect_consistent_failure_params, parameters):
    """Checks the WfSwarmingTask entities to get consistently failed tests."""
    steps = parameters['steps']
    callback_count = parameters['callback_count'] + 1
    try:
      consistent_failures = test_swarming.CollectSwarmingTaskResults(
          collect_consistent_failure_params, steps)
      if not consistent_failures:
        self.SaveCallbackParameters({
            'steps': steps,
            'callback_count': callback_count
        })
        self.ScheduleCallbackTask(countdown=_GetCountDown(callback_count))
        return None
      return None, consistent_failures
    except exceptions.RetryException as e:  # Indicate an error to retry.
      return ('Error on updating swarming task result: %s' % e.error_message,
              None)
