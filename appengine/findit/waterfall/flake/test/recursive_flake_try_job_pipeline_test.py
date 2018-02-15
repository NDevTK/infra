# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import copy
from datetime import datetime
import mock

from google.appengine.ext import ndb

from common import constants
from common.waterfall import failure_type
from dto.test_location import TestLocation
from gae_libs.gitiles.cached_gitiles_repository import CachedGitilesRepository
from gae_libs.pipelines import CreateInputObjectInstance
from gae_libs.pipelines import pipeline_handlers
from libs import analysis_status
from libs.gitiles.change_log import ChangeLog
from model.flake.flake_culprit import FlakeCulprit
from model.flake.flake_try_job import FlakeTryJob
from model.flake.flake_try_job_data import FlakeTryJobData
from model.flake.master_flake_analysis import DataPoint
from model.flake.master_flake_analysis import MasterFlakeAnalysis
from pipelines.flake_failure.create_bug_for_flake_pipeline import (
    CreateBugForFlakePipeline)
from pipelines.flake_failure.create_bug_for_flake_pipeline import (
    CreateBugForFlakePipelineInputObject)
from services import swarmbot_util
from services import swarmed_test_util
from services.flake_failure import flake_try_job
from waterfall.flake import confidence
from waterfall.flake import flake_constants
from waterfall.flake import recursive_flake_try_job_pipeline
from waterfall.flake.recursive_flake_try_job_pipeline import (
    _GetNextCommitPositionAndRemainingSuspects)
from waterfall.flake.recursive_flake_try_job_pipeline import (
    _GetNormalizedTryJobDataPoints)
from waterfall.flake.recursive_flake_try_job_pipeline import (
    NextCommitPositionPipeline)
from waterfall.flake.recursive_flake_try_job_pipeline import (
    RecursiveFlakeTryJobPipeline)
from waterfall.flake.send_notification_for_flake_culprit_pipeline import (
    SendNotificationForFlakeCulpritPipeline)
from waterfall.flake.update_flake_bug_pipeline import UpdateFlakeBugPipeline
from waterfall.test import wf_testcase
from waterfall.test.wf_testcase import DEFAULT_CONFIG_DATA

_DEFAULT_CACHE_NAME = swarmbot_util.GetCacheName(None, None)


class RecursiveFlakeTryJobPipelineTest(wf_testcase.WaterfallTestCase):
  app_module = pipeline_handlers._APP

  def testRecursiveFlakeTryJobPipeline(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    upper_bound_commit_position = 1000
    start_commit_position = 999
    revision = 'r999'
    try_job_id = 'try_job_id'
    lower_bound_commit_position = 998
    user_specified_iterations = None
    remaining_suggested_commits = []

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.status = analysis_status.COMPLETED
    analysis.algorithm_parameters = copy.deepcopy(
        DEFAULT_CONFIG_DATA['check_flake_settings'])
    analysis.Save()

    iterations_to_rerun = analysis.algorithm_parameters.get(
        'try_job_rerun', {}).get('iterations_to_rerun')

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, revision)

    try_job_result = {
        revision: {
            step_name: {
                'status': 'failed',
                'failures': [test_name],
                'valid': True,
                'pass_fail_counts': {
                    'test_name': {
                        'pass_count': 28,
                        'fail_count': 72,
                    }
                }
            }
        }
    }

    self.MockPipeline(
        recursive_flake_try_job_pipeline.ScheduleFlakeTryJobPipeline,
        try_job_id,
        expected_args=[
            master_name, builder_name, step_name, test_name, revision,
            analysis.key.urlsafe(), _DEFAULT_CACHE_NAME, None,
            iterations_to_rerun
        ])
    self.MockPipeline(
        recursive_flake_try_job_pipeline.MonitorTryJobPipeline,
        try_job_result,
        expected_args=[
            try_job.key.urlsafe(), failure_type.FLAKY_TEST, try_job_id
        ])
    self.MockPipeline(
        recursive_flake_try_job_pipeline.ProcessFlakeTryJobResultPipeline,
        None,
        expected_args=[
            revision, start_commit_position,
            try_job.key.urlsafe(),
            analysis.key.urlsafe()
        ])
    self.MockPipeline(
        recursive_flake_try_job_pipeline.NextCommitPositionPipeline,
        '',
        expected_args=[
            analysis.key.urlsafe(),
            try_job.key.urlsafe(), remaining_suggested_commits,
            start_commit_position, lower_bound_commit_position,
            upper_bound_commit_position, user_specified_iterations,
            _DEFAULT_CACHE_NAME, None, False
        ])

    pipeline_job = RecursiveFlakeTryJobPipeline(
        analysis.key.urlsafe(), remaining_suggested_commits,
        start_commit_position, revision, lower_bound_commit_position,
        upper_bound_commit_position, None, _DEFAULT_CACHE_NAME, None, False)
    pipeline_job.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()

    self.assertIsNotNone(
        FlakeTryJob.Get(master_name, builder_name, step_name, test_name,
                        revision))
    self.assertEqual(analysis.last_attempted_revision, revision)
    self.assertIsNone(analysis.last_attempted_swarming_task_id)

  @mock.patch.object(flake_try_job, 'GetSwarmingTaskIdForTryJob')
  @mock.patch.object(
      recursive_flake_try_job_pipeline, '_NeedANewTryJob', return_value=False)
  def testRecursiveFlakeTryJobPipelineUserRerunRangeReuseDataPoint(
      self, _, mocked_get_swarming_task_id):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    upper_bound_commit_position = 1000
    lower_bound_commit_position = 998
    start_commit_position = 999
    user_specified_iterations = 200
    revision = 'r999'
    try_job_id = 'try_job_id'
    remaining_suggested_commits = []
    mocked_get_swarming_task_id.return_value = try_job_id

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.status = analysis_status.COMPLETED
    analysis.algorithm_parameters = copy.deepcopy(
        DEFAULT_CONFIG_DATA['check_flake_settings'])
    analysis.Save()

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, revision)
    try_job_result = {
        revision: {
            step_name: {
                'status': 'failed',
                'failures': [test_name],
                'valid': True,
                'pass_fail_counts': {
                    test_name: {
                        'pass_count': 28,
                        'fail_count': 72,
                    }
                }
            }
        }
    }
    report = {'report': {'result': try_job_result}}
    try_job.flake_results.append(report)
    try_job.try_job_ids = [try_job_id]
    try_job.put()

    try_job_data = FlakeTryJobData.Create(try_job_id)
    try_job_data.start_time = datetime(2017, 10, 17, 1, 0, 0)
    try_job_data.end_time = datetime(2017, 10, 17, 2, 0, 0)
    try_job_data.try_job_key = try_job.key
    try_job_data.put()

    self.MockPipeline(
        recursive_flake_try_job_pipeline.NextCommitPositionPipeline,
        '',
        expected_args=[
            analysis.key.urlsafe(),
            try_job.key.urlsafe(), remaining_suggested_commits,
            start_commit_position, lower_bound_commit_position,
            upper_bound_commit_position, user_specified_iterations,
            _DEFAULT_CACHE_NAME, None, False
        ])

    pipeline_job = RecursiveFlakeTryJobPipeline(
        analysis.key.urlsafe(), remaining_suggested_commits,
        start_commit_position, revision, lower_bound_commit_position,
        upper_bound_commit_position, user_specified_iterations,
        _DEFAULT_CACHE_NAME, None, False)
    pipeline_job.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()
    self.assertIsNone(analysis.last_attempted_revision)
    self.assertIsNone(analysis.last_attempted_swarming_task_id)

  def testRecursiveFlakeTryJobPipelineRerunFromScratch(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    upper_bound_commit_position = 1000
    lower_bound_commit_position = 998
    start_commit_position = 999
    user_specified_iterations = None
    revision = 'r999'
    try_job_id = 'try_job_id'
    rerun = True
    remaining_suggested_commits = []

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.status = analysis_status.COMPLETED
    analysis.algorithm_parameters = copy.deepcopy(
        DEFAULT_CONFIG_DATA['check_flake_settings'])
    analysis.Save()

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, revision)

    try_job_result = {
        revision: {
            step_name: {
                'status': 'failed',
                'failures': [test_name],
                'valid': True,
                'pass_fail_counts': {
                    test_name: {
                        'pass_count': 28,
                        'fail_count': 72,
                    }
                }
            }
        }
    }
    report = {'report': {'result': try_job_result}}
    try_job.flake_results.append(report)
    try_job.put()

    self.MockPipeline(
        recursive_flake_try_job_pipeline.ScheduleFlakeTryJobPipeline,
        try_job_id,
        expected_args=[
            master_name, builder_name, step_name, test_name, revision,
            analysis.key.urlsafe(), _DEFAULT_CACHE_NAME, None, 100
        ])
    self.MockPipeline(
        recursive_flake_try_job_pipeline.MonitorTryJobPipeline,
        try_job_result,
        expected_args=[
            try_job.key.urlsafe(), failure_type.FLAKY_TEST, try_job_id
        ])
    self.MockPipeline(
        recursive_flake_try_job_pipeline.ProcessFlakeTryJobResultPipeline,
        None,
        expected_args=[
            revision, start_commit_position,
            try_job.key.urlsafe(),
            analysis.key.urlsafe()
        ])
    self.MockPipeline(
        recursive_flake_try_job_pipeline.NextCommitPositionPipeline,
        '',
        expected_args=[
            analysis.key.urlsafe(),
            try_job.key.urlsafe(), remaining_suggested_commits,
            start_commit_position, lower_bound_commit_position,
            upper_bound_commit_position, user_specified_iterations,
            _DEFAULT_CACHE_NAME, None, rerun
        ])

    pipeline_job = RecursiveFlakeTryJobPipeline(
        analysis.key.urlsafe(), remaining_suggested_commits,
        start_commit_position, revision, lower_bound_commit_position,
        upper_bound_commit_position, user_specified_iterations,
        _DEFAULT_CACHE_NAME, None, rerun)
    pipeline_job.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()

    self.assertIsNotNone(
        FlakeTryJob.Get(master_name, builder_name, step_name, test_name,
                        revision))
    self.assertEqual(revision, analysis.last_attempted_revision)
    self.assertIsNone(analysis.last_attempted_swarming_task_id)

  def testRecursiveFlakeTryJobPipelineDoNotStartIfError(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    commit_position = 1000
    revision = 'r1000'
    lower_bound_commit_position = 998
    remaining_suggested_commits = []

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.status = analysis_status.ERROR
    analysis.Save()

    pipeline_job = RecursiveFlakeTryJobPipeline(
        analysis.key.urlsafe(), remaining_suggested_commits, commit_position,
        revision, lower_bound_commit_position, commit_position, None,
        _DEFAULT_CACHE_NAME, None, False)
    pipeline_job.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()
    self.assertIsNone(analysis.try_job_status)

  def testNextCommitPositionPipelineNoHeuristicResults(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    revision = 'r99'
    try_job_id = '123'
    remaining_suggested_commits = []
    start_commit_position = 95
    start_revision = 'r95'
    lower_bound_commit_position = 90
    upper_bound_commit_position = 100

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, revision)
    try_job.try_job_ids.append(try_job_id)
    try_job.put()

    try_job_data = FlakeTryJobData.Create(try_job_id)
    try_job_data.try_job_key = try_job.key
    try_job_data.put()

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.status = analysis_status.COMPLETED
    analysis.try_job_status = analysis_status.RUNNING
    analysis.data_points = [
        DataPoint.Create(
            pass_rate=0.9,
            commit_position=100,
            build_number=12345,
            previous_build_commit_position=lower_bound_commit_position,
            blame_list=[
                'r91', 'r92', 'r93', 'r94', 'r95', 'r96', 'r97', 'r98', 'r99',
                'r100'
            ]),
        DataPoint.Create(
            pass_rate=0.9, commit_position=lower_bound_commit_position)
    ]
    analysis.suspected_flake_build_number = 12345
    analysis.algorithm_parameters = copy.deepcopy(
        DEFAULT_CONFIG_DATA['check_flake_settings'])
    analysis.Save()

    self.MockPipeline(
        recursive_flake_try_job_pipeline.RecursiveFlakeTryJobPipeline,
        '',
        expected_args=[
            analysis.key.urlsafe(), remaining_suggested_commits,
            start_commit_position, start_revision, lower_bound_commit_position,
            upper_bound_commit_position, None, _DEFAULT_CACHE_NAME, None, False
        ],
        expected_kwargs={
            'retries': 0
        })

    pipeline_job = NextCommitPositionPipeline(
        analysis.key.urlsafe(), try_job.key.urlsafe(),
        remaining_suggested_commits, start_commit_position,
        lower_bound_commit_position, upper_bound_commit_position, None,
        _DEFAULT_CACHE_NAME, None, False)
    pipeline_job.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()

  def testNextCommitPositionPipelineWithHeuristicResults(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    revision = 'r99'
    try_job_id = '123'
    suspected_commit = 94
    previous_suspected_commit = suspected_commit - 1
    remaining_suggested_commits = [previous_suspected_commit, suspected_commit]

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, revision)
    try_job.try_job_ids.append(try_job_id)
    try_job.put()

    try_job_data = FlakeTryJobData.Create(try_job_id)
    try_job_data.try_job_key = try_job.key
    try_job_data.put()

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.status = analysis_status.COMPLETED
    analysis.try_job_status = analysis_status.RUNNING
    analysis.data_points = [
        DataPoint.Create(
            pass_rate=0.9,
            commit_position=100,
            build_number=12345,
            previous_build_commit_position=90,
            blame_list=[
                'r91', 'r92', 'r93', 'r94', 'r95', 'r96', 'r97', 'r98', 'r99',
                'r100'
            ]),
        DataPoint.Create(pass_rate=0.9, commit_position=90),
        DataPoint.Create(
            pass_rate=1.0, commit_position=previous_suspected_commit)
    ]
    analysis.suspected_flake_build_number = 12345
    analysis.algorithm_parameters = copy.deepcopy(
        DEFAULT_CONFIG_DATA['check_flake_settings'])
    analysis.Save()

    self.MockPipeline(
        recursive_flake_try_job_pipeline.RecursiveFlakeTryJobPipeline,
        '',
        expected_args=[
            analysis.key.urlsafe(), [suspected_commit],
            previous_suspected_commit, 'r93', 90, 100, None,
            _DEFAULT_CACHE_NAME, None, False
        ],
        expected_kwargs={
            'retries': 0
        })

    pipeline_job = NextCommitPositionPipeline(
        analysis.key.urlsafe(), try_job.key.urlsafe(),
        remaining_suggested_commits, previous_suspected_commit, 90, 100, None,
        _DEFAULT_CACHE_NAME, None, False)
    pipeline_job.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()

  @mock.patch.object(CachedGitilesRepository, 'GetChangeLog')
  @mock.patch.object(swarmed_test_util, 'GetTestLocation')
  def testNextCommitPositionPipelineCompleted(self, mock_test_location,
                                              mock_change_log):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    git_hash = 'r95'
    commit_position = 95
    url = 'url'
    try_job_id = '123'
    remaining_suggested_commits = []
    change_log = ChangeLog(None, None, git_hash, commit_position, None, None,
                           url, None)
    mock_change_log.return_value = change_log

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, git_hash)
    try_job.try_job_ids.append(try_job_id)
    try_job.put()

    try_job_data = FlakeTryJobData.Create(try_job_id)
    try_job_data.try_job_key = try_job.key
    try_job_data.put()

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.start_time = datetime(2016, 6, 26, 23)
    analysis.status = analysis_status.COMPLETED
    analysis.try_job_status = analysis_status.RUNNING
    analysis.data_points = [
        DataPoint.Create(
            pass_rate=0.9,
            commit_position=100,
            build_number=12345,
            previous_build_commit_position=90,
            blame_list=[
                'r91', 'r92', 'r93', 'r94', 'r95', 'r96', 'r97', 'r98', 'r99',
                'r100'
            ]),
        DataPoint.Create(pass_rate=0.9, commit_position=99, try_job_url='u1'),
        DataPoint.Create(pass_rate=0.9, commit_position=97, try_job_url='u2'),
        DataPoint.Create(pass_rate=0.9, commit_position=95, try_job_url='u4'),
        DataPoint.Create(pass_rate=1.0, commit_position=94, try_job_url='u3')
    ]
    analysis.suspected_flake_build_number = 12345
    analysis.algorithm_parameters = copy.deepcopy(
        DEFAULT_CONFIG_DATA['check_flake_settings'])
    analysis.Save()

    test_location = TestLocation(file='foo/bar', line=1)
    mock_test_location.return_value = test_location

    self.MockPipeline(
        recursive_flake_try_job_pipeline.RecursiveFlakeTryJobPipeline,
        '',
        expected_args=[],
        expected_kwargs={})
    self.MockPipeline(
        UpdateFlakeBugPipeline,
        '',
        expected_args=[analysis.key.urlsafe()],
        expected_kwargs={})
    self.MockPipeline(
        SendNotificationForFlakeCulpritPipeline,
        '',
        expected_args=[analysis.key.urlsafe()],
        expected_kwargs={})
    input_obj = CreateInputObjectInstance(
        CreateBugForFlakePipelineInputObject,
        analysis_urlsafe_key=unicode(analysis.key.urlsafe()),
        test_location=test_location)

    self.MockGeneratorPipeline(CreateBugForFlakePipeline, input_obj, None)

    pipeline_job = NextCommitPositionPipeline(analysis.key.urlsafe(),
                                              try_job.key.urlsafe(),
                                              remaining_suggested_commits,
                                              commit_position, 90, 100, None,
                                              _DEFAULT_CACHE_NAME, None, False)
    pipeline_job.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()

    culprit = ndb.Key(urlsafe=analysis.culprit_urlsafe_key).get()
    self.assertEqual(git_hash, culprit.revision)
    self.assertEqual(95, culprit.commit_position)

  @mock.patch.object(CachedGitilesRepository, 'GetChangeLog')
  @mock.patch.object(swarmed_test_util, 'GetTestLocation')
  def testNextCommitPositionNewlyAddedFlakyTest(self, mocked_test_location,
                                                mock_change_log):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    git_hash = 'r100'
    try_job_id = '123'
    revision = 'r100'
    most_recently_run_commit_position = 100
    url = 'url'
    remaining_suggested_commits = []
    change_log = ChangeLog(None, None, revision,
                           most_recently_run_commit_position, None, None, url,
                           None)
    mock_change_log.return_value = change_log

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, revision)
    try_job.try_job_ids.append(try_job_id)
    try_job.put()

    try_job_data = FlakeTryJobData.Create(try_job_id)
    try_job_data.try_job_key = try_job.key
    try_job_data.put()

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.start_time = datetime(2016, 6, 26, 23)
    analysis.status = analysis_status.COMPLETED
    analysis.try_job_status = analysis_status.RUNNING
    analysis.data_points = [
        DataPoint.Create(
            pass_rate=0.9,
            commit_position=most_recently_run_commit_position,
            build_number=12345,
            previous_build_commit_position=98,
            blame_list=['r99', 'r100']),
        DataPoint.Create(pass_rate=-1, commit_position=99, try_job_url='id1'),
        DataPoint.Create(pass_rate=-1, commit_position=98)
    ]
    analysis.suspected_flake_build_number = 12345
    analysis.algorithm_parameters = copy.deepcopy(
        DEFAULT_CONFIG_DATA['check_flake_settings'])
    analysis.Save()

    test_location = TestLocation(file='foo/bar', line=1)
    mocked_test_location.return_value = test_location

    self.MockPipeline(
        recursive_flake_try_job_pipeline.RecursiveFlakeTryJobPipeline,
        '',
        expected_args=[])
    self.MockPipeline(
        SendNotificationForFlakeCulpritPipeline,
        '',
        expected_args=[analysis.key.urlsafe()],
        expected_kwargs={})

    input_obj = CreateInputObjectInstance(
        CreateBugForFlakePipelineInputObject,
        analysis_urlsafe_key=unicode(analysis.key.urlsafe()),
        test_location=test_location)
    self.MockGeneratorPipeline(CreateBugForFlakePipeline, input_obj, None)

    pipeline_job = NextCommitPositionPipeline(
        analysis.key.urlsafe(), try_job.key.urlsafe(),
        remaining_suggested_commits, most_recently_run_commit_position, 98, 100,
        None, _DEFAULT_CACHE_NAME, None, False)
    pipeline_job.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()

    culprit = ndb.Key(urlsafe=analysis.culprit_urlsafe_key).get()
    self.assertEqual(git_hash, culprit.revision)
    self.assertEqual(100, culprit.commit_position)

  @mock.patch(('waterfall.flake.recursive_flake_try_job_pipeline.'
               'RecursiveFlakeTryJobPipeline'))
  def testNextCommitPositionPipelineForFailedTryJob(self, mocked_pipeline):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    revision = 'r97'
    most_recently_run_commit_position = 97
    lower_bound_commit_position = 96
    upper_bound_commit_position = 100
    try_job_id = '123'
    error = {
        'code': 1,
        'message': 'some failure message',
    }
    remaining_suggested_commits = []
    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, revision)
    try_job.try_job_ids.append(try_job_id)
    try_job.put()

    try_job_data = FlakeTryJobData.Create(try_job_id)
    try_job_data.error = error
    try_job_data.try_job_key = try_job.key
    try_job_data.put()

    analysis = MasterFlakeAnalysis.Create(master_name, builder_name,
                                          build_number, step_name, test_name)
    analysis.start_time = datetime(2016, 6, 26, 23)
    analysis.data_points = [
        DataPoint.Create(commit_position=most_recently_run_commit_position)
    ]
    analysis.put()

    self.MockPipeline(
        UpdateFlakeBugPipeline,
        '',
        expected_args=[analysis.key.urlsafe()],
        expected_kwargs={})

    pipeline_job = NextCommitPositionPipeline(
        analysis.key.urlsafe(), try_job.key.urlsafe(),
        remaining_suggested_commits, most_recently_run_commit_position,
        lower_bound_commit_position, upper_bound_commit_position, None,
        _DEFAULT_CACHE_NAME, None, False)
    pipeline_job.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()

    self.assertFalse(mocked_pipeline.called)
    self.assertEqual(error, analysis.error)

  @mock.patch.object(CachedGitilesRepository, 'GetChangeLog')
  def testUpdateCulpritNewCulprit(self, mocked_fn):
    revision = 'a1b2c3d4'
    commit_position = 12345
    url = 'url'
    repo_name = 'repo_name'
    change_log = ChangeLog(None, None, revision, commit_position, None, None,
                           url, None)
    mocked_fn.return_value = change_log

    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')

    culprit = recursive_flake_try_job_pipeline.UpdateCulprit(
        analysis.key.urlsafe(), revision, commit_position, repo_name)

    self.assertIsNotNone(culprit)
    self.assertEqual([analysis.key.urlsafe()],
                     culprit.flake_analysis_urlsafe_keys)
    self.assertEqual(url, culprit.url)
    self.assertEqual(repo_name, culprit.repo_name)
    self.assertEqual(revision, culprit.revision)

  def testUpdateCulpritExistingCulprit(self):
    revision = 'a1b2c3d4'
    commit_position = 12345
    url = 'url'
    repo_name = 'repo_name'
    analysis_urlsafe_key = 'urlsafe_key'

    culprit = FlakeCulprit.Create(repo_name, revision, commit_position)
    culprit.flake_analysis_urlsafe_keys = ['another_analysis_urlsafe_key']
    culprit.url = url
    culprit.put()

    culprit = recursive_flake_try_job_pipeline.UpdateCulprit(
        analysis_urlsafe_key, revision, commit_position, repo_name)

    self.assertIsNotNone(culprit)
    self.assertEqual(2, len(culprit.flake_analysis_urlsafe_keys))
    self.assertIn(analysis_urlsafe_key, culprit.flake_analysis_urlsafe_keys)
    self.assertEqual(url, culprit.url)
    self.assertEqual(repo_name, culprit.repo_name)
    self.assertEqual(revision, culprit.revision)

  def testUpdateCulpritExistingCulpritAlreadyHasAnalyis(self):
    revision = 'a1b2c3d4'
    commit_position = 12345
    url = 'url'
    repo_name = 'repo_name'
    analysis_urlsafe_key = 'urlsafe_key'
    culprit = FlakeCulprit.Create(repo_name, revision, commit_position)
    culprit.flake_analysis_urlsafe_keys = [analysis_urlsafe_key]
    culprit.url = url
    culprit.put()

    culprit = recursive_flake_try_job_pipeline.UpdateCulprit(
        analysis_urlsafe_key, revision, commit_position, repo_name)

    self.assertIsNotNone(culprit)
    self.assertEqual(1, len(culprit.flake_analysis_urlsafe_keys))
    self.assertIn(analysis_urlsafe_key, culprit.flake_analysis_urlsafe_keys)
    self.assertEqual(url, culprit.url)
    self.assertEqual(repo_name, culprit.repo_name)
    self.assertEqual(revision, culprit.revision)

  @mock.patch.object(CachedGitilesRepository, 'GetChangeLog', return_value=None)
  def testUpdateCulpritNoLogs(self, _):
    revision = 'a1b2c3d4'
    commit_position = 12345
    repo_name = 'repo_name'
    analysis_urlsafe_key = 'urlsfe_key'
    culprit = recursive_flake_try_job_pipeline.UpdateCulprit(
        analysis_urlsafe_key, revision, commit_position, repo_name)

    self.assertIn(analysis_urlsafe_key, culprit.flake_analysis_urlsafe_keys)
    self.assertEqual(commit_position, culprit.commit_position)
    self.assertEqual(revision, culprit.revision)
    self.assertIsNone(culprit.url)
    self.assertEqual(repo_name, culprit.repo_name)

  def testGetTryJobNew(self):
    existing_try_job = FlakeTryJob.Create('m', 'b', 's', 't', 'a1b2c3d4')
    existing_try_job.put()
    self.assertEqual(existing_try_job,
                     recursive_flake_try_job_pipeline._GetTryJob(
                         'm', 'b', 's', 't', 'a1b2c3d4'))

  def testGetTryJobExisting(self):
    try_job = recursive_flake_try_job_pipeline._GetTryJob(
        'm', 'b', 's', 't', 'e5f6a1b2')
    self.assertIsNotNone(try_job)
    self.assertEqual(try_job.git_hash, 'e5f6a1b2')

  def testNeedANewTryJob(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    try_job = FlakeTryJob.Create('m', 'b', 's', 't', 'a1b2c3d4')
    self.assertTrue(
        recursive_flake_try_job_pipeline._NeedANewTryJob(
            analysis, try_job, 200, True))
    self.assertTrue(
        recursive_flake_try_job_pipeline._NeedANewTryJob(
            analysis, try_job, 200, False))

  @mock.patch.object(
      flake_try_job, 'IsTryJobResultAtRevisionValidForStep', return_value=True)
  def testNeedANewTryJobWithExistingFlakyTryJob(self, _):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.algorithm_parameters = {
        'try_job_rerun': {
            'lower_flake_threshold': 0.02,
            'upper_flake_threshold': 0.98
        }
    }
    try_job = FlakeTryJob.Create('m', 'b', 's', 't', 'a1b2c3d4')
    try_job.flake_results = [{
        'report': {
            'result': {
                'a1b2c3d4': {
                    's': {
                        'pass_fail_counts': {
                            't': {
                                'pass_count': 60,
                                'fail_count': 40
                            }
                        }
                    }
                }
            }
        }
    }]
    self.assertFalse(
        recursive_flake_try_job_pipeline._NeedANewTryJob(
            analysis, try_job, 200, False))

  @mock.patch.object(
      flake_try_job, 'IsTryJobResultAtRevisionValid', return_value=False)
  def testNeedANewTryJobWithInvalidExistingTryJobNoReport(self, _):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    try_job = FlakeTryJob.Create('m', 'b', 's', 't', 'a1b2c3d4')
    try_job.flake_results = [{'report': None}]
    self.assertTrue(
        recursive_flake_try_job_pipeline._NeedANewTryJob(
            analysis, try_job, 200, False))

  @mock.patch.object(
      flake_try_job, 'IsTryJobResultAtRevisionValidForStep', return_value=False)
  def testNeedANewTryJobWithInvalidExistingTryJob(self, _):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    try_job = FlakeTryJob.Create('m', 'b', 's', 't', 'a1b2c3d4')
    try_job.flake_results = [{
        'report': {
            'result': {
                'a1b2c3d4': {
                    's': {
                        'valid': False
                    }
                }
            }
        }
    }]
    self.assertTrue(
        recursive_flake_try_job_pipeline._NeedANewTryJob(
            analysis, try_job, 200, False))

  @mock.patch.object(
      flake_try_job, 'IsTryJobResultAtRevisionValidForStep', return_value=True)
  def testNeedANewTryJobWithExistingStableTryJobInsufficientIterations(self, _):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.algorithm_parameters = {
        'try_job_rerun': {
            'lower_flake_threshold': 0.02,
            'upper_flake_threshold': 0.98
        }
    }
    try_job = FlakeTryJob.Create('m', 'b', 's', 't', 'a1b2c3d4')
    try_job.flake_results = [{
        'report': {
            'result': {
                'a1b2c3d4': {
                    's': {
                        'pass_fail_counts': {
                            't': {
                                'pass_count': 99,
                                'fail_count': 1
                            }
                        }
                    }
                }
            }
        }
    }]
    self.assertTrue(
        recursive_flake_try_job_pipeline._NeedANewTryJob(
            analysis, try_job, 200, False))

  @mock.patch.object(
      flake_try_job, 'IsTryJobResultAtRevisionValidForStep', return_value=True)
  def testNeedANewTryJobWithExistingStableTryJobSufficientIterations(self, _):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.algorithm_parameters = {
        'try_job_rerun': {
            'lower_flake_threshold': 0.02,
            'upper_flake_threshold': 0.98,
        }
    }
    try_job = FlakeTryJob.Create('m', 'b', 's', 't', 'a1b2c3d4')
    try_job.flake_results = [{
        'report': {
            'result': {
                'a1b2c3d4': {
                    's': {
                        'pass_fail_counts': {
                            't': {
                                'pass_count': 200,
                                'fail_count': 0,
                            }
                        }
                    }
                }
            }
        }
    }]
    self.assertFalse(
        recursive_flake_try_job_pipeline._NeedANewTryJob(
            analysis, try_job, 200, False))

  @mock.patch.object(
      flake_try_job, 'IsTryJobResultAtRevisionValidForStep', return_value=True)
  def testNeedANewTryJobWithExistingTryJobNonexistentTest(self, _):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.algorithm_parameters = {
        'try_job_rerun': {
            'lower_flake_threshold': 0.02,
            'upper_flake_threshold': 0.98,
        }
    }
    try_job = FlakeTryJob.Create('m', 'b', 's', 't', 'a1b2c3d4')
    try_job.flake_results = [{'report': {'result': {'a1b2c3d4': {'s': {}}}}}]
    self.assertFalse(
        recursive_flake_try_job_pipeline._NeedANewTryJob(
            analysis, try_job, 200, False))

  def testSetAnalysisTryJobStatusRunning(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    recursive_flake_try_job_pipeline._SetAnalysisTryJobStatus(
        analysis, analysis_status.RUNNING)
    self.assertEqual(analysis.try_job_status, analysis_status.RUNNING)

  def testSetAnalysisTryJobStatusRunningAlreadyRunning(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.try_job_status = analysis_status.RUNNING
    recursive_flake_try_job_pipeline._SetAnalysisTryJobStatus(
        analysis, analysis_status.RUNNING)
    self.assertEqual(analysis.try_job_status, analysis_status.RUNNING)

  def testGetTryJobDataPointsNoTryJobsYet(self):
    suspected_flake_build_number = 12345
    suspected_flake_commit_position = 100,
    suspected_flake_lower_bound = 90
    data_points = [
        DataPoint.Create(
            pass_rate=0.8,
            commit_position=100,
            build_number=suspected_flake_build_number)
    ]
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.suspected_flake_build_number = suspected_flake_build_number
    analysis.data_points = data_points

    normalized_data_points = _GetNormalizedTryJobDataPoints(
        analysis, suspected_flake_lower_bound, suspected_flake_commit_position)
    self.assertEqual(normalized_data_points[0].run_point_number, 100)
    self.assertEqual(normalized_data_points[0].pass_rate, 0.8)
    self.assertEqual(len(normalized_data_points), 1)

  def testGetTryJobDataPointsWithTryJobs(self):
    suspected_flake_build_number = 12345
    all_data_points = [
        DataPoint.Create(
            pass_rate=0.8,
            commit_position=100,
            build_number=suspected_flake_build_number),
        DataPoint.Create(
            pass_rate=1.0,
            commit_position=90,
            build_number=suspected_flake_build_number - 1),
        DataPoint.Create(pass_rate=0.8, commit_position=99, try_job_url='url')
    ]

    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.suspected_flake_build_number = suspected_flake_build_number
    analysis.data_points = all_data_points
    normalized_data_points = _GetNormalizedTryJobDataPoints(analysis, 91, 100)

    self.assertEqual(normalized_data_points[0].run_point_number, 99)
    self.assertEqual(normalized_data_points[0].pass_rate, 0.8)
    self.assertEqual(normalized_data_points[1].run_point_number, 100)
    self.assertEqual(normalized_data_points[1].pass_rate, 0.8)

  @mock.patch.object(
      RecursiveFlakeTryJobPipeline, 'was_aborted', return_value=True)
  def testRecursiveFlakeTryJobPipelineAborted(self, _):
    master_name = 'm'
    builder_name = 'b'
    master_build_number = 100
    step_name = 's'
    test_name = 't'
    revision = 'rev'
    commit_position = 1
    build_id = 'b1'
    lower_bound_commit_position = 0
    remaining_suggested_commits = []

    analysis = MasterFlakeAnalysis.Create(
        master_name, builder_name, master_build_number, step_name, test_name)
    analysis.start_time = datetime(2016, 6, 26, 23)
    analysis.status = analysis_status.COMPLETED
    analysis.Save()

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, revision)
    try_job.try_job_ids = [build_id]
    try_job.put()

    try_job_data = FlakeTryJobData.Create(build_id)
    try_job_data.try_job_key = try_job.key
    try_job_data.put()

    rftp = RecursiveFlakeTryJobPipeline(
        analysis.key.urlsafe(), remaining_suggested_commits, commit_position,
        revision, lower_bound_commit_position, 100, None, _DEFAULT_CACHE_NAME,
        None, False)
    rftp._LogUnexpectedAbort()

    expected_error = {
        'error': 'RecursiveFlakeTryJobPipeline was aborted unexpectedly',
        'message': 'RecursiveFlakeTryJobPipeline was aborted unexpectedly'
    }

    self.assertEqual(analysis_status.ERROR, analysis.try_job_status)
    self.assertEqual(expected_error, analysis.error)
    self.assertEqual(analysis_status.ERROR, try_job.status)
    self.assertEqual(expected_error, try_job_data.error)

  @mock.patch.object(
      RecursiveFlakeTryJobPipeline, 'was_aborted', return_value=True)
  def testRecursiveFlakeTryJobPipelineAbortedNoUpdateCompletedTryJob(self, _):
    master_name = 'm'
    builder_name = 'b'
    master_build_number = 100
    suspected_build_commit_position = 1000
    step_name = 's'
    test_name = 't'
    revision = 'rev'
    commit_position = 1
    lower_bound_commit_position = 0
    remaining_suggested_commits = []

    analysis = MasterFlakeAnalysis.Create(
        master_name, builder_name, master_build_number, step_name, test_name)
    analysis.start_time = datetime(2016, 6, 26, 23)
    analysis.status = analysis_status.COMPLETED
    analysis.Save()

    try_job = FlakeTryJob.Create(master_name, builder_name, step_name,
                                 test_name, revision)
    try_job.status = analysis_status.COMPLETED
    try_job.put()

    pipeline_job = RecursiveFlakeTryJobPipeline(
        analysis.key.urlsafe(), remaining_suggested_commits, commit_position,
        revision, lower_bound_commit_position, suspected_build_commit_position,
        None, _DEFAULT_CACHE_NAME, None, False)

    pipeline_job._LogUnexpectedAbort()

    self.assertEqual(analysis_status.COMPLETED, try_job.status)

  @mock.patch.object(
      confidence, 'SteppinessForCommitPosition', return_value=0.6)
  def testGetSuspectedConmmitConfidenceScore(self, _):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 124, 's', 't')
    analysis.data_points = [
        DataPoint.Create(pass_rate=0.7, commit_position=123)
    ]
    self.assertEqual(
        0.6,
        recursive_flake_try_job_pipeline._GetSuspectedCommitConfidenceScore(
            analysis, 123))
    self.assertIsNone(
        recursive_flake_try_job_pipeline._GetSuspectedCommitConfidenceScore(
            analysis, None))

  def testGetSuspectedCommitConfidenceScoreIntroducedNewFlakyTest(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 124, 's', 't')
    analysis.data_points = [
        DataPoint.Create(pass_rate=0.7, commit_position=123),
        DataPoint.Create(
            pass_rate=flake_constants.PASS_RATE_TEST_NOT_FOUND,
            commit_position=122)
    ]
    self.assertEqual(
        1.0,
        recursive_flake_try_job_pipeline._GetSuspectedCommitConfidenceScore(
            analysis, 123))

  def testGetNextCommitPositionAndRemainingSuspectsNoHeuristicResults(self):
    previously_run_commit_position = 1000
    bisected_commit_position = 1005
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.Save()

    self.assertEqual((bisected_commit_position, []),
                     _GetNextCommitPositionAndRemainingSuspects(
                         analysis, [], previously_run_commit_position,
                         bisected_commit_position))

  def testGetNextCommitPositionAndRemainingSuspectsCorrectSuspect(self):
    # 1. Revision range: [r1 - r19].
    # 2. Commits to try: [r4, r5].
    # 4. r4 was just ran and was stable.
    # 5. r5 should be attempted.
    suspected_build_number = 2
    previously_run_commit_position = 4
    next_suggested_run_commit_position_1 = 5
    next_suggested_run_commit_position_2 = 6
    bisected_commit_position = 7
    remaining_suggested_commits = [
        next_suggested_run_commit_position_1,
        next_suggested_run_commit_position_2
    ]

    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.algorithm_parameters = {
        'try_job_rerun': {
            'lower_flake_threshold': 0.02,
            'upper_flake_threshold': 0.98,
        }
    }
    analysis.data_points = [
        DataPoint.Create(
            pass_rate=1.0, commit_position=previously_run_commit_position),
        DataPoint.Create(
            pass_rate=0.5,
            commit_position=20,
            build_number=suspected_build_number),
    ]
    analysis.Save()

    self.assertEqual((next_suggested_run_commit_position_1,
                      [next_suggested_run_commit_position_2]),
                     _GetNextCommitPositionAndRemainingSuspects(
                         analysis, remaining_suggested_commits,
                         previously_run_commit_position,
                         bisected_commit_position))

  def testGetNextCommitPositionAndRemainingSuspectsCulpritBeforeRange(self):
    # 1. Revision range: [r1 - r19].
    # 2. Commits to try: [r4, r5] where r5 incorrect and after the culprit.
    # 4. r4 was just ran and was already flaky.
    # 5. In this case, heuristic resulsts should be dropped and bisect used on
    #    the range before heuristic results, e.g. r1 - r4.
    suspected_build_number = 2
    previously_run_commit_position = 4
    next_suggested_run_commit_position = 5
    bisected_commit_position = 2
    remaining_suggested_commits = [next_suggested_run_commit_position]

    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.algorithm_parameters = {
        'try_job_rerun': {
            'lower_flake_threshold': 0.02,
            'upper_flake_threshold': 0.98,
        }
    }
    analysis.data_points = [
        DataPoint.Create(pass_rate=0.7, commit_position=4),
        DataPoint.Create(
            pass_rate=0.5,
            commit_position=20,
            build_number=suspected_build_number),
    ]
    analysis.Save()

    self.assertEqual((bisected_commit_position, []),
                     _GetNextCommitPositionAndRemainingSuspects(
                         analysis, remaining_suggested_commits,
                         previously_run_commit_position,
                         bisected_commit_position))
