# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd
"""This file sets up all the urls for monorail pages."""

import logging
import settings

import flask

from features import banspammer
from features import inboundemail
from features import hotlistcreate
from features import savedqueries
from features import userhotlists
from framework import banned
from framework import clientmon
from framework import warmup
from framework import reap
from framework import deleteusers
from framework import excessiveactivity
from framework import ts_mon_js
from framework import trimvisitedpages
from project import peopledetail
from project import peoplelist
from project import projectadmin
from project import projectadminadvanced
from project import projectexport
from project import projectsummary
from project import projectupdates
from project import project_constants
from project import redirects
from services import cachemanager_svc
from services import client_config_svc
from sitewide import custom_404
from sitewide import hostinghome
from sitewide import moved
from sitewide import userclearbouncing
from sitewide import userupdates
from sitewide import userprofile
from sitewide import projectcreate
from sitewide import usersettings
from sitewide import groupadmin
from sitewide import groupcreate
from sitewide import groupdetail
from sitewide import grouplist
from features import rerankhotlist
from features import hotlistdetails
from features import hotlistissues
from features import hotlistissuescsv
from features import hotlistpeople
from features import dateaction
from features import filterrules
from features import pubsub
from features import notify
from features import hotlistcreate
from features import savedqueries
from features import userhotlists
from features import banspammer
from search import backendnonviewable
from search import backendsearch
from tracker import componentcreate
from tracker import fltconversion
from tracker import fieldcreate
from tracker import fielddetail
from tracker import templatecreate
from tracker import templatedetail
from tracker import issueadmin
from tracker import issueadvsearch
from tracker import issueattachment
from tracker import issueattachmenttext
from tracker import issuebulkedit
from tracker import issuedetailezt
from tracker import issueentry
from tracker import issueentryafterlogin
from tracker import issueexport
from tracker import issueoriginal
from tracker import issuereindex
from tracker import issuetips
from tracker import issueimport

from tracker import webcomponentspage


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
  def RegisterGroupUrls(self, services):
    flaskapp_group = flask.Flask(__name__)
    _GROUP_URL = [
        (
            '/', grouplist.FlaskGroupList(services=services).GetGroupList,
            ['GET']),
        (
            '/<string:viewed_username>/',
            groupdetail.GroupDetail(services=services).GetGroupDetail, ['GET']),
        (
            '/<string:viewed_username>/edit.do',
            groupdetail.GroupDetail(services=services).PostGroupDetail,
            ['POST']),
        (
            '/<string:viewed_username>/groupadmin',
            groupadmin.GroupAdmin(services=services).GetGroupAdmin, ['GET']),
        (
            '/<string:viewed_username>/groupadmin.do',
            groupadmin.GroupAdmin(services=services).PostGroupAdmin, ['POST']),
    ]

    return self._AddFlaskUrlRules(flaskapp_group, _GROUP_URL)

  # pylint: disable=unused-argument
  def RegisterHostingUrl(self, service):
    flaskapp_hosting = flask.Flask(__name__)
    _HOSTING_URL = [
        (
            '/excessiveActivity',
            excessiveactivity.ExcessiveActivity(
                services=service).GetExcessiveActivity, ['GET']),
        (
            '/settings',
            usersettings.UserSettings(services=service).GetUserSetting, ['GET'
                                                                        ]),
        (
            '/settings.do',
            usersettings.UserSettings(services=service).PostUserSetting,
            ['POST']),
        ('/noAccess', banned.Banned(services=service).GetNoAccessPage, ['GET']),
        (
            '/moved', moved.ProjectMoved(services=service).GetProjectMoved,
            ['GET']),
        (
            '/createProject',
            projectcreate.ProjectCreate(services=service).GetCreateProject,
            ['GET']),
        (
            '/createProject.do',
            projectcreate.ProjectCreate(services=service).PostCreateProject,
            ['POST']),
        (
            '/createHotlist',
            hotlistcreate.HotlistCreate(services=service).GetCreateHotlist,
            ['GET']),
        (
            '/createHotlist.do',
            hotlistcreate.HotlistCreate(services=service).PostCreateHotlist,
            ['POST']),
        (
            '/createGroup',
            groupcreate.GroupCreate(services=service).GetGroupCreate, ['GET']),
        (
            '/createGroup.do',
            groupcreate.GroupCreate(services=service).PostGroupCreate, ['POST'
                                                                       ]),
        (
            '/deleteGroup',
            grouplist.FlaskGroupList(services=service).GetGroupDelete, ['GET']),
        (
            '/deleteGroup.do',
            grouplist.FlaskGroupList(services=service).PostGroupDelete,
            ['POST']),
    ]

    flaskapp_hosting = self._AddFlaskUrlRules(flaskapp_hosting, _HOSTING_URL)

    # pylint: disable=unused-variable
    # for url /hosting/
    @flaskapp_hosting.route('/')
    def DefaultToMainPage():
      url = flask.request.host_url
      return flask.redirect(url)

    return flaskapp_hosting

  def RegisterOldHostUrl(self, service):
    flaskapp_hosting_old = flask.Flask(__name__)

    # pylint: disable=unused-variable
    @flaskapp_hosting_old.route('/')
    def GetHostingOld():
      return hostinghome.HostingHome(services=service).GetOldHostingHome()

    return flaskapp_hosting_old

  def RegisterRedirectProjectUrl(self):
    flaskapp_project_redirect = flask.Flask(__name__)

    # pylint: disable=unused-variable
    @flaskapp_project_redirect.route('/')
    def GetRedirectProject():
      url = flask.request.host_url
      return flask.redirect(url)

    return flaskapp_project_redirect

  def RegisterProjectUrls(self, service):
    flaskapp_project = flask.Flask(__name__)
    _PROJECT_URLS = [
        # (
        #     '/<string:project_name>/<string:unrecognized>',
        #     custom_404.ErrorPage(services=service).Get404Page,
        #     ['GET'],
        # ),
        # (
        #     '/<string:project_name>/adminComponents',
        #     issueadmin.AdminComponents(
        #         services=service).GetAdminComponentsPage, ['GET']),
        # (
        #     '/<string:project_name>/adminComponents.do',
        #     issueadmin.AdminComponents(
        #         services=service).PostAdminComponentsPage, ['POST']),
        # (
        #     '/<string:project_name>/adminIntro',
        #     projectsummary.ProjectSummary(
        #         services=service).GetProjectSummaryPage, ['GET']),
        # (
        #     '/<string:project_name>/adminLabels',
        #     issueadmin.AdminLabels(services=service).GetAdminLabelsPage,
        #     ['GET']),
        # (
        #     '/<string:project_name>/adminLabels.do',
        #     issueadmin.AdminLabels(services=service).PostAdminLabelsPage,
        #     ['POST']),
        # (
        #     '/<string:project_name>/adminRules',
        #     issueadmin.AdminRules(services=service).GetAdminRulesPage,
        #     ['GET']),
        # (
        #     '/<string:project_name>/adminRules.do',
        #     issueadmin.AdminRules(services=service).PostAdminRulesPage,
        #     ['POST']),
        # (
        #     '/<string:project_name>/adminStatuses',
        #     issueadmin.AdminStatuses(services=service).GetAdminStatusesPage,
        #     ['GET']),
        # (
        #     '/<string:project_name>/adminStatuses.do',
        #     issueadmin.AdminStatuses(services=service).PostAdminStatusesPage,
        #     ['POST']),
        # (
        #     '/<string:project_name>/adminTemplates',
        #     issueadmin.AdminTemplates(services=service).GetAdminTemplatesPage,
        #     ['GET']),
        # (
        #     '/<string:project_name>/adminTemplates.do',
        #     issueadmin.AdminTemplates(
        #         services=service).PostAdminTemplatesPage, ['POST']),
        # (
        #     '/<string:project_name>/adminViews',
        #     issueadmin.AdminViews(services=service).GetAdminViewsPage,
        #     ['GET']),
        # (
        #     '/<string:project_name>/adminViews.do',
        #     issueadmin.AdminViews(services=service).PostAdminViewsPage,
        #     ['POST']),
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
        #     '/<string:project_name>/components/create',
        #     componentcreate.ComponentCreate(
        #         services=service).GetComponentCreatePage, ['GET']),
        # (
        #     '/<string:project_name>/components/create.do',
        #     componentcreate.ComponentCreate(
        #         services=service).PostComponentCreatePage, ['POST']),
        # (
        #     '/<string:project_name>/fields/create',
        #     fieldcreate.FieldCreate(
        #         services=service).GetFieldCreate, ['GET']),
        # (
        #     '/<string:project_name>/fields/create.do',
        #     fieldcreate.FieldCreate(
        #         services=service).PostFieldCreate, ['POST']),
        # (
        #     '/<string:project_name>/fields/detail',
        #     fielddetail.FieldDetail(
        #         services=service).GetFieldDetail, ['GET']),
        # (
        #     '/<string:project_name>/fields/detail.do',
        #     fielddetail.FieldDetail(
        #         services=service).PostFieldDetail, ['POST']),
        # (
        #     '/<string:project_name>/issues/advsearch',
        #     issueadvsearch.IssueAdvancedSearch(
        #         services=service).GetIssueAdvSearchPage, ['GET']),
        # (
        #     '/<string:project_name>/issues/advsearch.do',
        #     issueadvsearch.IssueAdvancedSearch(
        #         services=service).PostIssueAdvSearchPage, ['POST']),
        # (
        #     '/<string:project_name>/issues/detail',
        #     webcomponentspage.WebComponentsPage(
        #         services=service).GetWebComponentsIssueDetail, ['GET']),
        # (
        #     '/<string:project_name>/issues/export',
        #     issueexport.IssueExport(services=service).GetIssueExport,
        #     ['GET']),
        # (
        #     '/<string:project_name>/issues/export/json',
        #     issueexport.IssueExportJSON(services=service).GetIssueExportJSON,
        #     ['GET']),
        # (
        #     '/<string:project_name>/issues/export/json.do',
        #     issueexport.IssueExportJSON(services=service).PostIssueExportJSON,
        #     ['POST']),
        # (
        #     '/<string:project_name>/issues/import',
        #     issueimport.IssueImport(services=service).GetIssueImport,
        #     ['GET']),
        # (
        #     '/<string:project_name>/issues/import.do',
        #     issueimport.IssueImport(services=service).PostIssueImport, ['POST'
        #                                                                ]),
        # (
        #     '/<string:project_name>/issues/original',
        #     issueoriginal.IssueOriginal(services=service).GetIssueOriginal,
        #     ['GET']),
        # (
        #     '/<string:project_name>/issues/entry',
        #     issueentry.IssueEntry(
        #         services=service).GetIssueEntry, ['GET']),
        # (
        #     '/<string:project_name>/issues/entry.do',
        #     issueentry.IssueEntry(
        #         services=service).PostIssueEntry, ['POST']),
        # (
        #     '/<string:project_name>/issues/entry_new',
        #     webcomponentspage.WebComponentsPage(
        #         services=service).GetWebComponentsIssueNewEntry, ['GET']),
        # (
        #     '/<string:project_name>/issues/list',
        #     webcomponentspage.WebComponentsPage(
        #         services=service).GetWebComponentsIssueList, ['GET']),
        # (
        #     '/<string:project_name>/issues/reindex',
        #     issuereindex.IssueReindex(
        #         services=service).GetIssueReindex, ['GET']),
        # (
        #     '/<string:project_name>/issues/reindex.do',
        #     issuereindex.IssueReindex(
        #         services=service).PostIssueReindex, ['POST']),
        # (
        #     '/<string:project_name>/issues/detail/list',
        #     issuedetailezt.FlipperList(
        #         services=service).GetFlipperList, ['GET']),
        # (
        #     '/<string:project_name>/issues/detail/flipper',
        #     issuedetailezt.FlipperIndex(
        #         services=service).GetFlipperIndex, ['GET']),
        # (
        #     '/<string:project_name>/issues/detail/flipper.do',
        #     issuedetailezt.FlipperIndex(
        #         services=service).PostFlipperIndex, ['POST']),
        # (
        #     '/<string:project_name>/issues/wizard',
        #     webcomponentspage.WebComponentsPage(
        #         services=service).GetWebComponentsIssueWizard, ['GET']),
        # (
        #     '/<string:project_name>/templates/create',
        #     templatecreate.TemplateCreate(
        #         services=service).GetTemplateCreate, ['GET']),
        # (
        #     '/<string:project_name>/templates/create.do',
        #     templatecreate.TemplateCreate(
        #         services=service).PostTemplateCreate, ['POST']),
        # (
        #     '/<string:project_name>/templates/detail',
        #     templatedetail.TemplateDetail(
        #         services=service).GetTemplateDetail, ['GET']),
        # (
        #     '/<string:project_name>/templates/detail.do',
        #     templatedetail.TemplateDetail(
        #         services=service).PostTemplateDetail, ['POST']),
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
        #     '/<string:project_name>/projectExport',
        #     projectexport.ProjectExport(
        #         services=service).GetProjectExportPage, ['GET']),
        # (
        #     '/<string:project_name>/projectExport/json',
        #     projectexport.ProjectExportJSON(
        #         services=service).GetProjectExportJSONPage, ['GET']),
        # (
        #     '/<string:project_name>/projectExport/json.do',
        #     projectexport.ProjectExportJSON(
        #         services=service).PostProjectExportJSONPage, ['POST']),
        # (
        #     '/<string:project_name>/updates/list',
        #     projectupdates.ProjectUpdates(
        #         services=service).GetProjectUpdatesPage, ['GET']),
        # (
        #     '/<string:project_name>/w/list',
        #     redirects.WikiRedirect(
        #         services=service).GetWikiListRedirect, ['GET']),
        # (
        #     '/<string:project_name>/wiki/<string:wiki_page>',
        #     redirects.WikiRedirect(
        #         services=service).GetWikiRedirect, ['GET']),
        # (
        #     '/<string:project_name>/source/<string:source_page>',
        #     redirects.SourceRedirect(
        #         services=service).GetSourceRedirect, ['GET']),
        #     '/<string:project_name>/issues/entryafterlogin',
        #     issueentryafterlogin.IssueEntryAfterLogin(
        #         services=service).GetIssueEntryAfterLogin,
        #     ['GET'],
        # ),
        # (
        #     '/<string:project_name>/issues/searchtips',
        #     issuetips.IssueSearchTips(services=service).GetIssueSearchTips,
        #     ['GET'],
        # ),
        # (
        #    '/<string:project_name>/issues/attachment',
        #    issueattachment.AttachmentPage(services=service).GetAttachmentPage,
        #    ['GET'],
        # ),
        # (
        #     '/<string:project_name>/issues/attachmentText',
        #     issueattachmenttext.AttachmentText(
        #         services=service).GetAttachmentText,
        #     ['GET'],
        # ),
        # (
        #     '/<string:project_name>/issues/bulkedit',
        #     issuebulkedit.IssueBulkEdit(
        #         services=service).GetIssueBulkEdit, ['GET']),
        # (
        #     '/<string:project_name>/issues/bulkedit.do',
        #     issuebulkedit.IssueBulkEdit(
        #         services=service).PostIssueBulkEdit, ['POST']),
        # (
        #     '/<string:project_name>/issues/detail/next',
        #     issuedetailezt.FlipperNext(
        #         services=service).GetFlipperNextRedirectPage, ['GET']),
        # (
        #     '/<string:project_name>/issues/detail/previous',
        #     issuedetailezt.FlipperPrev(
        #         services=service).GetFlipperPrevRedirectPage, ['GET']),
    ]
    flaskapp_project = self._AddFlaskUrlRules(flaskapp_project, _PROJECT_URLS)

    # pylint: disable=unused-variable
    @flaskapp_project.route('/<string:project_name>/issues/approval')
    @flaskapp_project.route('/<string:project_name>/issues/detail_ezt')
    def ProjectRedirectToIssueDetail(project_name):
      host_url = flask.request.host_url
      url = host_url + 'p/' + project_name + '/issues/detail'
      query_string = flask.request.query_string
      if query_string:
        url = '%s?%s' % (url, query_string)
      return flask.redirect(url)

    # pylint: disable=unused-variable
    @flaskapp_project.route('/<string:project_name>/issues/list_new')
    @flaskapp_project.route('/<string:project_name>/')
    @flaskapp_project.route('/<string:project_name>/issues/')
    def ProjectRedirectToIssueList(project_name):
      host_url = flask.request.host_url
      url = host_url + 'p/' + project_name + '/issues/list'
      query_string = flask.request.query_string
      if query_string:
        url = '%s?%s' % (url, query_string)
      return flask.redirect(url)

    # pylint: disable=unused-variable
    @flaskapp_project.route('/')
    def ProjectRedirectToMainPage():
      url = flask.request.host_url
      return flask.redirect(url)

    # pylint: disable=unused-variable
    @flaskapp_project.route('/<string:project_name>/people/')
    def ProjectRedirectToPeopleList(project_name):
      host_url = flask.request.host_url
      url = host_url + 'p/' + project_name + '/people/list'
      return flask.redirect(url)

    return flaskapp_project

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
        # (
        #     '/<string:viewed_username>/hotlists/<string:hotlist_id>',
        #    hotlistissues.HotlistIssues(services=service).GetHotlistIssuesPage,
        #     ['GET']),
        # (
        #     '/<string:viewed_username>/hotlists/<string:hotlist_id>.do',
        #   hotlistissues.HotlistIssues(services=service).PostHotlistIssuesPage,
        #     ['POST']),
        # (
        #     '/<string:viewed_username>/hotlists/<string:hotlist_id>/csv',
        #     hotlistissuescsv.HotlistIssuesCsv(
        #         services=service).GetHotlistIssuesCsvPage, ['GET']),
        # (
        #     '/<string:viewed_username>/hotlists/<string:hotlist_id>/people',
        #     hotlistpeople.HotlistPeopleList(
        #         services=service).GetHotlistPeoplePage, ['GET']),
        # (
        #    '/<string:viewed_username>/hotlists/<string:hotlist_id>/people.do',
        #     hotlistpeople.HotlistPeopleList(
        #         services=service).PostHotlistPeoplePage, ['POST']),
        # (
        #     '/<string:viewed_username>/hotlists/<string:hotlist_id>/details',
        #     hotlistdetails.HotlistDetails(
        #         services=service).GetHotlistDetailsPage, ['GET']),
        # (
        #   '/<string:viewed_username>/hotlists/<string:hotlist_id>/details.do',
        #     hotlistdetails.HotlistDetails(
        #         services=service).PostHotlistDetailsPage, ['POST']),
        # (
        #     '/<string:viewed_username>/hotlists/<string:hotlist_id>/rerank',
        #     rerankhotlist.RerankHotlistIssue(
        #         services=service).GetRerankHotlistIssuePage, ['GET']),
        # (
        #   '/<string:viewed_username>/hotlists/<string:hotlist_id>/rerank.do',
        #     rerankhotlist.RerankHotlistIssue(
        #         services=service).PostRerankHotlistIssuePage, ['POST']),
    ]

    flaskapp_user = self._AddFlaskUrlRules(flaskapp_user, _USER_URLS)

    # pylint: disable=unused-variable
    # for url /u/
    @flaskapp_user.route('/')
    def UserRedirectToMainPage():
      url = flask.request.host_url
      return flask.redirect(url)

    return flaskapp_user

  # pylint: disable=unused-argument
  def RegisterTaskUrl(self, service):
    flaskapp_task = flask.Flask(__name__)
    _TASK_URL = [
        (
            '/banSpammer.do',
            banspammer.BanSpammerTask(services=service).PostBanSpammer,
            ['POST']),
        (
            '/sendWipeoutUserListsTask.do',
            deleteusers.SendWipeoutUserListsTask(
                services=service).PostSendWipeoutUserListsTask, ['POST']),
        (
            '/deleteWipeoutUsersTask.do',
            deleteusers.DeleteWipeoutUsersTask(
                services=service).PostDeleteWipeoutUsersTask, ['POST']),
        (
            '/deleteUsersTask.do',
            deleteusers.DeleteUsersTask(services=service).PostDeleteUsersTask,
            ['POST']),
        (
            '/notifyRulesDeleted.do',
            notify.NotifyRulesDeletedTask(
                services=service).PostNotifyRulesDeletedTask, ['POST']),
        (
            '/notifyIssueChange.do',
            notify.NotifyIssueChangeTask(
                services=service).PostNotifyIssueChangeTask, ['POST']),
        (
            '/notifyBlockingChange.do',
            notify.NotifyBlockingChangeTask(
                services=service).PostNotifyBlockingChangeTask, ['POST']),
        (
            '/notifyBulkEdit.do', notify.NotifyBulkChangeTask(
                services=service).PostNotifyBulkChangeTask, ['POST']),
        (
            '/notifyApprovalChange.do',
            notify.NotifyApprovalChangeTask(
                services=service).PostNotifyApprovalChangeTask, ['POST']),
        (
            '/publishPubsubIssueChange.do',
            pubsub.PublishPubsubIssueChangeTask(
                services=service).PostPublishPubsubIssueChangeTask, ['POST']),
        (
            '/issueDateAction.do',
            dateaction.IssueDateActionTask(
                services=service).PostIssueDateActionTask, ['POST']),
        (
            '/fltConversionTask.do',
            fltconversion.FLTConvertTask(services=service).PostFLTConvertTask,
            ['POST']),
        (
            '/outboundEmail.do',
            notify.OutboundEmailTask(services=service).PostOutboundEmailTask,
            ['POST']),
        (
            '/recomputeDerivedFields.do',
            filterrules.RecomputeDerivedFieldsTask(
                services=service).PostRecomputeDerivedFieldsTask, ['POST']),
    ]

    for rule in _TASK_URL:
      flaskapp_task.add_url_rule(rule[0], view_func=rule[1], methods=rule[2])

    return flaskapp_task

  # pylint: disable=unused-argument
  def RegisterCronUrl(self, service):
    flaskapp_cron = flask.Flask(__name__)
    _CRON_URL = [
        (
            '/wipeoutSync',
            deleteusers.WipeoutSyncCron(services=service).GetWipeoutSyncCron,
            ['GET']),
        (
            '/reindexQueue',
            filterrules.ReindexQueueCron(services=service).GetReindexQueueCron,
            ['GET']),
        (
            '/dateAction',
            dateaction.DateActionCron(services=service).GetDateActionCron,
            ['GET']),
        (
            '/ramCacheConsolidate',
            cachemanager_svc.RamCacheConsolidate(
                services=service).GetRamCacheConsolidate, ['GET']),
        ('/reap', reap.Reap(services=service).GetReap, ['GET']),
        (
            '/loadApiClientConfigs', client_config_svc.GetLoadApiClientConfigs,
            ['GET']),
        (
            '/trimVisitedPages',
            trimvisitedpages.TrimVisitedPages(
                services=service).GetTrimVisitedPages, ['GET']),
    ]

    for rule in _CRON_URL:
      flaskapp_cron.add_url_rule(rule[0], view_func=rule[1], methods=rule[2])

    return flaskapp_cron

  # pylint: disable=unused-argument
  def RegisterBackendUrl(self, service):
    flaskapp_backend = flask.Flask(__name__)
    _BACKEND_URL = [
        (
            '/search',
            backendsearch.BackendSearch(services=service).GetBackendSearch,
            ['GET']),
        (
            '/nonviewable',
            backendnonviewable.BackendNonviewable(
                services=service).GetBackendNonviewable, ['GET']),
    ]

    for rule in _BACKEND_URL:
      flaskapp_backend.add_url_rule(rule[0], view_func=rule[1], methods=rule[2])

    return flaskapp_backend

  # pylint: disable=unused-argument
  def RegisterMONSetUrl(self, service):
    flaskapp_mon = flask.Flask(__name__)
    _MON_URL = [
        (
            '/clientmon.do',
            clientmon.ClientMonitor(services=service).PostClientMonitor,
            ['POST']),
        (
            '/jstsmon.do',
            ts_mon_js.FlaskMonorailTSMonJSHandler(
                services=service).PostMonorailTSMonJSHandler,
            ['POST'],
        )
    ]

    flaskapp_mon = self._AddFlaskUrlRules(flaskapp_mon, _MON_URL)
    return flaskapp_mon

  def RegisterAHUrl(self, service):
    flaskapp_ah = flask.Flask(__name__)
    _AH_URL = [
        ('/warmup', warmup.Warmup, ['GET']), ('/start', warmup.Start, ['GET']),
        ('/stop', warmup.Stop, ['GET']),
        (
            '/bounce',
            inboundemail.BouncedEmail(services=service).postBouncedEmail,
            ['POST']),
        (
            '/mail/<string:project_addr>',
            inboundemail.InboundEmail(services=service).HandleInboundEmail,
            ['GET', 'POST'])
    ]

    flaskapp_ah = self._AddFlaskUrlRules(flaskapp_ah, _AH_URL)

    return flaskapp_ah
