# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock
import webapp2

from gae_libs.handlers.base_handler import BaseHandler
from handlers.code_coverage import create_author_coverage
from model.code_coverage import CoveragePercentage
from model.code_coverage import CoverageReportModifier
from services.code_coverage import author_coverage
from services.code_coverage import referenced_coverage
from waterfall.test.wf_testcase import WaterfallTestCase


class CreateAuthorCoverageMetricsTest(WaterfallTestCase):
  app_module = webapp2.WSGIApplication([
      ('/coverage/cron/author-coverage',
       create_author_coverage.CreateAuthorCoverageMetrics),
  ],
                                       debug=True)

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(author_coverage, 'CreateAuthorCoverage')
  @mock.patch.object(
      create_author_coverage.CreateAuthorCoverageMetrics,
      '_GetSourceBuilders',
      return_value=['builder1'])
  def testAuthorCoverageLogicInvoked(self, mock_detect, *_):
    url = '/coverage/cron/author-coverage'
    self.test_app.get(url, status=200)
    self.assertEqual(1, mock_detect.call_count)
