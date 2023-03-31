# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json

from common.findit_http_client import FinditHttpClient
from gae_libs.handlers.base_handler import BaseHandler, Permission

from model.code_coverage import BlockingStatus
from model.code_coverage import LowCoverageBlocking


class PostReviewToGerrit(BaseHandler):
  PERMISSION_LEVEL = Permission.APP_SELF

  def HandlePost(self):
    body = json.loads(self.request.body)
    tracking_entity = LowCoverageBlocking.Get(body['host'], body['change'],
                                              body['patchset'])
    assert (tracking_entity.blocking_status
            in [BlockingStatus.VERDICT_BLOCK, BlockingStatus.VERDICT_NOT_BLOCK])
    url = 'https://%s/changes/%d/revisions/%d/review' % (
        body['host'], body['change'], body['patchset'])
    headers = {'Content-Type': 'application/json; charset=UTF-8'}
    FinditHttpClient().Post(url, json.dumps(body['data']), headers=headers)
    return {'return_code': 200}
