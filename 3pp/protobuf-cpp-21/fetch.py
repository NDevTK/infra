#!/usr/bin/env python3
# Copyright 2019 The Chromium Authors
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

# TODO(crbug.com/1422445): Remove version limit once blocking issues are fixed.
VERSION_LIMIT = packaging.version.parse("22.0")


def do_latest():
  releases = json.load(urllib.request.urlopen(RELEASES))
  latest = packaging.version.parse('0')
  for r in releases:
    if r['prerelease']:
      continue
    v = packaging.version.parse(r['tag_name'][1:])  # Strip leading 'v'
    if v >= VERSION_LIMIT:
      continue
    latest = max(latest, v)
  print(latest)


def get_download_url(version):
  # Starting with v21, the language major version is decoupled from the
  # protocol buffers version.
  # Starting with v21.11, only zip files are provided. Prefer .tar.gz to
  # keep a consistent source location for older releases.
  if version >= '21.11':
    ext = '.zip'
  else:
    ext = '.tar.gz'
  name_re = r'protobuf-cpp-([0-9]*\.)?%s' % version + ext

  rsp = json.load(urllib.request.urlopen(TAGGED_RELEASE % version))
  actual_tag = rsp['tag_name'][1:]
  if version != actual_tag:
    raise ValueError('expected %s, actual is %s' % (version, actual_tag))

  for a in rsp['assets']:
    if re.match(name_re, a['name']):
      partial_manifest = {
          'url': [a['browser_download_url']],
          'ext': ext,
      }
      print(json.dumps(partial_manifest))
      return
  raise ValueError('missing release for protobuf-cpp')


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers()

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest())

  download = sub.add_parser("get_url")
  download.set_defaults(
      func=lambda opts: get_download_url(os.environ['_3PP_VERSION']))

  opts = ap.parse_args()
  return opts.func(opts)


if __name__ == '__main__':
  sys.exit(main())
