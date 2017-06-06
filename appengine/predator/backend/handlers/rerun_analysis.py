# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from datetime import datetime
from datetime import timedelta
import logging

from google.appengine.ext import ndb

from analysis.type_enums import CrashClient
from common.crash_pipeline import RerunPipeline
from common.crash_pipeline import FinditForClientID
from gae_libs import appengine_util
from gae_libs.handlers.base_handler import BaseHandler
from gae_libs.handlers.base_handler import Permission
from libs import time_util

RERUN_SERVICE = 'backend-process'
RERUN_QUEUE = 'rerun-queue'

DATETIME_FORMAT = '%Y-%m-%d'

from common.model.clusterfuzz_analysis import ClusterfuzzAnalysis
from common.model.cracas_crash_analysis import CracasCrashAnalysis
from common.model.crash_analysis import CrashAnalysis
from common.model.crash_config import CrashConfig
from common.model.fracas_crash_analysis import FracasCrashAnalysis

from gae_libs.gitiles.cached_gitiles_repository import CachedGitilesRepository
from gae_libs.http.http_client_appengine import HttpClientAppengine
from gae_libs.iterator import Iterate

CLIENT_ID_TO_CRASH_ANALYSIS = {
    CrashClient.FRACAS: FracasCrashAnalysis,
    CrashClient.CRACAS: CracasCrashAnalysis,
    CrashClient.CLUSTERFUZZ: ClusterfuzzAnalysis
}

_BATCH_SIZE = 500


def IterateCrashBatches(client_id, start_date, end_date,
                        batch_size=_BATCH_SIZE):
  """Iterates and re-initializes crash analyses in [start_date, end_date)."""
  analysis = CLIENT_ID_TO_CRASH_ANALYSIS.get(client_id)

  query = analysis.query()
  query = query.filter(
      analysis.requested_time >= start_date).filter(
          analysis.requested_time < end_date)

  client = FinditForClientID(
      client_id,
      CachedGitilesRepository.Factory(HttpClientAppengine()),
      CrashConfig.Get())

  for crash_batch in Iterate(query, batch_size=batch_size, batch_run=True):
    updated_crashes = []
    for crash in crash_batch:
      crash.ReInitialize(client)
      updated_crashes.append(crash)

    keys = ndb.put_multi(updated_crashes)
    yield [key.urlsafe() for key in keys]


class RerunAnalysis(BaseHandler):
  PERMISSION_LEVEL = Permission.ADMIN

  def HandleGet(self):
    """Update crash analysis models."""
    client_id = self.request.get('client_id') or CrashClient.CRACAS

    now = time_util.GetUTCNow()
    yesterday = time_util.GetUTCNow() - timedelta(days=7)

    start_date = self.request.get('start_date')
    start_date = datetime.strptime(
        start_date, DATETIME_FORMAT) if start_date else yesterday

    end_date = self.request.get('end_date')
    end_date = datetime.strptime(
        end_date, DATETIME_FORMAT) if end_date else now

    count = 0
    for crash_keys in IterateCrashBatches(client_id, start_date, end_date):
      pipeline = RerunPipeline(client_id, crash_keys)
      # Attribute defined outside __init__ - pylint: disable=W0201
      pipeline.target = appengine_util.GetTargetNameForModule(RERUN_SERVICE)
      pipeline.start(queue_name=RERUN_QUEUE)
      count += 1

    if count == 0:
      message = 'No rerun pipeline started.'
    elif count == 1:
      message = 'A rerun pipeline kicked off.'
    else:
      message = '%d rerun pipelines kicked off.' % count

    return {'data': {'message': message}}
