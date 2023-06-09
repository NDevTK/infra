# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import unittest
from redirect import redirect
from mock import patch
class TestRedirectApp(unittest.TestCase):

  def setUp(self):
    self.app = redirect.GenerateRedirectApp()
    self.app.config['TESTING'] = True

  def testNoRedirectIssueList(self):
    client = self.app.test_client()
    response = client.get('/p/project1/issues/list')
    self.assertEqual(response.status_code, 404)

  @patch("redirect.redirect_utils.GetRedirectURL")
  @patch("redirect.redirect_utils.GetSearchQuery")
  def testRedirectIssueList(self, fake_get_url, fake_get_search_query):
    client = self.app.test_client()
    response = client.get('/p/project1/issues/list')
    self.assertEqual(response.status_code, 302)

  def testNoRedirectCreateIssue(self):
    client = self.app.test_client()
    response = client.get('/p/project1/issues/entry')
    self.assertEqual(response.status_code, 404)

  @patch("redirect.redirect_utils.GetRedirectURL")
  def testRedirectCreateIssue(self, fake_get_url):
    fake_get_url.return_value = "test"
    client = self.app.test_client()
    response = client.get('/p/project1/issues/entry')
    self.assertEqual(response.status_code, 302)

  def testNoRedirectIssueDetail(self):
    client = self.app.test_client()
    response = client.get('/p/project1/issues/detail?id=1')
    self.assertEqual(response.status_code, 404)

  @patch("redirect.redirect_utils.GetRedirectURL")
  @patch("redirect.redirectissue.RedirectIssue.Get")
  def testRedirectIssueDetail(self, fake_get_url, fake_redirectIssue):
    fake_get_url.return_value = "test"
    fake_redirectIssue.return_value = "1"
    client = self.app.test_client()
    response = client.get('/p/project1/issues/detail?id=1')
    self.assertEqual(response.status_code, 200)

  @patch("redirect.redirect_utils.GetRedirectURL")
  def testRedirectIssueDetail(self, fake_get_url):
    fake_get_url.return_value = "test"
    client = self.app.test_client()
    response = client.get('/p/project1/issues/detail?id=10000001')
    self.assertEqual(response.status_code, 200)
