# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import sys


def main():
  parser = argparse.ArgumentParser()
  parser.add_argument('-c', help='compile')
  parser.add_argument('-o', help='output')
  options = parser.parse_args()

  data = ''
  with open(options.c) as f:
    data = f.read()
  with open(options.o, 'w') as f:
    f.write('compile of %s' % data)
  return 0


if __name__ == "__main__":
  sys.exit(main())
