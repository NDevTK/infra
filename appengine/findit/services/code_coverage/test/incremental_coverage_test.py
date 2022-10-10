# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock
from datetime import datetime
from datetime import timedelta
from libs import time_util

from parameterized import parameterized

from google.appengine.ext import ndb
from libs import time_util
from waterfall.test.wf_testcase import WaterfallTestCase
from model.code_coverage import PresubmitCoverageData
from model.code_coverage import CLPatchset
from services.code_coverage import incremental_coverage
from services import bigquery_helper
from services import test_tag_util

_DEFAULT_LUCI_PROJECT = 'chromium'


class IncrementalCoverageTest(WaterfallTestCase):

  @mock.patch.object(time_util, 'GetUTCNow', return_value=datetime(2020, 9, 21))
  #   @mock.patch.object(bigquery_helper, '_GetBigqueryClient')
  @mock.patch.object(bigquery_helper, 'ReportRowsToBigquery', return_value={})
  def testExportIncrementalCoverageMetrics(self, mocked_report_rows, *_):
    datastore_entity = PresubmitCoverageData(
        key=ndb.Key('PresubmitCoverageData', 'chromium$123$1'),
        cl_patchset=CLPatchset(
            change=123,
            patchset=1,
            server_host="chromium-review.googlesource.com"),
        incremental_percentages=[
            {
                "path": "//chrome/browser/abc",
                "covered_lines": 100,
                "total_lines": 100
            },
            {
                "path": "//chrome/browser/xyz",
                "covered_lines": 50,
                "total_lines": 100
            },
        ],
        data={})
    datastore_entity.put()
    run_id = 123
    incremental_coverage.ExportIncrementalCoverage(run_id)

    expected_bqrows = [{
        'cl_number': 123,
        'cl_patchset': 1,
        'server_host': "chromium-review.googlesource.com",
        'total_lines': 100,
        'covered_lines': 100,
        'path': 'chrome/browser/abc',
        'insert_timestamp': '2020-09-21T00:00:00',
        'run_id': run_id
    }, {
        'cl_number': 123,
        'cl_patchset': 1,
        'server_host': "chromium-review.googlesource.com",
        'total_lines': 100,
        'covered_lines': 50,
        'path': 'chrome/browser/xyz',
        'insert_timestamp': '2020-09-21T00:00:00',
        'run_id': run_id
    }]
    mocked_report_rows.assert_called_with(expected_bqrows, 'findit-for-me',
                                          'code_coverage_summaries',
                                          'incremental_coverage')
