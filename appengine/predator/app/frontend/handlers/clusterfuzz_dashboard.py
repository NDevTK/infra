# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from analysis.type_enums import CrashClient
from common.model.clusterfuzz_analysis import ClusterfuzzAnalysis
from frontend.handlers.dashboard import DashBoard
from gae_libs.handlers.base_handler import Permission


class ClusterfuzzDashBoard(DashBoard):
  PERMISSION_LEVEL = Permission.ADMIN

  @property
  def crash_analysis_cls(self):
    return ClusterfuzzAnalysis

  @property
  def client(self):
    return CrashClient.CLUSTERFUZZ

  @property
  def template(self):
    return 'clusterfuzz_dashboard.html'

  def Order(self):
    return -self.crash_analysis_cls.signature

  def CrashDataToDisplay(self, crash_analyses):
    """Gets the crash data to display."""
    if not crash_analyses:
      return []

    crashes = []
    for crash in crash_analyses:
      display_data = {
          'signature': crash.signature,
          'testcase': crash.testcase,
          'version': crash.crashed_version,
          'job_type': crash.job_type,
          'crash_type': crash.crashed_type,
          'platform': crash.platform,
          'regression_range': crash.regression_range,
          'commits': crash.commits_in_regression_range,
          'error_name': crash.error_name or '',
          'suspected_cls': (crash.result.get('suspected_cls', [])
                            if crash.result else []),
          'suspected_project': (crash.result.get('suspected_project', '')
                                if crash.result else ''),
          'suspected_components': (crash.result.get('suspected_components', [])
                                   if crash.result else []),
          'key': crash.key.urlsafe()
      }
      crashes.append(display_data)

    crashes.sort(key=lambda crash: crash['signature'])
    return crashes
