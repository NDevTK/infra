# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

RESTRICT_VIEW_PATTERN = 'restrict-view-%'


def GetPersonalAtRiskLabelIDs(
  cnxn, user, config_svc, effective_ids, project, perms):
  """Return list of label_ids for restriction labels that user can't view.

  Args:
    cnxn: An instance of MonorailConnection.
    user: User PB for the signed in user making the request, or None for anon.
    config_svc: An instance of ConfigService.
    effective_ids: The effective IDs of the current user.
    project: A project object for the current project.
    perms: A PermissionSet for the current user.
  Returns:
    A list of LabelDef IDs the current user is forbidden to access.
  """
  if user and user.is_site_admin:
    return []

  at_risk_label_ids = []
  label_def_rows = config_svc.GetLabelDefRowsAnyProject(
    cnxn, where=[('LOWER(label) LIKE %s', [RESTRICT_VIEW_PATTERN])])

  for label_id, _pid, _rank, label, _docstring, _hidden in label_def_rows:
    label_lower = label.lower()
    needed_perm = label_lower.split('-', 2)[-1]

    if not perms.CanUsePerm(needed_perm, effective_ids, project, []):
      at_risk_label_ids.append(label_id)

  return at_risk_label_ids


def InitializeAppEngineSearchAPI():
  import os
  import sys

  # Add third_party/ to sys.path to find antlr3.
  sys.path.append(os.path.join(os.path.dirname(__file__), 'third_party'))

  # Add third_party/google/ to the google namespace.
  # This makes Python look in this additional location for google.net.proto.
  import google
  package_path = os.path.join(
      os.path.dirname(__file__), 'third_party', 'google')
  google.__path__.append(package_path)

  # Add third_party/google/ to the google.appengine.api namespace. This makes
  # Python look in this additional location for google.appengine.api.search.
  import google.appengine.api
  package_path = os.path.join(
      os.path.dirname(__file__), 'third_party', 'google', 'appengine', 'api')
  google.appengine.api.__path__.append(package_path)

  # Add third_party/google/ to the google.appengine.datastore namespace.
  # This makes Python look in this additional location for document_pb.py.
  import google.appengine.datastore
  package_path = os.path.join(
      os.path.dirname(__file__), 'third_party', 'google', 'appengine',
      'datastore')
  google.appengine.datastore.__path__.append(package_path)
