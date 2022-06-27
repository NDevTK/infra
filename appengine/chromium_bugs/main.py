#!/usr/bin/env python
#
# Copyright 2012 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import flask

import gae_ts_mon

application = flask.Flask(__name__)


@application.route('/')
@application.route('/wizard.html')
@application.route('/wizard.do')
def Main():
  new_url = (
      "https://www.google.com/accounts/ServiceLogin?service=ah&"
      "passive=true&continue=https://appengine.google.com/_ah/conflogin%3F"
      "continue=https://bugs.chromium.org/p/chromium/issues/entryafterlogin"
      "&ltmpl=")
  return flask.redirect(new_url, code=302)

# TODO(crbug.com/1322775) Migrate away from the shared prodx-mon-chrome-infra
# service account and change to gae_ts_mon.initialize_prod()
gae_ts_mon.initialize_adhoc(application)
