# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Sets up Findit's Python3 test environment and runs tests."""

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
      '--log-level',
      'INFO',
      '-Wignore',
      '--ignore',
      'first_party/',
      '--ignore',
      'third_party/',
      '--ignore',
      'libs/',
      '--ignore',
      'components/',
      '--ignore',
      'gae_libs/',
      '--ignore',
      'local_libs/',
  ]
  sys.exit(pytest.main(args + sys.argv[1:]))
