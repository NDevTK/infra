# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Main program for Monorail.

Monorail is an issue tracking tool that is based on the code.google.com
issue tracker, but it has been ported to Google AppEngine and Google Cloud SQL.
"""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import logging
import webapp2
from werkzeug.middleware import dispatcher

from components import endpoints_webapp2
import gae_ts_mon

import flaskregisterpages
import registerpages
from framework import sorting
from services import api_svc_v1
from services import service_manager


services = service_manager.set_up_services()
sorting.InitializeArtValues(services)
registry = registerpages.ServletRegistry()
app_routes = registry.Register(services)
app = webapp2.WSGIApplication(
    app_routes, config={'services': services})
# TODO(crbug.com/1322775) Migrate away from the shared prodx-mon-chrome-infra
# service account and change to gae_ts_mon.initialize_prod()
gae_ts_mon.initialize_adhoc(app)

flask_regist = flaskregisterpages.ServletRegistry()
app = dispatcher.DispatcherMiddleware(
    app,
    {
        '/hosting_old': flask_regist.RegisterOldHostUrl(services),
        '/projects': flask_regist.RegisterRedirectProjectUrl(),
        # '/_': flask_regist.RegisterMONSetUrl(services),
        # '/hosting': flask_regist.RegisterHostingUrl(services),
        # '/p': flask_regist.RegisterProjectUrls(services),
        # '/u': flask_regist.RegisterUserUrls(services),
        # '/_task': flask_regist.RegisterTaskUrl(services),
        # '/_cron': flask_regist.RegisterCronUrl(services),
        # '/_backend': flask_regist.RegisterBackendUrl(services),
    })

endpoints = endpoints_webapp2.api_server(
    [api_svc_v1.MonorailApi, api_svc_v1.ClientConfigApi])

# TODO(crbug/monorail/8221): Remove this code during this milestone.
# It only serves as a safe way to begin connecting to redis without risking
# user facing problems.
try:
  logging.info('Starting initial redis connection verification.')
  from framework import redis_utils
  redis_utils.AsyncVerifyRedisConnection()
except:  # pylint: disable=bare-except
  logging.exception('Exception when instantiating redis connection.')
