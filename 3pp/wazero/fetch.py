#!/usr/bin/env python3
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import urllib.request


def do_latest():
  print(
      json.load(
          urllib.request.urlopen(
              'https://api.github.com/repos/tetratelabs/wazero/releases/latest'))
      ['tag_name'].lstrip('v'))


_PLATFORMS = {
    'linux-amd64': 'linux_amd64',
}


def get_download_url(version, platform):
  if platform not in _PLATFORMS:
    raise ValueError('unsupported platform {}'.format(platform))

  extension = '.tar.gz'
  url = ('https://github.com/tetratelabs/wazero/releases/download/v{version}/'
         'wazero_{version}_{platform}{extension}'.format(
             version=version,
             platform=_PLATFORMS[platform],
             extension=extension,
         ))
  partial_manifest = {
      'url': [url],
      'ext': extension,
  }
  print(json.dumps(partial_manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers()

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest())

  download = sub.add_parser("get_url")
  download.set_defaults(func=lambda opts: get_download_url(
      os.environ['_3PP_VERSION'], os.environ['_3PP_PLATFORM']))

  opts = ap.parse_args()
  opts.func(opts)


if __name__ == '__main__':
  main()
