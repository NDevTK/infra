# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Tests for monorail.tracker.issueattachment."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import unittest

from google.appengine.ext import testbed

try:
  from mox3 import mox
except ImportError:
  import mox

from framework import exceptions, gcs_helpers
from framework import permissions
from mrproto import tracker_pb2
from services import service_manager
from testing import fake
from testing import testing_helpers
from tracker import attachment_helpers
from tracker import issueattachment


class IssueattachmentTest(unittest.TestCase):

  def setUp(self):
    self.mox = mox.Mox()
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_memcache_stub()
    self.testbed.init_app_identity_stub()
    self.testbed.init_urlfetch_stub()
    self.attachment_data = ""

    services = service_manager.Services(
        project=fake.ProjectService(),
        config=fake.ConfigService(),
        issue=fake.IssueService(),
        user=fake.UserService())
    self.project = services.project.TestAddProject('proj')
    self.servlet = issueattachment.AttachmentPage(services=services)
    services.user.TestAddUser('commenter@example.com', 111)
    self.issue = fake.MakeTestIssue(
        self.project.project_id, 1, 'summary', 'New', 111)
    services.issue.TestAddIssue(self.issue)
    self.comment = tracker_pb2.IssueComment(
        id=123, issue_id=self.issue.issue_id,
        project_id=self.project.project_id, user_id=111,
        content='this is a comment')
    services.issue.TestAddComment(self.comment, self.issue.local_id)
    self.attachment = tracker_pb2.Attachment(
        attachment_id=54321, filename='hello.txt', filesize=23432,
        mimetype='text/plain', gcs_object_id='/pid/attachments/object_id')
    services.issue.TestAddAttachment(
        self.attachment, self.comment.id, self.issue.issue_id)
    self.orig_sign_attachment_id = attachment_helpers.SignAttachmentID
    attachment_helpers.SignAttachmentID = (
        lambda aid: 'signed_%d' % aid)

  def tearDown(self):
    self.mox.UnsetStubs()
    self.mox.ResetAll()
    self.testbed.deactivate()
    attachment_helpers.SignAttachmentID = self.orig_sign_attachment_id

  def testGatherPageData_NotFound(self):
    aid = 12345
    path = '/p/proj/issues/attachment?aid=%s&signed_aid=signed_%d' % (
        aid, aid)
    # But, no such attachment is in the database.
    _request, mr = testing_helpers.GetRequestObjects(
        project=self.project, path=path,
        perms=permissions.EMPTY_PERMISSIONSET)
    with self.assertRaises(Exception) as cm:
      self.servlet.GatherPageData(mr)
    self.assertEqual(404, cm.exception.code)

  # TODO(jrobbins): test cases for missing comment and missing issue.

  def testGatherPageData_PermissionDenied(self):
    aid = self.attachment.attachment_id
    path = '/p/proj/issues/attachment?aid=%s&signed_aid=signed_%d' % (
        aid, aid)
    _request, mr = testing_helpers.GetRequestObjects(
        project=self.project, path=path,
        perms=permissions.EMPTY_PERMISSIONSET)  # not even VIEW
    self.assertRaises(
        permissions.PermissionException,
        self.servlet.GatherPageData, mr)

    _request, mr = testing_helpers.GetRequestObjects(
        project=self.project, path=path,
        perms=permissions.READ_ONLY_PERMISSIONSET)  # includes VIEW

    # issue is now deleted
    self.issue.deleted = True
    self.assertRaises(
        permissions.PermissionException,
        self.servlet.GatherPageData, mr)
    self.issue.deleted = False

    # issue is now restricted
    self.issue.labels.extend(['Restrict-View-PermYouLack'])
    self.assertRaises(
        permissions.PermissionException,
        self.servlet.GatherPageData, mr)

  def testGatherPageData_Download_WithDisposition(self):
    aid = self.attachment.attachment_id
    self.mox.StubOutWithMock(gcs_helpers, 'MaybeCreateDownload')
    gcs_helpers.MaybeCreateDownload(
        'app_default_bucket',
        '/pid/attachments/object_id',
        self.attachment.filename).AndReturn(True)
    self.mox.StubOutWithMock(gcs_helpers, 'SignUrl')
    gcs_helpers.SignUrl(
        'app_default_bucket',
        '/pid/attachments/object_id-download'
        ).AndReturn('googleusercontent.com/...-download...')
    path = '/p/proj/issues/attachment?aid=%s&signed_aid=signed_%d' % (
        aid, aid)
    _request, mr = testing_helpers.GetRequestObjects(
        project=self.project, path=path,
        perms=permissions.READ_ONLY_PERMISSIONSET)  # includes VIEW
    self.mox.ReplayAll()
    with self.assertRaises(exceptions.RedirectException) as e:
      self.servlet.GatherPageData(mr)
    self.mox.VerifyAll()
    self.assertIn('googleusercontent.com', str(e.exception))
    self.assertIn('-download', str(e.exception))

  def testGatherPageData_Download_WithoutDisposition(self):
    aid = self.attachment.attachment_id
    path = '/p/proj/issues/attachment?aid=%s&signed_aid=signed_%d' % (
        aid, aid)
    self.mox.StubOutWithMock(gcs_helpers, 'MaybeCreateDownload')
    gcs_helpers.MaybeCreateDownload(
        'app_default_bucket',
        '/pid/attachments/object_id',
        self.attachment.filename).AndReturn(False)
    self.mox.StubOutWithMock(gcs_helpers, 'SignUrl')
    gcs_helpers.SignUrl(
        'app_default_bucket',
        '/pid/attachments/object_id'
        ).AndReturn('googleusercontent.com/...')
    _request, mr = testing_helpers.GetRequestObjects(
        project=self.project, path=path,
        perms=permissions.READ_ONLY_PERMISSIONSET)  # includes VIEW
    self.mox.ReplayAll()
    with self.assertRaises(exceptions.RedirectException) as e:
      self.servlet.GatherPageData(mr)
    self.mox.VerifyAll()
    self.assertIn('googleusercontent.com', str(e.exception))
    self.assertNotIn('-download', str(e.exception))

  def testGatherPageData_DownloadBadFilename(self):
    aid = self.attachment.attachment_id
    path = '/p/proj/issues/attachment?aid=%s&signed_aid=signed_%d' % (
        aid, aid)
    self.attachment.filename = '<script>alert("xsrf")</script>.txt';
    safe_filename = 'attachment-%d.dat' % aid
    self.mox.StubOutWithMock(gcs_helpers, 'MaybeCreateDownload')
    gcs_helpers.MaybeCreateDownload(
        'app_default_bucket',
        '/pid/attachments/object_id',
        safe_filename).AndReturn(True)
    self.mox.StubOutWithMock(gcs_helpers, 'SignUrl')
    gcs_helpers.SignUrl(
        'app_default_bucket',
        '/pid/attachments/object_id-download'
        ).AndReturn('googleusercontent.com/...-download...')
    _request, mr = testing_helpers.GetRequestObjects(
        project=self.project,
        path=path,
        perms=permissions.READ_ONLY_PERMISSIONSET)  # includes VIEW
    self.mox.ReplayAll()
    with self.assertRaises(exceptions.RedirectException) as e:
      self.servlet.GatherPageData(mr)
    self.mox.VerifyAll()
    self.assertIn('googleusercontent.com', str(e.exception))
