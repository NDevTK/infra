# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Base classes for Monorail servlets.

This base class provides HTTP get() and post() methods that
conveniently drive the process of parsing the request, checking base
permissions, gathering common page information, gathering
page-specific information, and adding on-page debugging information
(when appropriate).  Subclasses can simply implement the page-specific
logic.

Summary of page classes:
  Servlet: abstract base class for all Monorail servlets.
  _ContextDebugItem: displays page_data elements for on-page debugging.
"""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import gc
import httplib
import json
import logging
import os
import random
import time
import urllib

import ezt
from third_party import httpagentparser

from google.appengine.api import app_identity
from google.appengine.api import modules
from google.appengine.api import users
from oauth2client.client import GoogleCredentials

import webapp2
import flask

import settings
from businesslogic import work_env
from features import savedqueries_helpers
from features import features_bizobj
from features import hotlist_views
from framework import alerts
from framework import exceptions
from framework import framework_constants
from framework import framework_helpers
from framework import framework_views
from framework import monorailrequest
from framework import permissions
from framework import ratelimiter
from framework import servlet_helpers
from framework import template_helpers
from framework import urls
from framework import xsrf
from project import project_constants
from proto import project_pb2
from search import query2ast
from tracker import tracker_views

from infra_libs import ts_mon

NONCE_LENGTH = 32

if not settings.unit_test_mode:
  import MySQLdb

GC_COUNT = ts_mon.NonCumulativeDistributionMetric(
    'monorail/servlet/gc_count',
    'Count of objects in each generation tracked by the GC',
    [ts_mon.IntegerField('generation')])

GC_EVENT_REQUEST = ts_mon.CounterMetric(
    'monorail/servlet/gc_event_request',
    'Counts of requests that triggered at least one GC event',
    [])

# TODO(crbug/monorail:7084): Find a better home for this code.
trace_service = None
# TOD0(crbug/monorail:7082): Re-enable this once we have a solution that doesn't
# inur clatency, or when we're actively using Cloud Tracing data.
# if app_identity.get_application_id() != 'testing-app':
#   logging.warning('app id: %s', app_identity.get_application_id())
#   try:
#     credentials = GoogleCredentials.get_application_default()
#     trace_service = discovery.build(
#         'cloudtrace', 'v1', credentials=credentials)
#   except Exception as e:
#     logging.warning('could not get trace service: %s', e)


class MethodNotSupportedError(NotImplementedError):
  """An exception class for indicating that the method is not supported.

  Used by GatherPageData and ProcessFormData to indicate that GET and POST,
  respectively, are not supported methods on the given Servlet.
  """
  pass


class FlaskServlet(object):
  """Base class for all Monorail flask servlets.

  Defines a framework of methods that build up parts of the EZT page data.

  Subclasses should override GatherPageData and/or ProcessFormData to
  handle requests.
  """

  _MAIN_TAB_MODE = None  # Normally overriden in subclasses to be one of these:

  MAIN_TAB_NONE = 't0'
  MAIN_TAB_DASHBOARD = 't1'
  MAIN_TAB_ISSUES = 't2'
  MAIN_TAB_PEOPLE = 't3'
  MAIN_TAB_PROCESS = 't4'
  MAIN_TAB_UPDATES = 't5'
  MAIN_TAB_ADMIN = 't6'
  MAIN_TAB_DETAILS = 't7'
  PROCESS_TAB_SUMMARY = 'st1'
  PROCESS_TAB_STATUSES = 'st3'
  PROCESS_TAB_LABELS = 'st4'
  PROCESS_TAB_RULES = 'st5'
  PROCESS_TAB_TEMPLATES = 'st6'
  PROCESS_TAB_COMPONENTS = 'st7'
  PROCESS_TAB_VIEWS = 'st8'
  ADMIN_TAB_META = 'st1'
  ADMIN_TAB_ADVANCED = 'st9'
  HOTLIST_TAB_ISSUES = 'ht2'
  HOTLIST_TAB_PEOPLE = 'ht3'
  HOTLIST_TAB_DETAILS = 'ht4'

  # Most forms require a security token, however if a form is really
  # just redirecting to a search GET request without writing any data,
  # subclass can override this to allow anonymous use.
  CHECK_SECURITY_TOKEN = True

  # Some pages might be posted to by clients outside of Monorail.
  # ie: The issue entry page, by the issue filing wizard. In these cases,
  # we can allow an xhr-scoped XSRF token to be used to post to the page.
  ALLOW_XHR = False

  # Most forms just ignore fields that have value "".  Subclasses can override
  # if needed.
  KEEP_BLANK_FORM_VALUES = False

  # Most forms use regular forms, but subclasses that accept attached files can
  # override this to be True.
  MULTIPART_POST_BODY = False

  # This value should not typically be overridden.
  _TEMPLATE_PATH = framework_constants.TEMPLATE_PATH

  _PAGE_TEMPLATE = None  # Normally overriden in subclasses.
  _ELIMINATE_BLANK_LINES = False

  _MISSING_PERMISSIONS_TEMPLATE = 'sitewide/403-page.ezt'

  def __init__(self, services=None,
               content_type='text/html; charset=UTF-8'):
    """Load and parse the template, saving it for later use."""
    if self._PAGE_TEMPLATE:  # specified in subclasses
      template_path = self._TEMPLATE_PATH + self._PAGE_TEMPLATE
      self.template = template_helpers.GetTemplate(
          template_path, eliminate_blank_lines=self._ELIMINATE_BLANK_LINES)
    else:
      self.template = None

    self._missing_permissions_template = template_helpers.MonorailTemplate(
        self._TEMPLATE_PATH + self._MISSING_PERMISSIONS_TEMPLATE)
    self.services = services
    self.content_type = content_type
    self.mr = None
    self.request = flask.request
    self.ratelimiter = ratelimiter.RateLimiter()

  def dispatch(self):
    """Do common stuff then dispatch the request to get() or put() methods."""
    handler_start_time = time.time()

    self.mr = monorailrequest.MonorailRequest(self.services)
    logging.info('???')
    logging.info(repr(self.request.values))
    if 'X-Cloud-Trace-Context' in self.request.headers:
      self.mr.profiler.trace_context = (
          self.request.headers.get('X-Cloud-Trace-Context'))
    # TOD0(crbug/monorail:7082): Re-enable tracing.
    # if trace_service is not None:
    #   self.mr.profiler.trace_service = trace_service

    if self.services.cache_manager:
      # TODO(jrobbins): don't do this step if invalidation_timestep was
      # passed via the request and matches our last timestep
      try:
        with self.mr.profiler.Phase('distributed invalidation'):
          self.services.cache_manager.DoDistributedInvalidation(self.mr.cnxn)

      except MySQLdb.OperationalError as e:
        logging.exception(e)
        page_data = {
          'http_response_code': httplib.SERVICE_UNAVAILABLE,
          'requested_url': self.request.url,
        }
        self.template = template_helpers.GetTemplate(
            'templates/framework/database-maintenance.ezt',
            eliminate_blank_lines=self._ELIMINATE_BLANK_LINES)
        self.template.WriteResponse(
          flask.response, page_data, content_type='text/html')
        return

    # try:
    #   self.ratelimiter.CheckStart(self.request)

    #   with self.mr.profiler.Phase('parsing request and doing lookups'):
    #     self.mr.ParseRequest(self.request, self.services)

    # except exceptions as e:
    #   logging.warning('Trapped NoSuchUserException %s', e)
    #   flask.abort(404, 'user not found')
    #   return
    # except exceptions.NoSuchGroupException as e:
    #   logging.warning('Trapped NoSuchGroupException %s', e)
    #   flask.abort(404, 'user group not found')

    # except exceptions.InputException as e:
    #   logging.info('Rejecting invalid input: %r', e)
    #   flask.response.status = httplib.BAD_REQUEST

    # except exceptions.NoSuchProjectException as e:
    #   logging.info('Rejecting invalid request: %r', e)
    #   self.response.status = httplib.NOT_FOUND

    # except xsrf.TokenIncorrect as e:
    #   logging.info('Bad XSRF token: %r', e.message)
    #   self.response.status = httplib.BAD_REQUEST

    # except permissions.BannedUserException as e:
    #   logging.warning('The user has been banned')
    #   url = framework_helpers.FormatAbsoluteURL(
    #       self.mr, urls.BANNED, include_project=False, copy_params=False)
    #   self.redirect(url, abort=True)

    # except ratelimiter.RateLimitExceeded as e:
    #   logging.info('RateLimitExceeded Exception %s', e)
    #   self.response.status = httplib.BAD_REQUEST
    #   self.response.body = 'Slow your roll.'

    # finally:
    #   self.mr.CleanUp()
    #   self.ratelimiter.CheckEnd(self.request, time.time(), handler_start_time)

    total_processing_time = time.time() - handler_start_time
    logging.info(
        'Processed request in %d ms', int(total_processing_time * 1000))

    end_count0, end_count1, end_count2 = gc.get_count()
    logging.info('gc counts: %d %d %d', end_count0, end_count1, end_count2)

  def get(self, project_name):
    self.dispatch()
    template_path = framework_constants.TEMPLATE_PATH + 'framework/footer-shared.ezt'
    template = ezt.Template(
          fname=template_path,
          compress_whitespace=True,
          base_format=ezt.FORMAT_HTML)
    buf = template_helpers.cStringIOUnicodeWrapper()
    template.generate(buf, {
      'old_ui_url': 'old',
      'new_ui_url': None,
      'projectname': 't',
      'is_ezt': ezt.boolean(False),
      'app_version': 'version1'
    })
    whole_page = buf.getvalue()
    return whole_page
