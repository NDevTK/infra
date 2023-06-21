#!/usr/bin/env python3
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

RED = '\033[0;31m'
NC = '\033[0m'  # No Color

import sys
import os
import platform


def main():
  if platform.system() == 'Windows':
    os.system('color')  # Enable text colors.

  print(
      f'{RED}'
      '[ERROR] ACTION REQUIRED\n'
      'You are using an old infra gclient checkout and the migration '
      'deadline has passed. You may still try to migrate by '
      'following the directions at '
      'go/infra-superproject-migration-guide or '
      'https://bit.ly/41qLeVY. There is no gaurantee that these steps '
      'will still work and you may have to create a brand new checkout: '
      'https://chromium.googlesource.com/infra/infra/+/refs/heads/main/doc/source.md'
      f'{NC}')


if __name__ == '__main__':
  sys.exit(main())
