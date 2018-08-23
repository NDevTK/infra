# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Tests for the issues servicer."""

import logging
import unittest
from mock import ANY, Mock, patch

from google.protobuf import empty_pb2

from components.prpc import codes
from components.prpc import context
from components.prpc import server

from api import issues_servicer
from api.api_proto import common_pb2
from api.api_proto import issues_pb2
from api.api_proto import issue_objects_pb2
from api.api_proto import common_pb2
from businesslogic import work_env
from features import filterrules_helpers
from features import send_notifications
from framework import authdata
from framework import exceptions
from framework import monorailcontext
from framework import permissions
from proto import tracker_pb2
from testing import fake
from services import service_manager
from proto import tracker_pb2


class IssuesServicerTest(unittest.TestCase):

  NOW = 1234567890

  def setUp(self):
    self.cnxn = fake.MonorailConnection()
    self.services = service_manager.Services(
        config=fake.ConfigService(),
        issue=fake.IssueService(),
        issue_star=fake.IssueStarService(),
        user=fake.UserService(),
        usergroup=fake.UserGroupService(),
        project=fake.ProjectService(),
        features=fake.FeaturesService())
    self.project = self.services.project.TestAddProject(
        'proj', project_id=789, owner_ids=[111L], contrib_ids=[222L, 333L])
    self.user_1 = self.services.user.TestAddUser('owner@example.com', 111L)
    self.user_2 = self.services.user.TestAddUser('approver2@example.com', 222L)
    self.user_3 = self.services.user.TestAddUser('approver3@example.com', 333L)
    self.issue_1 = fake.MakeTestIssue(
        789, 1, 'sum', 'New', 111L, project_name='proj',
        opened_timestamp=self.NOW)
    self.issue_2 = fake.MakeTestIssue(
        789, 2, 'sum', 'New', 111L, project_name='proj')
    self.issue_1.blocked_on_iids.append(self.issue_2.issue_id)
    self.services.issue.TestAddIssue(self.issue_1)
    self.services.issue.TestAddIssue(self.issue_2)
    self.issues_svcr = issues_servicer.IssuesServicer(
        self.services, make_rate_limiter=False)
    self.prpc_context = context.ServicerContext()
    self.prpc_context.set_code(server.StatusCode.OK)
    self.auth = authdata.AuthData(user_id=333L, email='approver3@example.com')

    self.fd_1 = tracker_pb2.FieldDef(
        field_name='FirstField', field_id=1,
        field_type=tracker_pb2.FieldTypes.STR_TYPE,
        applicable_type='')
    self.fd_2 = tracker_pb2.FieldDef(
        field_name='SecField', field_id=2,
        field_type=tracker_pb2.FieldTypes.INT_TYPE,
        applicable_type='')
    self.fd_3 = tracker_pb2.FieldDef(
        field_name='LegalApproval', field_id=3,
        field_type=tracker_pb2.FieldTypes.APPROVAL_TYPE,
        applicable_type='')
    self.fd_4 = tracker_pb2.FieldDef(
        field_name='UserField', field_id=4,
        field_type=tracker_pb2.FieldTypes.USER_TYPE,
        applicable_type='')

  def CallWrapped(self, wrapped_handler, *args, **kwargs):
    return wrapped_handler.wrapped(self.issues_svcr, *args, **kwargs)

  def testCreateIssue_Normal(self):
    """We can create an issue."""
    request = issues_pb2.CreateIssueRequest(
        project_name='proj',
        issue=issue_objects_pb2.Issue(
            project_name='proj', local_id=1, summary='sum'))
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')

    response = self.CallWrapped(self.issues_svcr.CreateIssue, mc, request)

    self.assertEqual('proj', response.project_name)

  def testGetIssue_Normal(self):
    """We can get an issue."""
    request = issues_pb2.GetIssueRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    response = self.CallWrapped(self.issues_svcr.GetIssue, mc, request)

    actual = response.issue
    self.assertEqual('proj', actual.project_name)
    self.assertEqual(1, actual.local_id)
    self.assertEqual(1, len(actual.blocked_on_issue_refs))
    self.assertEqual('proj', actual.blocked_on_issue_refs[0].project_name)
    self.assertEqual(2, actual.blocked_on_issue_refs[0].local_id)

  def testListReferencedIssues(self):
    """We can get the referenced issues that exist."""
    self.services.project.TestAddProject(
        'other-proj', project_id=788, owner_ids=[111L])
    other_issue = fake.MakeTestIssue(
        788, 1, 'sum', 'Fixed', 111L, project_name='other-proj', issue_id=78801)
    self.services.issue.TestAddIssue(other_issue)
    # We ignore project_names or local_ids that don't exist in our DB.
    request = issues_pb2.ListReferencedIssuesRequest(
        issue_refs=[
            common_pb2.IssueRef(project_name='proj', local_id=1),
            common_pb2.IssueRef(project_name='other-proj', local_id=1),
            common_pb2.IssueRef(project_name='other-proj', local_id=2),
            common_pb2.IssueRef(project_name='ghost-proj', local_id=1)
            ]
        )
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    response = self.CallWrapped(
        self.issues_svcr.ListReferencedIssues, mc, request)
    self.assertEqual(len(response.closed_refs), 1)
    self.assertEqual(len(response.open_refs), 1)
    self.assertEqual(response.closed_refs[0],
                     common_pb2.IssueRef(project_name='other-proj', local_id=1))
    self.assertEqual(response.open_refs[0],
                     common_pb2.IssueRef(project_name='proj', local_id=1))

  def testListReferencedIssues_MissingInput(self):
    request = issues_pb2.ListReferencedIssuesRequest(
        issue_refs=[common_pb2.IssueRef(local_id=1)])
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    with self.assertRaises(exceptions.InputException):
      self.CallWrapped(self.issues_svcr.ListReferencedIssues, mc, request)

  def testUpdateIssue_Denied(self):
    """We reject requests to update an issue when the user lacks perms."""
    request = issues_pb2.UpdateIssueRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1

    # Anon user can never update.
    mc = monorailcontext.MonorailContext(self.services, cnxn=self.cnxn)
    mc.LookupLoggedInUserPerms(self.project)
    with self.assertRaises(permissions.PermissionException):
      self.CallWrapped(self.issues_svcr.UpdateIssue, mc, request)

    # Signed in user cannot view this issue.
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='approver3@example.com')
    mc.LookupLoggedInUserPerms(self.project)
    self.issue_1.labels = ['Restrict-View-CoreTeam']
    with self.assertRaises(permissions.PermissionException):
      self.CallWrapped(self.issues_svcr.UpdateIssue, mc, request)

    # Signed in user cannot edit this issue.
    self.issue_1.labels = ['Restrict-EditIssue-CoreTeam']
    with self.assertRaises(permissions.PermissionException):
      self.CallWrapped(self.issues_svcr.UpdateIssue, mc, request)

  @patch('features.send_notifications.PrepareAndSendIssueChangeNotification')
  def testUpdateIssue_Normal(self, fake_pasicn):
    """We can update an issue."""
    request = issues_pb2.UpdateIssueRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    request.delta.summary.value = 'New summary'
    request.delta.label_refs_add.extend([
        common_pb2.LabelRef(label='Hot')])
    request.comment_content = 'test comment'
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    response = self.CallWrapped(self.issues_svcr.UpdateIssue, mc, request)

    actual = response.issue
    # Intended stuff was changed.
    self.assertEqual(1, len(actual.label_refs))
    self.assertEqual('Hot', actual.label_refs[0].label)
    self.assertEqual('New summary', actual.summary)

    # Other stuff didn't change.
    self.assertEqual('proj', actual.project_name)
    self.assertEqual(1, actual.local_id)
    self.assertEqual(1, len(actual.blocked_on_issue_refs))
    self.assertEqual('proj', actual.blocked_on_issue_refs[0].project_name)
    self.assertEqual(2, actual.blocked_on_issue_refs[0].local_id)

    # A comment was added.
    fake_pasicn.assert_called_once()
    comments = self.services.issue.GetCommentsForIssue(
        self.cnxn, self.issue_1.issue_id)
    self.assertEqual(2, len(comments))
    self.assertEqual('test comment', comments[1].content)

  @patch('features.send_notifications.PrepareAndSendIssueChangeNotification')
  def testUpdateIssue_CommentOnly(self, fake_pasicn):
    """We can update an issue with a comment w/o making any other changes."""
    request = issues_pb2.UpdateIssueRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    request.comment_content = 'test comment'
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    self.CallWrapped(self.issues_svcr.UpdateIssue, mc, request)

    # A comment was added.
    fake_pasicn.assert_called_once()
    comments = self.services.issue.GetCommentsForIssue(
        self.cnxn, self.issue_1.issue_id)
    self.assertEqual(2, len(comments))
    self.assertEqual('test comment', comments[1].content)
    self.assertFalse(comments[1].is_description)

  @patch('features.send_notifications.PrepareAndSendIssueChangeNotification')
  def testUpdateIssue_CommentWithAttachments(self, fake_pasicn):
    """We can update an issue with a comment and attachments."""
    request = issues_pb2.UpdateIssueRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    request.comment_content = 'test comment'
    request.uploads.extend([
          issues_pb2.AttachmentUpload(filename='a.txt', content='aaaaa')])
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    self.CallWrapped(self.issues_svcr.UpdateIssue, mc, request)

    # A comment with an attachment was added.
    fake_pasicn.assert_called_once()
    comments = self.services.issue.GetCommentsForIssue(
        self.cnxn, self.issue_1.issue_id)
    self.assertEqual(2, len(comments))
    self.assertEqual('test comment', comments[1].content)
    self.assertFalse(comments[1].is_description)
    self.assertEqual(1, len(comments[1].attachments))
    self.assertEqual('a.txt', comments[1].attachments[0].filename)
    self.assertEqual(5, self.project.attachment_bytes_used)

  @patch('features.send_notifications.PrepareAndSendIssueChangeNotification')
  def testUpdateIssue_Description(self, fake_pasicn):
    """We can update an issue's description."""
    request = issues_pb2.UpdateIssueRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    request.comment_content = 'new description'
    request.is_description = True
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    self.CallWrapped(self.issues_svcr.UpdateIssue, mc, request)

    # A comment was added.
    fake_pasicn.assert_called_once()
    comments = self.services.issue.GetCommentsForIssue(
        self.cnxn, self.issue_1.issue_id)
    self.assertEqual(2, len(comments))
    self.assertEqual('new description', comments[1].content)
    self.assertTrue(comments[1].is_description)

  @patch('features.send_notifications.PrepareAndSendIssueChangeNotification')
  def testUpdateIssue_NoOp(self, fake_pasicn):
    """We gracefully ignore requests that have no delta or comment."""
    request = issues_pb2.UpdateIssueRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    response = self.CallWrapped(self.issues_svcr.UpdateIssue, mc, request)

    actual = response.issue
    # Other stuff didn't change.
    self.assertEqual('proj', actual.project_name)
    self.assertEqual(1, actual.local_id)
    self.assertEqual('sum', actual.summary)
    self.assertEqual('New', actual.status_ref.status)

    # No comment was added.
    fake_pasicn.assert_not_called()
    comments = self.services.issue.GetCommentsForIssue(
        self.cnxn, self.issue_1.issue_id)
    self.assertEqual(1, len(comments))

  def testStarIssue_Denied(self):
    """We reject requests to star an issue if the user lacks perms."""
    request = issues_pb2.StarIssueRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    request.starred = True

    # Anon user cannot star an issue.
    mc = monorailcontext.MonorailContext(self.services, cnxn=self.cnxn)
    mc.LookupLoggedInUserPerms(self.project)
    with self.assertRaises(permissions.PermissionException):
      self.CallWrapped(self.issues_svcr.StarIssue, mc, request)

    # User star an issue that they cannot view.
    self.issue_1.labels = ['Restrict-View-CoreTeam']
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='approver3@example.com')
    mc.LookupLoggedInUserPerms(self.project)
    with self.assertRaises(permissions.PermissionException):
      self.CallWrapped(self.issues_svcr.StarIssue, mc, request)

    # The issue was not actually starred.
    self.assertEqual(0, self.issue_1.star_count)

  def testStarIssue_Normal(self):
    """Users can star and unstar issues."""
    request = issues_pb2.StarIssueRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    request.starred = True
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='approver3@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    # First, star it.
    response = self.CallWrapped(self.issues_svcr.StarIssue, mc, request)
    self.assertEqual(1, response.star_count)

    # Then, unstar it.
    request.starred = False
    response = self.CallWrapped(self.issues_svcr.StarIssue, mc, request)
    self.assertEqual(0, response.star_count)

  def testIsIssueStared_Anon(self):
    """Anon users can't star issues, so they always get back False."""
    request = issues_pb2.IsIssueStarredRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    mc = monorailcontext.MonorailContext(self.services, cnxn=self.cnxn)
    mc.LookupLoggedInUserPerms(self.project)

    response = self.CallWrapped(self.issues_svcr.IsIssueStarred, mc, request)
    self.assertFalse(response.is_starred)

  def testIsIssueStared_Denied(self):
    """Users can't ask about an issue that they cannot currently view."""
    request = issues_pb2.IsIssueStarredRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='approver3@example.com')
    mc.LookupLoggedInUserPerms(self.project)
    self.issue_1.labels = ['Restrict-View-CoreTeam']

    with self.assertRaises(permissions.PermissionException):
      self.CallWrapped(self.issues_svcr.IsIssueStarred, mc, request)

  def testIsIssueStared_Normal(self):
    """Users can star and unstar issues."""
    request = issues_pb2.IsIssueStarredRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='approver3@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    # It is not initially starred by this user.
    response = self.CallWrapped(self.issues_svcr.IsIssueStarred, mc, request)
    self.assertFalse(response.is_starred)

    # If we star it, we get response True.
    self.services.issue_star.SetStar(
        self.cnxn, self.services, 'fake config', self.issue_1.issue_id,
        333L, True)
    response = self.CallWrapped(self.issues_svcr.IsIssueStarred, mc, request)
    self.assertTrue(response.is_starred)

  def testListComments_Normal(self):
    """We can get comments on an issue."""
    comment = tracker_pb2.IssueComment(
        user_id=111L, timestamp=self.NOW, content='second',
        project_id=789, issue_id=self.issue_1.issue_id, sequence=1)
    self.services.issue.TestAddComment(comment, self.issue_1.local_id)
    request = issues_pb2.ListCommentsRequest()
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    response = self.CallWrapped(self.issues_svcr.ListComments, mc, request)

    actual_0 = response.comments[0]
    actual_1 = response.comments[1]
    expected_0 = issue_objects_pb2.Comment(
        project_name='proj', local_id=1, sequence_num=0, is_deleted=False,
        commenter=common_pb2.UserRef(
            user_id=111L, display_name='owner@example.com'),
        timestamp=self.NOW, content='sum', is_spam=False,
        description_num=1)
    expected_1 = issue_objects_pb2.Comment(
        project_name='proj', local_id=1, sequence_num=1, is_deleted=False,
        commenter=common_pb2.UserRef(
            user_id=111L, display_name='owner@example.com'),
        timestamp=self.NOW, content='second')
    self.assertEqual(expected_0, actual_0)
    self.assertEqual(expected_1, actual_1)

  def testListActivities_Normal(self):
    """We can get issue activity."""
    comment = tracker_pb2.IssueComment(
        user_id=111L, timestamp=self.NOW, content='sum',
        project_id=789, issue_id=self.issue_1.issue_id, sequence=1)
    self.services.issue.TestAddComment(comment, self.issue_1.local_id)
    self.services.issue.TestAddIssue(self.issue_1)
    request = issues_pb2.ListActivitiesRequest()
    request.user_ref.user_id = 111L
    config = tracker_pb2.ProjectIssueConfig(project_id=789,
        field_defs=[self.fd_1])
    self.services.config.StoreConfig(self.cnxn, config)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)

    response = self.CallWrapped(self.issues_svcr.ListActivities, mc, request)
    actual_0 = response.comments[0]
    actual_1 = response.issue_summaries[0]
    expected_0 = issue_objects_pb2.Comment(
        project_name='proj', local_id=1, sequence_num=0, is_deleted=False,
        commenter=common_pb2.UserRef(
            user_id=111L, display_name='owner@example.com'),
        timestamp=self.NOW, content='sum', is_spam=False,
        description_num=1)
    expected_1 = issue_objects_pb2.IssueSummary(project_name='proj', local_id=1,
        summary='sum')
    self.assertEqual(expected_0, actual_0)
    self.assertEqual(expected_1, actual_1)

  @patch('testing.fake.IssueService.SoftDeleteComment')
  def testDeleteComment_Invalid(self, fake_softdeletecomment):
    """We reject requests to delete a non-existent comment."""
    # Note: no comments added to self.issue_1 after the description.
    request = issues_pb2.DeleteCommentRequest(
        issue_ref=common_pb2.IssueRef(project_name='proj', local_id=1),
        sequence_num=2, delete=True)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')

    with self.assertRaises(exceptions.NoSuchCommentException):
      self.CallWrapped(self.issues_svcr.DeleteComment, mc, request)

    fake_softdeletecomment.assert_not_called()

  def testDeleteComment_Normal(self):
    """An authorized user can delete and undelete a comment."""
    comment_1 = tracker_pb2.IssueComment(
        project_id=789, issue_id=self.issue_1.issue_id, content='one')
    self.services.issue.TestAddComment(comment_1, 1)
    comment_2 = tracker_pb2.IssueComment(
        project_id=789, issue_id=self.issue_1.issue_id, content='two')
    self.services.issue.TestAddComment(comment_2, 1)

    # Delete a comment.
    request = issues_pb2.DeleteCommentRequest(
        issue_ref=common_pb2.IssueRef(project_name='proj', local_id=1),
        sequence_num=2, delete=True)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')

    response = self.CallWrapped(self.issues_svcr.DeleteComment, mc, request)

    self.assertTrue(isinstance(response, empty_pb2.Empty))
    self.assertEqual(111L, comment_2.deleted_by)

    # Undelete a comment.
    request.delete=False

    response = self.CallWrapped(self.issues_svcr.DeleteComment, mc, request)

    self.assertTrue(isinstance(response, empty_pb2.Empty))
    self.assertEqual(None, comment_2.deleted_by)

  @patch('testing.fake.IssueService.SoftDeleteComment')
  def testDeleteComment_Denied(self, fake_softdeletecomment):
    """An unauthorized user cannot delete a comment."""
    comment_1 = tracker_pb2.IssueComment(
        project_id=789, issue_id=self.issue_1.issue_id, content='one')
    self.services.issue.TestAddComment(comment_1, 1)

    request = issues_pb2.DeleteCommentRequest(
        issue_ref=common_pb2.IssueRef(project_name='proj', local_id=1),
        sequence_num=1, delete=True)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='approver3@example.com')

    with self.assertRaises(permissions.PermissionException):
      self.CallWrapped(self.issues_svcr.DeleteComment, mc, request)

    fake_softdeletecomment.assert_not_called()
    self.assertIsNone(comment_1.deleted_by)

  def testUpdateApproval_MissingFieldDef(self):
    """Missing Approval Field Def throwns exception."""
    issue_ref = common_pb2.IssueRef(project_name='proj', local_id=1)
    field_ref = common_pb2.FieldRef(field_name='LegalApproval')
    approval_delta = issues_pb2.ApprovalDelta(
        status=issue_objects_pb2.REVIEW_REQUESTED)
    request = issues_pb2.UpdateApprovalRequest(
        issue_ref=issue_ref, field_ref=field_ref, approval_delta=approval_delta)

    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='approver3@example.com',
        auth=self.auth)

    with self.assertRaises(exceptions.NoSuchFieldDefException):
      self.CallWrapped(self.issues_svcr.UpdateApproval, mc, request)

  @patch('businesslogic.work_env.WorkEnv.UpdateIssueApproval')
  @patch('features.send_notifications.PrepareAndSendApprovalChangeNotification')
  def testUpdateApproval(self, _mockPrepareAndSend, mockUpdateIssueApproval):
    """We can update an approval."""

    av_3 = tracker_pb2.ApprovalValue(
            approval_id=3,
            status=tracker_pb2.ApprovalStatus.NEEDS_REVIEW,
            approver_ids=[333L]
    )
    self.issue_1.approval_values = [av_3]

    config = self.services.config.GetProjectConfig(
        self.cnxn, 789)
    config.field_defs = [self.fd_1, self.fd_3]

    self.services.config.StoreConfig(self.cnxn, config)

    issue_ref = common_pb2.IssueRef(project_name='proj', local_id=1)
    field_ref = common_pb2.FieldRef(field_name='LegalApproval')
    approval_delta = issues_pb2.ApprovalDelta(
        status=issue_objects_pb2.REVIEW_REQUESTED,
        approver_refs_add=[
          common_pb2.UserRef(user_id=222L, display_name='approver2@example.com')
          ],
        field_vals_add=[
          issue_objects_pb2.FieldValue(
              field_ref=common_pb2.FieldRef(field_name='FirstField'),
              value='string')
          ]
    )

    request = issues_pb2.UpdateApprovalRequest(
        issue_ref=issue_ref, field_ref=field_ref, approval_delta=approval_delta,
        comment_content='Well, actually'
    )
    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    request.uploads.extend([
          issues_pb2.AttachmentUpload(filename='a.txt', content='aaaaa')])

    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='approver3@example.com',
        auth=self.auth)

    mockUpdateIssueApproval.return_value = [
        tracker_pb2.ApprovalValue(
            approval_id=3,
            status=tracker_pb2.ApprovalStatus.REVIEW_REQUESTED,
            setter_id=333L,
            approver_ids=[333L, 222L]),
        'comment_pb']

    actual = self.CallWrapped(self.issues_svcr.UpdateApproval, mc, request)

    expected = issues_pb2.UpdateApprovalResponse()
    expected.approval.CopyFrom(
      issue_objects_pb2.Approval(
          field_ref=common_pb2.FieldRef(
              field_id=3,
              field_name='LegalApproval',
              type=common_pb2.APPROVAL_TYPE),
          approver_refs=[
              common_pb2.UserRef(
                  user_id=333, display_name='approver3@example.com'),
              common_pb2.UserRef(
                  user_id=222, display_name='approver2@example.com')
              ],
          status=issue_objects_pb2.REVIEW_REQUESTED,
          setter_ref=common_pb2.UserRef(
                  user_id=333, display_name='approver3@example.com'),
          phase_ref=issue_objects_pb2.PhaseRef()
      )
      )

    work_env.WorkEnv(mc, self.services).UpdateIssueApproval.\
    assert_called_once_with(
        self.issue_1.issue_id, 3, ANY, u'Well, actually', False,
        attachments=[(u'a.txt', 'aaaaa', 'text/plain')]
    )
    self.assertEqual(expected, actual)

  @patch('businesslogic.work_env.WorkEnv.UpdateIssueApproval')
  @patch('features.send_notifications.PrepareAndSendApprovalChangeNotification')
  def testUpdateApproval_IsDescription(
      self, _mockPrepareAndSend, mockUpdateIssueApproval):
    """We can update an approval survey."""

    av_3 = tracker_pb2.ApprovalValue(approval_id=3)
    self.issue_1.approval_values = [av_3]

    config = self.services.config.GetProjectConfig(self.cnxn, 789)
    config.field_defs = [self.fd_3]
    self.services.config.StoreConfig(self.cnxn, config)

    issue_ref = common_pb2.IssueRef(project_name='proj', local_id=1)
    field_ref = common_pb2.FieldRef(field_name='LegalApproval')
    approval_delta = issues_pb2.ApprovalDelta()

    request = issues_pb2.UpdateApprovalRequest(
        issue_ref=issue_ref, field_ref=field_ref, approval_delta=approval_delta,
        comment_content='Better response.', is_description=True)

    request.issue_ref.project_name = 'proj'
    request.issue_ref.local_id = 1
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='approver3@example.com',
        auth=self.auth)

    mockUpdateIssueApproval.return_value = [
        tracker_pb2.ApprovalValue(approval_id=3),
        'comment_pb']

    actual = self.CallWrapped(self.issues_svcr.UpdateApproval, mc, request)

    expected = issues_pb2.UpdateApprovalResponse()
    expected.approval.CopyFrom(
        issue_objects_pb2.Approval(
            field_ref=common_pb2.FieldRef(
                field_id=3,
                field_name='LegalApproval',
                type=common_pb2.APPROVAL_TYPE),
            setter_ref=common_pb2.UserRef(display_name='----'),
            phase_ref=issue_objects_pb2.PhaseRef()
        )
    )

    work_env.WorkEnv(mc, self.services
    ).UpdateIssueApproval.assert_called_once_with(
        self.issue_1.issue_id, 3,
        tracker_pb2.ApprovalDelta(),
        u'Better response.', True, attachments=[])
    self.assertEqual(expected, actual)


  @patch('businesslogic.work_env.WorkEnv.SnapshotCountsQuery')
  def testSnapshotCounts_RequiredFields(self, mockSnapshotCountsQuery):
    """Test that timestamp is required at all times.
    And that label_prefix is required when group_by is 'label'.
    """
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')

    # Test timestamp is required.
    request = issues_pb2.IssueSnapshotRequest(project_name='proj')
    with self.assertRaises(exceptions.InputException):
      self.CallWrapped(self.issues_svcr.IssueSnapshot, mc, request)

    # Test project_name is required.
    request = issues_pb2.IssueSnapshotRequest(timestamp=1531334109)
    with self.assertRaises(exceptions.InputException):
      self.CallWrapped(self.issues_svcr.IssueSnapshot, mc, request)

    # Test label_prefix is required when group_by is 'label'.
    request = issues_pb2.IssueSnapshotRequest(timestamp=1531334109,
        project_name='proj', group_by='label')
    with self.assertRaises(exceptions.InputException):
      self.CallWrapped(self.issues_svcr.IssueSnapshot, mc, request)

    mockSnapshotCountsQuery.assert_not_called()

  @patch('businesslogic.work_env.WorkEnv.SnapshotCountsQuery')
  def testSnapshotCounts_Basic(self, mockSnapshotCountsQuery):
    """Tests the happy path case."""
    request = issues_pb2.IssueSnapshotRequest(
        timestamp=1531334109, project_name='proj')
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mockSnapshotCountsQuery.return_value = ({'total': 123}, [])

    response = self.CallWrapped(self.issues_svcr.IssueSnapshot, mc, request)

    self.assertEqual(123, response.snapshot_count[0].count)
    self.assertEqual(0, len(response.unsupported_field))
    mockSnapshotCountsQuery.assert_called_once_with(self.project, 1531334109,
        '', '', '', None)

  @patch('businesslogic.work_env.WorkEnv.SnapshotCountsQuery')
  def testSnapshotCounts_GroupByLabel(self, mockSnapshotCountsQuery):
    """Tests grouping by label with label_prefix and a query.
    But no canned_query.
    """
    request = issues_pb2.IssueSnapshotRequest(timestamp=1531334109,
        project_name='proj', group_by='label', label_prefix='Type',
        query='rutabaga:rutabaga')
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mockSnapshotCountsQuery.return_value = (
        {'label1': 123, 'label2': 987},
        ['rutabaga'])

    response = self.CallWrapped(self.issues_svcr.IssueSnapshot, mc, request)

    self.assertEqual(2, len(response.snapshot_count))
    self.assertEqual('label1', response.snapshot_count[0].dimension)
    self.assertEqual(123, response.snapshot_count[0].count)
    self.assertEqual('label2', response.snapshot_count[1].dimension)
    self.assertEqual(987, response.snapshot_count[1].count)
    self.assertEqual(1, len(response.unsupported_field))
    self.assertEqual('rutabaga', response.unsupported_field[0])
    mockSnapshotCountsQuery.assert_called_once_with(self.project, 1531334109,
        'label', 'Type', 'rutabaga:rutabaga', None)

  @patch('businesslogic.work_env.WorkEnv.SnapshotCountsQuery')
  def testSnapshotCounts_GroupByComponent(self, mockSnapshotCountsQuery):
    """Tests grouping by component with a query and a canned_query."""
    request = issues_pb2.IssueSnapshotRequest(timestamp=1531334109,
        project_name='proj', group_by='component', query='rutabaga:rutabaga',
        canned_query=2)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mockSnapshotCountsQuery.return_value = (
        {'component1': 123, 'component2': 987},
        ['rutabaga'])

    response = self.CallWrapped(self.issues_svcr.IssueSnapshot, mc, request)

    self.assertEqual(2, len(response.snapshot_count))
    self.assertEqual('component1', response.snapshot_count[0].dimension)
    self.assertEqual(123, response.snapshot_count[0].count)
    self.assertEqual('component2', response.snapshot_count[1].dimension)
    self.assertEqual(987, response.snapshot_count[1].count)
    self.assertEqual(1, len(response.unsupported_field))
    self.assertEqual('rutabaga', response.unsupported_field[0])
    mockSnapshotCountsQuery.assert_called_once_with(self.project, 1531334109,
        'component', '', 'rutabaga:rutabaga', 'is:open')

  def AddField(self, name, **kwargs):
    if kwargs.get('needs_perm'):
      kwargs['needs_member'] = True
    kwargs.setdefault('cnxn', self.cnxn)
    kwargs.setdefault('project_id', self.project.project_id)
    kwargs.setdefault('field_name', name)
    kwargs.setdefault('field_type_str', 'USER_TYPE')
    for arg in ('applic_type', 'applic_pred', 'is_required', 'is_niche',
                'is_multivalued', 'min_value', 'max_value', 'regex',
                'needs_member', 'needs_perm', 'grants_perm', 'notify_on',
                'date_action_str', 'docstring', 'admin_ids'):
      kwargs.setdefault(arg, None)

    return self.services.config.CreateFieldDef(**kwargs)

  @patch('testing.fake.FeaturesService.GetFilterRules')
  def testPresubmitIssue_NoDerivedFields(self, mockGetFilterRules):
    """When no rules match, we respond with just owner availability."""
    issue_ref = common_pb2.IssueRef(project_name='proj', local_id=1)
    issue_delta = issues_pb2.IssueDelta(
        owner_ref=common_pb2.UserRef(user_id=111L),
        label_refs_add=[common_pb2.LabelRef(label='foo')])

    mockGetFilterRules.return_value = [
        filterrules_helpers.MakeRule('label:bar', add_labels=['baz'])]

    request = issues_pb2.PresubmitIssueRequest(
        issue_ref=issue_ref, issue_delta=issue_delta)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)
    response = self.CallWrapped(self.issues_svcr.PresubmitIssue, mc, request)

    self.assertEqual(
        issues_pb2.PresubmitIssueResponse(
            owner_availability="User never visited",
            owner_availability_state="never"),
        response)

  @patch('testing.fake.FeaturesService.GetFilterRules')
  def testPresubmitIssue_DerivedLabels(self, mockGetFilterRules):
    """Test that we can match label rules and return derived labels."""
    issue_ref = common_pb2.IssueRef(project_name='proj', local_id=1)
    issue_delta = issues_pb2.IssueDelta(
        owner_ref=common_pb2.UserRef(user_id=111L),
        label_refs_add=[common_pb2.LabelRef(label='foo')])

    mockGetFilterRules.return_value = [
        filterrules_helpers.MakeRule('label:foo', add_labels=['bar', 'baz'])]

    request = issues_pb2.PresubmitIssueRequest(
        issue_ref=issue_ref, issue_delta=issue_delta)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)
    response = self.CallWrapped(self.issues_svcr.PresubmitIssue, mc, request)

    self.assertEqual(
        [common_pb2.ValueAndWhy(
            value='bar',
            why='Added by rule: IF label:foo THEN ADD LABEL'),
         common_pb2.ValueAndWhy(
            value='baz',
            why='Added by rule: IF label:foo THEN ADD LABEL')],
        [vnw for vnw in response.derived_labels])

  @patch('testing.fake.FeaturesService.GetFilterRules')
  def testPresubmitIssue_DerivedOwner(self, mockGetFilterRules):
    """Test that we can match component rules and return derived owners."""
    self.services.config.CreateComponentDef(
        self.cnxn, self.project.project_id, 'Foo', 'Foo Docstring', False,
        [], [], 0, 111L, [])
    issue_ref = common_pb2.IssueRef(project_name='proj', local_id=1)
    issue_delta = issues_pb2.IssueDelta(
        comp_refs_add=[common_pb2.ComponentRef(path='Foo')])

    mockGetFilterRules.return_value = [
        filterrules_helpers.MakeRule('component:Foo', default_owner_id=222L)]

    request = issues_pb2.PresubmitIssueRequest(
        issue_ref=issue_ref, issue_delta=issue_delta)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)
    response = self.CallWrapped(self.issues_svcr.PresubmitIssue, mc, request)

    self.assertEqual(
        [common_pb2.ValueAndWhy(
            value='approver2@example.com',
            why='Added by rule: IF component:Foo THEN SET DEFAULT OWNER')],
        [vnw for vnw in response.derived_owners])

  @patch('testing.fake.FeaturesService.GetFilterRules')
  def testPresubmitIssue_DerivedCCs(self, mockGetFilterRules):
    """Test that we can match field rules and return derived cc emails."""
    field_id = self.AddField('Foo', field_type_str='ENUM_TYPE')
    issue_ref = common_pb2.IssueRef(project_name='proj', local_id=1)
    issue_delta = issues_pb2.IssueDelta(
        owner_ref=common_pb2.UserRef(user_id=111L),
        field_vals_add=[issue_objects_pb2.FieldValue(
            value='Bar', field_ref=common_pb2.FieldRef(field_id=field_id))])

    mockGetFilterRules.return_value = [
        filterrules_helpers.MakeRule('Foo=Bar', add_cc_ids=[222L, 333L])]

    request = issues_pb2.PresubmitIssueRequest(
        issue_ref=issue_ref, issue_delta=issue_delta)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)
    response = self.CallWrapped(self.issues_svcr.PresubmitIssue, mc, request)

    self.assertEqual(
        [common_pb2.ValueAndWhy(
            value='approver2@example.com',
            why='Added by rule: IF Foo=Bar THEN ADD CC'),
         common_pb2.ValueAndWhy(
            value='approver3@example.com',
            why='Added by rule: IF Foo=Bar THEN ADD CC')],
        [vnw for vnw in response.derived_ccs])

  @patch('testing.fake.FeaturesService.GetFilterRules')
  def testPresubmitIssue_Warnings(self, mockGetFilterRules):
    """Test that we can match owner rules and return warnings."""
    issue_ref = common_pb2.IssueRef(project_name='proj', local_id=1)
    issue_delta = issues_pb2.IssueDelta(
        owner_ref=common_pb2.UserRef(user_id=111L))

    mockGetFilterRules.return_value = [
        filterrules_helpers.MakeRule(
            'owner:owner@example.com', warning='Owner is too busy')]

    request = issues_pb2.PresubmitIssueRequest(
        issue_ref=issue_ref, issue_delta=issue_delta)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)
    response = self.CallWrapped(self.issues_svcr.PresubmitIssue, mc, request)

    self.assertEqual(
        [common_pb2.ValueAndWhy(
            value='Owner is too busy',
            why='Added by rule: IF owner:owner@example.com THEN ADD WARNING')],
        [vnw for vnw in response.warnings])

  @patch('testing.fake.FeaturesService.GetFilterRules')
  def testPresubmitIssue_Errors(self, mockGetFilterRules):
    """Test that we can match owner rules and return errors."""
    issue_ref = common_pb2.IssueRef(project_name='proj', local_id=1)
    issue_delta = issues_pb2.IssueDelta(
        owner_ref=common_pb2.UserRef(user_id=111L),
        cc_refs_add=[
            common_pb2.UserRef(user_id=222L),
            common_pb2.UserRef(user_id=333L)])

    mockGetFilterRules.return_value = [
        filterrules_helpers.MakeRule(
            'owner:owner@example.com', error='Owner is not to be disturbed')]

    request = issues_pb2.PresubmitIssueRequest(
        issue_ref=issue_ref, issue_delta=issue_delta)
    mc = monorailcontext.MonorailContext(
        self.services, cnxn=self.cnxn, requester='owner@example.com')
    mc.LookupLoggedInUserPerms(self.project)
    response = self.CallWrapped(self.issues_svcr.PresubmitIssue, mc, request)

    self.assertEqual(
        [common_pb2.ValueAndWhy(
            value='Owner is not to be disturbed',
            why='Added by rule: IF owner:owner@example.com THEN ADD ERROR')],
        [vnw for vnw in response.errors])
