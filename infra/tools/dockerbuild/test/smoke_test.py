# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import subprocess
import unittest

THIS_DIR = os.path.dirname(os.path.abspath(__file__))
SPEC = os.path.join(THIS_DIR, '..', 'standalone.vpython3')
TESTDATA = os.path.join(THIS_DIR, 'testdata')


class TestSmoke(unittest.TestCase):
  """Verifies some important wheels are actually usable.

  Assumes 'vpython3' is in PATH.
  """

  def test_check_requests(self):
    code, out = run_vpython(os.path.join(TESTDATA, 'check_requests.py'))
    if code:
      print(out)
      self.fail('Exit code %d' % code)


def run_vpython(script):
  """Runs the given script through vpython.

  Returns:
    (exit code, combined stdout+stderr).
  """
  try:
    return 0, subprocess.check_output([
        'vpython3',
        '-vpython-log-level',
        'debug',
        '-vpython-spec',
        SPEC,
        script,
    ],
                                      stderr=subprocess.STDOUT)
  except subprocess.CalledProcessError as e:
    return e.returncode, e.output


if __name__ == '__main__':
  unittest.main()
