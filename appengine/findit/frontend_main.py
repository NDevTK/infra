# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2

import gae_ts_mon

from gae_libs import appengine_util
from gae_libs.pipelines import pipeline_status_ui
from handlers import auto_revert_metrics
from handlers import check_duplicate_failures
from handlers import config
from handlers import culprit
from handlers import failure_log
from handlers import home
from handlers import list_analyses
from handlers import pipeline_errors_dashboard
from handlers import trooper
from handlers import try_job_dashboard
from handlers import url_redirect

# App Engine pipeline status pages.
pipeline_status_handler_mappings = [
    ('/_ah/pipeline/rpc/tree', pipeline_status_ui._TreeStatusHandler),
    ('/_ah/pipeline/rpc/class_paths', pipeline_status_ui._ClassPathListHandler),
    ('/_ah/pipeline/rpc/list', pipeline_status_ui._RootListHandler),
    ('/_ah/pipeline(/.+)', pipeline_status_ui._StatusUiHandler),
]
pipeline_status_application = webapp2.WSGIApplication(
    pipeline_status_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(pipeline_status_application)

# waterfall frontend.
waterfall_frontend_web_pages_handler_mappings = [
    ('/', home.Home),
    ('/waterfall/auto-revert-metrics', auto_revert_metrics.AutoRevertMetrics),
    ('/waterfall/check-duplicate-failures',
     check_duplicate_failures.CheckDuplicateFailures),
    ('/waterfall/config', config.Configuration),
    ('/waterfall/culprit', culprit.Culprit),
    ('/waterfall/failure-log', failure_log.FailureLog),
    ('/waterfall/list-failures', list_analyses.ListAnalyses),
    ('/waterfall/pipeline-errors-dashboard',
     pipeline_errors_dashboard.PipelineErrorsDashboard),
    ('/waterfall/trooper', trooper.Trooper),
    ('/waterfall/try-job-dashboard', try_job_dashboard.TryJobDashboard),
    (r'/.*', url_redirect.URLRedirect),
]
waterfall_frontend_web_application = webapp2.WSGIApplication(
    waterfall_frontend_web_pages_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(waterfall_frontend_web_application)

