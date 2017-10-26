# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import mock

from gae_libs.pipeline_wrapper import pipeline_handlers
from gae_libs.pipelines import CreateInputObjectInstance
from model.flake.master_flake_analysis import MasterFlakeAnalysis
from pipelines.flake_failure.get_sha_for_commit_position_pipeline import (
    GetShaForCommitPositionPipeline)
from pipelines.flake_failure.get_sha_for_commit_position_pipeline import (
    GetShaForCommitPositionPipelineInputObject)
from waterfall import build_util
from waterfall import swarming_util
from waterfall.build_info import BuildInfo
from waterfall.test.wf_testcase import WaterfallTestCase


class GetShaForCommitPositionPipelineTest(WaterfallTestCase):
  app_module = pipeline_handlers._APP

  @mock.patch.object(build_util, 'GetEarliestContainingBuild')
  @mock.patch.object(swarming_util, 'GetIsolatedShaForStep')
  def testGetShaForCommitPositionPipeline(self, mocked_get_isolated_sha,
                                          mocked_get_build_info):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    requested_commit_position = 1000

    build = BuildInfo(master_name, builder_name, build_number)
    build.commit_position = requested_commit_position
    mocked_get_build_info.return_value = build

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.Save()

    get_sha_input = CreateInputObjectInstance(
        GetShaForCommitPositionPipelineInputObject,
        analysis_urlsafe_key=unicode(analysis.key.urlsafe()),
        commit_position=requested_commit_position)

    pipeline_job = GetShaForCommitPositionPipeline(get_sha_input)
    pipeline_job.start()
    self.execute_queued_tasks()

    self.assertTrue(mocked_get_build_info.called)
    self.assertTrue(mocked_get_isolated_sha.called)
