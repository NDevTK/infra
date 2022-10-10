# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.from datetime import datetime

import mock
import webapp2

from gae_libs.handlers.base_handler import BaseHandler
from handlers.code_coverage import export_incremental_coverage
from services.code_coverage import incremental_coverage
from waterfall.test.wf_testcase import WaterfallTestCase


class ExportIncrementalCoverageMetricsCronTest(WaterfallTestCase):
  app_module = webapp2.WSGIApplication([
      ('/coverage/cron/incremental-coverage',
       export_incremental_coverage.ExportIncrementalCoverageMetricsCron),
  ],
                                       debug=True)

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testTaskAddedToQueue(self, mocked_is_request_from_appself):
    response = self.test_app.get('/coverage/cron/incremental-coverage')
    self.assertEqual(200, response.status_int)
    response = self.test_app.get('/coverage/cron/incremental-coverage')
    self.assertEqual(200, response.status_int)

    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='incremental-coverage-queue')
    self.assertEqual(2, len(tasks))
    self.assertTrue(mocked_is_request_from_appself.called)


class ExportIncrementalCoverageMetricsTest(WaterfallTestCase):
  app_module = webapp2.WSGIApplication([
      ('/coverage/task/incremental-coverage',
       export_incremental_coverage.ExportIncrementalCoverageMetrics),
  ],
                                       debug=True)

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(incremental_coverage, 'ExportIncrementalCoverage')
  def testAbsoluteCoverageFilesExported(self, mock_detect, _):
    response = self.test_app.get(
        '/coverage/task/incremental-coverage', status=200)
    self.assertEqual(1, mock_detect.call_count)
    self.assertEqual(200, response.status_int)
