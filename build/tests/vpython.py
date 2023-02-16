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
  res = subprocess.check_output([
      os.path.join('vpython'+EXE_SUFFIX),
      '-c',
      'print(1)',
  ])
  return 0 if res.strip() == b'1' else 1


if __name__ == '__main__':
  sys.exit(main())
