# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from analysis.type_enums import CrashClient
from common.base_handler import BaseHandler
from common.crash_pipeline import RerunPipeline
from gae_libs import appengine_util

RERUN_SERVICE = 'backend-process'
RERUN_QUEUE = 'rerun-queue'


class RerunAnalysis(BaseHandler):
  """Rerun analysis for a single crash."""

  def HandleGet(self):
    client_id = self.request.values.get('client_id', CrashClient.CRACAS)
    key = self.request.values.get('key')
    if not key:
      return self.CreateError('Should provide key of the analysis to rerun.')

    pipeline = RerunPipeline(
        client_id, [key],
        publish_to_client=bool(self.request.values.get('publish')))
    # Attribute defined outside __init__ - pylint: disable=W0201
    pipeline.target = appengine_util.GetTargetNameForModule(RERUN_SERVICE)
    pipeline.start(queue_name=RERUN_QUEUE)

    return {'data': {'success': True}}
