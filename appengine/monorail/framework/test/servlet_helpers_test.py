# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Unit tests for servlet base class helper functions."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import unittest
import settings

from google.appengine.ext import testbed
from six.moves import urllib

from framework import permissions
from framework import servlet_helpers
from mrproto import project_pb2
from mrproto import tracker_pb2
from testing import testing_helpers


class EztDataTest(unittest.TestCase):

  def testGetBannerTime(self):
    """Tests GetBannerTime method."""
    timestamp = [2019, 6, 13, 18, 30]

    banner_time = servlet_helpers.GetBannerTime(timestamp)
    self.assertEqual(1560450600, banner_time)


class AssertBasePermissionTest(unittest.TestCase):

  def testAccessGranted(self):
    _, mr = testing_helpers.GetRequestObjects(path='/hosting')
    # No exceptions should be raised.
    servlet_helpers.AssertBasePermission(mr)

    mr.auth.user_id = 123
    # No exceptions should be raised.
    servlet_helpers.AssertBasePermission(mr)
    servlet_helpers.AssertBasePermissionForUser(
        mr.auth.user_pb, mr.auth.user_view)

  def testBanned(self):
    _, mr = testing_helpers.GetRequestObjects(path='/hosting')
    mr.auth.user_pb.banned = 'spammer'
    self.assertRaises(
        permissions.BannedUserException,
        servlet_helpers.AssertBasePermissionForUser,
        mr.auth.user_pb, mr.auth.user_view)
    self.assertRaises(
        permissions.BannedUserException,
        servlet_helpers.AssertBasePermission, mr)

  def testPlusAddressAccount(self):
    _, mr = testing_helpers.GetRequestObjects(path='/hosting')
    mr.auth.user_pb.email = 'mailinglist+spammer@chromium.org'
    self.assertRaises(
        permissions.BannedUserException,
        servlet_helpers.AssertBasePermissionForUser,
        mr.auth.user_pb, mr.auth.user_view)
    self.assertRaises(
        permissions.BannedUserException,
        servlet_helpers.AssertBasePermission, mr)

  def testNoAccessToProject(self):
    project = project_pb2.Project()
    project.project_name = 'proj'
    project.access = project_pb2.ProjectAccess.MEMBERS_ONLY
    _, mr = testing_helpers.GetRequestObjects(path='/p/proj/', project=project)
    mr.perms = permissions.EMPTY_PERMISSIONSET
    self.assertRaises(
        permissions.PermissionException,
        servlet_helpers.AssertBasePermission, mr)

  def testPermForProject(self):
    project = project_pb2.Project()
    project.project_name = 'proj'
    project.access = project_pb2.ProjectAccess.MEMBERS_ONLY
    _, mr = testing_helpers.GetRequestObjects(path='/p/proj/', project=project)
    mr.auth.user_pb.email = settings.borg_service_account
    project_perm = servlet_helpers.CheckPermForProject(
        mr, permissions.CREATE_GROUP, project)
    self.assertTrue(project_perm)


FORM_URL = 'http://example.com/issues/form.php'


class ComputeIssueEntryURLTest(unittest.TestCase):

  def setUp(self):
    self.project = project_pb2.Project()
    self.project.project_name = 'proj'
    self.config = tracker_pb2.ProjectIssueConfig()
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_user_stub()
    self.testbed.init_memcache_stub()
    self.testbed.init_datastore_v3_stub()

  def tearDown(self):
    self.testbed.deactivate()

  def testComputeIssueEntryURL_Normal(self):
    _request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/issues/detail?id=123&q=term',
        project=self.project)

    url = servlet_helpers.ComputeIssueEntryURL(mr)
    self.assertEqual('/p/proj/issues/entry', url)

  def testComputeIssueEntryURL_Chromium(self):
    _request, mr = testing_helpers.GetRequestObjects(
        path='/p/chromium/issues/detail?id=123&q=term', project=self.project)

    url = servlet_helpers.ComputeIssueEntryURL(mr)
    self.assertIn('/issues/wizard', url)

class IssueListURLTest(unittest.TestCase):

  def setUp(self):
    self.project = project_pb2.Project()
    self.project.project_name = 'proj'
    self.project.owner_ids = [111]
    self.config = tracker_pb2.ProjectIssueConfig()
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_user_stub()
    self.testbed.init_memcache_stub()
    self.testbed.init_datastore_v3_stub()

  def tearDown(self):
    self.testbed.deactivate()

  def testIssueListURL_NotCustomized(self):
    _request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/issues', project=self.project)

    url = servlet_helpers.IssueListURL(mr, self.config)
    self.assertEqual('/p/proj/issues/list', url)

  def testIssueListURL_Customized_Nonmember(self):
    _request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/issues', project=self.project)
    self.config.member_default_query = 'owner:me'

    url = servlet_helpers.IssueListURL(mr, self.config)
    self.assertEqual('/p/proj/issues/list', url)

  def testIssueListURL_Customized_Member(self):
    _request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/issues', project=self.project,
        user_info={'effective_ids': {111}})
    self.config.member_default_query = 'owner:me'

    url = servlet_helpers.IssueListURL(mr, self.config)
    self.assertEqual('/p/proj/issues/list?q=owner%3Ame', url)

  def testIssueListURL_Customized_RetainQS(self):
    _request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/issues', project=self.project,
        user_info={'effective_ids': {111}})
    self.config.member_default_query = 'owner:me'

    url = servlet_helpers.IssueListURL(mr, self.config, query_string='')
    self.assertEqual('/p/proj/issues/list?q=owner%3Ame', url)

    url = servlet_helpers.IssueListURL(mr, self.config, query_string='q=Pri=1')
    self.assertEqual('/p/proj/issues/list?q=Pri=1', url)

class ProjectIsRestrictedTest(unittest.TestCase):

  def setUp(self):
    self.project = project_pb2.Project()
    self.project.project_name = 'proj'
    self.config = tracker_pb2.ProjectIssueConfig()
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_user_stub()
    self.testbed.init_memcache_stub()
    self.testbed.init_datastore_v3_stub()

  def tearDown(self):
    self.testbed.deactivate()

  def testProjectIsRestricted(self):
    self.project.access = project_pb2.ProjectAccess.MEMBERS_ONLY
    _request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/issues/detail?id=123&q=term', project=self.project)
    isRestrict = servlet_helpers.ProjectIsRestricted(mr)
    self.assertTrue(isRestrict)

  def testProjectIsNotRestricted(self):
    self.project.access = project_pb2.ProjectAccess.ANYONE
    _request, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/issues/detail?id=123&q=term', project=self.project)
    isRestrict = servlet_helpers.ProjectIsRestricted(mr)
    self.assertFalse(isRestrict)


class ComputerCreateUrl(unittest.TestCase):

  def setUp(self):
    self.project = project_pb2.Project()
    self.project.project_name = 'proj'
    self.config = tracker_pb2.ProjectIssueConfig()
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_user_stub()
    self.testbed.init_memcache_stub()
    self.testbed.init_datastore_v3_stub()

  def tearDown(self):
    self.testbed.deactivate()

  def testCreateEscapedLoginUrlFromMR(self):
    _, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/issues/detail?id=123&q=term', project=self.project)
    mr.current_page_url_encoded = (
        'https%3A%2F%2Fbugs.chromium.org'
        '%2Fp%2Fchromium%2Fissues%2Fentry%3F'
        'template%3DBuild%2520Infrastructure%26'
        'labels%3DRestrict-View-Google%2CInfra-Troopers')
    url = servlet_helpers.SafeCreateLoginURL(mr)
    double_encoded_query = (
        'https%253A%252F%252Fbugs.chromium.org'
        '%252Fp%252Fchromium%252Fissues%252Fentry%253F'
        'template%253DBuild%252520Infrastructure%2526'
        'labels%253DRestrict-View-Google%252CInfra-Troopers')
    self.assertIn(double_encoded_query, url)

  def testCreateLogoutUrl(self):
    _, mr = testing_helpers.GetRequestObjects(
        path='/p/proj/issues/detail?id=123&q=term', project=self.project)
    url = servlet_helpers.SafeCreateLogoutURL(mr)
    self.assertIn('/Logout', url)
