# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import mock

from dto.flakiness import Flakiness
from gae_libs.pipeline_wrapper import pipeline_handlers
from libs import time_util
from libs.list_of_basestring import ListOfBasestring
from model.flake.analysis.master_flake_analysis import DataPoint
from model.flake.analysis.master_flake_analysis import MasterFlakeAnalysis
from pipelines.flake_failure.save_flakiness_verification_pipeline import (
    SaveFlakinessVerificationInput)
from pipelines.flake_failure.save_flakiness_verification_pipeline import (
    SaveFlakinessVerificationPipeline)
from waterfall.test.wf_testcase import WaterfallTestCase


class SaveFlakinessVerificationPipelineTest(WaterfallTestCase):
  app_module = pipeline_handlers._APP

  @mock.patch.object(time_util, 'GetUTCNow')
  def testSaveFlakinessVerificationPipeline(self, mocked_updated_time):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.Save()

    commit_position = 1000
    pass_rate = 0.5
    expected_updated_time = datetime(2018, 9, 18, 0, 0, 0)
    mocked_updated_time.return_value = expected_updated_time

    flakiness = Flakiness(
        build_number=None,
        build_url='url',
        commit_position=commit_position,
        total_test_run_seconds=100,
        error=None,
        failed_swarming_task_attempts=0,
        iterations=50,
        pass_rate=pass_rate,
        revision='r1000',
        try_job_url=None,
        task_ids=ListOfBasestring.FromSerializable(['task_id']))

    expected_data_point = DataPoint.Create(
        build_number=None,
        build_url='url',
        commit_position=commit_position,
        elapsed_seconds=100,
        error=None,
        failed_swarming_task_attempts=0,
        iterations=50,
        pass_rate=pass_rate,
        git_hash='r1000',
        try_job_url=None,
        task_ids=['task_id'],
        updated_time=expected_updated_time)

    data_point_input = SaveFlakinessVerificationInput(
        analysis_urlsafe_key=analysis.key.urlsafe(), flakiness=flakiness)

    pipeline_job = SaveFlakinessVerificationPipeline(data_point_input)
    pipeline_job.start()
    self.execute_queued_tasks()

    self.assertEqual(1, len(analysis.flakiness_verification_data_points))
    self.assertEqual(expected_data_point,
                     analysis.flakiness_verification_data_points[0])
