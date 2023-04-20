#!/usr/bin/env vpython3
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Pulls the most recent Intel GPU drivers and uploads them to CIPD.

This is done by grabbing the HTML for a download page and scraping it. The URL
for a particular platform/generation combination is fixed, but the particular
driver version and its download link is determined by a dropdown menu on that
page.
"""

import argparse
import hashlib
import json
import logging
import os
import re
import sys
import tempfile
from typing import (Optional, Tuple)

import bs4
import requests

BASE_URL = 'https://www.intel.com'
DOWNLOAD_PAGE_URL_SUFFIX = (
    '/content/www/us/en/download/762755/'
    'intel-6th-10th-gen-processor-graphics-windows.html')
ZIP_NAME = 'intel_gen_6_to_12_driver.zip'

DRIVER_VERSION_REGEX = re.compile(r'^\s*(\d+\.\d+\.\d+\.\d+)\s+.*$')
SHA1_REGEX = re.compile(r'^\s*SHA1:\s*([a-fA-F0-9]{40})\s*$')


def _get_html(url: str) -> str:
  """Gets the HTML response from a URL.

  Args:
    url: A string containing the URL to GET.

  Returns:
    None if an error occurs, otherwise a string containing the HTML response.
  """
  r = requests.get(url)
  if r.status_code != requests.codes.ok:
    raise RuntimeError(
        'Got non-OK response code %d getting %s. Response text: %s' %
        (r.status_code, url, r.text))
  return r.text


def _get_driver_version(soup: bs4.BeautifulSoup, url: str,
                        url_suffix: str) -> str:
  """Scrapes the driver version from the download page.

  Args:
    soup: A BeautifulSoup object that has parsed the HTML for an Intel driver
        download page.
    url: A string containing the URL that provided the HTML parsed by |soup|.
    url_suffix: A string containing the URL suffix that uniquely identifies the
        driver.

  Returns:
    None if an error occurs, otherwise a string containing the parsed driver
    version.
  """
  # Find the driver version for the most recent driver. This will be the
  # auto-selected option of the driver list dropdown, whose string value
  # should be in the format "1.2.3.4 (Latest)"
  selected_drivers = soup.find_all('option', selected=True, value=url_suffix)
  if not selected_drivers:
    raise RuntimeError('Unable to find selected driver on %s' % url)
  if len(selected_drivers) > 1:
    raise RuntimeError(
        'More than one option found for selected driver. Options: %s' %
        ', '.join([str(s) for s in selected_drivers]))
  match = DRIVER_VERSION_REGEX.match(selected_drivers[0].string)
  if not match:
    raise RuntimeError('Unable to extract driver version from %s' %
                       selected_drivers[0].string)
  return match.group(1)


def _verify_driver_version_matches_3pp(soup, url) -> None:
  """Asserts that the actual driver version is the same as the requested one.

  Args:
    soup: A BeautifulSoup object that has parsed the HTML for an Intel driver
        download page.
    url: A string containing the URL that provided the HTML parsed by |soup|.
  """
  # Ensure that there wasn't a race condition between us getting the driver
  # version and trying to get the download URL.
  requested_driver_version = os.environ.get('_3PP_VERSION')
  if not requested_driver_version:
    raise RuntimeError('Tried to get download URL without setting _3PP_VERSION '
                       'environment variable')
  actual_driver_version = _get_driver_version(soup, url,
                                              DOWNLOAD_PAGE_URL_SUFFIX)
  if requested_driver_version != actual_driver_version:
    raise RuntimeError('Mismatch between requested (%s) and actual (%s) driver '
                       'versions' %
                       (requested_driver_version, actual_driver_version))


def _get_zip_download_button(soup: bs4.BeautifulSoup, url: str) -> bs4.Tag:
  """Finds the relevant download button for the driver .zip

  Args:
    soup: A BeautifulSoup object that has parsed the HTML for an Intel driver
        download page.
    url: A string containing the URL that provided the HTML parsed by |soup|.

  Returns:
    A BeautifulSoup Tag object for the download button for the driver in .zip
    format.
  """
  download_buttons = soup.find_all(
      'button', attrs={
          'data-wap': True,
          'data-wap_ref': 'download-button'
      })
  if not download_buttons:
    raise RuntimeError('Unable to find any download buttons on %s' % url)

  # Pull the download URL from the download button.
  chosen_button = None
  for button in download_buttons:
    chosen_button = button
    if button['data-href'].endswith('.zip'):
      break
  else:
    raise RuntimeError('Unable to find .zip download button on %s' % url)
  return chosen_button


def _get_zip_download_url(soup: bs4.BeautifulSoup, url: str) -> str:
  """Scrapes the driver .zip download URL from the download page.

  Args:
    soup: A BeautifulSoup object that has parsed the HTML for an Intel driver
        download page.
    url: A string containing the URL that provided the HTML parsed by |soup|.

  Returns:
    A string containing the parsed driver download URL.
  """
  download_button = _get_zip_download_button(soup, url)
  return download_button['data-href']


def _get_checksum(soup: bs4.BeautifulSoup, url: str) -> str:
  """Scrapes the driver checksum from the download page.

  Args:
    soup: A BeautifulSoup object that has parsed the HTML for an Intel driver
        download page.
    url: A string containing the URL that provided the HTML parsed by |soup|.

  Returns:
    A string containing the parsed checksum.
  """
  download_button = _get_zip_download_button(soup, url)
  # Pull the SHA1 checksum from the button's sibling element.
  details = download_button.parent.parent.find_all(
      'div', attrs={'class': 'dc-page-available-downloads-hero__details'})
  if not details:
    raise RuntimeError('Unable to find details element on %s' % url)
  if len(details) > 1:
    raise RuntimeError('More than one details element on %s' % url)
  details = details[0]
  list_items = details.find_all('li')
  if not list_items:
    raise RuntimeError('Unable to find detail list items on %s' % url)

  checksum = None
  for li in list_items:
    match = SHA1_REGEX.match(li.string)
    if match:
      checksum = match.group(1).lower()
      break
  else:
    raise RuntimeError('Unable to find SHA1 checksum on %s' % url)

  return checksum


def _verify_file_checksum(data: bytes, checksum: str) -> None:
  """Verifies a SHA1 checksum.

  Args:
    data: Bytes containing the file data to verify.
    checksum: A string containing the SHA1 checksum of |data|.
  """
  hasher = hashlib.sha1()
  hasher.update(data)
  digest = hasher.hexdigest()
  if digest != checksum:
    raise RuntimeError('SHA1 digest %s did not match checksum %s.' %
                       (digest, checksum))


def _get_soup(url) -> bs4.BeautifulSoup:
  html = _get_html(url)
  return bs4.BeautifulSoup(html, 'html.parser')


def cmd_get_latest_version(url, _):
  soup = _get_soup(url)
  driver_version = _get_driver_version(soup, url, DOWNLOAD_PAGE_URL_SUFFIX)
  print(driver_version)


def cmd_get_url(url, _):
  soup = _get_soup(url)
  _verify_driver_version_matches_3pp(soup, url)

  download_url = _get_zip_download_url(soup, url)
  download_manifest = {
      'url': [download_url],
      'ext': '.zip',
      'name': [ZIP_NAME],
  }
  print(json.dumps(download_manifest))


def cmd_checkout(url, args):
  soup = _get_soup(url)
  _verify_driver_version_matches_3pp(soup, url)

  download_url = _get_zip_download_url(soup, url)
  checksum = _get_checksum(soup, url)

  r = requests.get(download_url)
  if r.status_code != requests.codes.ok:
    raise RuntimeError('Failed do get driver from download url %s' %
                       download_url)

  _verify_file_checksum(r.content, checksum)
  with open(os.path.join(args.checkout_path, ZIP_NAME), 'wb') as outfile:
    outfile.write(r.content)


def cmd_verify(url, args):
  soup = _get_soup(url)

  checksum = _get_checksum(soup, url)
  with open(os.path.join(args.package_path, ZIP_NAME), 'rb') as infile:
    _verify_file_checksum(infile.read(), checksum)


def _parse_args():
  parser = argparse.ArgumentParser()
  subparsers = parser.add_subparsers()

  get_latest_version_parser = subparsers.add_parser('latest')
  get_latest_version_parser.set_defaults(func=cmd_get_latest_version)

  get_url_parser = subparsers.add_parser('get_url')
  get_url_parser.set_defaults(func=cmd_get_url)

  checkout_parser = subparsers.add_argument('checkout')
  checkout_parser.add_argument('checkout_path')
  checkout_parser.set_defaults(func=cmd_checkout)

  verify_parser = subparsers.add_parser('verify')
  verify_parser.add_argument('package_path')
  verify_parser.set_defaults(func=cmd_verify)

  return parser.parse_args()


def main() -> int:
  args = _parse_args()
  url = BASE_URL + DOWNLOAD_PAGE_URL_SUFFIX
  args.func(url, args)


if __name__ == '__main__':
  sys.exit(main())
