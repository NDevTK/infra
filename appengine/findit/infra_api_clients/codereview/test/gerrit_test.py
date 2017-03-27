# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import mock
import os

from testing_utils import testing

from infra_api_clients.codereview.gerrit import Gerrit
from libs.http import retry_http_client


class DummyHttpClient(retry_http_client.RetryHttpClient):

  def __init__(self):
    super(DummyHttpClient, self).__init__()
    self.responses = {}
    self.requests = []

  def SetResponse(self, url, result):
    self.responses[url] = result

  def _SetChangeIdResponse(self, host, change_id, response_file):
    with open(
      os.path.join(os.path.dirname(__file__), 'data', response_file)) as f:
      response = f.read()
    url = 'https://%s/a/changes/%s/detail' % (host, change_id)
    query = '?o=%5B%27CURRENT_REVISION%27%2C+%27CURRENT_COMMIT%27%5D'
    self.SetResponse(url + query, (200, response))

  def GetBackoff(self, *_):  # pragma: no cover
    """Override to avoid sleep."""
    return 0

  def _Get(self, url, _, headers):
    self.requests.append((url, None, headers))
    return self.responses.get(url, (404, 'Not Found'))

  def _Post(self, url, data, _, headers):  # pragma: no cover
    self.requests.append((url, data, headers))
    return self.responses.get(url, (404, 'Not Found'))

  def _Put(self, *_):  # pragma: no cover
    pass


class GerritTest(testing.AppengineTestCase):

  def setUp(self):
    super(GerritTest, self).setUp()
    self.http_client = DummyHttpClient()
    self.server_hostname = 'server.host.name'
    self.gerrit = Gerrit(self.server_hostname)
    self.gerrit.HTTP_CLIENT = self.http_client
    self.maxDiff = None

  def testGetClInfoCQCommit(self):
    change_id = 'I40bc1e744806f2c4aadf0ce6609aaa61b4019fa7'
    response_file = 'cq_and_commit.json'
    self.http_client._SetChangeIdResponse(self.server_hostname, change_id,
                                          response_file)
    cl_info = self.gerrit.GetClDetails(change_id)
    self.assertNotEqual(
      cl_info.commit_attempts.values()[0].last_cq_timestamp,
      cl_info.commits[0].timestamp)

  def testGetClInfoManualCommit(self):
    change_id = 'I7ecd56d7d0c3fef90cfe998a29b948c5032980e4'
    response_file = 'manual_commit.json'
    self.http_client._SetChangeIdResponse(self.server_hostname, change_id,
                                          response_file)
    cl_info = self.gerrit.GetClDetails(change_id)
    self.assertEqual(
      cl_info.commit_attempts.values()[0].last_cq_timestamp,
      cl_info.commits[0].timestamp)

  def testGetClInfoRevertedCommit(self):
    change_id = 'I4303e1b7166aaab873587a3fda0ec907d3d8ace0'
    response_file = 'reverted_commit.json'
    self.http_client._SetChangeIdResponse(self.server_hostname, change_id,
                                          response_file)
    reverting_change_id = 'If02ca1cd494579d6bb92a157bf1819e3689cd6b1'
    reverting_response_file = 'reverting_commit.json'
    self.http_client._SetChangeIdResponse(self.server_hostname,
                                          reverting_change_id,
                                          reverting_response_file)
    cl_info = self.gerrit.GetClDetails(change_id)
    self.assertEqual(cl_info.serialize(),{
      'server_hostname': 'server.host.name',
      'reviewers': [u'agable@chromium.org',
                    u'commit-bot@chromium.org',
                    u'bpastene@chromium.org'],
      'closed': True,
      'commits': [{'patchset_id': 2,
                    'timestamp': '2017-02-27 18:56:54 UTC',
                     'revision': u'edda1046ce724695004242e943f59f5e1b2d00ff'}],
      'cc': [u'chromium-reviews@chromium.org'],
      'change_id': 'I4303e1b7166aaab873587a3fda0ec907d3d8ace0',
      'reverts': [{'patchset_id': 2,
                   'reverting_user_email': u'agable@chromium.org',
                   'timestamp': '2017-02-27 19:04:51 UTC',
                   'reverting_cl': {
                       'server_hostname': 'server.host.name',
                       'reviewers': [u'agable@chromium.org',
                                     u'commit-bot@chromium.org',
                                     u'chromium-reviews@chromium.org',
                                     u'bpastene@chromium.org'],
                       'closed': True,
                       'commits': [{
                           'patchset_id': 2,
                           'timestamp': '2017-02-27 19:05:03 UTC',
                           'revision':
                                 u'bd1db4534d7dc3f3f9c693ca0ac3e67caf484824'}],
                       'cc': [],
                       'change_id':
                             u'If02ca1cd494579d6bb92a157bf1819e3689cd6b1',
                             'reverts': [],
                             'commit_attempts': [{
                                 'patchset_id': 1,
                                 'timestamp': '2017-02-27 19:04:53 UTC',
                                 'committing_user_email':
                                     u'agable@chromium.org'}]}}],
      'commit_attempts': [{'patchset_id': 1,
                           'timestamp': '2017-02-27 18:47:15 UTC',
                           'committing_user_email': u'agable@chromium.org'}]})
