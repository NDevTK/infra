# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2

import gae_ts_mon

from gae_libs import appengine_util
from gae_libs.pipelines import pipeline_status_ui
from handlers import config
from handlers import home
from handlers import url_redirect
from handlers.code_coverage import serve_ci_coverage
from handlers.code_coverage import serve_cq_coverage

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
    ('/waterfall/config', config.Configuration),
    (r'/.*', url_redirect.URLRedirect),
]
waterfall_frontend_web_application = webapp2.WSGIApplication(
    waterfall_frontend_web_pages_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(waterfall_frontend_web_application)

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
