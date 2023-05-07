# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Redirect Middleware for Monorail.

Handles traffic redirection before hitting main monorail app.
"""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import flask
from redirect import redirect_utils

class RedirectMiddleware(object):

  def __init__(self, main_app, redirect_app):
    self._main_app = main_app
    self._redirect_app = redirect_app

  def __call__(self, environ, start_response):
    # Run the redirect app first.
    response = flask.Response.from_app(self._redirect_app, environ)
    if response.status_code == 404:
      # If it returns 404, run the main app.
      return self._main_app(environ, start_response)
    # Otherwise, return the response from the redirect app.
    app_iter, status, headers = response.get_wsgi_response(environ)
    start_response(status, headers)
    return app_iter


def GenerateRedirectApp():
  redirect_app = flask.Flask(__name__)

  def PreCheckHandler():
    # Should not redirect away from monorail if param set.
    r = flask.request
    no_redirect = r.args.get('no_tracker_redirect')
    if no_redirect:
      flask.abort(404)
  redirect_app.before_request(PreCheckHandler)

  def IssueList(project_name):
    redirect_url = redirect_utils.GetRedirectURL(project_name)
    if redirect_url:
      return flask.redirect(redirect_url)
    flask.abort(404)
  redirect_app.route('/p/<string:project_name>/issues/')(IssueList)
  redirect_app.route('/p/<string:project_name>/issues/list')(IssueList)

  def IssueDetail(project_name):
    del project_name
    # TODO(crbug.com/monorail/12012): Add issue redirect logic
    flask.abort(404)
  redirect_app.route('/p/<string:project_name>/issues/detail')(IssueDetail)

  def IssueCreate(project_name):
    redirect_url = redirect_utils.GetRedirectURL(project_name)
    if redirect_url:
      return flask.redirect(redirect_url + '/new')
    flask.abort(404)
  redirect_app.route('/p/<string:project_name>/issues/entry')(IssueCreate)
  redirect_app.route('/p/<string:project_name>/issues/entry_new')(IssueCreate)

  return redirect_app
