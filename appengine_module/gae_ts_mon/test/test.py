# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Sets up gae_ts_mon's test environment and runs tests."""

import os
import sys

import pytest

if __name__ == '__main__':
  os.environ['GAE_RUNTIME'] = 'python3'
  os.environ['GAE_APPLICATION'] = 'testing-app'
  os.environ['SERVER_SOFTWARE'] = 'test'

  # Ignore deprecation warnings in some old modules.
  ignore_deprecation_warnings = [
      'infra_libs.ts_mon.protos.' + s for s in [
          'acquisition_network_device_pb2',
          'acquisition_task_pb2',
          'any_pb2',
          'metrics_pb2',
          'timestamp_pb2',
      ]
  ]
  ignore_deprecation_warnings += [
      'jinja2.runtime',
      'jinja2.utils',
  ]

  args = ['-Werror']
  args += ['--ignore', 'test_support']
  args += ['--ignore', 'instrument_endpoint_test.py']
  args += ['--ignore', 'instrument_webapp2_test.py']
  for mod in ignore_deprecation_warnings:
    args += ['-W', 'ignore::DeprecationWarning:' + mod]

  os.chdir(os.path.dirname(__file__))
  sys.exit(
      pytest.main(
          args +
          ['--cov', '--cov-fail-under=100', '--cov-report', 'term-missing'] +
          sys.argv[1:]))
