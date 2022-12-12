# Copyright 2020 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

from google.protobuf import empty_pb2

from api import resource_name_converters as rnc
from api.v3 import api_constants
from api.v3 import monorail_servicer
from api.v3 import paginator
from api.v3.api_proto import projects_pb2
from api.v3.api_proto import projects_prpc_pb2
from businesslogic import work_env


class ProjectsServicer(monorail_servicer.MonorailServicer):
  """Handle API requests related to Project objects.
  Each API request is implemented with a method as defined in the
  .proto file. Each method does any request-specific validation, uses work_env
  to safely operate on business objects, and returns a response proto.
  """

  DESCRIPTION = projects_prpc_pb2.ProjectsServiceDescription

  @monorail_servicer.PRPCMethod
  def ListIssueTemplates(self, mc, request):
    # type: (MonorailContext, ListIssueTemplatesRequest) ->
    #   ListIssueTemplatesResponse
    """pRPC API method that implements ListIssueTemplates.

      Raises:
        InputException if the request.parent is invalid.
        NoSuchProjectException if no project exists with the given name.
    """
    project_id = rnc.IngestProjectName(mc.cnxn, request.parent, self.services)

    with work_env.WorkEnv(mc, self.services) as we:
      # TODO(crbug/monorail/7614): Eliminate the need to do this lookup.
      project = we.GetProject(project_id)
      mc.LookupLoggedInUserPerms(project)
      templates = we.ListProjectTemplates(project_id)

    return projects_pb2.ListIssueTemplatesResponse(
        templates=self.converter.ConvertIssueTemplates(project_id, templates))

  @monorail_servicer.PRPCMethod
  def GetComponentDef(self, mc, request):
    # type: (MonorailContext, GetComponentDefRequest) ->
    #   ComponentDef
    """pRPC API method that implements GetComponentDef.

      Raises:
        InputException if the request.parent is invalid.
        NoSuchProjectException if the parent project is not found.
    """
    project_id, component_id = rnc.IngestComponentDefNames(
        mc.cnxn, [request.name], self.services)[0]

    with work_env.WorkEnv(mc, self.services) as we:
      # TODO(crbug/monorail/7614): Eliminate the need to do this lookup.
      project = we.GetProject(project_id)
      mc.LookupLoggedInUserPerms(project)

      component_def = we.GetComponentDef(project_id, component_id)
      api_component_def = self.converter.ConvertComponentDef(component_def)

    return api_component_def

  @monorail_servicer.PRPCMethod
  def ListComponentDefs(self, mc, request):
    # type: (MonorailContext, ListComponentDefsRequest) ->
    #   ListComponentDefsResponse
    """pRPC API method that implements ListComponentDefs.

      Raises:
        InputException if the request.parent is invalid.
        NoSuchProjectException if the parent project is not found.
    """
    project_id = rnc.IngestProjectName(mc.cnxn, request.parent, self.services)

    with work_env.WorkEnv(mc, self.services) as we:
      # TODO(crbug/monorail/7614): Eliminate the need to do this lookup.
      project = we.GetProject(project_id)
      mc.LookupLoggedInUserPerms(project)

      page_size = paginator.CoercePageSize(
        request.page_size, api_constants.MAX_COMPONENTS_PER_PAGE)
      pager = paginator.Paginator(
          parent=request.parent, page_size=page_size)
      list_result = we.ListComponentDefs(
          project_id, page_size, pager.GetStart(request.page_token))

      api_component_defs = self.converter.ConvertComponentDefs(
          list_result.items, project_id)

    return projects_pb2.ListComponentDefsResponse(
        component_defs=api_component_defs,
        next_page_token=pager.GenerateNextPageToken(list_result.next_start))

  @monorail_servicer.PRPCMethod
  def CreateComponentDef(self, mc, request):
    # type: (MonorailContext, CreateComponentDefRequest) ->
    #   ComponentDef
    """pRPC API method that implements CreateComponentDef.

      Raises:
        InputException if the request is invalid.
        NoSuchUserException if any given component admins or ccs do not exist.
        NoSuchProjectException if the parent project does not exist.
        PermissionException if the requester is not allowed to create
          this component.
    """
    project_id = rnc.IngestProjectName(mc.cnxn, request.parent, self.services)
    admin_ids = rnc.IngestUserNames(
        mc.cnxn, request.component_def.admins, self.services)
    cc_ids = rnc.IngestUserNames(
        mc.cnxn, request.component_def.ccs, self.services)

    with work_env.WorkEnv(mc, self.services) as we:
      component_def = we.CreateComponentDef(
          project_id, request.component_def.value,
          request.component_def.docstring, admin_ids, cc_ids,
          request.component_def.labels)

    return self.converter.ConvertComponentDef(component_def)

  @monorail_servicer.PRPCMethod
  def DeleteComponentDef(self, mc, request):
    # type: (MonorailContext, DeleteComponentDefRequest) -> Empty
    """pRPC API method that implements DeleteComponentDef.

      Raises:
        InputException if the request in invalid.
        NoSuchComponentException if the component does not exist.
        PermissionException if the requester is not allowed to delete
          this component.
        NoSuchProjectException if the parent project does not exist.
    """
    project_id, component_id = rnc.IngestComponentDefNames(
        mc.cnxn, [request.name], self.services)[0]

    with work_env.WorkEnv(mc, self.services) as we:
      we.DeleteComponentDef(project_id, component_id)

    return empty_pb2.Empty()

  @monorail_servicer.PRPCMethod
  def ListProjects(self, mc, _):
    # type: (MonorailContext, ListProjectsRequest) -> ListProjectsResponse
    """pRPC API method that implements ListProjects.

      Raises:
        InputException if the request.page_token is invalid or the request does
          not match the previous request that provided the given page_token.
    """
    with work_env.WorkEnv(mc, self.services) as we:
      # NOTE(crbug/monorail/7614): Until the referenced cleanup is complete,
      # all servicer methods that are scoped to a single Project need to call
      # mc.LookupLoggedInUserPerms.
      #  This method does not because it may be scoped to multiple projects.
      allowed_project_ids = we.ListProjects()
      projects_dict = we.GetProjects(allowed_project_ids)
      projects = [projects_dict[proj_id] for proj_id in allowed_project_ids]

    # TODO(crbug.com/monorail/7505): Add pagination logic.
    return projects_pb2.ListProjectsResponse(
        projects=self.converter.ConvertProjects(projects))
