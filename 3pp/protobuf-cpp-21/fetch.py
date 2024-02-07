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

# Protobuf v21 is no longer being updated, so freeze the release version.
LATEST = '21.12'

# https://developer.github.com/v3/repos/releases/#get-a-release-by-tag-name
# Returns a JSON loadable text response like LATEST, but for a specific tag.
TAGGED_RELEASE = (
    'https://api.github.com/repos/protocolbuffers/protobuf/releases/tags/v%s')


def do_latest():
  print(LATEST)


def get_download_url(version):
  ext = '.zip'
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
