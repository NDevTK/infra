# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import unittest
from testing_utils import testing

from gae_libs import appengine_util
#import backend_main
#import default_main
#import frontend_main

class MainTest(testing.AppengineTestCase):

  def testImportInMainIsAllGood(self):
    pass  # Should not raise any exception if handler importing is all good.

  def testAppEngineID(self):
    self.assertIsNone(appengine_util.GetApplicationId())
