# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Test vpython.cipd package by running 'vpython -c'."""

import subprocess
import sys
import os


# .exe on Windows.
EXE_SUFFIX = '.exe' if sys.platform == 'win32' else ''


def main():
  env = os.environ.copy()
  env['VPYTHON_LOG_TRACE'] = '1'

  with open(os.path.join(os.getcwd(), '.vpython'), 'w') as f:
    f.write('python_version: "2.7"')
  with open(os.path.join(os.getcwd(), '.vpython3'), 'w') as f:
    f.write('python_version: "3.8"')

  res = subprocess.check_output([
      os.path.join(os.getcwd(), 'vpython'+EXE_SUFFIX),
      '-c',
      'import sys; print sys.version_info.major',
  ], env=env).strip().decode()
  if res != '2':
    print('sys.version_info.major: expected 2, actual {}'.format(res))
    return 1

  res = subprocess.check_output([
      os.path.join(os.getcwd(), 'vpython3'+EXE_SUFFIX),
      '-c',
      'import sys; print(sys.version_info.major)',
  ], env=env).strip().decode()
  if res != '3':
    print('sys.version_info.major: expected 3, actual {}'.format(res))
    return 1

  with open(os.path.join(os.getcwd(), '.vpython3'), 'w') as f:
    f.write('python_version: "2.7"')
  try:
    res = subprocess.check_output([
        os.path.join(os.getcwd(), 'vpython3'+EXE_SUFFIX),
        '-c',
        'import sys; print(sys.version_info.major)',
    ], env=env)
  except subprocess.CalledProcessError:
    pass
  else:
    print('Python2 should only work with calling vpython')
    return 1

  return 0

if __name__ == '__main__':
  sys.exit(main())
