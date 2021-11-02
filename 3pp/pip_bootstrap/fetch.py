#!/usr/bin/env python
# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import re
import ssl
import sys
import urllib

import pkg_resources
import certifi

# Make sure up-to-date root certificates are used.
urllib._urlopener = urllib.FancyURLopener(
    context=ssl.create_default_context(cafile=certifi.where()))


def _get_wheel_url(pkgname, version):
  for filedata in json.load(
      urllib.urlopen('https://pypi.org/pypi/%s/json' %
                     pkgname))['releases'][version]:
    if filedata['packagetype'] == 'bdist_wheel':
      return filedata['url'], filedata['filename']
  raise AssertionError('could not find wheel for %s @ %s' % (pkgname, version))


def _get_version(pkgname, bad_versions=()):
  # Find the latest python2-compatible version.
  releases = json.load(
      urllib.urlopen('https://pypi.org/pypi/%s/json' % pkgname))['releases']
  versions = [pkg_resources.parse_version(v) for v in releases.keys()]
  for version in sorted(versions, reverse=True):
    version_str = str(version)
    for filedata in releases[version_str]:
      if (version_str not in bad_versions and
          filedata['packagetype'] == 'bdist_wheel' and
          not filedata['yanked'] and filedata['python_version'] != 'py3'):
        return version_str
  raise AssertionError('could not find a compatible version for %s' % pkgname)


def do_latest():
  setuptools_bad_versions = frozenset([
      '45.0.0',  # Advertises python_version='py2.py3', but also requires >= 3.5
  ])

  print 'pip%s.setuptools%s.wheel%s' % (_get_version(
      'pip'), _get_version('setuptools', bad_versions=setuptools_bad_versions),
                                        _get_version('wheel'))


def get_download_url(version):
  # split version pip<vers>.setuptools<vers>.wheel<vers>
  m = re.match(
    r'^pip(.*)\.setuptools(.*)\.wheel(.*)$',
    version)
  versions = {
    'pip': m.group(1),
    'setuptools': m.group(2),
    'wheel': m.group(3),
  }
  download_urls, name = [], []
  for pkgname, vers in versions.iteritems():
    url, filename = _get_wheel_url(pkgname, vers)
    download_urls.append(url)
    name.append(filename)
  partial_manifest = {
      'url': download_urls,
      'name': name,
      'ext': '.whl',
  }
  print(json.dumps(partial_manifest))


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
