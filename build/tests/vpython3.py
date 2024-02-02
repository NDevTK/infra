# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Test vpython3.cipd package by running 'vpython3 -c'."""

import os
import subprocess
import sys
import tempfile


# .exe on Windows.
EXE_SUFFIX = '.exe' if sys.platform == 'win32' else ''


def main():
  with tempfile.TemporaryDirectory() as tmp:
    res = subprocess.check_output([
        os.path.join(os.getcwd(), 'vpython3'+EXE_SUFFIX),
        '-vpython-root',
        tmp,
        '-c',
        'print(1)',
    ])
  return 0 if res.strip() == b'1' else 1


if __name__ == '__main__':
  sys.exit(main())
