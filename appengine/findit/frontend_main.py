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


# waterfall frontend.
waterfall_frontend_web_pages_handler_mappings = [
    ('/waterfall/config', config.Configuration),
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
