# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2
from flask import Flask

import gae_ts_mon

from gae_libs import appengine_util
from handlers import config
from handlers.code_coverage import serve_ci_coverage
from handlers.code_coverage import serve_cq_coverage

# "code-coverage-frontend" module.
handler_mappings = [
    # TODO(crbug.com/924573): Migrate to '.*/coverage/api/coverage-data'.
    ('/coverage/api/coverage-data', 'CQServeCodeCoverageData',
     serve_cq_coverage.ServeCodeCoverageData().Handle, ['GET']),
    # These mappings are separated so that ts_mon data (e.g. latency) is
    # groupable by view. (instead of a single entry like /coverage/p/.*)
    ('/coverage/p/<string:project>/component',
     'CIServeCodeCoverageDataComponent',
     serve_ci_coverage.ServeCodeCoverageData().Handle, ['GET']),
    ('/coverage/p/<string:project>/dir', 'CIServeCodeCoverageDataDir',
     serve_ci_coverage.ServeCodeCoverageData().Handle, ['GET']),
    ('/coverage/p/<string:project>/file', 'CIServeCodeCoverageDataFile',
     serve_ci_coverage.ServeCodeCoverageData().Handle, ['GET']),
    ('/coverage/p/<string:project>', 'CIServeCodeCoverageDataDefault',
     serve_ci_coverage.ServeCodeCoverageData().Handle, ['GET']),
    ('/waterfall/config', 'WaterfallConfig', config.Configuration().Handle,
     ['GET', 'POST'])
]
code_coverage_frontend_web_application = Flask(__name__)
for url, endpoint, view_func, methods in handler_mappings:
  code_coverage_frontend_web_application.add_url_rule(
      url, endpoint=endpoint, view_func=view_func, methods=methods)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(code_coverage_frontend_web_application)
