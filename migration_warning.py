#!/usr/bin/env python3
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

YELLOW = '\033[0;33m'
NC = '\033[0m'  # No Color

import sys
import os
import platform


def main():
  if platform.system() == 'Windows':
    os.system('color')  # Enable text colors.

  print(f'{YELLOW}'
        '[WARNING] ACTION REQUIRED\n'
        'You are using an old infra gclient checkout. Please migrate '
        'your checkout to infra_superproject by following directions '
        'at go/infra-superproject-migration-guide or '
        'https://bit.ly/41qLeVY.\n'
        'Migration deadline: May 12th, 2023'
        f'{NC}')


if __name__ == '__main__':
  sys.exit(main())
