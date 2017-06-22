# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import glob
import os
import unittest

import infra.path_hacks

class FullInfraPathTest(unittest.TestCase):
  def test_check_valid_full_infra_path(self):
    self.assertTrue(os.path.isdir(infra.path_hacks.full_infra_path))

  def test_path_stuff_on_windows
    infra_abs_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    gl = os.path.join(infra_abs_dir, 'tes*', '*_test.py')
    self.assertEqual(5, len(glob.glob(gl)))
