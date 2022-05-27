#!/usr/bin/env python
# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import ssl
import sys
import urllib.request

import certifi

# Make sure up-to-date root certificates are used.
urllib.request.install_opener(
    urllib.request.build_opener(
        urllib.request.HTTPSHandler(
            context=ssl.create_default_context(cafile=certifi.where()))))


# https://developer.github.com/v3/repos/releases/#get-the-latest-release
# Returns a JSON-loadable text response like:
# {
#   ...,
#   "assets": [
#     {
#       ...,
#       "browser_download_url": "...",
#       ...,
#       "name": "protoc-3.8.0-win32.zip",
#       ...,
#     },
#     ...
#   ],
#   ...
#   "tag_name": "v3.8.0",
#   ...
# }
#
# Of interest are tag_name, which contains the version, and assets, which
# details platform-specific binaries. Under assets, name indicates the platform
# and browser_download_url indicates where to download a zip file containing the
# prebuilt binary.
LATEST = 'https://api.github.com/repos/protocolbuffers/protobuf/releases/latest'

# https://developer.github.com/v3/repos/releases/#get-a-release-by-tag-name
# Returns a JSON loadable text response like LATEST, but for a specific tag.
TAGGED_RELEASE = (
  'https://api.github.com/repos/protocolbuffers/protobuf/releases/tags/v%s')


# A mapping of supported CIPD platforms to the name of the corresponding protoc
# platform.
PROTOC_PLATFORMS = {
    'linux-amd64': 'linux-x86_64',
    'linux-arm64': 'linux-aarch_64',
    'mac-amd64': 'osx-x86_64',
    'mac-arm64': 'osx-aarch_64',
    'windows-386': 'win32',
    'windows-amd64': 'win64',
}


def do_latest():
  print(json.load(
      urllib.request.urlopen(LATEST))['tag_name'][1:])  # e.g. v3.8.0 -> 3.8.0


def get_download_url(version, platform):
  if platform not in PROTOC_PLATFORMS:
    raise ValueError('unsupported platform %s' % platform)

  # The protobuf team made releases labelled "osx-aarch_64" for awhile that
  # were actually amd64 releases. That was resolved with v21. Around the same
  # time they changed their release number format, moving from "v3.20" to "v21".
  # Explicitly disable any "v##.*" releases from being packaged for arm64 where
  # "##" is less than 21.
  if platform == 'mac-arm64':
    major_version = int(version.split('.')[0])
    if major_version < 21:
      raise ValueError(
          'unsupported platform %s for version %s' % (platform, version)
      )

  name = 'protoc-%s-%s.zip' % (version, PROTOC_PLATFORMS[platform])

  rsp = json.load(urllib.request.urlopen(TAGGED_RELEASE % version))
  actual_tag = rsp['tag_name'][1:]
  if version != actual_tag:
    raise ValueError('expected %s, actual is %s' % (version, actual_tag))

  for a in rsp['assets']:
    if a['name'] == name:
      partial_manifest = {
        'url': [a['browser_download_url']],
        'ext': '.zip',
      }
      print(json.dumps(partial_manifest))
      return
  raise ValueError('missing release for supported platform %s' % platform)


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
