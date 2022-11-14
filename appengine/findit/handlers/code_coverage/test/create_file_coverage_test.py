# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import mock
import webapp2

from gae_libs.handlers.base_handler import BaseHandler
from handlers.code_coverage import create_file_coverage
from handlers.code_coverage import utils
from model.code_coverage import FileCoverageByAuthor
from waterfall.test.wf_testcase import WaterfallTestCase


class ExportCoverageMetricsTest(WaterfallTestCase):
  app_module = webapp2.WSGIApplication([
      ('/coverage/task/create-file-coverage.*',
       create_file_coverage.CreateFileCoverageMetrics),
  ],
                                       debug=True)

  # This test uses the following code:
  # 1|  abc@google.com     1| line1
  # 2|  abc@google.com     0| line2
  # 3|  ^abc@google.com    1| line3
  # 4|  def@google.com     0| line4
  #
  # Where the first column is the line number, second column is the author
  # and the third column is the expected number of times the line is executed.
  # The caret(^) at the third line means that it was changed outside the
  # desired blame list window.

  @mock.patch.object(BaseHandler, 'IsRequestFromAppSelf', return_value=True)
  @mock.patch.object(create_file_coverage, '_GetChromiumToGooglerMapping')
  @mock.patch.object(utils, 'GetFileContentFromGs')
  @mock.patch.object(create_file_coverage, 'GetV2Build')
  def testCoverageLogicInvoked(self, mock_get_build, mock_get_file_content,
                               mock_account_mapping, *_):
    # Mock buildbucket v2 API.
    build = mock.Mock()
    build.builder.project = 'chrome'
    build.builder.bucket = 'coverage'
    build.builder.builder = 'android-code-coverage'
    build.output.properties.items.return_value = [
        ('coverage_is_presubmit', False),
        ('coverage_gs_bucket', 'code-coverage-data'),
        ('coverage_metadata_gs_paths', [
            'postsubmit/chromium.googlesource.com/chromium/src/'
            'aaaaa/coverage/android-code-coverage/123456789/metadata',
        ]), ('mimic_builder_names', ['android-code-coverage'])
    ]
    build.input.gitiles_commit = mock.Mock(
        host='chromium.googlesource.com',
        project='chromium/src',
        ref='refs/heads/main',
        id='123')
    mock_get_build.return_value = build
    jacoco_data = """
          <report name="my-app">
            <package name="com/mycompany/app">
              <sourcefile name="App.java">
                <line cb="0" ci="1" mb="0" mi="3" nr="1" />
                <line cb="0" ci="0" mb="0" mi="3" nr="2" />
                <line cb="0" ci="1" mb="0" mi="3" nr="3" />
                <line cb="0" ci="0" mb="0" mi="1" nr="4" />
              </sourcefile>
            </package>
          </report>
      """
    blamelist = {
        'com/mycompany/app/App.java': {
            'abc@google.com': [1, 2],
            'def@chromium.org': [4]
        }
    }
    mock_get_file_content.side_effect = [jacoco_data, json.dumps(blamelist)]
    mock_account_mapping.return_value = {'def@chromium.org': 'def@google.com'}
    self.test_app.get(
        '/coverage/task/create-file-coverage?build_id=123', status=200)

    fetched_file_coverage_data = FileCoverageByAuthor.query().fetch()
    self.assertEqual(2, len(fetched_file_coverage_data))
