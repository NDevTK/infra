# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import webapp2

import gae_ts_mon

from gae_libs import appengine_util
from gae_libs.pipelines import pipeline_handlers
from gae_libs.pipelines import pipeline_status_ui
from handlers import completed_build_pubsub_ingestor
from handlers import home
from handlers import url_redirect
from handlers import not_available

from components import endpoints_webapp2

# Default module.
default_web_pages_handler_mappings = [
    ('/_ah/push-handlers/index-isolated-builds',
     completed_build_pubsub_ingestor.CompletedBuildPubsubIngestor),
    ('/', home.Home),
]
default_web_application = webapp2.WSGIApplication(
    default_web_pages_handler_mappings, debug=False)
if appengine_util.IsInProductionApp():
  gae_ts_mon.initialize_prod(default_web_application)

