#!/usr/bin/env python3
# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import re
import ssl
import sys
import urllib.request

import certifi

# Make sure up-to-date root certificates are used.
urllib.request.install_opener(
    urllib.request.build_opener(
        urllib.request.HTTPSHandler(
            context=ssl.create_default_context(cafile=certifi.where()))))


# A regex for a name of the release asset to package, available at
# https://github.com/git-for-windows/git/releases
WINDOWS_ASSET_RE = {
    '386': re.compile(r'^PortableGit-.*-32-bit\.7z\.exe$'),
    'amd64': re.compile(r'^PortableGit-.*-64-bit\.7z\.exe$'),
    'arm64': re.compile(r'^PortableGit-.*-64-bit\.7z\.exe$'),
}


def do_latest():
  print(
      json.load(
          urllib.request.urlopen(
              'https://api.github.com/repos/git-for-windows/git/releases/latest'
          ))['tag_name'].lstrip('v'))


def get_download_url(version, platform):
  arch = platform.split('-')[-1]
  download_url = None

  target_tag = 'v%s' % (version,)
  for release in json.load(
      urllib.request.urlopen(
          'https://api.github.com/repos/git-for-windows/git/releases')):
    if str(release['tag_name']) == target_tag:
      for asset in release['assets']:
        asset_name = str(asset['name'])
        if WINDOWS_ASSET_RE[arch].match(asset_name):
          download_url = asset['browser_download_url']
          break
      break
  if not download_url:
    raise Exception('could not find download_url')

  partial_manifest = {
    'url': [download_url],
    'ext': '.7z.exe',
  }
  print(json.dumps(partial_manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers()

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest())

  download = sub.add_parser("get_url")
  download.set_defaults(
    func=lambda opts: get_download_url(
      os.environ['_3PP_VERSION'], os.environ['_3PP_PLATFORM']))

  opts = ap.parse_args()
  return opts.func(opts)


if __name__ == '__main__':
  sys.exit(main())
