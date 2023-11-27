#!/usr/bin/env python3
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import sys

VERSION = '7.27.0'

_EXTENSION = {
    'windows': '.msi',
}

_PLATFORMS = {
    'windows-386': 'x86',
    'windows-amd64': 'x64',
    'windows-arm64': 'x64',
}


def do_latest():
  print(VERSION)


def get_download_url(version, platform):
  if platform not in _PLATFORMS:
    raise ValueError('unsupported platform {}'.format(platform))

  extension = _EXTENSION[platform.split('-')[0]]

  url = ('https://downloads.puppetlabs.com/windows/puppet7/'
         'puppet-agent-{version}-{platform}{extension}'.format(
             version=version,
             platform=_PLATFORMS[platform],
             extension=extension,
         ))
  manifest = {
      'url': [url],
      'ext': extension,
  }
  print(json.dumps(manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers()

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest())

  download = sub.add_parser("get_url")
  download.set_defaults(func=lambda opts: get_download_url(
      os.environ['_3PP_VERSION'], os.environ['_3PP_PLATFORM']))

  opts = ap.parse_args()
  return opts.func(opts)


if __name__ == '__main__':
  sys.exit(main())
