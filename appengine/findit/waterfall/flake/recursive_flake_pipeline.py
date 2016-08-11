# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
from common import appengine_util
from common import constants
from common.pipeline_wrapper import BasePipeline

from model import analysis_status
from model.flake.master_flake_analysis import MasterFlakeAnalysis
from model.flake.flake_swarming_task import FlakeSwarmingTask
from waterfall.trigger_flake_swarming_task_pipeline import (
    TriggerFlakeSwarmingTaskPipeline)
from waterfall.process_flake_swarming_task_result_pipeline import (
    ProcessFlakeSwarmingTaskResultPipeline)


class RecursiveFlakePipeline(BasePipeline):

  # Arguments number differs from overridden method - pylint: disable=W0221
  def run(self, master_name, builder_name, run_build_number, step_name,
          test_name, master_build_number, algo_dict,
          queue_name=constants.DEFAULT_QUEUE):
    # Call trigger pipeline (flake style).
    task_id = yield TriggerFlakeSwarmingTaskPipeline(
        master_name, builder_name, run_build_number, step_name, [test_name])
    # Pass the trigger pipeline into a process pipeline.
    test_result_future = yield ProcessFlakeSwarmingTaskResultPipeline(
        master_name, builder_name, run_build_number,
        step_name, task_id, master_build_number, test_name)
    yield NextBuildNumberPipeline(
        master_name, builder_name, master_build_number, run_build_number,
        step_name, test_name, test_result_future, queue_name, algo_dict)


class NextBuildNumberPipeline(BasePipeline):

  # Arguments number differs from overridden method - pylint: disable=W0221
  # Unused argument - pylint: disable=W0613
  def run(self, master_name, builder_name, master_build_number,
          run_build_number, step_name, test_name, test_result_future,
          queue_name, algo_dict):

    # Get MasterFlakeAnalysis success list corresponding to parameters.
    master = MasterFlakeAnalysis.Get(master_name, builder_name,
                                     master_build_number, step_name, test_name)

    # Don't call another pipeline if we fail.
    flake_swarming_task = FlakeSwarmingTask.Get(
        master_name, builder_name, run_build_number, step_name, test_name)
    if flake_swarming_task.status == analysis_status.ERROR:
      return

    # Figure out what build_number we should call, if any
    # This is a placeholder for testing:

    # Get the last result.
    last_result = master.success_rates[-1]
    if last_result < .02 or last_result > .98:
      algo_dict['stable_in_a_row'] += 1
      if algo_dict['stable_in_a_row'] > 4:
        algo_dict['stabled_out'] = True
      algo_dict['flakes_in_a_row'] = 0
      step_size = algo_dict['stable_in_a_row'] + 1
    else:
      algo_dict['flakes_in_a_row'] += 1
      if algo_dict['flakes_in_a_row'] > 4:
        algo_dict['flaked_out'] = True
      algo_dict['stable_in_a_row'] = 0
      step_size = algo_dict['flakes_in_a_row'] + 1
    next_run = min(master.build_numbers) - step_size

    if (next_run < algo_dict['last_build_number'] or
        (algo_dict['stabled_out'] and algo_dict['flaked_out'])):
      next_run = False

    if next_run:
      new_algo_dict = {
          'flakes_in_a_row': algo_dict['flakes_in_a_row'],
          'stable_in_a_row': algo_dict['stable_in_a_row'],
          'stabled_out': algo_dict['stabled_out'],
          'flaked_out': algo_dict['flaked_out'],
          'last_build_number': algo_dict['last_build_number']
      }
      pipeline_job = RecursiveFlakePipeline(
          master_name, builder_name, next_run, step_name, test_name,
          master_build_number, algo_dict=new_algo_dict)
      # pylint: disable=W0201
      pipeline_job.target = appengine_util.GetTargetNameForModule(
          constants.WATERFALL_BACKEND)
      pipeline_job.start(queue_name=queue_name)
