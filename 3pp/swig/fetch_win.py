#!/usr/bin/env python3
# Copyright 2023 The Chromium Authors
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

# https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-repository-tags
# Returns a JSON-loadable text response like:
# [
#  {
#   "name": "v4.1.1",
#   "zipball_url": "https://api.github.com/repos/swig/swig/zipball/refs/tags/v4.1.1",
#   "tarball_url": "https://api.github.com/repos/swig/swig/tarball/refs/tags/v4.1.1",
#   "commit": {
#     "sha": "77323a0f07562b7d90d36181697a72a909b9519a",
#     "url": "https://api.github.com/repos/swig/swig/commits/77323a0f07562b7d90d36181697a72a909b9519a"
#   },
#   "node_id": "MDM6UmVmNzM3MjUyODpyZWZzL3RhZ3MvdjQuMS4x"
#  },
# ]
#
# The only thing we use here are the tag names. These are sorted and we
# select the highest version number as 'latest'. We do this rather than
# depend on the 'latest' API, which only reports the most recently-created
# release and as such may include backported security fixes.
TAGS = 'https://api.github.com/repos/swig/swig/tags'

# Once we have the release tag, download the prebuilt binary
# from this URL.
SWIGWIN_URL = 'http://ixpeering.dl.sourceforge.net/project/swig/swigwin/swigwin-{0}/swigwin-{0}.zip'


def do_latest():
  tags = json.load(urllib.request.urlopen(TAGS))
  latest = packaging.version.parse('0')
  for t in tags:
    # Skip any prereleases, which will have a '-' in the name.
    if '-' in t['name']:
      continue
    v = packaging.version.parse(t['name'][1:])  # Strip leading 'v'
    latest = max(latest, v)
  print(latest)


def get_download_url(version):
  print(json.dumps({
      'url': [SWIGWIN_URL.format(version)],
      'ext': '.zip',
  }))


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
