#!/usr/bin/env vpython3
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Pulls the most recent Linux Mesa drivers and uploads them to CIPD.

This is done by looking for the most recent stable version in the Mesa download
archive.
"""

import argparse
import json
import os
import posixpath
import re
import sys

import bs4
import requests

BASE_ARCHIVE_URL = 'https://archive.mesa3d.org'

VERSION_REGEX = re.compile(r'^mesa-(\d+\.\d+\.\d+)\.tar\.xz$')


def _get_html(url: str) -> str:
  """Gets the HTML response from a URL.

  Args:
    url: A string containing the URL to GET.

  Returns:
    A string containing the HTML response.
  """
  r = requests.get(url)
  if r.status_code != requests.codes.ok:
    raise RuntimeError(
        f'Got non-OK response code {r.status_code} getting {url}. Response '
        f'text: {r.text}')
  return r.text


def _get_soup(url: str) -> bs4.BeautifulSoup:
  html = _get_html(url)
  return bs4.BeautifulSoup(html, 'html.parser')


def _get_most_recent_release(soup: bs4.BeautifulSoup) -> str:
  """Extracts the most recent Mesa release from the directory listing.

  Args:
    soup: A BeautifulSoup object made from the Mesa archive directory listing.

  Returns:
    A string containing the most recent stable driver release version.
  """

  def _convert_version_to_tuple(version_str):
    major, minor, patch = version_str.split('.')
    return (int(major), int(minor), int(patch))

  versions = []
  links = soup.find_all('a')
  for l in links:
    match = VERSION_REGEX.match(l.string)
    if match:
      versions.append(match.group(1))
  if not versions:
    raise RuntimeError('Did not extract any valid Mesa versions')
  versions.sort(key=_convert_version_to_tuple)
  return versions[-1]


def cmd_get_latest_version() -> None:
  soup = _get_soup(BASE_ARCHIVE_URL)
  version = _get_most_recent_release(soup)
  print(version)


def cmd_get_url() -> None:
  requested_driver_version = os.environ.get('_3PP_VERSION')
  if not requested_driver_version:
    raise RuntimeError('get_url command requires _3PP_VERSION to be set')
  download_url = posixpath.join(BASE_ARCHIVE_URL,
                                f'mesa-{requested_driver_version}.tar.xz')
  download_manifest = {
      'url': [download_url],
      'ext': '.tar.xz',
  }
  print(json.dumps(download_manifest))


def _parse_args() -> argparse.Namespace:
  parser = argparse.ArgumentParser()
  subparsers = parser.add_subparsers(dest='action', required=True)

  get_latest_version_parser = subparsers.add_parser('latest')
  get_latest_version_parser.set_defaults(func=cmd_get_latest_version)

  get_url_parser = subparsers.add_parser('get_url')
  get_url_parser.set_defaults(func=cmd_get_url)

  # 'checkout' and 'verify' commands not supported.

  return parser.parse_args()


def main() -> int:
  args = _parse_args()
  args.func()
  return 0


if __name__ == '__main__':
  sys.exit(main())
