# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""This module integrates buildbucket with resultdb."""

import json
import logging

from google.appengine.ext import ndb
import webapp2

from components import decorators
from components.prpc import client
from components.prpc import codes
from go.chromium.org.luci.buildbucket.proto import common_pb2
from go.chromium.org.luci.resultdb.proto.rpc.v1 import recorder_pb2
from go.chromium.org.luci.resultdb.proto.rpc.v1 import recorder_prpc_pb2
from go.chromium.org.luci.resultdb.proto.rpc.v1 import invocation_pb2

import model
import tq


def create_invocation(build):
  """Creates an invocation for the build.

  Modifies the build entity given and its infra property to record the new
  invocation and its update token.

  Returns a boolean indicating whether a new invocation was created (and the
  build's fields modified).
  """
  rdb = build.proto.infra.resultdb
  if not rdb.hostname or rdb.invocation:
    return False

  invocation_id = 'build:%d' % build.proto.id
  response_metadata = {}
  recorder = client.Client(
      build.proto.infra.resultdb.hostname,
      recorder_prpc_pb2.RecorderServiceDescription,
  )
  new_invocation = recorder.CreateInvocation(
      recorder_pb2.CreateInvocationRequest(
          invocation_id=invocation_id,
          invocation=invocation_pb2.Invocation(),
          request_id=invocation_id,
      ),
      credentials=client.service_account_credentials(),
      response_metadata=response_metadata,
  )
  update_token = response_metadata['update-token']

  assert new_invocation
  assert new_invocation.name, 'Empty invocation name in %s' % (new_invocation)
  assert update_token

  rdb.invocation = new_invocation.name
  build.resultdb_update_token = update_token
  return True


def enqueue_invocation_finalization_async(build):
  """Enqueues a task to call ResultDB to finalize the build's invocation."""
  assert ndb.in_transaction()
  assert build
  assert build.is_ended

  task_def = {
      'url': '/internal/task/resultdb/finalize/%d' % build.key.id(),
      'payload': {'id': build.key.id()},
      'retry_options': {'task_age_limit': model.BUILD_TIMEOUT.total_seconds()},
  }

  return tq.enqueue_async('backend-default', [task_def])


class FinalizeInvocation(webapp2.RequestHandler):  # pragma: no cover
  """Calls ResultDB to finalize the build's invocation."""

  @decorators.require_taskqueue('backend-default')
  def post(self, build_id):  # pylint: disable=unused-argument
    build_id = json.loads(self.request.body)['id']
    _finalize_invocation(build_id)


def _is_interrupted(build):
  # Treat canceled builds and infra failures as interrupted builds.
  return build.status in (common_pb2.CANCELED, common_pb2.INFRA_FAILURE)


def _finalize_invocation(build_id):
  bundle = model.BuildBundle.get(build_id, infra=True)
  rdb = bundle.infra.parse().resultdb
  if not rdb.hostname:
    # If there's no hostname, it means resultdb integration is not enabled
    # for this build.
    return

  if not rdb.invocation:
    # This is a problem, swarming._sync_build() should have created an
    # invocation and saved the name to this field.
    logging.error(
        'Cannot finalize invocation for build %s without an invocation name',
        build_id
    )
    return  # Avoid retry.

  try:
    _ = _call_finalize_rpc(
        rdb.hostname,
        recorder_pb2.FinalizeInvocationRequest(
            name=rdb.invocation, interrupted=_is_interrupted(bundle.build)
        ),
        {'update-token': bundle.build.resultdb_update_token},
    )
  except client.RpcError as rpce:
    if rpce.status_code in (codes.StatusCode.FAILED_PRECONDITION,
                            codes.StatusCode.PERMISSION_DENIED):
      logging.error('RpcError when finalizing %s: %s', rdb.invocation, rpce)
    else:
      raise  # Retry other errors.


def _call_finalize_rpc(host, req, metadata):  # pragma: no cover
  recorder = client.Client(
      host,
      recorder_prpc_pb2.RecorderServiceDescription,
  )
  return recorder.FinalizeInvocation(
      req, credentials=client.service_account_credentials(), metadata=metadata
  )
