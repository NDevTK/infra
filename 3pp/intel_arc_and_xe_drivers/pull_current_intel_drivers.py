#!/usr/bin/env vpython3
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Pulls the most recent Intel GPU drivers and uploads them to CIPD.

This is done by grabbing the HTML for a download page and scraping it. The URL
for a particular platform/generation combination is fixed, but the particular
driver version and its download link is determined by a dropdown menu on that
page.

This file needs to be kept in sync with the version in
intel_gen_7_to_10_drivers/ except for the constants in main() until 3pp code
sharing is properly supported on Windows.
"""

import argparse
import hashlib
import json
import os
import re
import sys

import bs4
import requests

BASE_URL = 'https://www.intel.com'

DRIVER_VERSION_REGEX = re.compile(r'^\s*(\d+\.\d+\.\d+\.\d+)\s+.*$')
# This can end up being used if the driver download contains multiple slightly
# different drivers for different GPUs. For example, for the combined Arc/Xe
# download, the reported driver version can be something like
# "31.0.101.5186_101.5234", which means that 31.0.101.5186 would be used for
# one GPU and 31.0.101.5234 would be used for the other.
MIXED_DRIVER_VERSION_REGEX = re.compile(
    r'^\s*(\d+\.\d+\.\d+\.\d+_\d+\.\d+)\s+.*$')
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
  driver_string = selected_drivers[0].string
  match = DRIVER_VERSION_REGEX.match(driver_string)
  if not match:
    match = MIXED_DRIVER_VERSION_REGEX.match(driver_string)
    if not match:
      raise RuntimeError('Unable to extract driver version from %s' %
                         driver_string)
  return match.group(1)


def _verify_driver_version_matches_3pp(soup: bs4.BeautifulSoup, url: str,
                                       url_suffix: str) -> None:
  """Asserts that the actual driver version is the same as the requested one.

  Args:
    soup: A BeautifulSoup object that has parsed the HTML for an Intel driver
        download page.
    url: A string containing the URL that provided the HTML parsed by |soup|.
    url_suffix: A string containing the URL suffix that uniquely identifies the
        driver.
  """
  # Ensure that there wasn't a race condition between us getting the driver
  # version and trying to get the download URL.
  requested_driver_version = os.environ.get('_3PP_VERSION')
  if not requested_driver_version:
    raise RuntimeError('Tried to get download URL without setting _3PP_VERSION '
                       'environment variable')
  actual_driver_version = _get_driver_version(soup, url, url_suffix)
  if requested_driver_version != actual_driver_version:
    raise RuntimeError('Mismatch between requested (%s) and actual (%s) driver '
                       'versions' %
                       (requested_driver_version, actual_driver_version))


def _get_file_download_button(soup: bs4.BeautifulSoup, url: str,
                              file_extension: str) -> bs4.Tag:
  """Finds the relevant download button for the driver file.

  Args:
    soup: A BeautifulSoup object that has parsed the HTML for an Intel driver
        download page.
    url: A string containing the URL that provided the HTML parsed by |soup|.
    file_extension: A string containing the file extension to check for such as
        .zip.

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
    if button['data-href'].endswith(file_extension):
      break
  else:
    raise RuntimeError('Unable to find %s download button on %s' %
                       (file_extension, url))
  return chosen_button


def _get_file_download_url(soup: bs4.BeautifulSoup, url: str,
                           file_extension: str) -> str:
  """Scrapes the driver file download URL from the download page.

  Args:
    soup: A BeautifulSoup object that has parsed the HTML for an Intel driver
        download page.
    url: A string containing the URL that provided the HTML parsed by |soup|.
    file_extension: A string containing the file extension to check for such as
        .zip.

  Returns:
    A string containing the parsed driver download URL.
  """
  download_button = _get_file_download_button(soup, url, file_extension)
  return download_button['data-href']


def _get_checksum(soup: bs4.BeautifulSoup, url: str,
                  file_extension: str) -> str:
  """Scrapes the driver checksum from the download page.

  Args:
    soup: A BeautifulSoup object that has parsed the HTML for an Intel driver
        download page.
    url: A string containing the URL that provided the HTML parsed by |soup|.
    file_extension: A string containing the file extension to check for such as
        .zip.

  Returns:
    A string containing the parsed checksum.
  """
  download_button = _get_file_download_button(soup, url, file_extension)
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


def cmd_get_latest_version(url: str, _: argparse.Namespace,
                           download_page_url_suffix: str, __: str, ___: str):
  soup = _get_soup(url)
  driver_version = _get_driver_version(soup, url, download_page_url_suffix)
  print(driver_version)


def cmd_get_url(url: str, _: argparse.Namespace, download_page_url_suffix: str,
                file_name_base: str, file_extension: str):
  soup = _get_soup(url)
  _verify_driver_version_matches_3pp(soup, url, download_page_url_suffix)

  download_url = _get_file_download_url(soup, url, file_extension)
  download_manifest = {
      'url': [download_url],
      'ext': file_extension,
      'name': [file_name_base + file_extension],
  }
  print(json.dumps(download_manifest))


def cmd_checkout(url: str, args: argparse.Namespace,
                 download_page_url_suffix: str, file_name_base: str,
                 file_extension: str):
  soup = _get_soup(url)
  _verify_driver_version_matches_3pp(soup, url, download_page_url_suffix)

  download_url = _get_file_download_url(soup, url, file_extension)
  checksum = _get_checksum(soup, url, file_extension)

  r = requests.get(download_url)
  if r.status_code != requests.codes.ok:
    raise RuntimeError('Failed do get driver from download url %s' %
                       download_url)

  _verify_file_checksum(r.content, checksum)
  filename = file_name_base + file_extension
  with open(os.path.join(args.checkout_path, filename), 'wb') as outfile:
    outfile.write(r.content)


def cmd_verify(url: str, args: argparse.Namespace, _: str, __: str,
               file_extension: str):
  soup = _get_soup(url)

  checksum = _get_checksum(soup, url, file_extension)
  with open(os.path.join(args.package_path, filename), 'rb') as infile:
    _verify_file_checksum(infile.read(), checksum)


def _parse_args() -> argparse.Namespace:
  parser = argparse.ArgumentParser()
  subparsers = parser.add_subparsers(dest='action', required=True)

  get_latest_version_parser = subparsers.add_parser('latest')
  get_latest_version_parser.set_defaults(func=cmd_get_latest_version)

  get_url_parser = subparsers.add_parser('get_url')
  get_url_parser.set_defaults(func=cmd_get_url)

  checkout_parser = subparsers.add_parser('checkout')
  checkout_parser.add_argument('checkout_path')
  checkout_parser.set_defaults(func=cmd_checkout)

  verify_parser = subparsers.add_parser('verify')
  verify_parser.add_argument('package_path')
  verify_parser.set_defaults(func=cmd_verify)

  return parser.parse_args()


def main() -> int:
  download_page_url_suffix = ('/content/www/us/en/download/785597/'
                              'intel-arc-iris-xe-graphics-windows.html')
  file_name_base = 'intel_arc_and_xe_driver'
  file_extension = '.exe'
  args = _parse_args()
  url = BASE_URL + download_page_url_suffix
  args.func(url, args, download_page_url_suffix, file_name_base, file_extension)


if __name__ == '__main__':
  sys.exit(main())
