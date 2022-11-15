# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock
import webapp2

from gae_libs.handlers.base_handler import BaseHandler
from handlers.code_coverage import create_author_coverage
from model.code_coverage import CoverageReportModifier
from services.code_coverage import author_coverage
from waterfall.test.wf_testcase import WaterfallTestCase


class CreateAuthorCoverageMetricsCronTest(WaterfallTestCase):
  app_module = webapp2.WSGIApplication([
      ('/coverage/cron/author-coverage',
       create_author_coverage.CreateAuthorCoverageMetricsCron),
  ],
                                       debug=True)

  @mock.patch.object(
      create_author_coverage.CreateAuthorCoverageMetricsCron,
      '_GetSourceBuilders',
      return_value=['android-code-coverage', 'android-code-coverage_unit'])
  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testTaskAddedToQueue(self, mocked_is_request_from_appself, _):
    response = self.test_app.get('/coverage/cron/author-coverage')
    self.assertEqual(200, response.status_int)
    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='experimental-coverage-queue')
    self.assertEqual(2, len(tasks))
    self.assertTrue(mocked_is_request_from_appself.called)


class CreateAuthorCoverageMetricsTest(WaterfallTestCase):
  app_module = webapp2.WSGIApplication([
      ('/coverage/task/author-coverage',
       create_author_coverage.CreateAuthorCoverageMetrics),
  ],
                                       debug=True)

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(author_coverage, 'CreateAuthorCoverage')
  def testAuthorCoverageLogicInvoked(self, mock_detect, *_):
    url = ('/coverage/task/author-coverage' '?builder=android-code-coverage')
    response = self.test_app.get(url, status=200)
    self.assertEqual(1, mock_detect.call_count)
    self.assertEqual(200, response.status_int)
