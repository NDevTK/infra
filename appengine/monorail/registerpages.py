# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""This file sets up all the urls for monorail pages."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import logging
import webapp2
import settings

from features import autolink

from framework import csp_report
from framework import urls

from project import project_constants


from tracker import webcomponentspage


class ServletRegistry(object):

  _PROJECT_NAME_REGEX = project_constants.PROJECT_NAME_PATTERN
  _USERNAME_REGEX = r'[-+\w=.%]+(@([-a-z0-9]+\.)*[a-z0-9]+)?'
  _HOTLIST_ID_NAME_REGEX = r'\d+|[a-zA-Z][-0-9a-zA-Z\.]*'

  def __init__(self):
    self.routes = []

  def _AddRoute(self, path_regex, servlet_class, method, does_write=False):
    """Add a GET or POST handler to our webapp2 route list.

    Args:
      path_regex: string with webapp2 URL template regex.
      servlet_class: a subclass of class Servlet.
      method: string 'GET' or 'POST'.
      does_write: True if the servlet could write to the database, we skip
          registering such servlets when the site is in read_only mode. GET
          handlers never write. Most, but not all, POST handlers do write.
    """
    if settings.read_only and does_write:
      logging.info('Not registring %r because site is read-only', path_regex)
      # TODO(jrobbins): register a helpful error page instead.
    else:
      self.routes.append(
          webapp2.Route(path_regex, handler=servlet_class, methods=[method]))

  def _SetupServlets(self, spec_dict, base='', post_does_write=True):
    """Register each of the given servlets."""
    for get_uri, servlet_class in spec_dict.items():
      self._AddRoute(base + get_uri, servlet_class, 'GET')
      post_uri = get_uri + ('edit.do' if get_uri.endswith('/') else '.do')
      self._AddRoute(base + post_uri, servlet_class, 'POST',
                     does_write=post_does_write)

  def Register(self, services):
    """Register all the monorail request handlers."""
    self._RegisterWebComponentsHanders()
    self._RegisterRedirects()

    autolink.RegisterAutolink(services)
    return self.routes

  def _RegisterFrameworkHandlers(self):
    """Register page and form handlers for framework functionality."""
    self._SetupServlets(
        {
            urls.CSP_REPORT:
                csp_report.CSPReportPage,
        })

  def _RegisterWebComponentsHanders(self):
    """Register page handlers that are handled by WebComponentsPage."""
    self._AddRoute('/', webcomponentspage.ProjectListPage, 'GET')
    self._AddRoute(
        '/hotlists<unused:.*>', webcomponentspage.WebComponentsPage, 'GET')
    self._AddRoute('/users<unused:.*>', webcomponentspage.WebComponentsPage,
                   'GET')
