# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from flask import Flask

# Fix imports before importing gae_ts_mon.
import import_utils

import_utils.FixImports()
import gae_ts_mon

# Setup cloud logging
import google.cloud.logging

client = google.cloud.logging.Client()
client.setup_logging()

import google.appengine.api

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
auxiliary_application = Flask(__name__)
auxiliary_application.wsgi_app = google.appengine.api.wrap_wsgi_app(
    auxiliary_application.wsgi_app)
for url, endpoint, view_func, methods in handler_mappings:
  auxiliary_application.add_url_rule(
      url, endpoint=endpoint, view_func=view_func, methods=methods)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(auxiliary_application)
