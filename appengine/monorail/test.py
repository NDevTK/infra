# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Sets up Monorail's test environment and runs tests."""

import os
import sys

import pytest

# gae_ts_mon's __init__.py does some import magic to create an infra_libs
# package, so we need to import it before importing Monorail packages.
import gae_ts_mon

import import_utils

if __name__ == '__main__':
  os.environ['GAE_RUNTIME'] = 'python3'
  os.environ['GAE_APPLICATION'] = 'testing-app'
  os.environ['SERVER_SOFTWARE'] = 'test'

  import_utils.FixImports()

  args = ['--ignore', 'components', '--ignore', 'gae_ts_mon'] + sys.argv[1:]
  sys.exit(pytest.main(args))
