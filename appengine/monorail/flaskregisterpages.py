# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd
"""This file sets up all the urls for monorail pages."""

import logging
import settings

import flask

from framework import banned
from framework import excessiveactivity
from project import peoplelist
from project import project_constants
from sitewide import moved
from sitewide import projectcreate
from sitewide import usersettings
from sitewide import groupcreate
from sitewide import grouplist
from features import hotlistcreate


class ServletRegistry(object):

  _PROJECT_NAME_REGEX = project_constants.PROJECT_NAME_PATTERN
  _USERNAME_REGEX = r'[-+\w=.%]+(@([-a-z0-9]+\.)*[a-z0-9]+)?'
  _HOTLIST_ID_NAME_REGEX = r'\d+|[a-zA-Z][-0-9a-zA-Z\.]*'

  def __init__(self):
    self.routes = []

  def _AddRoute(
      self, path_regex, servlet_handler, method='GET', does_write=False):
    """Add a GET or POST handler to our flask route list.

    Args:
      path_regex: string with flask URL template regex.
      servlet_handler: a servlet handler function.
      method: string 'GET' or 'POST'.
      does_write: True if the servlet could write to the database, we skip
          registering such servlets when the site is in read_only mode. GET
          handlers never write. Most, but not all, POST handlers do write.
    """
    if settings.read_only and does_write:
      logging.info('Not registring %r because site is read-only', path_regex)
    else:
      self.routes.append([path_regex, servlet_handler, [method]])

  def _SetupServlets(self, spec_dict, base='', post_does_write=True):
    """Register each of the given servlets."""
    for get_uri, servlet_handler in spec_dict.items():
      self._AddRoute(base + get_uri, servlet_handler, 'GET')
      post_uri = get_uri + ('edit.do' if get_uri.endswith('/') else '.do')
      self._AddRoute(
          base + post_uri, servlet_handler, 'POST', does_write=post_does_write)

  def Register(self):
    """Register all the monorail request handlers."""
    return self.routes

  # pylint: disable=unused-argument
  def RegisterHostingUrl(self, service):
    flaskapp_hosting = flask.Flask(__name__)
    _HOSTING_URL = [
        # (
        #     '/excessiveActivity',
        #     excessiveactivity.ExcessiveActivity(
        #         services=service).GetExcessiveActivity, ['GET']),
        # (
        #     '/settings',
        #     usersettings.UserSettings(services=service).GetUserSetting, ['GET'
        #                                                                 ]),
        # (
        #     '/settings.do',
        #     usersettings.UserSettings(services=service).PostUserSetting,
        #     ['POST']),
        # ('/noAccess', banned.Banned(services=service).GetNoAccessPage,
        #   ['GET']),
        # (
        #     '/moved', moved.ProjectMoved(services=service).GetProjectMoved,
        #     ['GET']),
        # (
        #     '/createProject',
        #     projectcreate.ProjectCreate(services=service).GetCreateProject,
        #     ['GET']),
        # (
        #     '/createProject.do',
        #     projectcreate.ProjectCreate(services=service).PostCreateProject,
        #     ['POST']),
        # (
        #     '/createHotlist',
        #     hotlistcreate.HotlistCreate(services=service).GetCreateHotlist,
        #     ['GET']),
        # (
        #     '/createHotlist.do',
        #     hotlistcreate.HotlistCreate(services=service).PostCreateHotlist,
        #     ['POST']),
        # (
        #     '/createGroup',
        #     groupcreate.GroupCreate(services=service).GetGroupCreate,
        #     ['GET']),
        # (
        #     '/createGroup.do',
        #     groupcreate.GroupCreate(services=service).PostGroupCreate,
        #     ['POST']),
        # (
        #     '/deleteGroup',
        #     grouplist.GroupDelete(services=service).GetGroupDelete,
        #     ['GET']),
        # (
        #     '/deleteGroup.do',
        #     grouplist.GroupDelete(services=service).PostGroupDelete,
        #     ['POST']),
    ]

    for rule in _HOSTING_URL:
      flaskapp_hosting.add_url_rule(rule[0], view_func=rule[1], methods=rule[2])

    # pylint: disable=unused-variable
    @flaskapp_hosting.route('/')
    def DefaultToMainPage():
      url = flask.request.host_url
      return flask.redirect(url)

    return flaskapp_hosting

  def RegisterProjectUrls(self, service):
    flaskapp_project = flask.Flask(__name__)
    _PROJECT_URLS = [
        # (
        #     '/people/list',
        #     peoplelist.PeopleList(services=service).GetPeopleListPage,
        #     ['GET']),
        # (
        #     '/people/list.do',
        #     peoplelist.PeopleList(services=service).PostPeopleListPage,
        #     ['Post'])
    ]

    for rule in _PROJECT_URLS:
      flaskapp_project.add_url_rule(rule[0], view_func=rule[1], methods=rule[2])
    return flaskapp_project
