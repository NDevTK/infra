#!/usr/bin/env python
# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os

httpd_version="2.4.53"
php_version="8.1.4"

def do_latest():
  print('httpd{0}.php{1}.chromium.1'.format(httpd_version, php_version))


def get_download_url():
  urls = [
      "https://www.apachelounge.com/download/VS16/binaries/"
      "httpd-2.4.53-win64-VS16.zip",
      "https://windows.php.net/downloads/releases/"
      "php-8.1.4-Win32-vs16-x64.zip",
  ]

  packages = [
      "httpd-2.4.53-win64-VS16.zip",
      "php-8.1.4-Win32-vs16-x64.zip",
  ]

  partial_manifest = {
      'url': urls,
      'name': packages,
      'ext': 'zip',
  }
  print(json.dumps(partial_manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers()

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest())

  download = sub.add_parser("get_url")
  download.set_defaults(func=lambda _opts: get_download_url())

  opts = ap.parse_args()
  opts.func(opts)


if __name__ == '__main__':
  main()
