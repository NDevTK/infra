# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock
from flask import Flask

from common.base_handler import BaseHandler
from handlers.code_coverage import export_incremental_coverage
from services.code_coverage import incremental_coverage
from waterfall.test.wf_testcase import WaterfallTestCase


class ExportIncrementalCoverageMetricsCronTest(WaterfallTestCase):
  app_module = Flask(__name__)
  app_module.add_url_rule(
      '/coverage/cron/incremental-coverage',
      view_func=export_incremental_coverage
      .ExportIncrementalCoverageMetricsCron().Handle,
      methods=['GET'])

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
  app_module = Flask(__name__)
  app_module.add_url_rule(
      '/coverage/task/incremental-coverage',
      view_func=export_incremental_coverage.ExportIncrementalCoverageMetrics()
      .Handle,
      methods=['GET'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(incremental_coverage, 'ExportIncrementalCoverage')
  def testAbsoluteCoverageFilesExported(self, mock_detect, _):
    response = self.test_app.get(
        '/coverage/task/incremental-coverage', status=200)
    self.assertEqual(1, mock_detect.call_count)
    self.assertEqual(200, response.status_int)
