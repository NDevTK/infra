#!/usr/bin/env python3
# Copyright 2024 The Chromium Authors
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
              'https://api.github.com/repos/PowerShell/PowerShell/releases/latest'
          ))['tag_name'])


_PLATFORMS = {
    'windows-386': 'win-x86',
    'windows-amd64': 'win-x64',
    'windows-arm64': 'win-arm64',
}

_EXTENSION = {
    'windows-386': '.msi',
    'windows-amd64': '.msi',
    'windows-arm64': '.zip',
}


def get_download_url(version, platform):
  if platform not in _PLATFORMS:
    raise ValueError('unsupported platform {}'.format(platform))

  extension = _EXTENSION[platform]

  url = ('https://github.com/PowerShell/PowerShell/releases/download/{version}/'
         'PowerShell-{version_short}-{platform}{extension}'.format(
             version=version,
             version_short=version[1:],
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
