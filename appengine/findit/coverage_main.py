# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2

import gae_ts_mon

from gae_libs import appengine_util

from handlers.code_coverage import export_absolute_coverage
from handlers.code_coverage import export_incremental_coverage
from handlers.code_coverage import export_gerrit_filter_coverage
from handlers.code_coverage import fetch_source_file
from handlers.code_coverage import post_review_to_gerrit
from handlers.code_coverage import process_coverage
from handlers.code_coverage import serve_ci_coverage
from handlers.code_coverage import serve_cq_coverage
from handlers.code_coverage import update_postsubmit_report

# Feaure coverage worker module.
experimental_coverage_worker_handler_mappings = [
    ('.*/coverage/task/gerrit-filter-coverage.*',
     export_gerrit_filter_coverage.ExportCoverageMetrics),
]
experimental_coverage_worker_application = webapp2.WSGIApplication(
    experimental_coverage_worker_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(experimental_coverage_worker_application)

# Referenced coverage worker module.
referenced_coverage_worker_handler_mappings = [
    ('.*/coverage/task/incremental-coverage',
     export_incremental_coverage.ExportIncrementalCoverageMetrics),
    ('.*/coverage/task/low-coverage-blocking',
     post_review_to_gerrit.PostReviewToGerrit),
]
referenced_coverage_worker_application = webapp2.WSGIApplication(
    referenced_coverage_worker_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(referenced_coverage_worker_application)


# "code-coverage-backend" module.
code_coverage_backend_handler_mappings = [
    ('.*/coverage/task/fetch-source-file', fetch_source_file.FetchSourceFile),
    ('.*/coverage/task/process-data/.*',
     process_coverage.ProcessCodeCoverageData),
    ('.*/coverage/cron/files-absolute-coverage',
     export_absolute_coverage.ExportFilesAbsoluteCoverageMetricsCron),
    ('.*/coverage/task/files-absolute-coverage',
     export_absolute_coverage.ExportFilesAbsoluteCoverageMetrics),
    ('.*/coverage/cron/incremental-coverage',
     export_incremental_coverage.ExportIncrementalCoverageMetricsCron),
    ('.*/coverage/cron/all-gerrit-filter-coverage',
     export_gerrit_filter_coverage.ExportAllCoverageMetricsCron),
    ('.*/coverage/task/all-gerrit-filter-coverage',
     export_gerrit_filter_coverage.ExportAllCoverageMetrics),
    ('.*/coverage/task/postsubmit-report/update',
     update_postsubmit_report.UpdatePostsubmitReport),
]
code_coverage_backend_web_application = webapp2.WSGIApplication(
    code_coverage_backend_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(code_coverage_backend_web_application)

# "code-coverage-frontend" module.
code_coverage_frontend_handler_mappings = [
    # TODO(crbug.com/924573): Migrate to '.*/coverage/api/coverage-data'.
    ('/coverage/api/coverage-data', serve_cq_coverage.ServeCodeCoverageData),
    # These mappings are separated so that ts_mon data (e.g. latency) is
    # groupable by view. (instead of a single entry like /coverage/p/.*)
    ('/coverage/p/.*/referenced', serve_ci_coverage.ServeCodeCoverageData),
    ('/coverage/p/.*/component', serve_ci_coverage.ServeCodeCoverageData),
    ('/coverage/p/.*/dir', serve_ci_coverage.ServeCodeCoverageData),
    ('/coverage/p/.*/file', serve_ci_coverage.ServeCodeCoverageData),
    ('/coverage/p/.*', serve_ci_coverage.ServeCodeCoverageData)
]
code_coverage_frontend_web_application = webapp2.WSGIApplication(
    code_coverage_frontend_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(code_coverage_frontend_web_application)
