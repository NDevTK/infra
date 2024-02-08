#!/usr/bin/env python3
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import urllib.request


def do_latest():
  # Get all releases (pre-release or not), look at the most recent one, and
  # grab the tag.
  print(
      json.load(
          urllib.request.urlopen(
              'https://api.github.com/repos/rr-debugger/rr/releases'
          )
      )[0]['tag_name']
  )


def get_download_url(version, platform):
  if platform != 'linux-amd64':
    raise ValueError('unsupported platform {}'.format(platform))

  url = (
      'https://github.com/rr-debugger/rr/releases/download/'
      '{0}/rr-{0}-Linux-x86_64.tar.gz'
  ).format(version)

  manifest = {
      'url': [url],
      'ext': '.tar.gz',
  }

  print(json.dumps(manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers(dest='action', required=True)

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest())

  download = sub.add_parser("get_url")
  download.set_defaults(func=lambda opts: get_download_url(
      os.environ['_3PP_VERSION'], os.environ['_3PP_PLATFORM']))

  opts = ap.parse_args()
  opts.func(opts)


if __name__ == '__main__':
  main()