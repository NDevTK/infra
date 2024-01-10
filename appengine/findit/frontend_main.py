# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from flask import Flask

# Fix imports before importing gae_ts_mon.
import import_utils
import os

import_utils.FixImports()
import gae_ts_mon

# Setup cloud logging
import google.cloud.logging

if os.getenv('GAE_ENV') == 'standard':
  # If this isn't a local server, set up cloud logging.
  client = google.cloud.logging.Client()
  client.setup_logging()

import google.appengine.api

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
frontend_application = Flask(__name__)
frontend_application.wsgi_app = google.appengine.api.wrap_wsgi_app(
    frontend_application.wsgi_app)
for url, endpoint, view_func, methods in handler_mappings:
  frontend_application.add_url_rule(
      url, endpoint=endpoint, view_func=view_func, methods=methods)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(frontend_application)
