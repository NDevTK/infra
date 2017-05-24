# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.ext import ndb

from analysis.type_enums import CrashClient
from common.model.crash_analysis import CrashAnalysis

_CLUSTERFUZZ_TESTCASE_URL_TEMPLATE = (
    'https://clusterfuzz.com/v2/testcase-detail/%s')


class ClusterfuzzAnalysis(CrashAnalysis):
  """Represents an analysis of a Clusterfuzz crash."""
  # Customized properties for Fracas crash.
  crashed_type = ndb.StringProperty()
  crashed_address = ndb.StringProperty()
  sanitizer = ndb.StringProperty()
  job_type = ndb.StringProperty()
  testcase = ndb.StringProperty()

  def Reset(self):
    super(ClusterfuzzAnalysis, self).Reset()
    self.crashed_type = None
    self.crashed_address = None
    self.sanitizer = None
    self.job_type = None
    self.testcase = None

  def Initialize(self, crash_data):
    """(Re)Initializes a CrashAnalysis ndb.Model from ``ClusterfuzzData``."""
    super(ClusterfuzzAnalysis, self).Initialize(crash_data)
    self.crashed_type = crash_data.crashed_type
    self.crashed_address = crash_data.crashed_address
    self.sanitizer = crash_data.sanitizer
    self.job_type = crash_data.job_type
    self.testcase = crash_data.testcase

  @property
  def client_id(self):  # pragma: no cover
    return CrashClient.CLUSTERFUZZ

  @property
  def crash_url(self):  # pragma: no cover
    return (_CLUSTERFUZZ_TESTCASE_URL_TEMPLATE % self.testcase
            if self.testcase else '')

  @property
  def customized_data(self):
    return {
        'regression_range': self.regression_range,
        'dependencies': self.dependencies,
        'dependency_rolls': self.dependency_rolls,
        'crashed_type': self.crashed_type,
        'crashed_address': self.crashed_address,
        'sanitizer': self.sanitizer,
        'job_type': self.job_type,
        'testcase': self.testcase
    }

  def ToJson(self):
    crash_json = super(ClusterfuzzAnalysis, self).ToJson()
    crash_json['customized_data'] = self.customized_data
    return crash_json
