# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.ext import ndb


class RedirectProjectTemplate(ndb.Model):
  """Represents a template redirect information."""
  ProjectName = ndb.StringProperty()
  MonorailTemplateName = ndb.StringProperty()
  RedirectComponentID = ndb.StringProperty()
  RedirectTemplateID = ndb.StringProperty()

  @classmethod
  def Get(cls, project, template_name):
    key = project + ':' + template_name
    entity = ndb.Key('RedirectProjectTemplate', key).get()
    if not entity:
      return None, None
    return entity.RedirectComponentID, entity.RedirectTemplateID
