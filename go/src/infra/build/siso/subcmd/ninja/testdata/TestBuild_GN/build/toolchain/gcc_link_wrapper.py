# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import sys


def main():
  parser = argparse.ArgumentParser()
  parser.add_argument('--output', help='output file')
  parser.add_argument('rsp')
  options = parser.parse_args()

  data = ''
  if options.rsp.startswith('@'):
    rsp = options.rsp.removeprefix('@')
    with open(rsp) as f:
      data = f.read()
  with open(options.output, 'w') as f:
    f.write('link of %s' % data)
  return 0


if __name__ == "__main__":
  sys.exit(main())
