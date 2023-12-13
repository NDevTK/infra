# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2
from flask import Flask

import gae_ts_mon

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

default_web_application = Flask(__name__)
for url, endpoint, view_func, methods in handler_mappings:
  default_web_application.add_url_rule(
      url, endpoint=endpoint, view_func=view_func, methods=methods)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(default_web_application)
