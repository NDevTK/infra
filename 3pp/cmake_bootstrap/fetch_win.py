#!/usr/bin/env python3
# Copyright 2022 The Chromium Authors
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
# https://github.com/activescott/lessmsi
WINDOWS_ASSET_RE = re.compile(r'^cmake-.*-windows-x86_64\.zip$')


def do_latest():
  print(
      json.load(
          urllib.request.urlopen(
              'https://api.github.com/repos/Kitware/CMake/releases/latest'))
      ['tag_name'].lstrip('v'))


def get_download_url(version):
  download_url = None

  target_tag = 'v%s' % (version,)
  for release in json.load(
      urllib.request.urlopen(
          'https://api.github.com/repos/Kitware/CMake/releases')):
    if str(release['tag_name']) == target_tag:
      for asset in release['assets']:
        asset_name = str(asset['name'])
        if WINDOWS_ASSET_RE.match(asset_name):
          download_url = asset['browser_download_url']
          break
      break
  if not download_url:
    raise Exception('could not find download_url')

  partial_manifest = {
      'url': [download_url],
      'ext': '.zip',
  }
  print(json.dumps(partial_manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers(dest='action', required=True)

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest())

  download = sub.add_parser("get_url")
  download.set_defaults(
      func=lambda opts: get_download_url(os.environ['_3PP_VERSION']))

  opts = ap.parse_args()
  return opts.func(opts)


if __name__ == '__main__':
  sys.exit(main())
