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

from components import prpc

from features import autolink
from features import dateaction
from features import banspammer
from features import hotlistcreate
from features import hotlistdetails
from features import hotlistissues
from features import hotlistissuescsv
from features import hotlistpeople
from features import filterrules
from features import pubsub
from features import userhotlists
from features import inboundemail
from features import notify
from features import rerankhotlist
from features import savedqueries
from features import spammodel
from features import spamtraining
from features import componentexport

from framework import banned, flaskservlet
from framework import clientmon
from framework import csp_report
from framework import deleteusers
from framework import excessiveactivity
from framework import trimvisitedpages
from framework import framework_bizobj
from framework import reap
from framework import registerpages_helpers
from framework import ts_mon_js
from framework import urls
from framework import warmup

from project import peopledetail
from project import peoplelist
from project import project_constants
from project import projectadmin
from project import projectadminadvanced
from project import projectexport
from project import projectsummary
from project import projectupdates
from project import redirects

from search import backendnonviewable
from search import backendsearch

from services import cachemanager_svc
from services import client_config_svc

from sitewide import custom_404
from sitewide import groupadmin
from sitewide import groupcreate
from sitewide import groupdetail
from sitewide import grouplist
from sitewide import hostinghome
from sitewide import moved
from sitewide import projectcreate
from sitewide import userprofile
from sitewide import usersettings
from sitewide import userclearbouncing
from sitewide import userupdates

from tracker import componentcreate
from tracker import componentdetail
from tracker import fieldcreate
from tracker import fielddetail
from tracker import issueadmin
from tracker import issueadvsearch
from tracker import issueattachment
from tracker import issueattachmenttext
from tracker import issuebulkedit
from tracker import webcomponentspage
from tracker import issuedetailezt
from tracker import issueentry
from tracker import issueentryafterlogin
from tracker import issueexport
from tracker import issueimport
from tracker import issueoriginal
from tracker import issuereindex
from tracker import issuetips
from tracker import spam
from tracker import templatecreate
from tracker import templatedetail
from tracker import fltconversion

from api import api_routes as api_routes_v0
from api.v3 import api_routes as api_routes_v3


class FlaskServletRegistry(object):

  _PROJECT_NAME_REGEX = project_constants.PROJECT_NAME_PATTERN
  _USERNAME_REGEX = r'[-+\w=.%]+(@([-a-z0-9]+\.)*[a-z0-9]+)?'
  _HOTLIST_ID_NAME_REGEX = r'\d+|[a-zA-Z][-0-9a-zA-Z\.]*'

  def __init__(self):
    self.routes = []

  def _AddRoute(self, path_regex, func, does_write=False):
    if settings.read_only and does_write:
      logging.info('Not registring %r because site is read-only', path_regex)
      # TODO(jrobbins): register a helpful error page instead.
    else:
      self.routes.append([path_regex, func])

  def _SetupServlets(self, spec_dict, base=''):
    """Register each of the given servlets."""
    for get_uri, func in spec_dict.items():
      self._AddRoute(base + get_uri, func)

  def _SetupProjectServlets(self, spec_dict):
    """Register each of the given servlets in the project URI space."""
    self._SetupServlets(
        spec_dict, base='/p/<string:project_name>')

  def Register(self, services):
    self._RegisterProjectHandlers(services)

    return self.routes

  def _RegisterProjectHandlers(self, services):
    """Register page and form handlers that operate within a project."""
    self._SetupProjectServlets({'/test': flaskservlet.FlaskServlet(services).get})
