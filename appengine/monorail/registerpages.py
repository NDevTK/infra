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
from features import hotlistdetails
from features import hotlistissues
from features import hotlistissuescsv
from features import hotlistpeople
from features import filterrules
from features import pubsub
from features import userhotlists
from features import notify
from features import rerankhotlist
from features import savedqueries

from framework import csp_report
from framework import deleteusers
from framework import trimvisitedpages
from framework import reap
from framework import registerpages_helpers
from framework import urls

from project import peopledetail
from project import peoplelist
from project import project_constants
from project import projectadmin
from project import projectadminadvanced
from project import projectexport
from project import projectsummary
from project import projectupdates
from project import redirects

from services import cachemanager_svc
from services import client_config_svc

from sitewide import custom_404
from sitewide import userprofile
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
from tracker import templatecreate
from tracker import templatedetail
from tracker import fltconversion

from api import api_routes as api_routes_v0
from api.v3 import api_routes as api_routes_v3


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

  def _SetupProjectServlets(self, spec_dict, post_does_write=True):
    """Register each of the given servlets in the project URI space."""
    self._SetupServlets(
        spec_dict, base='/p/<project_name:%s>' % self._PROJECT_NAME_REGEX,
        post_does_write=post_does_write)

  def _SetupUserServlets(self, spec_dict, post_does_write=True):
    """Register each of the given servlets in the user URI space."""
    self._SetupServlets(
        spec_dict, base='/u/<viewed_username:%s>' % self._USERNAME_REGEX,
        post_does_write=post_does_write)

  def _SetupGroupServlets(self, spec_dict, post_does_write=True):
    """Register each of the given servlets in the user group URI space."""
    self._SetupServlets(
        spec_dict, base='/g/<viewed_username:%s>' % self._USERNAME_REGEX,
        post_does_write=post_does_write)

  def _SetupUserHotlistServlets(self, spec_dict, post_does_write=True):
    """ Register given user hotlist servlets in the user URI space."""
    self._SetupServlets(
        spec_dict,
        base ='/u/<viewed_username:%s>/hotlists/<hotlist_id:%s>'
        % (self._USERNAME_REGEX, self._HOTLIST_ID_NAME_REGEX),
        post_does_write=post_does_write)

  def Register(self, services):
    """Register all the monorail request handlers."""
    self._RegisterSitewideHandlers()
    self._RegisterProjectHandlers()
    self._RegisterIssueHandlers()
    self._RegisterWebComponentsHanders()
    self._RegisterRedirects()

    # Register pRPC API routes
    prpc_server = prpc.Server(
        allowed_origins=client_config_svc.GetAllowedOriginsSet())
    api_routes_v0.RegisterApiHandlers(prpc_server, services)
    api_routes_v3.RegisterApiHandlers(prpc_server, services)
    self.routes.extend(prpc_server.get_routes())

    autolink.RegisterAutolink(services)
    # Error pages should be the last to register.
    self._RegisterErrorPages()
    return self.routes

  def _RegisterProjectHandlers(self):
    """Register page and form handlers that operate within a project."""

    self._SetupProjectServlets(
        {
            urls.ADMIN_INTRO: projectsummary.ProjectSummary,
            urls.PEOPLE_LIST: peoplelist.PeopleList,
            urls.PEOPLE_DETAIL: peopledetail.PeopleDetail,
            urls.UPDATES_LIST: projectupdates.ProjectUpdates,
            urls.ADMIN_META: projectadmin.ProjectAdmin,
            urls.ADMIN_ADVANCED: projectadminadvanced.ProjectAdminAdvanced,
            urls.ADMIN_EXPORT: projectexport.ProjectExport,
            urls.ADMIN_EXPORT_JSON: projectexport.ProjectExportJSON,
        })

  def _RegisterIssueHandlers(self):
    """Register page and form handlers for the issue tracker."""

    self._SetupProjectServlets(
        {
            urls.ISSUE_APPROVAL:
                registerpages_helpers.MakeRedirectInScope(
                    urls.ISSUE_DETAIL, 'p', keep_qs=True),
            urls.ISSUE_LIST:
                webcomponentspage.WebComponentsPage,
            urls.ISSUE_LIST_NEW_TEMP:
                registerpages_helpers.MakeRedirectInScope(
                    urls.ISSUE_LIST, 'p', keep_qs=True),
            urls.ISSUE_REINDEX:
                issuereindex.IssueReindex,
            urls.ISSUE_DETAIL_FLIPPER_NEXT:
                issuedetailezt.FlipperNext,
            urls.ISSUE_DETAIL_FLIPPER_PREV:
                issuedetailezt.FlipperPrev,
            urls.ISSUE_DETAIL_FLIPPER_LIST:
                issuedetailezt.FlipperList,
            urls.ISSUE_DETAIL_FLIPPER_INDEX:
                issuedetailezt.FlipperIndex,
            urls.ISSUE_DETAIL_LEGACY:
                registerpages_helpers.MakeRedirectInScope(
                    urls.ISSUE_DETAIL, 'p', keep_qs=True),
            urls.ISSUE_WIZARD:
                webcomponentspage.WebComponentsPage,
            urls.ISSUE_ENTRY:
                issueentry.IssueEntry,
            urls.ISSUE_ENTRY_NEW:
                webcomponentspage.WebComponentsPage,
            urls.ISSUE_ENTRY_AFTER_LOGIN:
                issueentryafterlogin.IssueEntryAfterLogin,
            urls.ISSUE_TIPS:
                issuetips.IssueSearchTips,
            urls.ISSUE_ATTACHMENT:
                issueattachment.AttachmentPage,
            urls.ISSUE_ATTACHMENT_TEXT:
                issueattachmenttext.AttachmentText,
            urls.ISSUE_BULK_EDIT:
                issuebulkedit.IssueBulkEdit,
            urls.COMPONENT_CREATE:
                componentcreate.ComponentCreate,
            urls.COMPONENT_DETAIL:
                componentdetail.ComponentDetail,
            urls.FIELD_CREATE:
                fieldcreate.FieldCreate,
            urls.FIELD_DETAIL:
                fielddetail.FieldDetail,
            urls.TEMPLATE_CREATE:
                templatecreate.TemplateCreate,
            urls.TEMPLATE_DETAIL:
                templatedetail.TemplateDetail,
            urls.WIKI_LIST:
                redirects.WikiRedirect,
            urls.WIKI_PAGE:
                redirects.WikiRedirect,
            urls.SOURCE_PAGE:
                redirects.SourceRedirect,
            urls.ADMIN_STATUSES:
                issueadmin.AdminStatuses,
            urls.ADMIN_LABELS:
                issueadmin.AdminLabels,
            urls.ADMIN_RULES:
                issueadmin.AdminRules,
            urls.ADMIN_TEMPLATES:
                issueadmin.AdminTemplates,
            urls.ADMIN_COMPONENTS:
                issueadmin.AdminComponents,
            urls.ADMIN_VIEWS:
                issueadmin.AdminViews,
            urls.ISSUE_ORIGINAL:
                issueoriginal.IssueOriginal,
            urls.ISSUE_EXPORT:
                issueexport.IssueExport,
            urls.ISSUE_EXPORT_JSON:
                issueexport.IssueExportJSON,
            urls.ISSUE_IMPORT:
                issueimport.IssueImport,
        })

    # GETs for /issues/detail are now handled by the web components page.
    base = '/p/<project_name:%s>' % self._PROJECT_NAME_REGEX
    self._AddRoute(base + urls.ISSUE_DETAIL,
                   webcomponentspage.WebComponentsPage, 'GET')

    self._SetupUserServlets({
        urls.SAVED_QUERIES: savedqueries.SavedQueries,
        urls.HOTLISTS: userhotlists.UserHotlists,
        })

    user_hotlists_redir = registerpages_helpers.MakeRedirectInScope(
        urls.HOTLISTS, 'u', keep_qs=True)
    self._SetupUserServlets({
        '/hotlists/': user_hotlists_redir,
        })

    # These servlets accept POST, but never write to the database, so they can
    # still be used when the site is read-only.
    self._SetupProjectServlets({
        urls.ISSUE_ADVSEARCH: issueadvsearch.IssueAdvancedSearch,
        }, post_does_write=False)

    list_redir = registerpages_helpers.MakeRedirectInScope(
        urls.ISSUE_LIST, 'p', keep_qs=True)
    self._SetupProjectServlets({
        '': list_redir,
        '/': list_redir,
        '/issues': list_redir,
        '/issues/': list_redir,
        })

  def _RegisterSitewideHandlers(self):
    """Register page and form handlers that aren't associated with projects."""

    self._SetupUserServlets({
        urls.USER_PROFILE: userprofile.UserProfile,
        urls.BAN_USER: userprofile.BanUser,
        urls.BAN_SPAMMER: banspammer.BanSpammer,
        urls.USER_CLEAR_BOUNCING: userclearbouncing.UserClearBouncing,
        urls.USER_UPDATES_PROJECTS: userupdates.UserUpdatesProjects,
        urls.USER_UPDATES_DEVELOPERS: userupdates.UserUpdatesDevelopers,
        urls.USER_UPDATES_MINE: userupdates.UserUpdatesIndividual,
        })

    self._SetupUserHotlistServlets(
        {
            urls.HOTLIST_ISSUES: hotlistissues.HotlistIssues,
            urls.HOTLIST_ISSUES_CSV: hotlistissuescsv.HotlistIssuesCsv,
            urls.HOTLIST_PEOPLE: hotlistpeople.HotlistPeopleList,
            urls.HOTLIST_DETAIL: hotlistdetails.HotlistDetails,
            urls.HOTLIST_RERANK_JSON: rerankhotlist.RerankHotlistIssue,
        })

    profile_redir = registerpages_helpers.MakeRedirectInScope(
        urls.USER_PROFILE, 'u')
    self._SetupUserServlets({'': profile_redir})

  def _RegisterWebComponentsHanders(self):
    """Register page handlers that are handled by WebComponentsPage."""
    self._AddRoute('/', webcomponentspage.ProjectListPage, 'GET')
    self._AddRoute(
        '/hotlists<unused:.*>', webcomponentspage.WebComponentsPage, 'GET')
    self._AddRoute('/users<unused:.*>', webcomponentspage.WebComponentsPage,
                   'GET')

  def _RegisterRedirects(self):
    """Register redirects among pages inside monorail."""
    redirect = registerpages_helpers.MakeRedirect('/')
    self._SetupServlets(
        {
            '/p': redirect,
            '/p/': redirect,
            '/u': redirect,
            '/u/': redirect,
            '/': redirect,
        })

    redirect = registerpages_helpers.MakeRedirectInScope(
        urls.PEOPLE_LIST, 'p')
    self._SetupProjectServlets({
        '/people': redirect,
        '/people/': redirect,
        })

  def _RegisterErrorPages(self):
    """Register handlers for errors."""
    self._AddRoute(
        '/p/<project_name:%s>/<unrecognized:.+>' % self._PROJECT_NAME_REGEX,
        custom_404.ErrorPage, 'GET')
