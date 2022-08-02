# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Standalone python script to return the comment from a ZIP archive. Intended
to be used by 'zip' recipe module internally. Should not be used elsewhere.
"""

from __future__ import print_function

import json
import os
import sys
import zipfile


def main():
  # See zip/api.py, def unzip(...) for format of |data|.
  data = json.load(sys.stdin)
  zip_file = data['zip_file']

  # Archive path should exist and be an absolute path to a file.
  assert os.path.exists(zip_file), zip_file
  assert os.path.isfile(zip_file), zip_file

  with zipfile.ZipFile(zip_file) as zip_file_obj:
    sys.stdout.buffer.write(zip_file_obj.comment)
  return 0


if __name__ == '__main__':
  sys.exit(main())
