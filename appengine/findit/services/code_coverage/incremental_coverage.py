# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging
from datetime import datetime
from datetime import timedelta

from google.appengine.ext import ndb

from libs import time_util
from model.code_coverage import PresubmitCoverageData
from services import bigquery_helper

_PAGE_SIZE = 100

# Time period for which coverage report is to fetched and processed
_NUM_REPORT_DAYS = 14

_BQ_ROWS_BATCH_SIZE = 500


def ExportIncrementalCoverage(run_id):
  """Exports incremental coverage metrics to Bigquery for last _NUM_REPORT_DAYS.

  Reads presubmit coverage data from Datastore, add few other dimensions to it
  and exports it to a Bigquery table.

  """
  # NDB caches each result in the in-context cache while accessing.
  # This is problematic as due to the size of the result set,
  # cache grows beyond the memory quota. Turn this off to prevent oom errors.
  #
  # Read more at:
  # https://cloud.google.com/appengine/docs/standard/python/ndb/cache#incontext
  # https://github.com/googlecloudplatform/datastore-ndb-python/issues/156#issuecomment-110869490
  context = ndb.get_context()
  context.set_cache_policy(False)
  update_limit = datetime.now() - timedelta(days=_NUM_REPORT_DAYS)
  query = PresubmitCoverageData.query(
      PresubmitCoverageData.cl_patchset.server_host == \
        'chromium-review.googlesource.com',
       PresubmitCoverageData.update_timestamp >= update_limit)
  total_patchsets = 0
  more = True
  cursor = None
  bqrows = []
  while more:
    results, cursor, more = query.fetch_page(_PAGE_SIZE, start_cursor=cursor)
    for result in results:
      bqrows_new = _CreateBigqueryRows(result, run_id)
      if bqrows_new:
        bqrows.extend(bqrows_new)
      if len(bqrows) > _BQ_ROWS_BATCH_SIZE:
        bigquery_helper.ReportRowsToBigquery(bqrows, 'findit-for-me',
                                             'code_coverage_summaries',
                                             'incremental_coverage')
        bqrows = []
      total_patchsets += 1
  bigquery_helper.ReportRowsToBigquery(bqrows, 'findit-for-me',
                                       'code_coverage_summaries',
                                       'incremental_coverage')
  logging.info('Total patchsets processed = %d', total_patchsets)


def _CreateBigqueryRows(presubmit_coverage, run_id):
  """Create a bigquery row for incremental coverage.

  Returns a list of dict whose keys are column names and values are column
  values corresponding to the schema of the bigquery table.

  Args:
    presubmit_coverage (PresubmitCoverageData): The PresubmitCoverageData
    fetched from Datastore
  """
  total_lines_inc_unit = {
      x.path: x.total_lines
      for x in presubmit_coverage.incremental_percentages_unit
  }
  covered_lines_inc_unit = {
      x.path: x.covered_lines
      for x in presubmit_coverage.incremental_percentages_unit
  }
  total_lines_abs_unit = {
      x.path: x.total_lines
      for x in presubmit_coverage.absolute_percentages_unit
  }
  covered_lines_abs_unit = {
      x.path: x.covered_lines
      for x in presubmit_coverage.absolute_percentages_unit
  }

  total_lines_inc_overall = {
      x.path: x.total_lines for x in presubmit_coverage.incremental_percentages
  }
  covered_lines_inc_overall = {
      x.path: x.covered_lines
      for x in presubmit_coverage.incremental_percentages
  }
  total_lines_abs_overall = {
      x.path: x.total_lines for x in presubmit_coverage.absolute_percentages
  }
  covered_lines_abs_overall = {
      x.path: x.covered_lines for x in presubmit_coverage.absolute_percentages
  }

  paths = set(total_lines_inc_unit.keys()).union(
      covered_lines_inc_unit.keys()).union(total_lines_abs_unit.keys()).union(
          covered_lines_abs_unit.keys()).union(
              total_lines_inc_overall.keys()).union(
                  covered_lines_inc_overall.keys()).union(
                      total_lines_abs_overall.keys()).union(
                          covered_lines_abs_overall.keys())

  if not paths:
    return None
  coverage = []
  for path in paths:
    coverage.append({
        'run_id': run_id,
        'total_lines_inc_unit': total_lines_inc_unit.get(path, None),
        'covered_lines_inc_unit': covered_lines_inc_unit.get(path, None),
        'total_lines_abs_unit': total_lines_abs_unit.get(path, None),
        'covered_lines_abs_unit': covered_lines_abs_unit.get(path, None),
        'total_lines_inc_overall': total_lines_inc_overall.get(path, None),
        'covered_lines_inc_overall': covered_lines_inc_overall.get(path, None),
        'total_lines_abs_overall': total_lines_abs_overall.get(path, None),
        'covered_lines_abs_overall': covered_lines_abs_overall.get(path, None),

        # ignore the leading double slash(//)
        'path': path[2:],
        'cl_number': presubmit_coverage.cl_patchset.change,
        'cl_patchset': presubmit_coverage.cl_patchset.patchset,
        'server_host': presubmit_coverage.cl_patchset.server_host,
        'insert_timestamp': time_util.GetUTCNow().isoformat()
    })
  return coverage
