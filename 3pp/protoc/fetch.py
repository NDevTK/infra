#!/usr/bin/env python3
# Copyright 2019 The Chromium Authors
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

# Make sure up-to-date root certificates are used.
urllib.request.install_opener(
    urllib.request.build_opener(
        urllib.request.HTTPSHandler(
            context=ssl.create_default_context(cafile=certifi.where()))))


# https://docs.github.com/en/rest/releases/releases#list-releases
# Returns a JSON-loadable text response like:
# [
#  {
#   "url": "https://api.github.com/repos/protocolbuffers/protobuf/releases/78625517",
#   ...
#   "tag_name": "v21.7",
#   ...
#   "assets": [
#     {
#       ...,
#       "browser_download_url": "...",
#       ...,
#       "name": "protoc-21.7-win32.zip",
#       ...,
#     },
#     ...
#   ],
#   ...
# }
#
# The only thing we use here are the tag names. These are sorted and we
# select the highest version number as 'latest'. We do this rather than
# depend on the 'latest' API, which only reports the most recently-created
# release and as such may include backported security fixes.
RELEASES = 'https://api.github.com/repos/protocolbuffers/protobuf/releases'

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
    # Use emulated mode for windows-arm64 for now.
    'windows-arm64': 'win64',
}


def do_latest():
  releases = json.load(urllib.request.urlopen(RELEASES))
  latest = packaging.version.parse('0')
  for r in releases:
    if r['prerelease']:
      continue
    latest = max(latest, packaging.version.parse(
        r['tag_name'][1:]))  # Strip leading 'v'
  print(latest)


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
