#!/usr/bin/env vpython3
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import hashlib
import os
import unittest
from unittest import mock

import pull_current_intel_drivers as pcid

import bs4
from pyfakefs import fake_filesystem_unittest


class GetHtmlUnittest(unittest.TestCase):

  def setUp(self):
    self._get_patcher = mock.patch.object(pcid.requests, 'get')
    self._get_mock = self._get_patcher.start()
    self.addCleanup(self._get_patcher.stop)

  def test_success(self):
    """Tests the happy path/success case when getting HTML."""
    self._get_mock.return_value = mock.Mock(status_code=200, text='text')
    self.assertEqual(pcid._get_html('foo.com'), 'text')

  def test_non_ok_status_code(self):
    """Tests behavior when a non-OK status code is reported."""
    self._get_mock.return_value = mock.Mock(status_code=404, text='error')
    with self.assertRaisesRegex(
        RuntimeError,
        'Got non-OK response code 404 getting foo.com. Response text: error'):
      pcid._get_html('foo.com')


class GetDriverVersionUnittest(unittest.TestCase):

  def test_success(self):
    """Tests the happy path/success case when getting the driver version."""
    html = """\
<body>
<select id="driver-select">
  <option selected value="suffix">1.2.3.4 (Latest)</option>
  <option value="suffix">0.1.2.3</option>
</select>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    self.assertEqual(pcid._get_driver_version(soup, '', 'suffix'), '1.2.3.4')

  def test_no_found_option(self):
    """Tests behavior when no valid option tag is found."""
    # Missing selected attribute.
    html = """\
<body>
<select id="driver-select">
  <option value="suffix">1.2.3.4 (Latest)</option>
  <option value="suffix">0.1.2.3</option>
</select>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find selected driver on url'):
      pcid._get_driver_version(soup, 'url', 'suffix')

    # Value attribute does not match what we're looking for.
    html = """\
<body>
<select id="driver-select">
  <option selected value="not-suffix">1.2.3.4 (Latest)</option>
  <option value="suffix">0.1.2.3</option>
</select>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find selected driver on url'):
      pcid._get_driver_version(soup, 'url', 'suffix')

  def test_multiple_found_options(self):
    """Tests behavior when multiple valid option tags are found."""
    html = """\
<body>
<select id="driver-select">
  <option selected value="suffix">1.2.3.4 (Latest)</option>
  <option value="suffix">0.1.2.3</option>
</select>
<select id="second-select">
  <option selected value="suffix">2.3.4.5 (Latest)</option>
  <option value="suffix">0.1.2.3</option>
</select>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(
        RuntimeError,
        'More than one option found for selected driver. Options: .*'):
      pcid._get_driver_version(soup, 'url', 'suffix')

  def test_non_matching_driver_syntax(self):
    """Tests behavior when the driver syntax does not match expectations."""
    html = """\
<body>
<select id="driver-select">
  <option selected value="suffix">1.2.3 (Latest)</option>
  <option value="suffix">0.1.2.3</option>
</select>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to extract driver version from .*'):
      pcid._get_driver_version(soup, 'url', 'suffix')


class VerifyDriverVersionMatches3ppUnittest(unittest.TestCase):

  @mock.patch.dict(os.environ, {'_3PP_VERSION': '1.2.3.4'})
  def test_success(self):
    """Tests the happy path/success case when verifying driver version."""
    html = """\
<body>
<select id="driver-select">
  <option selected value="suffix">1.2.3.4 (Latest)</option>
  <option value="suffix">0.1.2.3</option>
</select>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    pcid._verify_driver_version_matches_3pp(soup, 'url', 'suffix')

  @mock.patch.dict(os.environ, {}, clear=True)
  def test_no_3pp_version(self):
    """Tests that trying to verify without a 3pp version is an error."""
    soup = bs4.BeautifulSoup('<html></html>', 'html.parser')
    with self.assertRaisesRegex(
        RuntimeError,
        'Tried to get download URL without setting _3PP_VERSION environment '
        'variable'):
      pcid._verify_driver_version_matches_3pp(soup, 'url', 'suffix')

  @mock.patch.dict(os.environ, {'_3PP_VERSION': '2.3.4.5'})
  def test_mismatch(self):
    """Tests that verification fails if the versions do not match."""
    html = """\
<body>
<select id="driver-select">
  <option selected value="suffix">1.2.3.4 (Latest)</option>
  <option value="suffix">0.1.2.3</option>
</select>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(
        RuntimeError,
        'Mismatch between requested \\(2\\.3\\.4\\.5\\) and actual '
        '\\(1\\.2\\.3\\.4\\) driver versions'):
      pcid._verify_driver_version_matches_3pp(soup, 'url', 'suffix')


class GetZipDownloadButtonUnittest(unittest.TestCase):

  def test_success(self):
    """Tests the happy path/success case when getting the download button."""
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF5562068 </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    button = pcid._get_zip_download_button(soup, 'url')
    button_only_soup = bs4.BeautifulSoup(
        '<button data-wap data-wap_ref="download-button" '
        'data-href="foo.com/driver.zip"></button>', 'html.parser')
    self.assertEqual(button['data-href'], button_only_soup.button['data-href'])
    self.assertEqual(button['data-href'], 'foo.com/driver.zip')

  def test_no_buttons(self):
    """Tests behavior when no download buttons are found."""
    # No data-wap attribute.
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF5562068 </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find any download buttons on url'):
      pcid._get_zip_download_button(soup, 'url')

    # Incorrect data-wap_ref
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap_ref="upload-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF5562068 </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find any download buttons on url'):
      pcid._get_zip_download_button(soup, 'url')

  def test_non_zip_extension(self):
    """Tests behavior when the only files available are non-zip ones."""
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.exe">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF5562068 </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find \\.zip download button on url'):
      pcid._get_zip_download_button(soup, 'url')


class GetZipDownloadUrlUnittest(unittest.TestCase):

  def test_success(self):
    """Tests the happy path/success case when getting the URL."""
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF5562068 </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    download_url = pcid._get_zip_download_url(soup, 'url')
    self.assertEqual(download_url, 'foo.com/driver.zip')


class GetChecksumUnittest(unittest.TestCase):

  def test_success(self):
    """Tests the happy path/success case when getting the checksum."""
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF5562068 </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    checksum = pcid._get_checksum(soup, 'url')
    self.assertEqual(checksum, '1aff661be1111cfbff9dae54d51ece6ef5562068')

  def test_no_details_div(self):
    """Tests behavior when no details div is found."""
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find details element on url'):
      pcid._get_checksum(soup, 'url')

  def test_multiple_details_divs(self):
    """Tests behavior when multiple details divs are found."""
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF5562068 </li>
    </ul>
  </div>
  <div id="another-details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 400 MB </li>
      <li> SHA1: 1111111111111111111111111111111111111111 </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'More than one details element on url'):
      pcid._get_checksum(soup, 'url')

  def test_no_list_items(self):
    """Tests behavior when no li tags are found."""
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find detail list items on url'):
      pcid._get_checksum(soup, 'url')

  def test_no_valid_checksum(self):
    """Tests behavior when no valid checksum is present."""
    # Missing.
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find SHA1 checksum on url'):
      pcid._get_checksum(soup, 'url')

    # Too short.
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF556206 </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find SHA1 checksum on url'):
      pcid._get_checksum(soup, 'url')

    # Too long.
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF55620680 </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find SHA1 checksum on url'):
      pcid._get_checksum(soup, 'url')

    # Non-hex.
    html = """\
<body>
<div id="info-div">
  <div id="button-div">
    <button data-wap data-wap_ref="download-button" data-href="foo.com/driver.zip">
    </button>
  </div>
  <div id="details-div" class="dc-page-available-downloads-hero__details">
    <ul>
      <li> Size: 500 MB </li>
      <li> SHA1: 1AFF661BE1111CFBFF9DAE54D51ECE6EF556206G </li>
    </ul>
  </div>
</div>
</body>
"""
    soup = bs4.BeautifulSoup(html, 'html.parser')
    with self.assertRaisesRegex(RuntimeError,
                                'Unable to find SHA1 checksum on url'):
      pcid._get_checksum(soup, 'url')


if __name__ == '__main__':
  unittest.main()
