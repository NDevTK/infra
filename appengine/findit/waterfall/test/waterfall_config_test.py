# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import mock

from waterfall import waterfall_config
from waterfall.test import wf_testcase


class MastersTest(wf_testcase.WaterfallTestCase):

  def testGetCodeCoverageSettings(self):
    self.assertEqual({
        'serve_presubmit_coverage_data': True
    }, waterfall_config.GetCodeCoverageSettings())

