#!/usr/bin/env python3
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import ssl
import sys
import urllib.request

import certifi
import packaging.version

BASE_URL = 'https://nodejs.org/dist/'

# Follow the LTS release track.
_VERSION_LIMIT = packaging.version.parse("19.0")

# Make sure up-to-date root certificates are used.
urllib.request.install_opener(
    urllib.request.build_opener(
        urllib.request.HTTPSHandler(
            context=ssl.create_default_context(cafile=certifi.where()))))


def do_latest():
  data = json.load(urllib.request.urlopen(BASE_URL + 'index.json'))
  max_version, max_string = packaging.version.parse('0'), '0'
  for release in data:
    s = release['version'].lstrip('v')
    v = packaging.version.parse(s)
    if max_version < v < _VERSION_LIMIT:
      max_version = v
      max_string = s

  print(str(max_string))


def get_download_url(version, platform):
  targ_os, targ_arch = platform.split('-')
  ext = '.zip' if targ_os == 'windows' else '.tar.gz'
  fragment = {
      ('mac', 'amd64'): 'darwin-x64',
      ('mac', 'arm64'): 'darwin-arm64',
      ('linux', 'amd64'): 'linux-x64',
      ('linux', 'armv6l'): 'linux-armv6l',
      ('linux', 'arm64'): 'linux-arm64',
      ('windows', 'amd64'): 'win-x64',
      # Package emulated binary for now.
      ('windows', 'arm64'): 'win-x64',
  }[(targ_os, targ_arch)]
  download_url = (
      '%(base)s/v%(version)s/node-v%(version)s-%(fragment)s%(ext)s' % {
          'base': BASE_URL,
          'version': version,
          'fragment': fragment,
          'ext': ext
      })
  partial_manifest = {
      'url': [download_url],
      'ext': ext,
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
  return opts.func(opts)


if __name__ == '__main__':
  sys.exit(main())
