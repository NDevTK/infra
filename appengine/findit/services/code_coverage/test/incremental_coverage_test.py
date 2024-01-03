# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock
from datetime import datetime
from datetime import timedelta
from libs import time_util
import six

from google.appengine.ext import ndb
from libs import time_util
from waterfall.test.wf_testcase import WaterfallTestCase
from model.code_coverage import PresubmitCoverageData
from model.code_coverage import CLPatchset
from services.code_coverage import incremental_coverage
from services import bigquery_helper


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
        absolute_percentages=[
            {
                "path": "//chrome/browser/abc",
                "covered_lines": 1000,
                "total_lines": 1000
            },
            {
                "path": "//chrome/browser/xyz",
                "covered_lines": 500,
                "total_lines": 1000
            },
        ],
        incremental_percentages_unit=[
            {
                "path": "//chrome/browser/xyz",
                "covered_lines": 5,
                "total_lines": 10
            },
        ],
        absolute_percentages_unit=[{
            "path": "//chrome/browser/abc",
            "covered_lines": 10,
            "total_lines": 10
        }],
        data={})
    datastore_entity.put()
    run_id = 123
    incremental_coverage.ExportIncrementalCoverage(run_id)

    expected_bqrows = [{
        'cl_number': 123,
        'cl_patchset': 1,
        'server_host': "chromium-review.googlesource.com",
        'total_lines_inc_unit': None,
        'covered_lines_inc_unit': None,
        'total_lines_abs_unit': 10,
        'covered_lines_abs_unit': 10,
        'total_lines_inc_overall': 100,
        'covered_lines_inc_overall': 100,
        'total_lines_abs_overall': 1000,
        'covered_lines_abs_overall': 1000,
        'path': 'chrome/browser/abc',
        'insert_timestamp': '2020-09-21T00:00:00',
        'run_id': run_id
    }, {
        'cl_number': 123,
        'cl_patchset': 1,
        'server_host': "chromium-review.googlesource.com",
        'total_lines_inc_unit': 10,
        'covered_lines_inc_unit': 5,
        'total_lines_abs_unit': None,
        'covered_lines_abs_unit': None,
        'total_lines_inc_overall': 100,
        'covered_lines_inc_overall': 50,
        'total_lines_abs_overall': 1000,
        'covered_lines_abs_overall': 500,
        'path': 'chrome/browser/xyz',
        'insert_timestamp': '2020-09-21T00:00:00',
        'run_id': run_id
    }]
    calls = mocked_report_rows.call_args_list
    self.assertEqual(len(calls), 1)
    args, _ = calls[0]
    if six.PY2:
      self.assertItemsEqual(args[0], expected_bqrows)
    else:
      self.assertCountEqual(args[0], expected_bqrows)
    self.assertEqual(args[1], 'findit-for-me')
    self.assertEqual(args[2], 'code_coverage_summaries')
    self.assertEqual(args[3], 'incremental_coverage')
