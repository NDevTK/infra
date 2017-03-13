# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
from collections import OrderedDict
from datetime import datetime
from datetime import time
from datetime import timedelta
import json

from google.appengine.ext import ndb

from common.base_handler import BaseHandler, Permission
from libs import time_util

_DATE_FORMAT = '%Y-%m-%d'
_EQUAL_FILTER_TO_CONVERTER = OrderedDict([
    ('found_suspects', lambda x: x == 'yes'),
    ('has_regression_range', lambda x: x == 'yes'),
    ('suspected_cls_triage_status', int),
    ('regression_range_triage_status', int),
    ('signature', lambda x: x.strip().replace('%20', ' ')),
])

_DEFAULT_DISPLAY_COUNT = 500


class DashBoard(BaseHandler):
  PERMISSION_LEVEL = Permission.ANYONE

  @property
  def crash_analysis_cls(self):
    raise NotImplementedError()

  @property
  def client(self):
    raise NotImplementedError()

  def _GetStartDateAndEndDate(self):
    """Gets the start date and end date from dashboard filter."""
    midnight_today = datetime.combine(time_util.GetUTCNow(), time.min)
    midnight_yesterday = midnight_today - timedelta(days=1)
    midnight_tomorrow = midnight_today + timedelta(days=1)

    start = self.request.get('start_date')
    end = self.request.get('end_date')
    start_date = (datetime.strptime(start, _DATE_FORMAT) if start else
                  midnight_yesterday)
    end_date = (datetime.strptime(end, _DATE_FORMAT) if end else
                midnight_tomorrow)

    return start_date, end_date

  def _FilterCrashAnalysis(self, start_date, end_date):
    """Filters crash analysis by both unequal and equal filters."""
    # Filter by time range.
    query = self.crash_analysis_cls.query(
        self.crash_analysis_cls.requested_time >= start_date)
    query = self.crash_analysis_cls.query(
        self.crash_analysis_cls.requested_time < end_date)

    # Filter by property values of crash analyses.
    for equal_filter, converter in _EQUAL_FILTER_TO_CONVERTER.iteritems():
      if not self.request.get(equal_filter):
        continue

      query = query.filter(
          getattr(self.crash_analysis_cls, equal_filter) ==
          converter(self.request.get(equal_filter)))

    return query

  def HandleGet(self):
    """Shows crash analysis results in an HTML page."""
    start_date, end_date = self._GetStartDateAndEndDate()
    query = self._FilterCrashAnalysis(start_date, end_date)

    if self.request.get('count'):
      count = int(self.request.get('count'))
    else:
      count = _DEFAULT_DISPLAY_COUNT

    # TODO(katesonia): Add pagination here.
    crash_list = query.order(
        -self.crash_analysis_cls.requested_time).fetch(count)

    crashes = []
    for crash in crash_list:
      display_data = {
          'signature': crash.signature,
          'version': crash.crashed_version,
          'channel': crash.channel,
          'platform': crash.platform,
          'regression_range': ('' if not crash.has_regression_range else
                               crash.result['regression_range']),
          'suspected_cls': (crash.result.get('suspected_cls', [])
                            if crash.result else []),
          'suspected_project': (crash.result.get('suspected_project', '')
                                if crash.result else ''),
          'suspected_components': (crash.result.get('suspected_components', [])
                                   if crash.result else []),
          'stack_trace': crash.stack_trace,
          'historical_metadata': json.dumps(crash.historical_metadata),
          'key': crash.key.urlsafe()
      }
      crashes.append(display_data)

    data = {
        'start_date': start_date.strftime(_DATE_FORMAT),
        'end_date': end_date.strftime(_DATE_FORMAT),
        'found_suspects': self.request.get('found_suspects', '-1'),
        'has_regression_range': self.request.get('has_regression_range', '-1'),
        'suspected_cls_triage_status': self.request.get(
            'suspected_cls_triage_status', '-1'),
        'regression_range_triage_status': self.request.get(
            'regression_range_triage_status', '-1'),
        'client': self.client,
        'crashes': crashes,
        'signature': self.request.get('signature')
    }

    return {
        'template': 'crash/dashboard.html',
        'data': data
    }
