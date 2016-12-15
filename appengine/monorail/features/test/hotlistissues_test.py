# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Unit tests for issuelist module."""

import mox
import unittest

from google.appengine.ext import testbed
from third_party import ezt

from features import hotlistissues
from framework import permissions
from framework import sorting
from services import service_manager
from testing import fake
from testing import testing_helpers


class HotlistIssuesUnitTest(unittest.TestCase):

  def setUp(self):
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_memcache_stub()
    self.testbed.init_datastore_v3_stub()
    self.services = service_manager.Services(
        issue_star=fake.IssueStarService(),
        config=fake.ConfigService(),
        user=fake.UserService(),
        issue=fake.IssueService(),
        project=fake.ProjectService(),
        features=fake.FeaturesService(),
        cache_manager=fake.CacheManager())
    self.servlet = hotlistissues.HotlistIssues(
        'req', 'res', services=self.services)
    self.user1 = self.services.user.TestAddUser('testuser', 111L)
    self.user2 = self.services.user.TestAddUser('testuser2', 222L)
    self.services.project.TestAddProject('ProjectName', project_id=001)
    self.issue1 = fake.MakeTestIssue(
        001, 1, 'issue_summary', 'New', 111L, project_name='ProjectName')
    self.services.issue.TestAddIssue(self.issue1)
    self.issue2 = fake.MakeTestIssue(
        001, 2, 'issue_summary2', 'New', 111L, project_name='ProjectName')
    self.services.issue.TestAddIssue(self.issue2)
    self.issue3 = fake.MakeTestIssue(
        001, 3, 'issue_summary3', 'New', 222L, project_name='ProjectName')
    self.services.issue.TestAddIssue(self.issue3)
    self.issues = [self.issue1, self.issue2, self.issue3]
    self.iid_rank_pairs = [
        (issue.issue_id, rank) for rank, issue in enumerate(self.issues)]
    self.test_hotlist = self.services.features.TestAddHotlist(
        'hotlist', hotlist_id=123, owner_ids=[222L], editor_ids=[111L],
        iid_rank_pairs=self.iid_rank_pairs)
    self.hotlistissues = self.test_hotlist.iid_rank_pairs
    self.mr = testing_helpers.MakeMonorailRequest(hotlist = self.test_hotlist)
    self.mr.hotlist_id = self.test_hotlist.hotlist_id
    self.mr.auth.user_id = 111
    sorting.InitializeArtValues(self.services)

  def testAssertBasePermissions(self):
    private_hotlist = self.services.features.TestAddHotlist(
        'privateHotlist', hotlist_id=321, owner_ids=[222L],
        iid_rank_pairs=self.iid_rank_pairs, is_private=True)
    # non-members cannot view private hotlists
    mr = testing_helpers.MakeMonorailRequest(
        hotlist=private_hotlist)
    mr.auth.effective_ids = {333L}
    mr.hotlist_id = private_hotlist.hotlist_id
    self.assertRaises(permissions.PermissionException,
                      self.servlet.AssertBasePermission, mr)

    # members can view private hotlists
    mr = testing_helpers.MakeMonorailRequest(
        hotlist=private_hotlist)
    mr.auth.effective_ids = {222L, 444L}
    mr.hotlist_id = private_hotlist.hotlist_id
    self.servlet.AssertBasePermission(mr)

    # non-members can view public hotlists
    mr = testing_helpers.MakeMonorailRequest(
        hotlist=self.test_hotlist)
    mr.auth.effective_ids = {333L, 444L}
    mr.hotlist_id = self.test_hotlist.hotlist_id
    self.servlet.AssertBasePermission(mr)

    # members can view public hotlists
    mr = testing_helpers.MakeMonorailRequest(
        hotlist=self.test_hotlist)
    mr.auth.effective_ids = {111L, 333L}
    mr.hotlist_id = self.test_hotlist.hotlist_id
    self.servlet.AssertBasePermission(mr)

  def testGatherPageData(self):
    self.mr.mode = 'list'
    self.mr.auth.effective_ids = {111L}
    self.mr.sort_spec = 'rank stars'
    page_data = self.servlet.GatherPageData(self.mr)
    self.assertEqual(ezt.boolean(False), page_data['owner_permissions'])
    self.assertEqual(ezt.boolean(True), page_data['editor_permissions'])
    self.assertEqual(ezt.boolean(False), page_data['grid_mode'])
    self.assertEqual(ezt.boolean(True), page_data['allow_rerank'])

    self.mr.sort_spec = 'stars ranks'
    page_data = self.servlet.GatherPageData(self.mr)
    self.assertEqual(ezt.boolean(False), page_data['allow_rerank'])

  def testGetTableViewData(self):
    # TODO(jojwang): Write this test
    pass

  def testGetGridViewData(self):
    # TODO(jojwang): Write this test
    pass

  def testProcessFormData_NoNewIssues(self):
    post_data = fake.PostData(remove=['false'], add_local_ids=[''])
    url = self.servlet.ProcessFormData(self.mr, post_data)
    self.assertTrue(url.endswith('u/111/hotlists/123'))
    self.assertEqual(self.test_hotlist.iid_rank_pairs, self.hotlistissues)

  def testProcessFormData_NormalEditIssues(self):
    issue4 = fake.MakeTestIssue(
        001, 4, 'issue_summary4', 'New', 222L, project_name='ProjectName')
    self.services.issue.TestAddIssue(issue4)
    issue5 = fake.MakeTestIssue(
        001, 5, 'issue_summary5', 'New', 222L, project_name='ProjectName')
    self.services.issue.TestAddIssue(issue5)

    post_data = fake.PostData(remove=['false'],
                              add_local_ids=['ProjectName:4, ProjectName:5'])
    url = self.servlet.ProcessFormData(self.mr, post_data)
    self.assertTrue('u/111/hotlists/123' in url)
    self.assertEqual(len(self.test_hotlist.iid_rank_pairs), 5)
    self.assertEqual(
        self.test_hotlist.iid_rank_pairs[3].issue_id, issue4.issue_id)
    self.assertEqual(
        self.test_hotlist.iid_rank_pairs[4].issue_id, issue5.issue_id)

    post_data = fake.PostData(remove=['true'], remove_local_ids=[
        'ProjectName:4, ProjectName:1, ProjectName:2'])
    url = self.servlet.ProcessFormData(self.mr, post_data)
    self.assertTrue('u/111/hotlists/123' in url)
    self.assertTrue(len(self.test_hotlist.iid_rank_pairs), 2)
    issue_ids = [issue.issue_id for issue in self.test_hotlist.iid_rank_pairs]
    self.assertTrue(issue5.issue_id in issue_ids)
    self.assertTrue(self.issue3.issue_id in issue_ids)
