# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file.


"""Tests for pubsub.py."""

import unittest

from google.appengine.api import taskqueue
from google.appengine.ext import testbed

from features import notify
from proto import tracker_pb2
from services import service_manager
from testing import fake
from testing import testing_helpers
from tracker import attachment_helpers

from third_party import cloudstorage


def MakeTestIssue(project_id, local_id, owner_id, reporter_id):
  issue = tracker_pb2.Issue()
  issue.project_id = project_id
  issue.local_id = local_id
  issue.issue_id = 1000 * project_id + local_id
  issue.owner_id = owner_id
  issue.reporter_id = reporter_id
  return issue


class PublishPubsubIssueChangeTaskTest(unittest.TestCase):

  def setUp(self):
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_taskqueue_stub()
    self.taskqueue_stub = self.testbed.get_stub(testbed.TASKQUEUE_SERVICE_NAME)
    self.taskqueue_stub._root_path = os.path.dirname(
        os.path.dirname(os.path.dirname(__file__)))
    self.services = service_manager.Services(
        user=fake.UserService(),
        project=fake.ProjectService(),
        config=fake.ConfigService(),
        issue=fake.IssueService(),
        features=fake.FeaturesService())
    self.services.user.TestAddUser('requester@example.com', 1)
    self.services.user.TestAddUser('user@example.com', 2)
    self.services.user.TestAddUser('member@example.com', 3)
    self.services.project.TestAddProject(
        'test-project', owner_ids=[1, 3],
        project_id=12345)
    self.issue1 = MakeTestIssue(
        project_id=12345, local_id=1, owner_id=2, reporter_id=1)
    self.services.issue.TestAddIssue(self.issue1)

    self._old_gcs_open = cloudstorage.open
    cloudstorage.open = fake.gcs_open
    self.orig_sign_attachment_id = attachment_helpers.SignAttachmentID
    attachment_helpers.SignAttachmentID = (
        lambda aid: 'signed_%d' % aid)

  def tearDown(self):
    self.testbed.deactivate()
    cloudstorage.open = self._old_gcs_open
    attachment_helpers.SignAttachmentID = self.orig_sign_attachment_id

  def VerifyParams(self, result, params):
    if 'issue_id' in params:
      self.assertEqual(params['issue_id'], result['params']['issue_id'])
    if 'issue_ids' in params:
      self.assertEqual([int(p) for p in params['issue_ids'].split(',')],
                       result['params']['issue_ids'])

  def testPublishPubsubIssueChangeTask(self):
    task = notify.PublishPubsubIssueChangeTask(
        request=None, response=None, services=self.services, ps_topic='MetabugTest', ps_project='MonorailStaging')
    params = {'issue_id': 12345001, 'seq': 0, 'commenter_id': 2}
    mr = testing_helpers.MakeMonorailRequest(
        user_info={'user_id': 1},
        params=params,
        method='POST',
        services=self.services)
    result = task.HandleRequest(mr)
    self.VerifyParams(result, params)

