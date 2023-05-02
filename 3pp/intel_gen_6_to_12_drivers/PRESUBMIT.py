# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

USE_PYTHON3 = True

PRESUBMIT_VERSION = '2.0.0'


def _GetPCIDEnv(input_api):
  """Gets the common environment for running PCID tests."""
  pcid_env = dict(input_api.environ)
  current_path = input_api.PresubmitLocalPath()
  pcid_env.update({
      'PYTHONPATH': current_path,
      'PYTHONDONTWRITEBYTECODE': '1',
  })
  return pcid_env


def CheckScriptsUnittests(input_api, output_api):
  """Runs the unittests in the scripts/ directory."""
  return input_api.canned_checks.RunUnitTestsInDirectory(
      input_api,
      output_api,
      input_api.PresubmitLocalPath(),
      [r'^.+_unittest\.py$'],
      env=_GetPCIDEnv(input_api),
      run_on_python2=False,
      run_on_python3=True,
      skip_shebang_check=True,
  )
