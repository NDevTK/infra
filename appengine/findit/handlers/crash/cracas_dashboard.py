# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
from datetime import datetime
from datetime import time
from datetime import timedelta
import json

from common.base_handler import BaseHandler, Permission
from handlers.crash.dashboard import DashBoard
from libs import time_util
from model.crash.cracas_crash_analysis import CracasCrashAnalysis


_EQUAL_FILTER_TO_CONVERTER = {
    'found_suspects': lambda x: x == 'yes',
    'has_regression_range': lambda x: x == 'yes',
    'suspected_cls_triage_status': int,
    'regression_range_triage_status': int,
}

_DEFAULT_DISPLAY_COUNT = 500


class CracasDashBoard(DashBoard):
  PERMISSION_LEVEL = Permission.ANYONE

  @property
  def crash_analysis_cls(self):
    return CracasCrashAnalysis

  @property
  def client(self):
    return 'cracas'
