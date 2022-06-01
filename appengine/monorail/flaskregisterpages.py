# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd
"""This file sets up all the urls for monorail pages."""

import logging
import settings

import flask

from features import banspammer
from features import hotlistcreate
from features import savedqueries
from features import userhotlists
from framework import banned
from framework import excessiveactivity
from project import peopledetail
from project import peoplelist
from project import projectadmin
from project import projectadminadvanced
from project import projectsummary
from project import project_constants
from project import projectupdates
from sitewide import moved
from sitewide import userclearbouncing
from sitewide import userupdates
from sitewide import userprofile
from sitewide import projectcreate
from sitewide import usersettings
from sitewide import groupcreate
from sitewide import grouplist


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

  def _AddFlaskUrlRules(self, flask_instance, rule_tuple):
    """Add url rules to a given Flask instance.

    Args:
      flask_instance: The Flask app to add URLs to.
      rule_tuple: List of tuple of path, module and method to call, HTTP method

    Returns:
      The Flask instance.
    """
    for rule in rule_tuple:
      flask_instance.add_url_rule(rule[0], view_func=rule[1], methods=rule[2])
    return flask_instance


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

    flaskapp_hosting = self._AddFlaskUrlRules(flaskapp_hosting, _HOSTING_URL)

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
        #     '/<string:project_name>/adminIntro',
        #     projectsummary.ProjectSummary(
        #         services=service).GetProjectSummaryPage, ['GET']),
        # (
        #     '/<string:project_name>/admin',
        #     projectadmin.ProjectAdmin(services=service).GetProjectAdminPage,
        #     ['GET']),
        # (
        #     '/<string:project_name>/admin.do',
        #     projectadmin.ProjectAdmin(services=service).PostProjectAdminPage,
        #     ['POST']),
        # (
        #     '/<string:project_name>/adminAdvanced',
        #     projectadminadvanced.ProjectAdminAdvanced(
        #         services=service).GetProjectAdminAdvancedPage, ['GET']),
        # (
        #     '/<string:project_name>/adminAdvanced.do',
        #     projectadminadvanced.ProjectAdminAdvanced(
        #         services=service).PostProjectAdminAdvancedPage, ['POST']),
        # (
        #     '/<string:project_name>/people/list',
        #     peoplelist.PeopleList(services=service).GetPeopleListPage,
        #     ['GET']),
        # (
        #     '/<string:project_name>/people/list.do',
        #     peoplelist.PeopleList(services=service).PostPeopleListPage,
        #     ['POST']),
        # (
        #     '/<string:project_name>/people/detail',
        #     peopledetail.PeopleDetail(services=service).GetPeopleDetailPage,
        #     ['GET']),
        # (
        #     '/<string:project_name>/people/detail.do',
        #     peopledetail.PeopleDetail(services=service).PostPeopleDetailPage,
        #     ['POST']),
        # (
        #     '/<string:project_name>/updates/list',
        #     projectupdates.ProjectUpdates(
        #         services=service).GetProjectUpdatesPage, ['GET']),
    ]
    return self._AddFlaskUrlRules(flaskapp_project, _PROJECT_URLS)


  def RegisterUserUrls(self, service):
    flaskapp_user = flask.Flask(__name__)
    _USER_URLS = [
        # (
        #     '/<string:viewed_username>/queries',
        #     savedqueries.SavedQueries(services=service).GetSavedQueriesPage,
        #     ['GET']),
        # (
        #     '/<string:viewed_username>/queries.do',
        #     savedqueries.SavedQueries(services=service).PostSavedQueriesPage,
        #     ['Post']),
        # (
        #     '/<string:viewed_username>/hotlists',
        #     userhotlists.UserHotlists(services=service).GetUserHotlistsPage,
        #     ['GET']),
        # (
        #     '/<string:viewed_username>/hotlists.do',
        #     userhotlists.UserHotlists(services=service).PostUserHotlistsPage,
        #     ['Post']),
        # (
        #     '/<string:viewed_username>/',
        #     userprofile.UserProfile(services=service).GetUserProfilePage,
        #     ['GET']),
        # (
        #     '/<string:viewed_username>/edit.do',
        #     userprofile.UserProfile(services=service).PostUserProfilePage,
        #     ['POST']),
        # (
        #     '/<string:viewed_username>/ban.do',
        #     userprofile.BanUser(services=service).PostBanUserPage,
        #     ['POST']),
        # (
        #     '/<string:viewed_username>/banSpammer.do',
        #     banspammer.BanSpammer(services=service).PostBanSpammerPage,
        #     ['POST']),
        # (
        #     '/<string:viewed_username>/clearBouncing',
        #     userclearbouncing.UserClearBouncing(
        #         services=service).GetUserClearBouncingPage, ['GET']),
        # (
        #     '/<string:viewed_username>/clearBouncing.do',
        #     userclearbouncing.UserClearBouncing(
        #         services=service).PostUserClearBouncingPage, ['Post']),
        # (
        #     '/<string:viewed_username>/updates/projects',
        #     userupdates.UserUpdatesProjects(
        #         services=service).GetUserUpdatesProjectsPage, ['GET']),
        # (
        #     '/<string:viewed_username>/updates/developers',
        #     userupdates.UserUpdatesDevelopers(
        #         services=service).GetUserUpdatesDevelopersPage, ['GET']),
        # (
        #     '/<string:viewed_username>/updates',
        #     userupdates.UserUpdatesIndividual(
        #         services=service).GetUserUpdatesPage, ['GET']),
    ]

    return self._AddFlaskUrlRules(flaskapp_user, _USER_URLS)


  # pylint: disable=unused-argument
  def RegisterTaskUrl(self, service):
    flaskapp_task = flask.Flask(__name__)
    _TASK_URL = [
        # (
        #     '/banSpammer',
        #     banspammer.BanSpammerTask(services=service).GetBanSpammer,
        #     ['GET']),
        # (
        #     '/banSpammer.do',
        #     banspammer.BanSpammerTask(services=service).PostBanSpammer,
        #     ['POST']),
    ]

    for rule in _TASK_URL:
      flaskapp_task.add_url_rule(rule[0], view_func=rule[1], methods=rule[2])

    return flaskapp_task
