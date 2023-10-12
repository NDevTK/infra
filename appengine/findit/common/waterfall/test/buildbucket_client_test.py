# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import collections
from datetime import datetime
import json
import mock

from go.chromium.org.luci.buildbucket.proto.build_pb2 import Build
from go.chromium.org.luci.buildbucket.proto.builder_common_pb2 import BuilderID
from go.chromium.org.luci.buildbucket.proto.builds_service_pb2 import (
    SearchBuildsResponse)
from go.chromium.org.luci.buildbucket.proto.common_pb2 import GitilesCommit
from testing_utils import testing

from gae_libs.http import http_client_appengine
from common.findit_http_client import FinditHttpClient
from common.waterfall import buildbucket_client

_Result = collections.namedtuple('Result',
                                 ['content', 'status_code', 'headers'])


class BuildBucketClientTest(testing.AppengineTestCase):

  def setUp(self):
    super(BuildBucketClientTest, self).setUp()
    self.maxDiff = None

    with self.mock_urlfetch() as urlfetch:
      self.mocked_urlfetch = urlfetch


  @mock.patch.object(FinditHttpClient, 'Post')
  def testGetV2Build(self, mock_post):
    build_id = '8945610992972640896'
    mock_build = Build()
    mock_build.id = int(build_id)
    mock_build.status = 12
    mock_build.output.properties['builder_group'] = 'chromium.linux'
    mock_build.output.properties['buildername'] = 'Linux Builder'
    mock_build.output.properties.get_or_create_struct(
        'swarm_hashes_ref/heads/mockmaster(at){#123}'
    )['mock_target'] = 'mock_hash'
    gitiles_commit = mock_build.input.gitiles_commit
    gitiles_commit.host = 'gitiles.host'
    gitiles_commit.project = 'gitiles/project'
    gitiles_commit.ref = 'refs/heads/mockmaster'
    mock_build.builder.project = 'mock_luci_project'
    mock_build.builder.bucket = 'mock_bucket'
    mock_build.builder.builder = 'Linux Builder'
    mock_headers = {'X-Prpc-Grpc-Code': '0'}
    binary_data = mock_build.SerializeToString()
    mock_post.return_value = (200, binary_data, mock_headers)
    build = buildbucket_client.GetV2Build(build_id)
    self.assertIsNotNone(build)
    self.assertEqual(mock_build.id, build.id)

    mock_headers = {'X-Prpc-Grpc-Code': '4'}
    binary_data = mock_build.SerializeToString()
    mock_post.return_value = (404, binary_data, mock_headers)
    self.assertIsNone(buildbucket_client.GetV2Build(build_id))
