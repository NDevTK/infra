#!/usr/bin/env python3
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from __future__ import print_function

import argparse
import json
import os
import sys
import urllib.request

import packaging.version


def _gae_platform():
  osname, arch = os.environ['_3PP_PLATFORM'].split('-')
  osname = {'mac': 'darwin'}.get(osname, osname)
  return '%s_%s' % (osname, arch)


# SDKs are like `go_appengine_sdk_darwin_386-1.9.77.zip`
ZIP_PREFIX = 'go_appengine_sdk_' + _gae_platform() + '-'


def do_latest():
  BASE_URL = 'https://www.googleapis.com/storage/v1/b/appengine-sdks/o/'
  url = BASE_URL+'?prefix=featured/%s&delimiter=/' % ZIP_PREFIX
  print("Hitting %r" % url, file=sys.stderr)
  data = json.load(urllib.request.urlopen(url))
  max_ver, max_string = packaging.version.parse('0'), '0'
  for obj in data['items']:
    ver_string = obj['name'].split('/')[-1].lstrip(ZIP_PREFIX).rstrip('.zip')
    ver = packaging.version.parse(ver_string)
    if ver > max_ver:
      max_ver = ver
      max_string = ver_string

  if max_string == '0':
    print("GOT DATA")
    for obj in data['items']:
      print(obj)
    raise Exception('failed to find a version')

  print(max_string)


def get_download_url(version):
  URL = (
    'https://www.googleapis.com/download/storage/v1/b/appengine-sdks/'
    'o/featured%%2F%s%s.zip?alt=media' % (ZIP_PREFIX, version)
  )
  partial_manifest = {
    'url': [URL],
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
