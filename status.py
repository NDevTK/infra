# Copyright (c) 2011 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Status management pages."""

import datetime
import re

import simplejson as json
from google.appengine.api import memcache
from google.appengine.ext import db

from base_page import BasePage
import utils


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


def parse_date(date):
  """Parses a date."""
  match = re.match(r'^(\d\d\d\d)-(\d\d)-(\d\d)$', date)
  if match:
    return datetime.datetime(
        int(match.group(1)), int(match.group(2)), int(match.group(3)))
  if date.isdigit():
    return datetime.datetime.utcfromtimestamp(int(date))
  return None


class AllStatusPage(BasePage):
  """Displays a big chunk, 1500, status values."""
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
      beyond_end_of_range_status = Status.gql(
          'WHERE date < :end_date ORDER BY date DESC LIMIT 1',
          end_date=end_date).get()

    out_format = self.request.get('format', 'csv')
    if out_format == 'csv':
      # It's not really an html page.
      self.response.headers['Content-Type'] = 'text/plain'
      template_values = self.InitializeTemplate(self.app_name + ' Tree Status')
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
    """Displays the current message and nothing else."""
    # Module 'google.appengine.api.memcache' has no 'get' member
    # pylint: disable=E1101
    out_format = self.request.get('format', 'html')
    status = memcache.get('last_status')
    if status is None:
      status = Status.gql('ORDER BY date DESC').get()
      # Cache 2 seconds.
      memcache.add('last_status', status, 2)
    if not status:
      self.error(501)
    elif out_format == 'raw':
      self.response.headers['Content-Type'] = 'text/plain'
      self.response.out.write(status.message)
    elif out_format == 'json':
      self.response.headers['Content-Type'] = 'application/json'
      if self.request.get('with_credentials'):
        self.response.headers['Access-Control-Allow-Origin'] = (
            'gerrit-int.chromium.org, gerrit.chromium.org')
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
      template_values = self.InitializeTemplate(self.app_name + ' Tree Status')
      template_values['message'] = status.message
      template_values['state'] = status.general_state
      self.DisplayTemplate('current.html', template_values, use_cache=True)
    else:
      self.error(400)


class StatusPage(BasePage):
  """Displays the /status page."""

  def get(self):
    """Displays 1 if the tree is open, and 0 if the tree is closed."""
    status = Status.gql('ORDER BY date DESC').get()
    if status:
      self.response.headers['Cache-Control'] = 'no-cache, private, max-age=0'
      self.response.headers['Content-Type'] = 'text/plain'
      self.response.out.write(str(int(status.can_commit_freely)))

  @utils.admin_only
  def post(self):
    """Adds a new message from a backdoor.

    The main difference with MainPage.post() is that it doesn't look for
    conflicts and doesn't redirect to /.
    """
    message = self.request.get('message')
    username = self.request.get('username')
    if message and username:
      status = Status(message=message, username=username)
      status.put()
      # Cache the status.
      # Module 'google.appengine.api.memcache' has no 'set' member
      # pylint: disable=E1101
      memcache.set('last_status', status)
    self.response.out.write('OK')


class StatusViewerPage(BasePage):
  """Displays the /status_viewer page."""

  def get(self):
    """Displays status_viewer.html template."""
    template_values = self.InitializeTemplate(self.app_name + ' Tree Status')
    self.DisplayTemplate('status_viewer.html', template_values)


class MainPage(BasePage):
  """Displays the main page containing the last 25 messages."""

  @utils.require_user
  def get(self):
    return self._handle()

  def _handle(self, error_message='', last_message=''):
    """Sets the information to be displayed on the main page."""
    try:
      limit = min(max(int(self.request.get('limit')), 1), 1000)
    except ValueError:
      limit = 25
    status = Status.gql('ORDER BY date DESC LIMIT %d' % limit)
    current_status = status.get()
    if not last_message and current_status:
      last_message = current_status.message

    template_values = self.InitializeTemplate(self.app_name + ' Tree Status')
    template_values['status'] = status
    template_values['message'] = last_message
    # If the DB is empty, current_status is None.
    if current_status:
      template_values['last_status_key'] = current_status.key()
    template_values['error_message'] = error_message
    self.DisplayTemplate('main.html', template_values)

  @utils.require_user
  @utils.admin_only
  def post(self):
    """Adds a new message."""
    # We pass these variables back into get(), prepare them.
    last_message = ''
    error_message = ''

    # Get the posted information.
    new_message = self.request.get('message')
    last_status_key = self.request.get('last_status_key')
    if new_message:
      current_status = Status.gql('ORDER BY date DESC').get()
      if current_status and (last_status_key != str(current_status.key())):
        error_message = ('Message not saved, mid-air collision detected, '
                         'please resolve any conflicts and try again!')
        last_message = new_message
      else:
        status = Status(message=new_message, username=self.user.email())
        status.put()
        # Cache the status.
        # Module 'google.appengine.api.memcache' has no 'set' member
        # pylint: disable=E1101
        memcache.set('last_status', status)

    return self._handle(error_message, last_message)


def bootstrap():
  if db.GqlQuery('SELECT __key__ FROM Status').get() is None:
    Status(username='none', message='welcome to status').put()
