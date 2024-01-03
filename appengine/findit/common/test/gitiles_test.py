#!/usr/bin/env vpython
# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import base64
import datetime
import six
import sys
import unittest

import mock

from testing_utils import testing

from google.appengine.ext import ndb

# from components import auth
from common import gerrit
from common import gitiles

HOSTNAME = 'chromium.googlesource.com'
PROJECT = 'project'
REVISION = '404d1697dca23824bc1130061a5bd2be4e073922'
PATH = '/dir'


class GitilesTestCase(testing.AppengineTestCase):

  def setUp(self):
    super(GitilesTestCase, self).setUp()

  def mock_fetch(self, result):
    self.mock(gerrit, 'fetch_async', mock.Mock(return_value=ndb.Future()))
    gerrit.fetch_async.return_value.set_result(result)

  def test_get_file_content(self):
    req_path = 'project/+/main/a.txt'
    self.mock_fetch(base64.b64encode(six.ensure_binary('content')))

    content = gitiles.get_file_content(HOSTNAME, 'project', 'main', '/a.txt')
    gerrit.fetch_async.assert_called_once_with(
        HOSTNAME, req_path, headers={'Accept': 'text/plain'})
    self.assertEqual(content, six.ensure_binary('content'))


if __name__ == '__main__':
  if '-v' in sys.argv:
    unittest.TestCase.maxDiff = None
  unittest.main()
