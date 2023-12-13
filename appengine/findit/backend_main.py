# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from flask import Flask

import gae_ts_mon

from gae_libs import appengine_util

from handlers.code_coverage import export_absolute_coverage
from handlers.code_coverage import export_incremental_coverage
from handlers.code_coverage import export_gerrit_filter_coverage
from handlers.code_coverage import fetch_source_file
from handlers.code_coverage import process_coverage
from handlers.code_coverage import update_postsubmit_report

# "code-coverage-backend" module.
handler_mappings = [
    ('/coverage/task/fetch-source-file', 'FetchSourceFile',
     fetch_source_file.FetchSourceFile().Handle, ['POST']),
    ('/coverage/task/process-data/build/<build_id>', 'ProcessCodeCoverageData',
     process_coverage.ProcessCodeCoverageData().Handle, ['GET', 'POST']),
    ('/coverage/cron/files-absolute-coverage',
     'ExportFilesAbsoluteCoverageMetricsCron',
     export_absolute_coverage.ExportFilesAbsoluteCoverageMetricsCron().Handle,
     ['GET']),
    ('/coverage/task/files-absolute-coverage',
     'ExportFilesAbsoluteCoverageMetrics',
     export_absolute_coverage.ExportFilesAbsoluteCoverageMetrics().Handle,
     ['GET']),
    ('/coverage/cron/incremental-coverage',
     'ExportIncrementalCoverageMetricsCron',
     export_incremental_coverage.ExportIncrementalCoverageMetricsCron().Handle,
     ['GET']),
    ('/coverage/cron/all-gerrit-filter-coverage',
     'ExportAllCoverageMetricsCron',
     export_gerrit_filter_coverage.ExportAllCoverageMetricsCron().Handle,
     ['GET']),
    ('/coverage/task/all-gerrit-filter-coverage', 'ExportAllCoverageMetrics',
     export_gerrit_filter_coverage.ExportAllCoverageMetrics().Handle, ['GET']),
    ('/coverage/task/postsubmit-report/update', 'UpdatePostsubmitReport',
     update_postsubmit_report.UpdatePostsubmitReport().Handle, ['POST']),
]
code_coverage_backend_web_application = Flask(__name__)
for url, endpoint, view_func, methods in handler_mappings:
  code_coverage_backend_web_application.add_url_rule(
      url, endpoint=endpoint, view_func=view_func, methods=methods)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(code_coverage_backend_web_application)
