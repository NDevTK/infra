# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Tests for the WorkEnv class."""

import logging
import unittest
from mock import Mock, patch

from google.appengine.api import memcache
from google.appengine.ext import testbed

from businesslogic import work_env
from framework import exceptions
from framework import permissions
from features import send_notifications
from proto import project_pb2
from proto import tracker_pb2
from services import service_manager
from services import template_svc
from testing import fake
from testing import testing_helpers
from tracker import tracker_bizobj


class WorkEnvTest(unittest.TestCase):

  def setUp(self):
    self.cnxn = 'fake connection'
    self.services = service_manager.Services(
        config=fake.ConfigService(),
        issue=fake.IssueService(),
        user=fake.UserService(),
        project=fake.ProjectService(),
        issue_star=fake.IssueStarService(),
        project_star=fake.ProjectStarService(),
        template=Mock(spec=template_svc.TemplateService),
        spam=fake.SpamService())
    self.project = self.services.project.TestAddProject(
        'proj', project_id=789)
    self.admin_user = self.services.user.TestAddUser(
        'admin@example.com', 444L)
    self.admin_user.is_site_admin = True
    self.mr = testing_helpers.MakeMonorailRequest(project=self.project)
    self.mr.perms = permissions.READ_ONLY_PERMISSIONSET

    self.work_env = work_env.WorkEnv(
      self.mr, self.services, 'Testing phase')

  def SignIn(self, user_id=111L):
    self.mr.auth.user_pb = self.services.user.GetUser(self.cnxn, user_id)
    self.mr.auth.user_id = user_id
    self.mr.auth.effective_ids = {user_id}
    self.mr.perms = permissions.GetPermissions(
        self.mr.auth.user_pb, self.mr.auth.effective_ids, self.project)

  # FUTURE: GetSiteReadOnlyState()
  # FUTURE: SetSiteReadOnlyState()
  # FUTURE: GetSiteBannerMessage()
  # FUTURE: SetSiteBannerMessage()

  def testCreateProject_Normal(self):
    """We can create a project."""
    self.SignIn(user_id=self.admin_user.user_id)
    with self.work_env as we:
      project_id = we.CreateProject(
          'newproj', [111L], [222L], [333L], 'summary', 'desc')
      actual = we.GetProject(project_id)

    self.assertEqual('summary', actual.summary)
    self.assertEqual('desc', actual.description)
    self.services.template.CreateDefaultProjectTemplates\
        .assert_called_once_with(self.mr.cnxn, project_id)

  def testCreateProject_AlreadyExists(self):
    """We can create a project."""
    self.SignIn(user_id=self.admin_user.user_id)
    # Project 'proj' is created in setUp().
    with self.assertRaises(exceptions.ProjectAlreadyExists):
      with self.work_env as we:
        we.CreateProject('proj', [111L], [222L], [333L], 'summary', 'desc')

    self.assertFalse(
        self.services.template.CreateDefaultProjectTemplates.called)

  def testListProjects(self):
    """We can get the project IDs of projects visible to the current user."""
    # Project 789 is created in setUp()
    self.services.project.TestAddProject(
        'proj2', project_id=2, access=project_pb2.ProjectAccess.MEMBERS_ONLY)
    self.services.project.TestAddProject('proj3', project_id=3)
    with self.work_env as we:
      actual = we.ListProjects()

    self.assertEqual([3, 789], actual)

  def testGetProject_Normal(self):
    """We can get an existing project by project_id."""
    with self.work_env as we:
      actual = we.GetProject(789)

    self.assertEqual(self.project, actual)

  def testGetProject_NoSuchProject(self):
    """We reject attempts to get a non-existent project."""
    with self.assertRaises(exceptions.NoSuchProjectException):
      with self.work_env as we:
        _actual = we.GetProject(999)

  def testGetProjectByName_Normal(self):
    """We can get an existing project by project_name."""
    with self.work_env as we:
      actual = we.GetProjectByName('proj')

    self.assertEqual(self.project, actual)

  def testGetProjectByName_NoSuchProject(self):
    """We reject attempts to get a non-existent project."""
    with self.assertRaises(exceptions.NoSuchProjectException):
      with self.work_env as we:
        _actual = we.GetProjectByName('huh-what')

  def testUpdateProject_Normal(self):
    """We can update an existing project."""
    self.SignIn(user_id=self.admin_user.user_id)
    with self.work_env as we:
      we.UpdateProject(789, read_only_reason='test reason')
      project = we.GetProject(789)

    self.assertEqual('test reason', project.read_only_reason)

  def testUpdateProject_NoSuchProject(self):
    """Updating a nonexistent project raises an exception."""
    self.SignIn(user_id=self.admin_user.user_id)
    with self.assertRaises(exceptions.NoSuchProjectException):
      with self.work_env as we:
        we.UpdateProject(999, summary='new summary')

  def testDeleteProject_Normal(self):
    """We can mark an existing project as deletable."""
    self.SignIn(user_id=self.admin_user.user_id)
    with self.work_env as we:
      we.DeleteProject(789)

    self.assertEqual(project_pb2.ProjectState.DELETABLE, self.project.state)

  def testDeleteProject_NoSuchProject(self):
    """Changing a nonexistent project raises an exception."""
    self.SignIn(user_id=self.admin_user.user_id)
    with self.assertRaises(exceptions.NoSuchProjectException):
      with self.work_env as we:
        we.DeleteProject(999)

  def testStarProject_Normal(self):
    """We can star and unstar a project."""
    self.SignIn()
    with self.work_env as we:
      self.assertFalse(we.IsProjectStarred(789))
      we.StarProject(789, True)
      self.assertTrue(we.IsProjectStarred(789))
      we.StarProject(789, False)
      self.assertFalse(we.IsProjectStarred(789))

  def testStarProject_NoSuchProject(self):
    """We can't star a nonexistent project."""
    self.SignIn()
    with self.assertRaises(exceptions.NoSuchProjectException):
      with self.work_env as we:
        we.StarProject(999, True)

  def testStarProject_Anon(self):
    """Anon user can't star a project."""
    with self.assertRaises(permissions.PermissionException):
      with self.work_env as we:
        we.StarProject(789, True)

  def testIsProjectStarred_Normal(self):
    """We can check if a project is starred."""
    # Tested by method testStarProject_Normal().
    pass

  def testIsProjectStarred_NoProjectSpecified(self):
    """A project ID must be specified."""
    with self.work_env as we:
      with self.assertRaises(exceptions.InputException):
        self.assertFalse(we.IsProjectStarred(None))

  def testIsProjectStarred_NoSuchProject(self):
    """We can't check for stars on a nonexistent project."""
    self.SignIn()
    with self.assertRaises(exceptions.NoSuchProjectException):
      with self.work_env as we:
        we.IsProjectStarred(999)

  def testListStarredProjects_ViewingSelf(self):
    """A user can view their own starred projects, if they still have access."""
    project1 = self.services.project.TestAddProject('proj1', project_id=1)
    project2 = self.services.project.TestAddProject('proj2', project_id=2)
    with self.work_env as we:
      self.SignIn()
      we.StarProject(project1.project_id, True)
      we.StarProject(project2.project_id, True)
      self.assertItemsEqual(
        [project1, project2], we.ListStarredProjects())
      project2.access = project_pb2.ProjectAccess.MEMBERS_ONLY
      self.assertItemsEqual(
        [project1], we.ListStarredProjects())

  def testListStarredProjects_ViewingOther(self):
    """A user can view their own starred projects, if they still have access."""
    project1 = self.services.project.TestAddProject('proj1', project_id=1)
    project2 = self.services.project.TestAddProject('proj2', project_id=2)
    with self.work_env as we:
      self.SignIn(user_id=222L)
      we.StarProject(project1.project_id, True)
      we.StarProject(project2.project_id, True)
      self.SignIn(user_id=111L)
      self.assertEqual([], we.ListStarredProjects())
      self.assertItemsEqual(
        [project1, project2], we.ListStarredProjects(viewed_user_id=222L))
      project2.access = project_pb2.ProjectAccess.MEMBERS_ONLY
      self.assertItemsEqual(
        [project1], we.ListStarredProjects(viewed_user_id=222L))

  def testGetProjectConfig_Normal(self):
    """We can get an existing config by project_id."""
    config = fake.MakeTestConfig(789, ['LabelOne'], ['New'])
    self.services.config.StoreConfig('cnxn', config)
    with self.work_env as we:
      actual = we.GetProjectConfig(789)

    self.assertEqual(config, actual)

  def testGetProjectConfig_NoSuchProject(self):
    """We reject attempts to get a non-existent config."""
    self.services.config.strict = True
    with self.assertRaises(exceptions.NoSuchProjectException):
      with self.work_env as we:
        _actual = we.GetProjectConfig(789)

  # FUTURE: labels, statuses, fields, components, rules, templates, and views.
  # FUTURE: project saved queries.
  # FUTURE: GetProjectPermissionsForUser()

  def testCreateIssue_Normal(self):
    """We can create an issue."""
    self.SignIn(user_id=111L)
    approval_values = [tracker_pb2.ApprovalValue(approval_id=23, phase_id=3)]
    phases = [tracker_pb2.Phase(name='Canary', phase_id=3)]
    with self.work_env as we:
      actual_issue, comment = we.CreateIssue(
          789, 'sum', 'New', 222L, [333L], ['Hot'], [], [], 'desc',
          phases=phases, approval_values=approval_values)
    self.assertEqual(789, actual_issue.project_id)
    self.assertEqual('sum', actual_issue.summary)
    self.assertEqual('New', actual_issue.status)
    self.assertEqual(111L, actual_issue.reporter_id)
    self.assertEqual(222L, actual_issue.owner_id)
    self.assertEqual([333L], actual_issue.cc_ids)
    self.assertEqual([], actual_issue.field_values)
    self.assertEqual([], actual_issue.component_ids)
    self.assertEqual(approval_values, actual_issue.approval_values)
    self.assertEqual(phases, actual_issue.phases)
    self.assertEqual('desc', comment.content)
    loaded_comments = self.services.issue.GetCommentsForIssue(
        self.cnxn, actual_issue.issue_id)
    self.assertEqual('desc', loaded_comments[0].content)

    # Verify that an indexing task was enqueued for this issue:
    self.assertTrue(self.services.issue.enqueue_issues_called)
    self.assertEqual(1, len(self.services.issue.enqueued_issues))
    self.assertEqual(actual_issue.issue_id,
        self.services.issue.enqueued_issues[0])

  def testListIssues_Normal(self):
    """We can do a query that generates some results."""
    pass  # TODO(jrobbins): add unit test

  def testListIssues_Error(self):
    """Errors are safely reported."""
    pass  # TODO(jrobbins): add unit test

  def testFindIssuePositionInSearch_Normal(self):
    """We can find an issue position for the flipper."""
    pass  # TODO(jrobbins): add unit test

  def testFindIssuePositionInSearch_Error(self):
    """Errors are safely reported."""
    pass  # TODO(jrobbins): add unit test

  def testGetIssue_Normal(self):
    """We can get an existing issue by issue_id."""
    issue = fake.MakeTestIssue(789, 1, 'sum', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue)
    with self.work_env as we:
      actual = we.GetIssue(78901)

    self.assertEqual(issue, actual)

  def testGetIssue_NoSuchIssue(self):
    """We reject attempts to get a non-existent issue."""
    with self.assertRaises(exceptions.NoSuchIssueException):
      with self.work_env as we:
        _actual = we.GetIssue(78901)

  def testGetIssueByLocalID_Normal(self):
    """We can get an existing issue by project_id and local_id."""
    issue = fake.MakeTestIssue(789, 1, 'sum', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue)
    with self.work_env as we:
      actual = we.GetIssueByLocalID(789, 1)

    self.assertEqual(issue, actual)

  def testGetIssueByLocalID_ProjectNotSpecified(self):
    """We reject calls with missing information."""
    with self.assertRaises(exceptions.InputException):
      with self.work_env as we:
        _actual = we.GetIssueByLocalID(None, 1)

  def testGetIssueByLocalID_IssueNotSpecified(self):
    """We reject calls with missing information."""
    with self.assertRaises(exceptions.InputException):
      with self.work_env as we:
        _actual = we.GetIssueByLocalID(789, None)

  def testGetIssueByLocalID_NoSuchIssue(self):
    """We reject attempts to get a non-existent issue."""
    with self.assertRaises(exceptions.NoSuchIssueException):
      with self.work_env as we:
        _actual = we.GetIssueByLocalID(789, 1)

  @patch('features.send_notifications.PrepareAndSendIssueChangeNotification')
  def testUpdateIssue_Normal(self, fake_pasicn):
    """We can update an issue."""
    self.SignIn()
    issue = fake.MakeTestIssue(789, 1, 'summary', 'Available', 111L)
    self.services.issue.TestAddIssue(issue)
    delta = tracker_pb2.IssueDelta(
        owner_id=222L, summary='New summary', cc_ids_add=[333L])

    with self.work_env as we:
      we.UpdateIssue(issue, delta, 'Getting started')

    self.assertEqual(222L, issue.owner_id)
    self.assertEqual('New summary', issue.summary)
    self.assertEqual([333L], issue.cc_ids)
    self.assertEqual([issue.issue_id], self.services.issue.enqueued_issues)
    comments = self.services.issue.GetCommentsForIssue('cnxn', issue.issue_id)
    comment_pb = comments[-1]
    fake_pasicn.assert_called_with(
        issue.issue_id, 'testing-app.appspot.com', 111L, send_email=True,
        old_owner_id=111L, comment_id=comment_pb.id)

  def testDeleteIssue(self):
    """We can mark and unmark an issue as deleted."""
    self.SignIn(user_id=self.admin_user.user_id)
    issue = fake.MakeTestIssue(789, 1, 'sum', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue)
    with self.work_env as we:
      _actual = we.DeleteIssue(issue, True)
    self.assertTrue(issue.deleted)
    with self.work_env as we:
      _actual = we.DeleteIssue(issue, False)
    self.assertFalse(issue.deleted)

  # FUTURE: GetIssuePermissionsForUser()

  # FUTURE: CreateComment()

  def testGetIssueComments_Normal(self):
    """We can get an existing issue by project_id and local_id."""
    issue = fake.MakeTestIssue(789, 1, 'sum', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue)
    comment = tracker_pb2.IssueComment(
        project_id=789, content='more info', user_id=111L,
        issue_id=issue.issue_id)
    self.services.issue.TestAddComment(comment, 1)

    with self.work_env as we:
      actual_comments = we.ListIssueComments(issue)

    self.assertEqual(2, len(actual_comments))
    self.assertEqual('sum', actual_comments[0].content)
    self.assertEqual('more info', actual_comments[1].content)

  # FUTURE: UpdateComment()

  def testDeleteComment_Normal(self):
    """We can mark and unmark a comment as deleted."""
    self.SignIn(user_id=111L)
    issue = fake.MakeTestIssue(789, 1, 'sum', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue)
    comment = tracker_pb2.IssueComment(
        project_id=789, content='soon to be deleted', user_id=111L,
        issue_id=issue.issue_id)
    self.services.issue.TestAddComment(comment, 1)
    with self.work_env as we:
      we.DeleteComment(issue, comment, True)
      self.assertEqual(111L, comment.deleted_by)
      we.DeleteComment(issue, comment, False)
      self.assertEqual(None, comment.deleted_by)

  def testStarIssue_Normal(self):
    """We can star and unstar issues."""
    issue = fake.MakeTestIssue(789, 1, 'sum', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue)
    self.SignIn(user_id=111L)

    with self.work_env as we:
      we.StarIssue(issue, True)
      self.assertEqual(1, issue.star_count)
      we.StarIssue(issue, False)
      self.assertEqual(0, issue.star_count)

  def testStarIssue_Anon(self):
    """A signed out user cannot star or unstar issues."""
    issue = fake.MakeTestIssue(789, 1, 'sum', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue)
    # Don't sign in.

    with self.assertRaises(exceptions.InputException):
      with self.work_env as we:
        we.StarIssue(issue, True)

  def testIsIssueStarred_Normal(self):
    """We can check if the current user starred an issue or not."""
    issue = fake.MakeTestIssue(789, 1, 'sum', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue)
    self.SignIn(user_id=111L)

    with self.work_env as we:
      self.assertFalse(we.IsIssueStarred(issue))
      we.StarIssue(issue, True)
      self.assertTrue(we.IsIssueStarred(issue))
      we.StarIssue(issue, False)
      self.assertFalse(we.IsIssueStarred(issue))

  def testIsIssueStarred_Anon(self):
    """A signed out user has never starred anything."""
    issue = fake.MakeTestIssue(789, 1, 'sum', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue)
    # Don't sign in.

    with self.work_env as we:
      self.assertFalse(we.IsIssueStarred(issue))

  def testListStarredIssueIDs_Anon(self):
    """A signed out users has no starred issues."""
    # Don't sign in.
    with self.work_env as we:
      self.assertEqual([], we.ListStarredIssueIDs())

  def testListStarredIssueIDs_Normal(self):
    """We can get the list of issues starred by a user."""
    issue1 = fake.MakeTestIssue(789, 1, 'sum1', 'New', 111L, issue_id=78901)
    self.services.issue.TestAddIssue(issue1)
    issue2 = fake.MakeTestIssue(789, 2, 'sum2', 'New', 111L, issue_id=78902)
    self.services.issue.TestAddIssue(issue2)

    self.SignIn(user_id=111L)
    with self.work_env as we:
      # User has not starred anything yet.
      self.assertEqual([], we.ListStarredIssueIDs())

      # Now, star a couple of issues.
      we.StarIssue(issue1, True)
      we.StarIssue(issue2, True)
      self.assertItemsEqual(
          [issue1.issue_id, issue2.issue_id],
          we.ListStarredIssueIDs())

    # Check that there is no cross-talk between users.
    self.SignIn(user_id=222L)
    with self.work_env as we:
      # User has not starred anything yet.
      self.assertEqual([], we.ListStarredIssueIDs())

      # Now, star an issue as that other user.
      we.StarIssue(issue1, True)
      self.assertEqual([issue1.issue_id], we.ListStarredIssueIDs())


  # FUTURE: GetUser()
  # FUTURE: UpdateUser()
  # FUTURE: DeleteUser()
  # FUTURE: StarUser()
  # FUTURE: IsUserStarred()
  # FUTURE: ListStarredUsers()

  # FUTURE: CreateGroup()
  # FUTURE: ListGroups()
  # FUTURE: UpdateGroup()
  # FUTURE: DeleteGroup()

  # FUTURE: CreateHotlist()
  # FUTURE: ListHotlistsByUser()
  # FUTURE: UpdateHotlist()
  # FUTURE: DeleteHotlist()
