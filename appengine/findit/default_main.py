# Copyright 2018 The Chromium Authors. All rights reserved.
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
from handlers import completed_build_pubsub_ingestor
from handlers import home
from handlers import url_redirect

# Default module.
handler_mappings = [
    ('/_ah/push-handlers/index-isolated-builds', 'CompletedBuildPubsubIngestor',
     completed_build_pubsub_ingestor.CompletedBuildPubsubIngestor().Handle,
     ['POST']),
    ('/', 'Home', home.Home().Handle, ['GET']),
    ('/<path:rest_of_url>', 'URLRedirect', url_redirect.URLRedirect,
     ['GET', 'POST']),
]

default_application = Flask(__name__)
default_application.wsgi_app = google.appengine.api.wrap_wsgi_app(
    default_application.wsgi_app)
for url, endpoint, view_func, methods in handler_mappings:
  default_application.add_url_rule(
      url, endpoint=endpoint, view_func=view_func, methods=methods)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(default_application)
