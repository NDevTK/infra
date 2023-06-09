# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.ext import ndb


class RedirectCustomValue(ndb.Model):
  """Represents a project custome value redirect information."""
  ProjectName = ndb.StringProperty()
  MonorailType = ndb.StringProperty()
  MonorailValue = ndb.StringProperty()
  RedirectType = ndb.StringProperty()
  RedirectValue = ndb.StringProperty()

  @classmethod
  def Get(cls, project, custom_type, value):
    # TODO(b/283983843): add function to handle multiple values.
    entity = cls.query(
        RedirectCustomValue.ProjectName == project,
        RedirectCustomValue.MonorailType == custom_type,
        RedirectCustomValue.MonorailValue == value).get()
    if not entity:
      return None, None
    return entity.RedirectType, entity.RedirectValue
