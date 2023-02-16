# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Unit tests for servlet base class module."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import time
import mock
import unittest
import logging

from google.appengine.ext import testbed

from framework import framework_constants
from framework import servlet
from framework import servlet_helpers
from framework import xsrf
from mrproto import project_pb2
from mrproto import tracker_pb2
from mrproto import user_pb2
from services import service_manager
from testing import fake
from testing import testing_helpers


class TestableServlet(servlet.Servlet):
  """A tiny concrete subclass of abstract class Servlet."""

  def __init__(self, services=None, do_post_redirect=True):
    super(TestableServlet, self).__init__(services=services)
    self.do_post_redirect = do_post_redirect
    self.seen_post_data = None


class ServletTest(unittest.TestCase):

  def setUp(self):
    services = service_manager.Services(
        project=fake.ProjectService(),
        project_star=fake.ProjectStarService(),
        user=fake.UserService(),
        usergroup=fake.UserGroupService())
    services.user.TestAddUser('user@example.com', 111)
    self.page_class = servlet.Servlet(services=services)
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_user_stub()
    self.testbed.init_memcache_stub()
    self.testbed.init_datastore_v3_stub()

  def tearDown(self):
    self.testbed.deactivate()

  def testDefaultValues(self):
    self.assertEqual(None, self.page_class._MAIN_TAB_MODE)
    self.assertTrue(self.page_class._TEMPLATE_PATH.endswith('/templates/'))
    self.assertEqual(None, self.page_class._PAGE_TEMPLATE)

  @mock.patch('flask.abort')
  def testCheckForMovedProject_NoRedirect(self, mock_abort):
    project = fake.Project(
        project_name='proj', state=project_pb2.ProjectState.LIVE)
    request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj', project=project)
    self.page_class._CheckForMovedProject(mr, request)
    mock_abort.assert_not_called()

    request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/source/browse/p/adminAdvanced', project=project)
    self.page_class._CheckForMovedProject(mr, request)
    mock_abort.assert_not_called()

  @mock.patch('flask.redirect')
  def testCheckForMovedProject_Redirect(self, mock_redirect):
    project = fake.Project(project_name='proj', moved_to='http://example.com')
    request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj', project=project)
    self.page_class.request_path = '/p/test'
    self.page_class._CheckForMovedProject(mr, request)
    mock_redirect.assert_called_once_with(
        'http://127.0.0.1/hosting/moved?project=proj', code=302)

  def testGatherBaseData(self):
    project = self.page_class.services.project.TestAddProject(
        'testproj', state=project_pb2.ProjectState.LIVE)
    project.cached_content_timestamp = 12345

    (_request, mr) = testing_helpers.GetRequestObjects(
        path='/p/testproj/feeds', project=project)
    nonce = '1a2b3c4d5e6f7g'

    base_data = self.page_class.GatherBaseData(mr, nonce)

    self.assertEqual(base_data['nonce'], nonce)
    self.assertEqual(base_data['projectname'], 'testproj')
    self.assertEqual(base_data['project'].cached_content_timestamp, 12345)
    self.assertEqual(base_data['project_alert'], None)

    self.assertTrue(base_data['currentPageURL'].endswith('/p/testproj/feeds'))
    self.assertTrue(
        base_data['currentPageURLEncoded'].endswith('%2Fp%2Ftestproj%2Ffeeds'))

  def testGatherHelpData_Normal(self):
    project = fake.Project(project_name='proj')
    _request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj', project=project)
    help_data = self.page_class.GatherHelpData(mr, {})
    self.assertEqual(None, help_data['cue'])
    self.assertEqual(None, help_data['account_cue'])
