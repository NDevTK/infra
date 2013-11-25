# coding=utf-8
# Copyright (c) 2012 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Status management pages."""

import datetime
import json
import re

from google.appengine.api import memcache
from google.appengine.ext import db

from base_page import BasePage
import utils


ALLOWED_ORIGINS = [
    'https://gerrit-int.chromium.org',
    'https://gerrit.chromium.org',
    'https://chrome-internal-review.googlesource.com',
    'https://chromium-review.googlesource.com',
]


class Status(db.Model):
  """Description for the status table."""
  # The username who added this status.
  username = db.StringProperty(required=True)
  # The date when the status got added.
  date = db.DateTimeProperty(auto_now_add=True)
  # The message. It can contain html code.
  message = db.StringProperty(required=True)

  @property
  def general_state(self):
    """Returns a string representing the state that the status message
    describes.
    """
    message = self.message
    closed = re.search('close', message, re.IGNORECASE)
    if closed and re.search('maint', message, re.IGNORECASE):
      return 'maintenance'
    if re.search('throt', message, re.IGNORECASE):
      return 'throttled'
    if closed:
      return 'closed'
    return 'open'

  @property
  def can_commit_freely(self):
    return self.general_state == 'open'

  def AsDict(self):
    data = super(Status, self).AsDict()
    data['general_state'] = self.general_state
    data['can_commit_freely'] = self.can_commit_freely
    return data


def get_status():
  """Returns the current Status, e.g. the most recent one."""
  status = memcache.get('last_status')
  if status is None:
    status = Status.all().order('-date').get()
    # Use add instead of set(); must not change it if it was already set.
    memcache.add('last_status', status)
  return status


def put_status(status):
  """Sets the current Status, e.g. append a new one."""
  status.put()
  memcache.set('last_status', status)
  memcache.delete('last_statuses')


def get_last_statuses(limit):
  """Returns the last |limit| statuses."""
  statuses = memcache.get('last_statuses')
  if not statuses or len(statuses) < limit:
    statuses = Status.all().order('-date').fetch(limit)
    memcache.add('last_statuses', statuses)
  return statuses[:limit]


def parse_date(date):
  """Parses a date."""
  match = re.match(r'^(\d\d\d\d)-(\d\d)-(\d\d)$', date)
  if match:
    return datetime.datetime(
        int(match.group(1)), int(match.group(2)), int(match.group(3)))
  if date.isdigit():
    return datetime.datetime.utcfromtimestamp(int(date))
  return None


def limit_length(string, length):
  """Limits the string |string| at length |length|.

  Inserts an ellipsis if it is cut.
  """
  string = unicode(string.strip())
  if len(string) > length:
    string = string[:length - 1] + u'…'
  return string


class AllStatusPage(BasePage):
  """Displays a big chunk, 1500, status values."""
  @utils.requires_read_access
  def get(self):
    query = db.Query(Status).order('-date')
    start_date = self.request.get('startTime')
    if start_date:
      query.filter('date <', parse_date(start_date))
    try:
      limit = int(self.request.get('limit'))
    except ValueError:
      limit = 1000
    end_date = self.request.get('endTime')
    beyond_end_of_range_status = None
    if end_date:
      query.filter('date >=', parse_date(end_date))
      # We also need to get the very next status in the range, otherwise
      # the caller can't tell what the effective tree status was at time
      # |end_date|.
      beyond_end_of_range_status = Status.all(
          ).filter('date <', end_date).order('-date').get()

    out_format = self.request.get('format', 'csv')
    if out_format == 'csv':
      # It's not really an html page.
      self.response.headers['Content-Type'] = 'text/plain'
      template_values = self.InitializeTemplate(self.APP_NAME + ' Tree Status')
      template_values['status'] = query.fetch(limit)
      template_values['beyond_end_of_range_status'] = beyond_end_of_range_status
      self.DisplayTemplate('allstatus.html', template_values)
    elif out_format == 'json':
      self.response.headers['Content-Type'] = 'application/json'
      self.response.headers['Access-Control-Allow-Origin'] = '*'
      statuses = [s.AsDict() for s in query.fetch(limit)]
      if beyond_end_of_range_status:
        statuses.append(beyond_end_of_range_status.AsDict())
      data = json.dumps(statuses)
      callback = self.request.get('callback')
      if callback:
        if re.match(r'^[a-zA-Z$_][a-zA-Z$0-9._]*$', callback):
          data = '%s(%s);' % (callback, data)
      self.response.out.write(data)
    else:
      self.response.headers['Content-Type'] = 'text/plain'
      self.response.out.write('Invalid format')


class CurrentPage(BasePage):
  """Displays the /current page."""

  def get(self):
    # Show login link on current status bar when login is required.
    out_format = self.request.get('format', 'html')
    if out_format == 'html' and not self.read_access and not self.user:
      template_values = self.InitializeTemplate(self.APP_NAME + ' Tree Status')
      template_values['show_login'] = True
      self.DisplayTemplate('current.html', template_values, use_cache=True)
    else:
      self._handle()

  @utils.requires_read_access
  def _handle(self):
    """Displays the current message in various formats."""
    out_format = self.request.get('format', 'html')
    status = get_status()
    if out_format == 'raw':
      self.response.headers['Content-Type'] = 'text/plain'
      self.response.out.write(status.message)
    elif out_format == 'json':
      self.response.headers['Content-Type'] = 'application/json'
      origin = self.request.headers.get('Origin')
      if self.request.get('with_credentials') and origin in ALLOWED_ORIGINS:
        self.response.headers['Access-Control-Allow-Origin'] = origin
        self.response.headers['Access-Control-Allow-Credentials'] = 'true'
      else:
        self.response.headers['Access-Control-Allow-Origin'] = '*'
      data = json.dumps(status.AsDict())
      callback = self.request.get('callback')
      if callback:
        if re.match(r'^[a-zA-Z$_][a-zA-Z$0-9._]*$', callback):
          data = '%s(%s);' % (callback, data)
      self.response.out.write(data)
    elif out_format == 'html':
      template_values = self.InitializeTemplate(self.APP_NAME + ' Tree Status')
      template_values['show_login'] = False
      template_values['message'] = status.message
      template_values['state'] = status.general_state
      self.DisplayTemplate('current.html', template_values, use_cache=True)
    else:
      self.error(400)


class StatusPage(BasePage):
  """Displays the /status page."""

  def get(self):
    """Displays 1 if the tree is open, and 0 if the tree is closed."""
    # NOTE: This item is always public to allow waterfalls to check it.
    status = get_status()
    self.response.headers['Cache-Control'] = 'no-cache, private, max-age=0'
    self.response.headers['Content-Type'] = 'text/plain'
    self.response.out.write(str(int(status.can_commit_freely)))

  @utils.requires_bot_login
  @utils.requires_write_access
  def post(self):
    """Adds a new message from a backdoor.

    The main difference with MainPage.post() is that it doesn't look for
    conflicts and doesn't redirect to /.
    """
    message = self.request.get('message')
    message = limit_length(message, 500)
    username = self.request.get('username')
    if message and username:
      put_status(Status(message=message, username=username))
    self.response.out.write('OK')


class StatusViewerPage(BasePage):
  """Displays the /status_viewer page."""

  @utils.requires_read_access
  def get(self):
    """Displays status_viewer.html template."""
    template_values = self.InitializeTemplate(self.APP_NAME + ' Tree Status')
    self.DisplayTemplate('status_viewer.html', template_values)


class MainPage(BasePage):
  """Displays the main page containing the last 25 messages."""

  # NOTE: This is require_login in order to ensure that authentication doesn't
  # happen while changing the tree status.
  @utils.requires_login
  @utils.requires_read_access
  def get(self):
    return self._handle()

  def _handle(self, error_message='', last_message=''):
    """Sets the information to be displayed on the main page."""
    try:
      limit = min(max(int(self.request.get('limit')), 1), 1000)
    except ValueError:
      limit = 25
    status = get_last_statuses(limit)
    current_status = get_status()
    if not last_message:
      last_message = current_status.message

    template_values = self.InitializeTemplate(self.APP_NAME + ' Tree Status')
    template_values['status'] = status
    template_values['message'] = last_message
    template_values['last_status_key'] = current_status.key()
    template_values['error_message'] = error_message
    template_values['limit'] = limit
    self.DisplayTemplate('main.html', template_values)

  @utils.requires_login
  @utils.requires_write_access
  def post(self):
    """Adds a new message."""
    # We pass these variables back into get(), prepare them.
    last_message = ''
    error_message = ''

    # Get the posted information.
    new_message = self.request.get('message')
    new_message = limit_length(new_message, 500)
    last_status_key = self.request.get('last_status_key')
    if not new_message:
      # A submission contained no data. It's a better experience to redirect
      # in this case.
      self.redirect("/")
      return

    current_status = get_status()
    if current_status and (last_status_key != str(current_status.key())):
      error_message = ('Message not saved, mid-air collision detected, '
                       'please resolve any conflicts and try again!')
      last_message = new_message
      return self._handle(error_message, last_message)
    else:
      put_status(Status(message=new_message, username=self.user.email()))
      self.redirect("/")


def bootstrap():
  # Guarantee that at least one instance exists.
  if db.GqlQuery('SELECT __key__ FROM Status').get() is None:
    Status(username='none', message='welcome to status').put()
