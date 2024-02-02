#!/usr/bin/env vpython3
# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import unittest
from unittest import mock

import pull_current_mesa_drivers as pcmd


class RequestsMockingUnittest(unittest.TestCase):

  def setUp(self):
    self._get_patcher = mock.patch.object(pcmd.requests, 'get')
    self._get_mock = self._get_patcher.start()
    self.addCleanup(self._get_patcher.stop)

  def _set_html(self, html_content) -> None:
    self._get_mock.return_value = mock.Mock(status_code=200, text=html_content)


class GetHtmlUnittest(RequestsMockingUnittest):

  def test_success(self):
    """Tests the happy path/success case when getting HTML."""
    self._set_html('foo')
    self.assertEqual(pcmd._get_html('foo.com'), 'foo')

  def test_non_okay_status_code(self):
    """Tests behavior when a non-OK status code is reported."""
    self._get_mock.return_value = mock.Mock(status_code=404, text='error')
    with self.assertRaisesRegex(
        RuntimeError,
        'Got non-OK response code 404 getting foo.com. Response text: error'):
      pcmd._get_html('foo.com')


class GetMostRecentReleaseUnittest(RequestsMockingUnittest):

  def test_success(self):
    """Tests the happy path/success case when getting the download info."""
    html = """\
<body>
<tr>
  <td>
    <a href="mesa-1.0.0.tar.gz">mesa-1.0.0.tar.xz</a>
  </td>
</tr>
<tr>
  <td>
    <a href="mesa-2.0.0.tar.gz">mesa-2.0.0.tar.xz</a>
  </td>
</tr>
</body>
"""
    self._set_html(html)
    soup = pcmd._get_soup('foo.com')
    version = pcmd._get_most_recent_release(soup)
    self.assertEqual(version, '2.0.0')

  def test_sorting(self):
    """Tests that version numbers are properly sorted."""
    html = """\
<body>
<tr>
  <td>
    <a href="mesa-1.1.0.tar.gz">mesa-1.1.0.tar.xz</a>
  </td>
</tr>
<tr>
  <td>
    <a href="mesa-1.0.0.tar.gz">mesa-1.0.0.tar.xz</a>
  </td>
</tr>
</body>
"""
    self._set_html(html)
    soup = pcmd._get_soup('foo.com')
    version = pcmd._get_most_recent_release(soup)
    self.assertEqual(version, '1.1.0')

    html = """\
<body>
<tr>
  <td>
    <a href="mesa-1.0.1.tar.gz">mesa-1.0.1.tar.xz</a>
  </td>
</tr>
<tr>
  <td>
    <a href="mesa-1.0.0.tar.gz">mesa-1.0.0.tar.xz</a>
  </td>
</tr>
</body>
"""
    self._set_html(html)
    soup = pcmd._get_soup('foo.com')
    version = pcmd._get_most_recent_release(soup)
    self.assertEqual(version, '1.0.1')

    html = """\
<body>
<tr>
  <td>
    <a href="mesa-9.0.0.tar.gz">mesa-9.0.0.tar.xz</a>
  </td>
</tr>
<tr>
  <td>
    <a href="mesa-10.0.0.tar.gz">mesa-10.0.0.tar.xz</a>
  </td>
</tr>
</body>
"""
    self._set_html(html)
    soup = pcmd._get_soup('foo.com')
    version = pcmd._get_most_recent_release(soup)
    self.assertEqual(version, '10.0.0')

  def test_no_valid_versions(self):
    """Tests behavior when no valid versions can be extracted."""
    html = """\
<body>
<tr>
  <td>
    <a href="mesa-1.0.tar.gz">mesa-1.0.tar.xz</a>
  </td>
</tr>
<tr>
  <td>
    <a href="mesa-2.0.0.0.tar.gz">mesa-2.0.0.0.tar.xz</a>
  </td>
</tr>
<tr>
  <td>
    <a href="mesa-3.0.0.zip">mesa-3.0.0.zip</a>
  </td>
</tr>
</body>
"""
    self._set_html(html)
    soup = pcmd._get_soup('foo.com')
    with self.assertRaisesRegex(RuntimeError,
                                'Did not extract any valid Mesa versions'):
      pcmd._get_most_recent_release(soup)


class CmdLatestVersionUnittest(RequestsMockingUnittest):

  def test_success(self):
    """Tests the happy path/success case when getting the latest version."""
    html = """\
<body>
<tr>
  <td>
    <a href="mesa-1.0.0.tar.gz">mesa-1.0.0.tar.xz</a>
  </td>
</tr>
<tr>
  <td>
    <a href="mesa-2.0.0.tar.gz">mesa-2.0.0.tar.xz</a>
  </td>
</tr>
</body>
"""
    self._set_html(html)
    pcmd.cmd_get_latest_version()


class CmdGetUrlUnittest(RequestsMockingUnittest):

  @mock.patch.dict(os.environ, {'_3PP_VERSION': '2.0.0'}, clear=True)
  def test_success(self):
    """Tests the happy path/success case when getting the download URL."""
    pcmd.cmd_get_url()

  @mock.patch.dict(os.environ, {}, clear=True)
  def test_no_3pp_version(self):
    """Tests behavior when no _3PP_VERSION is set."""
    with self.assertRaisesRegex(
        RuntimeError, 'get_url command requires _3PP_VERSION to be set'):
      pcmd.cmd_get_url()


if __name__ == '__main__':
  unittest.main()
