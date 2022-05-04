# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Helper functions used by the Monorail servlet base class."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import settings
import calendar
import datetime
import logging
import urllib

from framework import framework_bizobj
from framework import framework_helpers
from framework import permissions
from framework import template_helpers
from framework import urls
from framework import xsrf
from proto import project_pb2

from google.appengine.api import app_identity
from google.appengine.api import modules
from google.appengine.api import users

_ZERO = datetime.timedelta(0)


class MethodNotSupportedError(NotImplementedError):
  """An exception class for indicating that the method is not supported.

  Used by GatherPageData and ProcessFormData in Servlet.
  """
  pass


class _ContextDebugItem(object):
  """Wrapper class to generate on-screen debugging output."""

  def __init__(self, key, val):
    """Store the key and generate a string for the value."""
    self.key = key
    if isinstance(val, list):
      nested_debug_strs = [self.StringRep(v) for v in val]
      self.val = '[%s]' % ', '.join(nested_debug_strs)
    else:
      self.val = self.StringRep(val)

  def StringRep(self, val):
    """Make a useful string representation of the given value."""
    try:
      return val.DebugString()
    except Exception:
      try:
        return str(val.__dict__)
      except Exception:
        return repr(val)


class ContextDebugCollection(object):
  """Attach a title to a dictionary for exporting as a table of debug info."""

  def __init__(self, title, collection):
    self.title = title
    self.collection = [
        _ContextDebugItem(key, collection[key])
        for key in sorted(collection.keys())
    ]


class _UTCTimeZone(datetime.tzinfo):
  """UTC"""

  def utcoffset(self, _dt):
    return _ZERO

  def tzname(self, _dt):
    return "UTC"

  def dst(self, _dt):
    return _ZERO


_UTC = _UTCTimeZone()


def GetBannerTime(timestamp):
  """Converts a timestamp into EZT-ready data so it can appear in the banner.

  Args:
    timestamp: timestamp expressed in the following format:
         [year,month,day,hour,minute,second]
         e.g. [2009,3,20,21,45,50] represents March 20 2009 9:45:50 PM

  Returns:
    EZT-ready data used to display the time inside the banner message.
  """
  if timestamp is None:
    return None

  ts = datetime.datetime(*timestamp, tzinfo=_UTC)
  return calendar.timegm(ts.timetuple())


def AssertBasePermissionForUser(user, user_view):
  """Verify user permissions and state.

  Args:
    user: user_pb2.User protocol buffer for the user
    user_view: framework.views.UserView for the user
  """
  if permissions.IsBanned(user, user_view):
    raise permissions.BannedUserException(
        'You have been banned from using this site')


def AssertBasePermission(mr):
  """Make sure that the logged in user can view the requested page.

  Args:
    mr: common information parsed from the HTTP request.

  Returns:
    Nothing

  Raises:
    BannedUserException: If the user is banned.
    PermissionException: If the user does not have permisssion to view.
  """
  AssertBasePermissionForUser(mr.auth.user_pb, mr.auth.user_view)

  if mr.project_name and not CheckPerm(mr, permissions.VIEW):
    logging.info('your perms are %r', mr.perms)
    raise permissions.PermissionException(
        'User is not allowed to view this project')


def CheckPerm(mr, perm, art=None, granted_perms=None):
  """Convenience method that makes permission checks easier.

  Args:
    mr: common information parsed from the HTTP request.
    perm: A permission constant, defined in module framework.permissions
    art: Optional artifact pb
    granted_perms: optional set of perms granted specifically in that artifact.

  Returns:
    A boolean, whether the request can be satisfied, given the permission.
  """
  return mr.perms.CanUsePerm(
      perm, mr.auth.effective_ids, mr.project,
      permissions.GetRestrictions(art), granted_perms=granted_perms)


def CheckPermForProject(mr, perm, project, art=None):
  """Convenience method that makes permission checks for projects easier.

  Args:
    mr: common information parsed from the HTTP request.
    perm: A permission constant, defined in module framework.permissions
    project: The project to enforce permissions for.
    art: Optional artifact pb

  Returns:
    A boolean, whether the request can be satisfied, given the permission.
  """
  perms = permissions.GetPermissions(
      mr.auth.user_pb, mr.auth.effective_ids, project)
  return perms.CanUsePerm(
      perm, mr.auth.effective_ids, project, permissions.GetRestrictions(art))


def ComputeIssueEntryURL(mr, config):
  """Compute the URL to use for the "New issue" subtab.

  Args:
    mr: commonly used info parsed from the request.
    config: ProjectIssueConfig for the current project.

  Returns:
    A URL string to use.  It will be simply "entry" in the non-customized
    case. Otherewise it will be a fully qualified URL that includes some
    query string parameters.
  """
  # TODO: remove the custom_issue_entry_url since its no longer
  if not config.custom_issue_entry_url:
    return '/p/%s/issues/entry' % (mr.project_name)

  return '/p/chromium/issues/wizard'


def IssueListURL(mr, config, query_string=None):
  """Make an issue list URL for non-members or members."""
  url = '/p/%s%s' % (mr.project_name, urls.ISSUE_LIST)
  if query_string:
    url += '?' + query_string
  elif framework_bizobj.UserIsInProject(mr.project, mr.auth.effective_ids):
    if config and config.member_default_query:
      url += '?q=' + urllib.quote_plus(config.member_default_query)
  return url


def ProjectIsRestricted(mr):
  """Return True if the mr has a 'private' project."""
  return (mr.project and mr.project.access != project_pb2.ProjectAccess.ANYONE)


def SafeCreateLoginURL(mr, continue_url=None):
  """Make a login URL w/ a detailed continue URL, otherwise use a short one."""
  continue_url = continue_url or mr.current_page_url
  try:
    url = users.create_login_url(continue_url)
  except users.RedirectTooLongError:
    if mr.project_name:
      url = users.create_login_url('/p/%s' % mr.project_name)
    else:
      url = users.create_login_url('/')

  # Give the user a choice of existing accounts in their session
  # or the option to add an account, even if they are currently
  # signed in to exactly one account.
  if mr.auth.user_id:
    # Notice: this makes assuptions about the output of users.create_login_url,
    # which can change at any time. See https://crbug.com/monorail/3352.
    url = url.replace('/ServiceLogin', '/AccountChooser', 1)
  return url


def SafeCreateLogoutURL(mr):
  """Make a logout URL w/ a detailed continue URL, otherwise use a short one."""
  try:
    return users.create_logout_url(mr.current_page_url)
  except users.RedirectTooLongError:
    if mr.project_name:
      return users.create_logout_url('/p/%s' % mr.project_name)
    else:
      return users.create_logout_url('/')


def VersionBaseURL(request):
  """Return a version-specific URL that we use to load static assets."""
  if settings.local_mode:
    version_base = '%s://%s' % (request.scheme, request.host)
  else:
    version_base = '%s://%s-dot-%s' % (
        request.scheme, modules.get_current_version_name(),
        app_identity.get_default_version_hostname())

  return version_base
