# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from flask import Flask

import gae_ts_mon

from gae_libs import appengine_util

from handlers.code_coverage import export_incremental_coverage
from handlers.code_coverage import export_gerrit_filter_coverage
from handlers.code_coverage import post_review_to_gerrit

# auxiliary coverage worker module.
handler_mappings = [
    ('/coverage/task/gerrit-filter-coverage', 'ExportCoverageMetrics',
     export_gerrit_filter_coverage.ExportCoverageMetrics().Handle, ['GET']),
    ('/coverage/task/incremental-coverage', 'ExportIncrementalCoverageMetrics',
     export_incremental_coverage.ExportIncrementalCoverageMetrics().Handle,
     ['GET']),
    ('/coverage/task/low-coverage-blocking', 'PostReviewToGerrit',
     post_review_to_gerrit.PostReviewToGerrit().Handle, ['POST']),
]
auxiliary_coverage_worker_application = Flask(__name__)
for url, endpoint, view_func, methods in handler_mappings:
  auxiliary_coverage_worker_application.add_url_rule(
      url, endpoint=endpoint, view_func=view_func, methods=methods)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(auxiliary_coverage_worker_application)
