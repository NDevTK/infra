# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Backfill existing analysis data to BigQuery events tables.

Take the existing analysis data and upload it to BigQuery. A full design for
BigQuery events and the backfill itself can be found at go/findit-bq-events.
"""
import os
import sys

_FINDIT_DIR = os.path.join(
    os.path.dirname(__file__), os.path.pardir, os.path.pardir)
_THIRD_PARTY_DIR = os.path.join(
    os.path.dirname(__file__), os.path.pardir, os.path.pardir, 'third_party')
_FIRST_PARTY_DIR = os.path.join(
    os.path.dirname(__file__), os.path.pardir, os.path.pardir, 'first_party')
sys.path.insert(1, _FINDIT_DIR)
sys.path.insert(0, _THIRD_PARTY_DIR)

import google
google.__path__.insert(0,
                       os.path.join(
                           os.path.dirname(os.path.realpath(__file__)),
                           'third_party', 'google'))

import datetime
from local_libs import remote_api
from common.waterfall import failure_type
from libs import analysis_status
from model import suspected_cl_status
from model.flake.flake_culprit import FlakeCulprit
from model.flake.master_flake_analysis import MasterFlakeAnalysis
from model.proto.gen import findit_pb2
from model.wf_try_job import WfTryJob
from model.wf_suspected_cl import WfSuspectedCL
from model.wf_analysis import WfAnalysis
from services import event_reporting
from services.event_reporting import ReportCompileFailureAnalysisCompletionEvent
from services.event_reporting import ReportTestFailureAnalysisCompletionEvent
from services.event_reporting import ReportTestFlakeAnalysisCompletionEvent
from waterfall.flake import triggering_sources

# Active script for Findit production.
remote_api.EnableRemoteApi(app_id='findit-for-me')


def CanReportAnalysis(analysis):
  """Returns True if the analysis can be reported, False otherwise."""
  return analysis.start_time and analysis.end_time


def ReportTestFlakesForRange(start_time, end_time, start_cursor=None):
  """Report test flakes to BQ for the given range.

  Optional cursor can be specific to continue.

  Args:
    (datetime) start_time: Start of the range.
    (datetime) end_time: End of the range.
    (Cursor) start_cursor: Marker on where to start the query at.
  """
  analyses_query = MasterFlakeAnalysis.query(
      MasterFlakeAnalysis.request_time > start_time,
      MasterFlakeAnalysis.request_time < end_time)

  print "reporting events from {}  -->  {}".format(start_time, end_time)

  cursor = start_cursor
  more = True
  page_size = 100

  while more:
    print "fetching {} results...".format(page_size)
    analyses, cursor, more = analyses_query.fetch_page(
        page_size, start_cursor=cursor)
    print "starting batch at cursor: {}".format(cursor.to_websafe_string())

    offset = 0
    for analysis in analyses:
      offset += 1
      if not CanReportAnalysis(analysis):
        continue
      ReportTestFlakeAnalysisCompletionEvent(analysis)
      print "offset: {}".format(offset)


def ReportTestFailuresForRange(start_time, end_time, start_cursor=None):
  """Report test failures to BQ for the given range.

  Optional cursor can be specific to continue.

  Args:
    (datetime) start_time: Start of the range.
    (datetime) end_time: End of the range.
    (Cursor) start_cursor: Marker on where to start the query at.
  """
  analyses_query = WfAnalysis.query(WfAnalysis.build_start_time > start_time,
                                    WfAnalysis.build_start_time < end_time)

  print "reporting events from {}  -->  {}".format(start_time, end_time)

  cursor = start_cursor
  more = True
  page_size = 100

  while more:
    print "fetching {} results...".format(page_size)
    analyses, cursor, more = analyses_query.fetch_page(
        page_size, start_cursor=cursor)
    print "starting batch at cursor: {}".format(cursor.to_websafe_string())

    offset = 0
    for analysis in analyses:
      offset += 1
      if analysis.build_failure_type != failure_type.TEST:
        continue
      if not CanReportAnalysis(analysis):
        continue
      ReportTestFailureAnalysisCompletionEvent(analysis)
      print "offset: {}".format(offset)


def ReportCompileFailuresForRange(start_time, end_time, start_cursor=None):
  """Report compile failures to BQ for the given range.

  Optional cursor can be specific to continue.

  Args:
    (datetime) start_time: Start of the range.
    (datetime) end_time: End of the range.
    (Cursor) start_cursor: Marker on where to start the query at.
  """
  analyses_query = WfAnalysis.query(WfAnalysis.build_start_time > start_time,
                                    WfAnalysis.build_start_time < end_time)

  print "reporting events from {}  -->  {}".format(start_time, end_time)

  cursor = start_cursor
  more = True
  page_size = 100

  while more:
    print "fetching {} results...".format(page_size)
    analyses, cursor, more = analyses_query.fetch_page(
        page_size, start_cursor=cursor)
    print "starting batch at cursor: {}".format(cursor.to_websafe_string())

    offset = 0
    for analysis in analyses:
      offset += 1
      if analysis.build_failure_type != failure_type.COMPILE:
        continue
      if not CanReportAnalysis(analysis):
        continue
      ReportCompileFailureAnalysisCompletionEvent(analysis)
      print "offset: {}".format(offset)


def main():
  start_time = datetime.datetime(2017, 12, 1)
  end_time = datetime.datetime(2017, 12, 31)

  ReportTestFlakesForRange(start_time, end_time)
  # ReportTestFailuresForRange(start_time, end_time)


if __name__ == '__main__':
  main()
