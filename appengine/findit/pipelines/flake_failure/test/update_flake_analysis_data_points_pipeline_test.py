# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import mock

from dto.flake_swarming_task_output import FlakeSwarmingTaskOutput
from gae_libs.pipeline_wrapper import pipeline_handlers
from model.flake.master_flake_analysis import MasterFlakeAnalysis
from pipelines.flake_failure.determine_approximate_pass_rate_pipeline import (
    UpdateFlakeAnalysisDataPointsInput)
from pipelines.flake_failure.determine_approximate_pass_rate_pipeline import (
    UpdateFlakeAnalysisDataPointsPipeline)
from services import step_util
from services.flake_failure import data_point_util
from waterfall.build_info import BuildInfo
from waterfall.test.wf_testcase import WaterfallTestCase


class UpdateFlakeAnalysisDataPointsPipelineTest(WaterfallTestCase):
  app_module = pipeline_handlers._APP

  @mock.patch.object(data_point_util, 'UpdateAnalysisDataPoints')
  @mock.patch.object(step_util, 'GetValidBoundingBuildsForStep')
  def testUpdateFlakeAnalysisDataPointsPipeline(self, mocked_builds,
                                                mocked_update):
    master_name = 'm'
    builder_name = 'b'
    build_number = 123
    step_name = 's'
    test_name = 't'

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.Save()
    commit_position = 1000
    revision = 'r1000'
    swarming_task_output = FlakeSwarmingTaskOutput(
        error=None,
        pass_count=5,
        iterations=10,
        started_time=datetime(2018, 1, 1, 0, 0, 0),
        completed_time=datetime(2018, 1, 1, 1, 0, 0),
        task_id='task_id')

    lower_bound_build = BuildInfo(master_name, builder_name, build_number - 1)
    lower_bound_build.commit_position = 998
    lower_bound_build.chromium_revision = 'r998'
    upper_bound_build = BuildInfo(master_name, builder_name, build_number)
    upper_bound_build.commit_position = 1000
    upper_bound_build.chromium_revision = 'r1000'
    mocked_builds.return_value = (lower_bound_build, upper_bound_build)

    update_data_points_input = UpdateFlakeAnalysisDataPointsInput(
        analysis_urlsafe_key=analysis.key.urlsafe(),
        commit_position=commit_position,
        revision=revision,
        swarming_task_output=swarming_task_output)

    pipeline_job = UpdateFlakeAnalysisDataPointsPipeline(
        update_data_points_input)
    pipeline_job.start()
    self.execute_queued_tasks()

    mocked_builds.assert_called_once_with(master_name, builder_name, step_name,
                                          None, build_number, commit_position)
    mocked_update.assert_called_once_with(
        analysis.key.urlsafe(), commit_position, revision, swarming_task_output,
        lower_bound_build, upper_bound_build)
