# Copyright 2021 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock
from flask import Flask

from common.base_handler import BaseHandler
from handlers.code_coverage import export_gerrit_filter_coverage
from model.code_coverage import CoverageReportModifier
from services.code_coverage import gerrit_filter_coverage
from waterfall.test.wf_testcase import WaterfallTestCase


class ExportAllCoverageMetricsCronTest(WaterfallTestCase):
  app_module = Flask(__name__)
  app_module.add_url_rule(
      '/coverage/cron/all-gerrit-filter-coverage',
      view_func=export_gerrit_filter_coverage.ExportAllCoverageMetricsCron()
      .Handle,
      methods=['GET'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testTaskAddedToQueue(self, mocked_is_request_from_appself):
    response = self.test_app.get('/coverage/cron/all-gerrit-filter-coverage')
    self.assertEqual(200, response.status_int)
    response = self.test_app.get('/coverage/cron/all-gerrit-filter-coverage')
    self.assertEqual(200, response.status_int)

    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='all-gerrit-filter-coverage-queue')
    self.assertEqual(2, len(tasks))
    self.assertTrue(mocked_is_request_from_appself.called)


class ExportAllCoverageMetricsTest(WaterfallTestCase):
  app_module = Flask(__name__)
  app_module.add_url_rule(
      '/coverage/task/all-gerrit-filter-coverage',
      view_func=export_gerrit_filter_coverage.ExportAllCoverageMetrics().Handle,
      methods=['GET'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  def testCoverageFilesExported(self, mocked_is_request_from_appself):
    CoverageReportModifier(gerrit_hashtag='f1', id=123).put()
    CoverageReportModifier(gerrit_hashtag='f2', id=456).put()

    response = self.test_app.get('/coverage/task/all-gerrit-filter-coverage')
    self.assertEqual(200, response.status_int)

    tasks = self.taskqueue_stub.get_filtered_tasks(
        queue_names='experimental-coverage-queue')
    self.assertEqual(2, len(tasks))
    self.assertTrue(mocked_is_request_from_appself.called)


class ExportCoverageMetricsTest(WaterfallTestCase):
  app_module = Flask(__name__)
  app_module.add_url_rule(
      '/coverage/task/gerrit-filter-coverage',
      view_func=export_gerrit_filter_coverage.ExportCoverageMetrics().Handle,
      methods=['GET'])

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(gerrit_filter_coverage, 'ExportCoverage')
  def testCoverageLogicInvoked(self, mock_detect, _):
    CoverageReportModifier(gerrit_hashtag='f1', id=123).put()
    response = self.test_app.get(
        '/coverage/task/gerrit-filter-coverage?modifier_id=123', status=200)
    self.assertEqual(1, mock_detect.call_count)
    self.assertEqual(200, response.status_int)
