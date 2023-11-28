#!/usr/bin/env python3
# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import argparse
import json
import os
import re
import ssl
import subprocess
import sys
import urllib.request

import certifi
import packaging.version


# TODO: Find these files dynamically.
# List of files to download for installation.
_FILES = (
    'core.msi',
    'core_d.msi',
    'core_pdb.msi',
    'dev.msi',
    'dev_d.msi',
    'doc.msi',
    'exe.msi',
    'exe_d.msi',
    'exe_pdb.msi',
    'lib.msi',
    'lib_d.msi',
    'lib_pdb.msi',
    'path.msi',
    'pip.msi',
    'tcltk.msi',
    'tcltk_d.msi',
    'tcltk_pdb.msi',
    'test.msi',
    'test_d.msi',
    'test_pdb.msi',
    'tools.msi',
    'ucrt.msi',
)

# Make sure up-to-date root certificates are used.
urllib.request.install_opener(
    urllib.request.build_opener(
        urllib.request.HTTPSHandler(
            context=ssl.create_default_context(cafile=certifi.where()))))


def get_installer_suffix(platform):
  if platform == 'windows-386':
    return '.exe'
  # Package windows-amd64 to run in emulated mode on windows-arm64.
  if platform in ['windows-amd64', 'windows-arm64']:
    return '-amd64.exe'
  raise ValueError('fetch.py is only supported for windows-(386|amd64|arm64)')


# Only look at versions in 3.11.x for now.
_VERSION_LIMIT = packaging.version.parse("3.12.0a0")


def do_latest(platform):
  """This is pretty janky, but the apache generic Index page hasn't changed
  since forever. It contains links (a tags with href's) to the different
  version folders."""
  suf = get_installer_suffix(platform)
  # Find the highest version e.g. 3.8.0.
  page_data = urllib.request.urlopen('https://www.python.org/ftp/python/')
  highest = None
  href_re = re.compile(r'href="(\d+\.\d+\.\d+)/"')
  for m in href_re.finditer(page_data.read().decode('utf-8')):
    v = packaging.version.parse(m.group(1))
    if v < _VERSION_LIMIT:
      if not highest or v > highest:
        highest = v
  page_data = urllib.request.urlopen('https://www.python.org/ftp/python/%s/' %
                                     highest)
  # Find the highest release e.g. 3.8.0a4.
  highest = None
  href_re = re.compile(r'href="python-(\d+\.\d+\.\d+((a|b|rc)\d+)?)%s"' % suf)
  for m in href_re.finditer(page_data.read().decode('utf-8')):
    v = packaging.version.parse(m.group(1))
    if v < _VERSION_LIMIT:
      if not highest or v > highest:
        highest = v
  print(highest)


def get_download_url(version, platform):
  # e.g. 3.8.0a4 -> 3.8.0
  short = version
  short_re = re.compile(r'(\d+\.\d+\.\d+)')
  m = short_re.match(version)
  if m:
    short = m.group(0)
  path = 'amd64' if platform in ['windows-amd64', 'windows-arm64'] else 'win32'
  base_download_url = (
    'https://www.python.org/ftp/python/%(short)s/%(path)s/'
    % {'short': short, 'path': path}
  )
  download_urls, artifact_names = [], []
  for filename in _FILES:
    download_urls.append(base_download_url + filename)
    artifact_names.append(filename)
  partial_manifest = {
    'url': download_urls,
    'name': artifact_names,
    'ext': '.msi',
  }
  print(json.dumps(partial_manifest))


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers()

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest(os.environ['_3PP_PLATFORM']))

  download = sub.add_parser("get_url")
  download.set_defaults(
    func=lambda opts: get_download_url(
      os.environ['_3PP_VERSION'], os.environ['_3PP_PLATFORM']))

  opts = ap.parse_args()
  return opts.func(opts)


if __name__ == '__main__':
  sys.exit(main())
