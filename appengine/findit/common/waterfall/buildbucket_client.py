# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
""""Serves as a client for selected APIs in Buildbucket."""

import collections
import json
import logging

from go.chromium.org.luci.buildbucket.proto import common_pb2
from go.chromium.org.luci.buildbucket.proto.build_pb2 import Build
from go.chromium.org.luci.buildbucket.proto.builder_common_pb2 import BuilderID
from go.chromium.org.luci.buildbucket.proto.builds_service_pb2 import (
    BuildPredicate, GetBuildRequest, ScheduleBuildRequest, SearchBuildsRequest,
    SearchBuildsResponse)
from common.findit_http_client import FinditHttpClient
from libs.math.integers import constrain

# https://github.com/grpc/grpc-go/blob/master/codes/codes.go
GRPC_OK = '0'

# TODO: save these settings in datastore and create a role account.
_BUILDBUCKET_HOST = 'cr-buildbucket.appspot.com'
_BUILDBUCKET_PUT_GET_ENDPOINT = (
    'https://{hostname}/_ah/api/buildbucket/v1/builds'.format(
        hostname=_BUILDBUCKET_HOST))
_LUCI_PREFIX = 'luci.'
_BUILDBUCKET_V2_GET_BUILD_ENDPOINT = (
    'https://{hostname}/prpc/buildbucket.v2.Builds/GetBuild'.format(
        hostname=_BUILDBUCKET_HOST))
_BUILDBUCKET_V2_SEARCH_BUILDS_ENDPOINT = (
    'https://{hostname}/prpc/buildbucket.v2.Builds/SearchBuilds'.format(
        hostname=_BUILDBUCKET_HOST))
_BUILDBUCKET_V2_SCHEDULE_BUILD_ENDPOINT = (
    'https://{hostname}/prpc/buildbucket.v2.Builds/ScheduleBuild'.format(
        hostname=_BUILDBUCKET_HOST))



def GetV2Build(build_id, fields=None):
  """Get a buildbucket build from the v2 API.

  Args:
    build_id (str): Buildbucket id of the build to get.
    fields (google.protobuf.FieldMask): Mask for the paths to get, as not all
        fields are populated by default (such as steps).

  Returns:
    A buildbucket_proto.build_pb2.Build proto.
  """
  request = GetBuildRequest(id=int(build_id), fields=fields)
  status_code, content, response_headers = FinditHttpClient().Post(
      _BUILDBUCKET_V2_GET_BUILD_ENDPOINT,
      request.SerializeToString(),
      headers={'Content-Type': 'application/prpc; encoding=binary'})
  if status_code == 200 and response_headers.get('X-Prpc-Grpc-Code') == GRPC_OK:
    result = Build()
    result.ParseFromString(content)
    return result
  logging.warning('Unexpected prpc code: %s',
                  response_headers.get('X-Prpc-Grpc-Code'))
  return None

