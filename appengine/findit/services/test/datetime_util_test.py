# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import datetime

from services import datetime_util
from waterfall.test import wf_testcase


class DatetimeUtilTest(wf_testcase.WaterfallTestCase):

  def testConvertDateTime(self):
    fmt = '%Y-%m-%dT%H:%M:%S.%f'
    time_string = '2016-02-10T18:32:06.538220'
    test_time = datetime_util.ConvertDateTime(time_string)
    time = datetime.datetime.strptime(time_string, fmt)
    self.assertEqual(test_time, time)

  def testConvertDateTimeNone(self):
    time_string = ''
    test_time = datetime_util.ConvertDateTime(time_string)
    self.assertIsNone(test_time)

  def testConvertDateTimefailure(self):
    with self.assertRaises(ValueError):
      datetime_util.ConvertDateTime('abc')