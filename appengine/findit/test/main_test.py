# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import unittest

import backend_main
import default_main
import frontend_main

class MainTest(unittest.TestCase):

  def testImportInMainIsAllGood(self):
    pass  # Should not raise any exception if handler importing is all good.
