#!/usr/bin/env python3
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import sys
import urllib.request


def do_latest():
  print(
      urllib.request.urlopen(
          'https://raw.githubusercontent.com/'
          'GoogleCloudPlatform/gsutil/master/VERSION').read().decode(
              'utf-8').strip())


def get_download_url(version):
  download_url = (
    'https://storage.googleapis.com/pub/gsutil_%s.tar.gz' % version)
  partial_manifest = {
    'url': [download_url],
    'ext': '.tar.gz',
  }
  print(json.dumps(partial_manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers()

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest())

  download = sub.add_parser("get_url")
  download.set_defaults(
    func=lambda opts: get_download_url(os.environ['_3PP_VERSION']))

  opts = ap.parse_args()
  return opts.func(opts)


if __name__ == '__main__':
  sys.exit(main())
