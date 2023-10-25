# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
import mock
from flask import Flask

from google.appengine.api import users
from google.appengine.ext import ndb

from analysis.type_enums import CrashClient
from backend.handlers.rerun_analysis import RerunAnalysis
from common.appengine_testcase import AppengineTestCase
from common.crash_pipeline import RerunPipeline
from common.model.cracas_crash_analysis import CracasCrashAnalysis


class RerunAnalysisTest(AppengineTestCase):
  """Tests utility functions and ``RerunAnalysis`` handler."""
  app_module = Flask(__name__)
  app_module.add_url_rule(
      '/process/rerun-analysis',
      view_func=RerunAnalysis().Handle,
      methods=['GET'])

  def testHandleErrorWhenThereIsNoKeyProvided(self):
    """Tests that handler returns 500 when there is no crash key provided."""
    self.mock_current_user(user_email='test@chromium.org', is_admin=True)
    with self.assertRaises(Exception):
      self.test_app.get('/process/rerun-analysis')

  @mock.patch('common.crash_pipeline.RerunPipeline.start')
  def testHandleGetWhenThereIsNoCrashAnalysis(self, mock_run):
    """Tests that RerunPipeline didn't run if there is no crash analysis."""
    mock_run.return_value = None
    self.mock_current_user(user_email='test@chromium.org', is_admin=True)
    response = self.test_app.get('/process/rerun-analysis?key=dummy_key')
    self.assertEqual(response.status_int, 200)
    self.assertTrue(mock_run.called)

  @mock.patch('common.crash_pipeline.RerunPipeline.start')
  def testHandleGetRerunOneCrashAnalysis(self, mock_run):
    """Tests that RerunPipeline ran once if there is one crash analysis."""
    mock_run.return_value = None
    crash_analysis = CracasCrashAnalysis.Create({'signature': 'sig'})
    crash_analysis.put()

    self.mock_current_user(user_email='test@chromium.org', is_admin=True)
    response = self.test_app.get('/process/rerun-analysis?key=%s' %
                                 crash_analysis.key.urlsafe().decode('utf-8'))
    self.assertEqual(response.status_int, 200)
    self.assertTrue(mock_run.called)
