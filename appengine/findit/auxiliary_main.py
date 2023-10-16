# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2

import gae_ts_mon

from gae_libs import appengine_util

from handlers.code_coverage import export_incremental_coverage
from handlers.code_coverage import export_gerrit_filter_coverage
from handlers.code_coverage import post_review_to_gerrit

# auxiliary coverage worker module.
auxiliary_coverage_worker_handler_mappings = [
    ('.*/coverage/task/gerrit-filter-coverage.*',
     export_gerrit_filter_coverage.ExportCoverageMetrics),
    ('.*/coverage/task/incremental-coverage',
     export_incremental_coverage.ExportIncrementalCoverageMetrics),
    ('.*/coverage/task/low-coverage-blocking',
     post_review_to_gerrit.PostReviewToGerrit),
]
auxiliary_coverage_worker_application = webapp2.WSGIApplication(
    auxiliary_coverage_worker_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(auxiliary_coverage_worker_application)
