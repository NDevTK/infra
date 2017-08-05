# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import mock

from google.appengine.ext import ndb

from common.waterfall import buildbucket_client
from common.waterfall import failure_type
from model.flake.flake_try_job import FlakeTryJob
from model.flake.flake_try_job_data import FlakeTryJobData
from model.wf_build import WfBuild
from waterfall import schedule_try_job_pipeline
from waterfall.flake.schedule_flake_try_job_pipeline import (
    ScheduleFlakeTryJobPipeline)
from waterfall import swarming_util
from waterfall.test import wf_testcase


class ScheduleFlakeTryJobPipelineTest(wf_testcase.WaterfallTestCase):

  def setUp(self):
    super(ScheduleFlakeTryJobPipelineTest, self).setUp()
    self.mock_select = mock.patch('waterfall.swarming_util.AssignWarmCacheHost')
    self.mock_select.start()

  def tearDown(self):
    self.mock_select.stop()
    super(ScheduleFlakeTryJobPipelineTest, self).tearDown()

  def testGetBuildProperties(self):
    self.UpdateUnitTestConfigSettings(
        config_property='check_flake_settings',
        override_data={'iterations_to_rerun': 100})
    master_name = 'm'
    builder_name = 'b'
    step_name = 's'
    test_name = 't'
    git_hash = 'a1b2c3d4'
    iterations = 200

    expected_properties = {
        'recipe': 'findit/chromium/flake',
        'target_mastername': master_name,
        'target_testername': builder_name,
        'test_revision': git_hash,
        'test_repeat_count': 200,
        'tests': {
            step_name: [test_name]
        }
    }

    try_job_pipeline = ScheduleFlakeTryJobPipeline()
    properties = try_job_pipeline._GetBuildProperties(
        master_name, builder_name, step_name, test_name, git_hash, iterations)

    self.assertEqual(properties, expected_properties)

  def testCreateTryJobData(self):
    master_name = 'm'
    builder_name = 'b'
    step_name = 's'
    test_name = 't'
    git_hash = 'a1b2c3d4'
    build_id = 'build_id'
    analysis_key = ndb.Key('key', 1)

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, git_hash)
    ScheduleFlakeTryJobPipeline()._CreateTryJobData(build_id, try_job.key,
                                                    analysis_key.urlsafe())

    try_job_data = FlakeTryJobData.Get(build_id)

    self.assertEqual(try_job_data.try_job_key, try_job.key)

  @mock.patch.object(
      swarming_util,
      'GetETAToStartAnalysis',
      return_value=datetime.datetime(1, 1, 1))
  @mock.patch.object(schedule_try_job_pipeline, 'buildbucket_client')
  def testScheduleFlakeTryJob(self, mock_module, *_):
    master_name = 'm'
    builder_name = 'b'
    build_number = 1
    step_name = 's'
    test_name = 't'
    git_hash = 'a1b2c3d4'
    build_id = '1'
    url = 'url'
    analysis_key = ndb.Key('key', 1)
    build = WfBuild.Create(master_name, builder_name, build_number)
    build.data = {
        'properties': {
            'parent_mastername': 'pm',
            'parent_buildername': 'pb'
        }
    }
    build.put()
    response = {
        'build': {
            'id': build_id,
            'url': url,
            'status': 'SCHEDULED',
        }
    }
    results = [(None, buildbucket_client.BuildbucketBuild(response['build']))]
    mock_module.TriggerTryJobs.return_value = results

    FlakeTryJob.Create(master_name, builder_name, step_name, test_name,
                       git_hash).put()

    try_job_pipeline = ScheduleFlakeTryJobPipeline()
    try_job_pipeline.start_test()
    try_job_pipeline.run(master_name, builder_name, step_name, test_name,
                         git_hash, analysis_key.urlsafe(), None, None)
    try_job_pipeline.finalized()
    # Reload from ID to get all internal properties in sync.
    try_job_pipeline = ScheduleFlakeTryJobPipeline.from_id(
        try_job_pipeline.pipeline_id)
    try_job_id = try_job_pipeline.outputs.default.value

    try_job = FlakeTryJob.Get(master_name, builder_name, step_name, test_name,
                              git_hash)
    try_job_data = FlakeTryJobData.Get(build_id)

    self.assertEqual(build_id, try_job_id)
    self.assertEqual(build_id, try_job.flake_results[-1]['try_job_id'])
    self.assertTrue(build_id in try_job.try_job_ids)
    self.assertEqual(try_job_data.try_job_key, try_job.key)
    self.assertEqual(analysis_key, try_job_data.analysis_key)

  @mock.patch.object(schedule_try_job_pipeline, 'buildbucket_client')
  def testTriggerTryJob(self, mock_module):
    master_name = 'm'
    builder_name = 'b'
    build_number = 1
    build = WfBuild.Create(master_name, builder_name, build_number)
    build.data = {
        'properties': {
            'parent_mastername': 'pm',
            'parent_buildername': 'pb'
        }
    }
    build.put()
    response = {
        'build': {
            'id': '1',
            'url': 'url',
            'status': 'SCHEDULED',
        }
    }
    results = [(None, buildbucket_client.BuildbucketBuild(response['build']))]
    mock_module.TriggerTryJobs.return_value = results

    build_id = ScheduleFlakeTryJobPipeline()._TriggerTryJob(
        master_name, builder_name, {}, [],
        failure_type.GetDescriptionForFailureType(failure_type.FLAKY_TEST),
        None, None)

    self.assertEqual(build_id, '1')

  @mock.patch.object(
      buildbucket_client, 'IsSwarmbucketMaster', return_value=True)
  @mock.patch.object(
      swarming_util,
      'GetSwarmingBotCounts',
      side_effect=[{
          'available': 1
      }, {
          'available': 3
      }])
  @mock.patch.object(
      swarming_util,
      'GetETAToStartAnalysis',
      side_effect=[datetime.datetime(2100, 1, 1),
                   datetime.datetime(1, 1, 1)])
  @mock.patch.object(schedule_try_job_pipeline, 'buildbucket_client')
  def testDelayedScheduleFlakeTryJob(self, mock_module, *_):
    master_name = 'm'
    builder_name = 'b'
    build_number = 1
    step_name = 's'
    test_name = 't'
    git_hash = 'a1b2c3d4'
    build_id = '1'
    url = 'url'
    analysis_key = ndb.Key('key', 1)
    build = WfBuild.Create(master_name, builder_name, build_number)
    build.data = {
        'properties': {
            'parent_mastername': 'pm',
            'parent_buildername': 'pb'
        }
    }
    build.put()
    response = {
        'build': {
            'id': build_id,
            'url': url,
            'status': 'SCHEDULED',
        }
    }
    results = [(None, buildbucket_client.BuildbucketBuild(response['build']))]
    mock_module.TriggerTryJobs.return_value = results

    FlakeTryJob.Create(master_name, builder_name, step_name, test_name,
                       git_hash).put()

    try_job_pipeline = ScheduleFlakeTryJobPipeline()
    try_job_pipeline.start_test()
    try_job_pipeline.run(master_name, builder_name, step_name, test_name,
                         git_hash, analysis_key.urlsafe(), None, None)
    try_job_pipeline.callback(master_name, builder_name, step_name, test_name,
                              git_hash, analysis_key.urlsafe(), None, None, 120)
    try_job_pipeline.finalized()
    # Reload from ID to get all internal properties in sync.
    try_job_pipeline = ScheduleFlakeTryJobPipeline.from_id(
        try_job_pipeline.pipeline_id)
    try_job_id = try_job_pipeline.outputs.default.value

    try_job = FlakeTryJob.Get(master_name, builder_name, step_name, test_name,
                              git_hash)
    try_job_data = FlakeTryJobData.Get(build_id)

    self.assertEqual(build_id, try_job_id)
    self.assertEqual(build_id, try_job.flake_results[-1]['try_job_id'])
    self.assertTrue(build_id in try_job.try_job_ids)
    self.assertEqual(try_job_data.try_job_key, try_job.key)
    self.assertEqual(analysis_key, try_job_data.analysis_key)
