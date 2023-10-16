# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2

import gae_ts_mon

from gae_libs import appengine_util

from handlers.code_coverage import export_absolute_coverage
from handlers.code_coverage import export_incremental_coverage
from handlers.code_coverage import export_gerrit_filter_coverage
from handlers.code_coverage import fetch_source_file
from handlers.code_coverage import process_coverage
from handlers.code_coverage import update_postsubmit_report

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
