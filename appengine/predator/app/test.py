# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Sets up Predator's test environment and runs tests."""

import os
import sys

import pytest

import import_utils

if __name__ == '__main__':
  os.environ['GAE_RUNTIME'] = 'python3'
  os.environ['GAE_APPLICATION'] = 'testing-app'
  os.environ['SERVER_SOFTWARE'] = 'test'

  import warnings
  warnings.simplefilter("ignore")

  import_utils.FixImports()

  args = [
      '-Wignore', '--ignore', 'first_party/', '--ignore', 'third_party/python2',
      '--ignore', 'third_party/pipeline'
  ]
  sys.exit(pytest.main(args + sys.argv[1:]))
