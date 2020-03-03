# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

from api import resource_name_converters as rnc
from api.v1 import converters
from api.v1 import monorail_servicer
from api.v1.api_proto import projects_pb2
from api.v1.api_proto import project_objects_pb2
from api.v1.api_proto import projects_prpc_pb2
from businesslogic import work_env
from framework import framework_views


class ProjectsServicer(monorail_servicer.MonorailServicer):
  """Handle API requests related to Project objects.
  Each API request is implemented with a method as defined in the
  .proto file. Each method does any request-specific validation, uses work_env
  to safely operate on business objects, and returns a response proto.
  """

  DESCRIPTION = projects_prpc_pb2.ProjectsServiceDescription

  @monorail_servicer.PRPCMethod
  def ListIssueTemplates(self, mc, request):
    # MonorailConnection, ListIssueTemplatesRequest ->
    # ListIssueTemplatesResponse
    """pRPC API method that implements ListIssueTemplates.

      Raises:
        InputException if the request.parent is invalid.
    """
    project_name = rnc.IngestProjectName(request.parent)

    with work_env.WorkEnv(mc, self.services) as we:
      # Fetch project
      project = we.GetProjectByName(project_name)
      # Fetch templates for project
      templates = we.ListProjectTemplates(project)

      # Fetch admins for all templates
      all_admin_ids = []
      for template in templates:
        all_admin_ids += template.admin_ids
      admins = self.services.user.GetUsersByIDs(mc.cnxn, all_admin_ids)

      # Make UserViews for admins, needed for display_name and email
      with mc.profiler.Phase('making user views'):
        users_by_id = framework_views.MakeAllUserViews(
            mc.cnxn, self.services.user, all_admin_ids)
        framework_views.RevealAllEmailsToMembers(
            mc.auth, project, users_by_id)

    return projects_pb2.ListIssueTemplatesResponse(
        templates=converters.ConvertIssueTemplates(
            templates, project_name, admins, users_by_id))
