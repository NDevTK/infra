# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import unittest
import werkzeug

from mock import patch

from redirect import redirect_utils
from mock import patch


class TestRedirectUtils(unittest.TestCase):

  @patch("redirect.redirect_project_template.RedirectProjectTemplate.Get")
  def testNewIssueParams(self, fake_redirectProjectTemplate):
    fake_redirectProjectTemplate.return_value = None, None
    params = werkzeug.datastructures.MultiDict(
        [
            ('summary', 'this is a summary'),
            ('owner', 'test@google.com'),
            ('description', 'task'),
            ('cc', 'c1@google.com,c2@google.com'),
        ])
    expected = ('title=this+is+a+summary&description=task&'
                'cc=c1%40google.com%2Cc2%40google.com&assignee=test')

    get = redirect_utils.GetNewIssueParams(params, 'project')
    self.assertEqual(expected, get)

  @patch("redirect.redirect_project_template.RedirectProjectTemplate.Get")
  def testNewIssueParamsWithComponent(self, fake_redirectProjectTemplate):
    fake_redirectProjectTemplate.return_value = '1', '2'
    params = werkzeug.datastructures.MultiDict(
        [('summary', 'this is a summary'), ('owner', 'test@google.com')])
    expected = 'component=1&template=2&title=this+is+a+summary&assignee=test'

    get = redirect_utils.GetNewIssueParams(params, 'project')
    self.assertEqual(expected, get)

  @patch("redirect.redirect_project_template.RedirectProjectTemplate.Get")
  def testNewIssueParamsWithNoValidValue(self, fake_redirectProjectTemplate):
    fake_redirectProjectTemplate.return_value = None, None
    params = werkzeug.datastructures.MultiDict([('test', 'this is a test')])
    expected = ''
    get = redirect_utils.GetNewIssueParams(params, 'project')
    self.assertEqual(expected, get)

  @patch("redirect.redirect_custom_value.RedirectCustomValue.Get")
  def testGetSearchQuery(self, fake_redirectcustomevalue):
    fake_redirectcustomevalue.return_value = None, None
    params = werkzeug.datastructures.MultiDict(
        [('q', 'owner%3Ame%20has%3ARollout-Type')])
    expected = 'q=is%3Aopen+assignee%3A%28me%29'

    get = redirect_utils.GetSearchQuery('project', params)
    self.assertEqual(expected, get)

  @patch("redirect.redirect_custom_value.RedirectCustomValue.Get")
  def testGetSearchQueryWithCanValue(self, fake_redirectcustomevalue):
    fake_redirectcustomevalue.return_value = None, None
    params = werkzeug.datastructures.MultiDict([('can', 4)])
    expected = 'q=is%3Aopen+reporter%3A%28me%29'

    get = redirect_utils.GetSearchQuery('project', params)
    self.assertEqual(expected, get)
