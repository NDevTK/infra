# Copyright (c) 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import logging
import json
import webapp2
from webapp2_extras import jinja2

from model import Badge, UserData


FAKE_BADGE_DEFS = {
  'lion': Badge(
      badge_name='lion', level_1=1, level_2=2, level_3=3,
      title='lion title',
      description='lion desc', icon='lion-face_1f981.png'),
  'shrug': Badge(
      badge_name='shrug', level_1=1, level_2=2, level_3=3,
      title='shrug title',
      description='shrug desc', icon='shrug_1f937.png'),
  'bottle': Badge(
      badge_name='bottle', level_1=1, level_2=2, level_3=3,
      title='bottle title',
      description='bottle desc', icon='bottle-with-popping-cork_1f37e.png'),                      
  }

FAKE_USER_DATA = [
  UserData(badge_name='lion', email='some email', value=1),
  UserData(badge_name='shrug', email='some email', value=2),
  UserData(badge_name='bottle', email='some email', value=3),
]


class BadgeView(object):

  def __init__(self, user_data, badge_def_dict):
    self.user_data = user_data
    self.badge_def = badge_def_dict.get(user_data.badge_name)
    if not self.badge_def:
      self.show = False
      return

    if self.badge_def.show_number:
      self.show = user_data.value > 0
      self.level = user_data.value
    else:
      self.show = user_data.value >= self.badge_def.level_1
      if user_data.value >= self.badge_def.level_3:
        self.level = 3
      elif user_data.value >= self.badge_def.level_2:
        self.level = 2
      else:
        self.level = None  # Don't show level for 1


class BaseHandler(webapp2.RequestHandler):
  """Provide a cached Jinja environment to each request."""

  def __init__(self, *args, **kwargs):
    webapp2.RequestHandler.__init__(self, *args, **kwargs)

  @staticmethod
  def jinja2_factory(app):
    template_dir = os.path.abspath(
        os.path.join(os.path.dirname(__file__), 'templates'))
    config = {'template_path': template_dir}
    jinja = jinja2.Jinja2(app, config=config)
    return jinja

  @webapp2.cached_property
  def jinja2(self):
    # Returns a Jinja2 renderer cached in the app registry.
    return jinja2.get_jinja2(app=self.app, factory=BaseHandler.jinja2_factory)

  def render_response(self, _template, **context):
    # Renders a template and writes the result to the response.
    context.update({
        'app_version': os.environ.get('CURRENT_VERSION_ID'),
        })
    rv = self.jinja2.render_template(_template, **context)
    self.response.write(rv)


class UserPage(BaseHandler):
  def get(self, viewed_user_email, *args):
    if '@' not in viewed_user_email:
      viewed_user_email += '@chromium.org'
    badge_views = [
      BadgeView(user_data, FAKE_BADGE_DEFS)
      for user_data in FAKE_USER_DATA]
    context = {
        'title': 'User Page',
        'viewed_user_email': viewed_user_email,
        'badges': badge_views,
        }
    self.render_response('user.html', **context)


class MainPage(BaseHandler):
  def get(self, *args):
    # TODO: redirect to signed in user... what if not signed in?
    self.redirect('/hinoka@chromium.org')


class BadgePage(BaseHandler):
  def get(self, badge_id, *args):
    context = {
        'title': 'Badge Details',
        'badge_id': badge_id,
        }
    self.render_response('badge.html', **context)

class Update(BaseHandler):
  """Update badge data.

  The expected format is:
  [
    {
      "badge_name": <str>,
      "level_1": <int>,     # Optional
      "level_2": <int>,     # Optional
      "level_3": <int>,     # Optional
      "show_number": <bool>,  # Optional
      "title": <str>,       # Optional
      "description": <str>, # Optional
      "icon": <str>,        # URI, Optional
      "data": {
        {
          "email": <str>,
          "value": <int>,
        }
      }
    }
  ]
  """
  def post(self):
    data = self.request.POST.getone('data')
    if not data:
      self.response.set_status(400)
      self.response.write('no data given')
      return
    o = json.loads(data)
    for badge in o:
      b = self.update_badge_entity(badge)
      self.update_user_data(badge, b)

  @staticmethod
  def update_badge_entity(badge):
    name = badge['badge_name']
    level_1 = badge.get('level_1', None)
    level_2 = badge.get('level_2', None)
    level_3 = badge.get('level_3', None)
    show_number = badge.get('show_number', None)
    title = badge.get('title', None)
    description = badge.get('description', None)
    icon = badge.get('icon', None)
    b = Badge.get_by_id(id=name)
    if not b:
      b = Badge(id=name, badge_name=name)
    if level_1 is not None:
      b.level_1 = level_1
    if level_2 is not None:
      b.level_2 = level_2
    if level_3 is not None:
      b.level_3 = level_3
    if show_number is not None:
      b.show_number = show_number
    if title is not None:
      b.title = title
    if description is not None:
      b.description = description
    if b.icon is not None:
      b.icon = icon
    b.put()
    return b

  @staticmethod
  def update_user_data(badge, b):
    data = badge.get('data', [])
    for item in data:
      email = item['email']
      value = int(item['value'])  # JSON might turn it into a float.
      uid = '%s:%s' % (b.badge_name, email)
      d = UserData.get_by_id(id=uid)
      if d and not d.visible:
        continue
      d = UserData(
          badge_name=b.badge_name, email=email, value=value,
          visible=True, id=uid)
      d.put()


app = webapp2.WSGIApplication([
    (r'/([a-z0-9]+(@[.a-z]+)?)', UserPage),
    (r'/b/([a-z0-9]+)', BadgePage),
    (r'/admin/update', Update),
    ('/', MainPage),
])


