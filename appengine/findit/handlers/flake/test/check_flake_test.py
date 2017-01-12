# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import mock
import re

import webapp2
import webtest

from google.appengine.api import users

from handlers.flake import check_flake
from handlers.flake.check_flake import CheckFlake
from model import analysis_status
from model.analysis_status import STATUS_TO_DESCRIPTION
from model.flake.flake_analysis_request import BuildStep
from model.flake.flake_analysis_request import FlakeAnalysisRequest
from model.flake.master_flake_analysis import DataPoint
from model.flake.master_flake_analysis import MasterFlakeAnalysis
from waterfall.flake import flake_analysis_service
from waterfall.test import wf_testcase


class CheckFlakeTest(wf_testcase.WaterfallTestCase):
  app_module = webapp2.WSGIApplication([
      ('/waterfall/flake', check_flake.CheckFlake),
  ], debug=True)

  @mock.patch.object(flake_analysis_service, 'ScheduleAnalysisForFlake',
                     return_value=True)
  def testCorpUserCanScheduleANewAnalysis(self, _):
    master_name = 'm'
    builder_name = 'b'
    build_number = '123'
    step_name = 's'
    test_name = 't'

    analysis = MasterFlakeAnalysis.Create(
        master_name, builder_name, build_number, step_name, test_name)
    analysis.Save()

    self.mock_current_user(user_email='test@google.com')

    response = self.test_app.get('/waterfall/flake', params={
        'master_name': master_name,
        'builder_name': builder_name,
        'build_number': build_number,
        'step_name': step_name,
        'test_name': test_name})

    self.assertEquals(200, response.status_int)

  def testNoneCorpUserCanNotScheduleANewAnalysis(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = '123'
    step_name = 's'
    test_name = 't'

    self.assertRaisesRegexp(
        webtest.app.AppError,
        re.compile('.*401 Unauthorized.*',
                   re.MULTILINE | re.DOTALL),
        self.test_app.get,
        '/waterfall/flake',
        params={
            'master_name': master_name,
            'builder_name': builder_name,
            'build_number': build_number,
            'step_name': step_name,
            'test_name': test_name
        })

  @mock.patch.object(check_flake, '_GetSuspectedFlakeInfo',
                     return_value={
                         'build_number': 100,
                         'commit_position': 12345,
                         'git_hash': 'git_hash_1',
                         'triage_result': 0})
  @mock.patch.object(check_flake, '_GetCoordinatesData',
                     return_value=[[12345, 0.9, '1', 100, 'git_hash_2',
                                    12344, 'git_hash_1']])
  def testAnyoneCanViewScheduledAnalysis(self, *_):
    master_name = 'm'
    builder_name = 'b'
    build_number = '123'
    step_name = 's'
    test_name = 't'
    success_rate = .9

    analysis = MasterFlakeAnalysis.Create(
        master_name, builder_name, build_number, step_name, test_name)
    data_point = DataPoint()
    data_point.build_number = int(build_number)
    data_point.pass_rate = success_rate
    data_point.task_id = '1'
    analysis.data_points.append(data_point)
    analysis.status = analysis_status.COMPLETED
    analysis.suspected_flake_build_number = 100
    analysis.request_time = datetime.datetime(2016, 10, 01, 12, 10, 00)
    analysis.start_time = datetime.datetime(2016, 10, 01, 12, 10, 05)
    analysis.end_time = datetime.datetime(2016, 10, 01, 13, 10, 00)
    analysis.algorithm_parameters = {'iterations_to_rerun': 100}
    analysis.Save()

    response = self.test_app.get('/waterfall/flake', params={
        'key': analysis.key.urlsafe(),
        'format': 'json'})

    expected_check_flake_result = {
        'key': analysis.key.urlsafe(),
        'pass_rates': [[12345, 0.9, '1', 100, 'git_hash_2', 12344,
                        'git_hash_1']],
        'analysis_status': STATUS_TO_DESCRIPTION.get(analysis.status),
        'master_name': master_name,
        'builder_name': builder_name,
        'build_number': int(build_number),
        'step_name': step_name,
        'test_name': test_name,
        'request_time': '2016-10-01 12:10:00 UTC',
        'task_number': 1,
        'error': None,
        'iterations_to_rerun': 100,
        'pending_time': '00:00:05',
        'duration': '00:59:55',
        'suspected_flake': {
            'build_number': 100,
            'commit_position': 12345,
            'git_hash': 'git_hash_1',
            'triage_result': 0
        },
        'version_number': 1,
        'show_debug_info': False
    }

    self.assertEquals(200, response.status_int)
    self.assertEqual(expected_check_flake_result, response.json_body)

  def testUnauthorizedUserCannotScheduleNewAnalysis(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 123
    step_name = 's'
    test_name = 't'

    self.assertRaisesRegexp(
        webtest.app.AppError,
        re.compile('.*401 Unauthorized.*', re.MULTILINE | re.DOTALL),
        self.test_app.get,
        '/waterfall/flake',
        params={
            'master_name': master_name,
            'builder_name': builder_name,
            'build_number': build_number,
            'step_name': step_name,
            'test_name': test_name,
            'format': 'json'})

  @mock.patch.object(flake_analysis_service, 'ScheduleAnalysisForFlake',
                     return_value=False)
  @mock.patch.object(check_flake, '_GetSuspectedFlakeInfo',
                     return_value={
                         'build_number': 100,
                         'commit_position': 12345,
                         'git_hash': 'a_git_hash',
                         'triage_result': 0})
  @mock.patch.object(check_flake, '_GetCoordinatesData',
                     return_value=[[12345, 0.9, '1', 100, 'git_hash_2',
                                    12344, 'git_hash_1']])
  def testRequestExistingAnalysis(self, *_):
    master_name = 'm'
    builder_name = 'b'
    build_number = 123
    step_name = 's'
    test_name = 't'
    success_rate = 0.9

    previous_analysis = MasterFlakeAnalysis.Create(
        master_name, builder_name, build_number - 1, step_name, test_name)
    data_point = DataPoint()
    data_point.build_number = build_number - 1
    data_point.pass_rate = success_rate
    previous_analysis.data_points.append(data_point)
    previous_analysis.status = analysis_status.COMPLETED
    previous_analysis.suspected_flake_build_number = 100
    previous_analysis.request_time = datetime.datetime(2016, 10, 01, 12, 10, 00)
    previous_analysis.start_time = datetime.datetime(2016, 10, 01, 12, 10, 05)
    previous_analysis.end_time = datetime.datetime(2016, 10, 01, 13, 10, 00)
    previous_analysis.algorithm_parameters = {'iterations_to_rerun': 100}
    previous_analysis.Save()

    previous_request = FlakeAnalysisRequest.Create(test_name, False, None)
    build_step = BuildStep.Create(
        master_name, builder_name, build_number, step_name, None)
    build_step.wf_master_name = build_step.master_name
    build_step.wf_builder_name = build_step.builder_name
    build_step.wf_build_number = build_step.build_number
    build_step.wf_step_name = build_step.step_name
    previous_request.build_steps.append(build_step)
    previous_request.analyses.append(previous_analysis.key)
    previous_request.Save()

    response = self.test_app.get('/waterfall/flake', params={
        'master_name': master_name,
        'builder_name': builder_name,
        'build_number': build_number,
        'step_name': step_name,
        'test_name': test_name,
        'format': 'json'})

    expected_check_flake_result = {
        'key': previous_analysis.key.urlsafe(),
        'pass_rates': [[12345, 0.9, '1', 100, 'git_hash_2', 12344,
                        'git_hash_1']],
        'analysis_status': STATUS_TO_DESCRIPTION.get(previous_analysis.status),
        'master_name': master_name,
        'builder_name': builder_name,
        'build_number': build_number - 1,
        'step_name': step_name,
        'test_name': test_name,
        'request_time': '2016-10-01 12:10:00 UTC',
        'task_number': 1,
        'error': None,
        'iterations_to_rerun': 100,
        'pending_time': '00:00:05',
        'duration': '00:59:55',
        'suspected_flake': {
            'build_number': 100,
            'commit_position': 12345,
            'git_hash': 'a_git_hash',
            'triage_result': 0
        },
        'version_number': 1,
        'show_debug_info': False
    }

    self.assertEqual(200, response.status_int)
    self.assertEqual(expected_check_flake_result, response.json_body)

  @mock.patch.object(flake_analysis_service, 'ScheduleAnalysisForFlake',
                     return_value=False)
  def testRequestUnsupportedAnalysis(self, _):
    master_name = 'm'
    builder_name = 'b'
    build_number = 123
    step_name = 's'
    test_name = 't'

    previous_request = FlakeAnalysisRequest.Create(test_name, False, None)
    previous_request.AddBuildStep(
        master_name, builder_name, build_number, step_name, None)
    previous_request.swarmed = False
    previous_request.supported = False

    self.assertRaisesRegexp(
        webtest.app.AppError,
        re.compile('.*not supported.*', re.MULTILINE | re.DOTALL),
        self.test_app.get,
        '/waterfall/flake',
        params={
            'master_name': master_name,
            'builder_name': builder_name,
            'build_number': build_number,
            'step_name': step_name,
            'test_name': test_name,
            'format': 'json'})

  @mock.patch.object(check_flake, '_GetSuspectedFlakeInfo',
                     return_value={
                         'build_number': 100,
                         'commit_position': 12345,
                         'git_hash': 'a_git_hash',
                         'triage_result': 0})
  @mock.patch.object(check_flake, '_GetCoordinatesData',
                     return_value=[[12345, 0.9, '1', 100, 'git_hash_2',
                                    12344, 'git_hash_1']])
  @mock.patch.object(users, 'is_current_user_admin', return_value=True)
  @mock.patch.object(flake_analysis_service, 'ScheduleAnalysisForFlake',
                     return_value=True)
  def testGetTriageHistory(self, *_):
    master_name = 'm'
    builder_name = 'b'
    build_number = '123'
    step_name = 's'
    test_name = 't'
    suspected_flake_build_number = 123
    triage_result = 2
    user_name = 'test'

    analysis = MasterFlakeAnalysis.Create(
        master_name, builder_name, build_number, step_name, test_name)
    analysis.status = analysis_status.COMPLETED
    analysis.suspected_flake_build_number = 100
    analysis.Save()
    analysis.UpdateTriageResult(
        triage_result, {'build_number': suspected_flake_build_number}, 'test')

    response = self.test_app.get('/waterfall/flake', params={
        'master_name': master_name,
        'builder_name': builder_name,
        'build_number': build_number,
        'step_name': step_name,
        'test_name': test_name,
        'format': 'json'})

    # Because TriagedResult uses auto_now=True, a direct dict comparison will
    # always fail. Instead only compare the relevant fields for trige_history.
    triage_history = response.json_body.get('triage_history')
    self.assertEqual(len(triage_history), 1)
    self.assertEqual(triage_history[0].get('triage_result'), 'Correct')
    self.assertEqual(triage_history[0].get('user_name'), user_name)
    self.assertEqual(
        triage_history[0].get('suspect_info', {}).get('build_number'),
        suspected_flake_build_number)

  def testValidateInput(self):
    self.assertIsNone(
        CheckFlake()._ValidateInput('m', 'b', '123', 's', 't', None))
    self.assertIsNone(
        CheckFlake()._ValidateInput('m', 'b', '123', 's', 't', '654321'))
    self.assertEqual(
        CheckFlake()._ValidateInput(
            None, 'b', '1', 's', 't', 'a').get('data', {}).get('error_message'),
        'Master name must be specified')
    self.assertEqual(
        CheckFlake()._ValidateInput(
            'm', None, '1', 's', 't', '').get('data', {}).get('error_message'),
        'Builder name must be specified')
    self.assertEqual(
        CheckFlake()._ValidateInput(
            'm', 'b', None, 's', 't', '').get('data', {}).get('error_message'),
        'Build number must be specified as an int')
    self.assertEqual(
        CheckFlake()._ValidateInput(
            'm', 'b', 'a', 's', 't', '').get('data', {}).get('error_message'),
        'Build number must be specified as an int')
    self.assertEqual(
        CheckFlake()._ValidateInput(
            'm', 'b', '1', None, 't', '').get('data', {}).get('error_message'),
        'Step name must be specified')
    self.assertEqual(
        CheckFlake()._ValidateInput(
            'm', 'b', '1', 's', None, '').get('data', {}).get('error_message'),
        'Test name must be specified')
    self.assertEqual(
        CheckFlake()._ValidateInput(
            'm', 'b', '1', 's', 't', 'a').get('data', {}).get('error_message'),
        'Bug id (optional) must be an int')

  def testGetSuspectedFlakeInfo(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 123, 's', 't')
    analysis.suspected_flake_build_number = 123
    data_point = DataPoint()
    data_point.build_number = 123
    data_point.pass_rate = 0.9
    data_point.commit_position = 2
    data_point.git_hash = 'git_hash_2'
    data_point.previous_build_commit_position = 1
    data_point.previous_build_git_hash = 'git_hash_1'
    analysis.data_points.append(data_point)
    analysis.Save()

    expected_result = {
        'build_number': analysis.suspected_flake_build_number,
        'commit_position': 2,
        'git_hash': 'git_hash_2',
        'previous_build_commit_position': 1,
        'previous_build_git_hash': 'git_hash_1',
        'triage_result': 0
    }
    self.assertEqual(expected_result,
                     check_flake._GetSuspectedFlakeInfo(analysis))

  def testGetCoordinatesData(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 123
    step_name = 's'
    test_name = 't'
    success_rate = .9
    try_job_url = 'try_job_url'
    analysis = MasterFlakeAnalysis.Create(
        master_name, builder_name, build_number, step_name, test_name)
    data_point = DataPoint()
    data_point.build_number = build_number
    data_point.pass_rate = success_rate
    data_point.commit_position = 2
    data_point.git_hash = 'git_hash_2'
    data_point.previous_build_commit_position = 1
    data_point.previous_build_git_hash = 'git_hash_1'
    data_point.try_job_url = try_job_url
    analysis.data_points.append(data_point)
    analysis.Save()

    expected_result = [
        {
            'commit_position': 2,
            'pass_rate': success_rate,
            'task_id': None,
            'build_number': build_number,
            'git_hash': 'git_hash_2',
            'previous_build_commit_position': 1,
            'previous_build_git_hash': 'git_hash_1',
            'try_job_url': try_job_url
        }
    ]
    self.assertEqual(expected_result, check_flake._GetCoordinatesData(analysis))
