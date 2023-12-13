# Copyright 2021 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import logging

from common.base_handler import BaseHandler, Permission

from handlers.code_coverage import utils
from model.code_coverage import PostsubmitReport


class UpdatePostsubmitReport(BaseHandler):
  PERMISSION_LEVEL = Permission.CORP_USER

  def HandlePost(self, **kwargs):
    luci_project = self.request.values.get('luci_project')
    platform = self.request.values.get('platform')
    platform_info_map = utils.GetPostsubmitPlatformInfoMap(luci_project)
    if platform not in platform_info_map:
      return BaseHandler.CreateError('Platform: %s is not supported' % platform,
                                     400)
    test_suite_type = self.request.values.get('test_suite_type', 'all')
    modifier_id = int(self.request.values.get('modifier_id', '0'))
    bucket = platform_info_map[platform]['bucket']

    builder = platform_info_map[platform]['builder']
    if test_suite_type == 'unit':
      builder += '_unit'

    project = self.request.values.get('project')
    host = self.request.values.get('host')
    ref = self.request.values.get('ref')
    revision = self.request.values.get('revision')
    visible = self.request.values.get('visible').lower() == 'true'

    logging.info("host = %s", host)
    logging.info("project = %s", project)
    logging.info("ref = %s", ref)
    logging.info("revision = %s", revision)
    logging.info("bucket = %s", bucket)
    logging.info("builder = %s", builder)
    logging.info("modifier_id = %d", modifier_id)

    report = PostsubmitReport.Get(
        server_host=host,
        project=project,
        ref=ref,
        revision=revision,
        bucket=bucket,
        builder=builder,
        modifier_id=modifier_id)

    if not report:
      return BaseHandler.CreateError('Report record not found', 404)

    # At present, we only update visibility
    report.visible = visible
    report.put()

    return {'return_code': 200}
