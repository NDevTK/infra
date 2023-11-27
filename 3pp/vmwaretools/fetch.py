#!/usr/bin/env python3
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import re
import sys
import urllib.request

_PLATFORMS = {
    'windows-386': 'x86',
    'windows-amd64': 'x64',
}

_PLATFORMS_ALT = {
    'windows-386': 'i386',
    'windows-amd64': 'x86_64',
}

_EXTENSION = {
    'windows': '.exe',
}


def do_latest():
  request = urllib.request.urlopen(
      'https://packages.vmware.com/tools/esx/latest/windows/')
  latest = re.findall(r'HREF=".*-(\d+.\d+.\d+-\d+).iso"',
                      request.read().decode('utf-8'))[0]
  print(latest)


def get_download_url(version, platform):
  if platform not in _PLATFORMS:
    raise ValueError(f'unsupported platform {platform}')

  platform_alt = _PLATFORMS_ALT[platform]
  extension = _EXTENSION[platform.split('-')[0]]
  platform = _PLATFORMS[platform]

  base_url = f'https://packages.vmware.com/tools/esx/latest/windows/{platform}/'
  request = urllib.request.urlopen(base_url)
  file_name = re.findall(rf'HREF="(.*-{version}-{platform_alt}{extension})"',
                         request.read().decode('utf-8'))[0]
  url = (f'{base_url}{file_name}')

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
