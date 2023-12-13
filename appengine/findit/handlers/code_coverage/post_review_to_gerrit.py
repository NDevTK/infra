# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import logging
import time

from common.base_handler import BaseHandler, Permission
from common.findit_http_client import FinditHttpClient

from model.code_coverage import BlockingStatus
from model.code_coverage import LowCoverageBlocking
from services import bigquery_helper


class PostReviewToGerrit(BaseHandler):
  PERMISSION_LEVEL = Permission.APP_SELF

  def HandlePost(self, **kwargs):
    body = json.loads(self.request.body)
    host = body['host']
    change = int(body['change'])
    patchset = int(body['patchset'])
    cohorts_matched = set(body['cohorts_matched'])
    cohorts_violated = set(body['cohorts_violated'])
    assert cohorts_violated.issubset(
        cohorts_matched), "Invalid state for blocking change=%d patchset=%d" % (
            body['change'], body['patchset'])
    tracking_entity = LowCoverageBlocking.Get(host, change, patchset)
    assert (tracking_entity.blocking_status
            in [BlockingStatus.VERDICT_BLOCK, BlockingStatus.VERDICT_NOT_BLOCK])
    url = 'https://%s/changes/%d/revisions/%d/review' % (host, change, patchset)
    headers = {'Content-Type': 'application/json; charset=UTF-8'}
    FinditHttpClient().Post(url, json.dumps(body['data']), headers=headers)

    bq_rows = self._CreateBigQueryRows(
        change, patchset, cohorts_matched, cohorts_violated,
        tracking_entity.blocking_status == BlockingStatus.VERDICT_BLOCK,
        int(time.time()))
    if bq_rows:
      bigquery_helper.ReportRowsToBigquery(bq_rows, 'findit-for-me',
                                           'code_coverage_summaries',
                                           'low_coverage_blocking_raw')
      logging.info('%d rows added for change %d, patchset %d' %
                   (len(bq_rows), change, patchset))

    return {'return_code': 200}

  def _CreateBigQueryRows(self, change, patchset, cohorts_matched,
                          cohorts_violated, is_cl_blocked, run_id):
    rows = []
    for cohort in cohorts_matched:
      rows.append({
          'cl_number': change,
          'cl_patchset': patchset,
          'cohort': cohort,
          'is_cohort_violated': cohort in cohorts_violated,
          'is_cl_blocked': is_cl_blocked,
          'run_id': run_id
      })
    return rows
