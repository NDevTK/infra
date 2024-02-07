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

# Make sure up-to-date root certificates are used.
urllib.request.install_opener(
    urllib.request.build_opener(
        urllib.request.HTTPSHandler(
            context=ssl.create_default_context(cafile=certifi.where()))))

RELEASES = 'https://api.github.com/repos/zephyrproject-rtos/sdk-ng/releases'
TAGGED_RELEASE = (
    'https://api.github.com/repos/zephyrproject-rtos/sdk-ng/releases/tags/v%s')


_PLATFORMS = {
    'linux-amd64': 'linux-x86_64',
    'linux-arm64': 'linux-aarch64',
    'mac-amd64': 'macos-x86_64',
    'mac-arm64': 'macos-aarch64',
    'windows-amd64': 'windows-x86_64',
}


def _get_extension(platform):
  return '.7z' if 'windows' in platform else '.tar.xz'


def _get_filename(version, platform):
  extension = _get_extension(platform)
  return f'zephyr-sdk-{version}_{_PLATFORMS[platform]}{extension}'


def do_latest(platform):
  if platform not in _PLATFORMS:
    raise ValueError('unsupported platform {}'.format(platform))

  releases = json.load(urllib.request.urlopen(RELEASES))
  latest = packaging.version.parse('0')
  for r in releases:
    if r['prerelease']:
      continue
    # Make sure the release contains an asset for the requested platform.
    version = r['tag_name'][1:]  # Strip leading 'v'
    filename = _get_filename(version, platform)
    found = False
    for a in r['assets']:
      if a['name'] == filename:
        found = True
        break
    if found:
      latest = max(latest, packaging.version.parse(version))
  print(latest)


def get_download_url(version, platform):
  if platform not in _PLATFORMS:
    raise ValueError('unsupported platform {}'.format(platform))

  name = _get_filename(version, platform)

  rsp = json.load(urllib.request.urlopen(TAGGED_RELEASE % version))
  actual_tag = rsp['tag_name'][1:]
  if version != actual_tag:
    raise ValueError('expected %s, actual is %s' % (version, actual_tag))

  for a in rsp['assets']:
    if a['name'] == name:
      partial_manifest = {
          'url': [a['browser_download_url']],
          'ext': _get_extension(platform),
      }
      print(json.dumps(partial_manifest))
      return
  raise ValueError('missing release for supported platform %s' % platform)


def main():
  ap = argparse.ArgumentParser()
  sub = ap.add_subparsers(dest='action', required=True)

  latest = sub.add_parser("latest")
  latest.set_defaults(func=lambda _opts: do_latest(os.environ['_3PP_PLATFORM']))

  download = sub.add_parser("get_url")
  download.set_defaults(func=lambda _opts: get_download_url(
      os.environ['_3PP_VERSION'], os.environ['_3PP_PLATFORM']))

  opts = ap.parse_args()
  opts.func(opts)


if __name__ == '__main__':
  main()
