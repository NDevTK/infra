# Copyright 2024 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.ext import ndb


class RedirectCustomLabelsToHotlists(ndb.Model):
  """Represents a project custom label to hotlist redirect information."""
  ProjectName = ndb.StringProperty()
  MonorailLabel = ndb.StringProperty()
  HotlistId = ndb.StringProperty()

  @classmethod
  def Get(cls, project, label):
    key = project + ':' + label
    entity = ndb.Key('RedirectCustomLabelsToHotlists', key).get()
    if not entity:
      return None
    return entity.HotlistId


class RedirectToCustomFields(ndb.Model):
  """Represents a project custom label to custom field redirect information."""
  ProjectName = ndb.StringProperty()
  MonorailPrefix = ndb.StringProperty()
  CustomFieldId = ndb.StringProperty()
  ExpectedValueType = ndb.StringProperty()
  ProcessRedirectValue = ndb.StringProperty()

  @classmethod
  def GetAll(cls):
    custom_fields_map = {}
    results = RedirectToCustomFields.query().fetch()
    for result in results:
      custom_fields_map.update(
          {
              result.key.id():
                  {
                      'monorail_prefix': result.MonorailPrefix,
                      'custom_field_id': result.CustomFieldId,
                      'expected_value_type': result.ExpectedValueType,
                      'process_redirect_value': result.ProcessRedirectValue
                  }
          })
    return custom_fields_map
