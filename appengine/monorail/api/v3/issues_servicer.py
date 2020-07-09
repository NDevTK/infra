# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

from api import resource_name_converters as rnc
from api.v3 import converters
from api.v3 import monorail_servicer
from api.v3 import paginator
from api.v3.api_proto import issues_pb2
from api.v3.api_proto import issue_objects_pb2
from api.v3.api_proto import issues_prpc_pb2
from businesslogic import work_env
from framework import exceptions
from tracker import tracker_constants


class IssuesServicer(monorail_servicer.MonorailServicer):
  """Handle API requests related to Issue objects.
  Each API request is implemented with a method as defined in the
  .proto file that does any request-specific validation, uses work_env
  to safely operate on business objects, and returns a response proto.
  """

  DESCRIPTION = issues_prpc_pb2.IssuesServiceDescription

  @monorail_servicer.PRPCMethod
  def GetIssue(self, mc, request):
    # type: (MonorailContext, GetIssueRequest) -> Issue
    """pRPC API method that implements GetIssue.

    Raises:
      InputException: the given name does not have a valid format.
      NoSuchIssueException: the issue is not found.
      PermissionException the user is not allowed to view the issue.
    """
    issue_id = rnc.IngestIssueName(mc.cnxn, request.name, self.services)
    with work_env.WorkEnv(mc, self.services) as we:
      issue = we.GetIssue(issue_id, allow_viewing_deleted=True)
    return self.converter.ConvertIssue(issue)

  @monorail_servicer.PRPCMethod
  def BatchGetIssues(self, mc, request):
    # type: (MonorailContext, BatchGetIssuesRequest) -> BatchGetIssuesResponse
    """pRPC API method that implements BatchGetIssues.

    Raises:
      InputException: If `names` is formatted incorrectly. Or if a parent
          collection in `names` does not match the value in `parent`.
      NoSuchIssueException: If any of the given issues do not exist.
      PermissionException If the requester does not have permission to view one
          (or more) of the given issues.
    """
    if request.parent:
      parent_match = rnc._GetResourceNameMatch(
          request.parent, rnc.PROJECT_NAME_RE)
      parent_project = parent_match.group('project_name')
      with exceptions.ErrorAggregator(exceptions.InputException) as err_agg:
        for name in request.names:
          try:
            name_match = rnc._GetResourceNameMatch(name, rnc.ISSUE_NAME_RE)
            issue_project = name_match.group('project')
            if issue_project != parent_project:
              err_agg.AddErrorMessage(
                  '%s is not a child issue of %s.' % (name, request.parent))
          except exceptions.InputException as e:
            err_agg.AddErrorMessage(e.message)

    with work_env.WorkEnv(mc, self.services) as we:
      issue_ids = rnc.IngestIssueNames(mc.cnxn, request.names, self.services)
      issues_by_iid = we.GetIssuesDict(issue_ids)
    return issues_pb2.BatchGetIssuesResponse(
        issues=self.converter.ConvertIssues(
            [issues_by_iid[issue_id] for issue_id in issue_ids]))

  @monorail_servicer.PRPCMethod
  def SearchIssues(self, mc, request):
    # type: (MonorailContext, SearchIssuesRequest) -> SearchIssuesResponse
    """pRPC API method that implements SearchIssue.

    Raises:
      InputException: if any given names in `projects` are invalid.
    """
    page_size = paginator.CoercePageSize(
        request.page_size, tracker_constants.MAX_ISSUES_PER_PAGE)
    pager = paginator.Paginator(projects=request.projects, page_size=page_size)

    project_names = []
    for resource_name in request.projects:
      match = rnc._GetResourceNameMatch(resource_name, rnc.PROJECT_NAME_RE)
      project_names.append(match.group('project_name'))

    # TODO(crbug.com/monorail/6758): Proto string fields are unicode types in
    # python 2. In python 3 these unicode strings will be represented with
    # string types. pager.GetStart requires a string token during validation
    # (compare_digest()). While in python 2, we're converting the unicode
    # page_token to a string so our existing type annotations can stay accurate
    # now and after the python 3 migration.
    token = str(request.page_token)
    with work_env.WorkEnv(mc, self.services) as we:
      list_result = we.SearchIssues(
          request.query, project_names, mc.auth.user_id, page_size,
          pager.GetStart(token), request.order_by)

    return issues_pb2.SearchIssuesResponse(
        issues=self.converter.ConvertIssues(list_result.items),
        next_page_token=pager.GenerateNextPageToken(list_result.next_start))

  @monorail_servicer.PRPCMethod
  def ListComments(self, mc, request):
    # type: (MonorailContext, ListCommentsRequest) -> ListCommentsResponse
    """pRPC API method that implements ListComments.

    Raises:
      InputException: the given name format or page_size are not valid.
      NoSuchIssue: the parent is not found.
      PermissionException: the user is not allowed to view the parent.
    """
    issue_id = rnc.IngestIssueName(mc.cnxn, request.parent, self.services)
    page_size = paginator.CoercePageSize(
        request.page_size, tracker_constants.MAX_COMMENTS_PER_PAGE)
    pager = paginator.Paginator(parent=request.parent, page_size=page_size)

    with work_env.WorkEnv(mc, self.services) as we:
      list_result = we.SafeListIssueComments(
          issue_id, page_size, pager.GetStart(request.page_token))
    return issues_pb2.ListCommentsResponse(
        comments=self.converter.ConvertComments(issue_id, list_result.items),
        next_page_token=pager.GenerateNextPageToken(list_result.next_start))

  @monorail_servicer.PRPCMethod
  def MakeIssueFromTemplate(self, _mc, _request):
    # type: (MonorailContext, MakeIssueFromTemplateRequest) -> Issue
    """pRPC API method that implements MakeIssueFromTemplate.

    Raises:
      TODO(crbug/monorail/7197): Document errors when implemented
    """
    # Phase 1: Gather info
    #   Get project id and template name from template resource name.
    #   Get template pb.
    #   Make tracker_pb2.IssueDelta from request.template_issue_delta, share
    #   code with v3/ModifyIssue

    # with work_env.WorkEnv(mc, self.services) as we:
    #   created_issue = we.MakeIssueFromTemplate(template, description, delta)

    # Return newly created API issue.
    # return converters.ConvertIssue(created_issue)

    return issue_objects_pb2.Issue()

  @monorail_servicer.PRPCMethod
  def MakeIssue(self, _mc, _request):
    # type: (MonorailContext, MakeIssueRequest) -> Issue
    """pRPC API method that implements MakeIssue.

    Raises:
      TODO(crbug/monorail/7919): Document errors when implemented
    """
    return issue_objects_pb2.Issue()
