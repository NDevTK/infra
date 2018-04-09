# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from libs.test_results.base_test_results import BaseTestResults
from waterfall.test import wf_testcase


class BaseTestResultsTest(wf_testcase.WaterfallTestCase):

  def testRemoveAllPrefixesFromTestName(self):
    self.assertEqual(
        'test',
        BaseTestResults(None).RemoveAllPrefixesFromTestName('test'))