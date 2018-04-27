# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""WorkEnv is a context manager and API for high-level operations.

A work environment is used by request handlers for the legacy UI, v1
API, and v2 API.  The WorkEnvironment operations are a common code
path that does permission checking, rate limiting, and other systemic
functionality so that that code is not duplicated in multiple request
handlers.

Responsibilities of request handers (legacy UI and external API) and associated
frameworks:
+ API: check oauth client whitelist
+ Rate-limiting
+ Create a MonorailContext (or MonorailRequest) object:
  - Parse the request, including syntaxtic validation, e.g, non-negative ints
  - Authenticate the requesting user
+ Call the WorkEnvironment to perform the requested action
  - Catch exceptions and generate error messages
+ UI: Decide screen flow, and on-page online-help
+ Render the result business objects as UI HTML or API response protobufs.

Responsibilities of WorkEnv:
+ Most monitoring, profiling, and logging
+ Apply business rules:
  - Check permissions
  - Detailed validation of request parameters
  - Raise exceptions to indicate problems
+ Call the services layer to make DB changes
+ Enqueue tasks for background follow-up work:
  - E.g., email notifications

Responsibilities of the Services layer:
+ CRUD operations on objects in the database
+ App-specific interface around external services:
  - E.g., GAE search, GCS, monorail-predict
"""

import logging

from features import send_notifications
from features import features_bizobj
from features import hotlist_helpers
from framework import exceptions
from framework import framework_helpers
from framework import permissions
from search import frontendsearchpipeline
from services import project_svc
from sitewide import sitewide_helpers
from tracker import tracker_bizobj
from proto import project_pb2


# TODO(jrobbins): rate limiting and permission checking in each method.

# TODO(jrobbins): break this file into one facade plus ~5
# implementation parts that roughly correspond to services files.


class WorkEnv(object):

  def __init__(self, mr, services, phase=None):
    self.mr = mr
    self.services = services
    self.phase = phase

  def __enter__(self):
    if self.mr.profiler and self.phase:
      self.mr.profiler.StartPhase(name=self.phase)
    return self  # The instance of this class is the context object.

  def __exit__(self, exception_type, value, traceback):
    if self.mr.profiler and self.phase:
      self.mr.profiler.EndPhase()
    return False  # Re-raise any exception in the with-block.

  ### Site methods

  # FUTURE: GetSiteReadOnlyState()
  # FUTURE: SetSiteReadOnlyState()
  # FUTURE: GetSiteBannerMessage()
  # FUTURE: SetSiteBannerMessage()

  ### Project methods

  def CreateProject(
      self, project_name, owner_ids, committer_ids, contributor_ids,
      summary, description, state=project_pb2.ProjectState.LIVE,
      access=None, read_only_reason=None, home_page=None, docs_url=None,
      source_url=None, logo_gcs_id=None, logo_file_name=None):
    """Create and store a Project with the given attributes.

    Args:
      cnxn: connection to SQL database.
      project_name: a valid project name, all lower case.
      owner_ids: a list of user IDs for the project owners.
      committer_ids: a list of user IDs for the project members.
      contributor_ids: a list of user IDs for the project contributors.
      summary: one-line explanation of the project.
      description: one-page explanation of the project.
      state: a project state enum defined in project_pb2.
      access: optional project access enum defined in project.proto.
      read_only_reason: if given, provides a status message and marks
        the project as read-only.
      home_page: home page of the project
      docs_url: url to redirect to for wiki/documentation links
      source_url: url to redirect to for source browser links
      logo_gcs_id: google storage object id of the project's logo
      logo_file_name: uploaded file name of the project's logo

    Returns:
      The int project_id of the new project.

    Raises:
      ProjectAlreadyExists: A project with that name already exists.
    """
    with self.mr.profiler.Phase('creating project %r' % project_name):
      project_id = self.services.project.CreateProject(
          self.mr.cnxn, project_name, owner_ids, committer_ids, contributor_ids,
          summary, description, state=state, access=access,
          read_only_reason=read_only_reason, home_page=home_page,
          docs_url=docs_url, source_url=source_url, logo_gcs_id=logo_gcs_id,
          logo_file_name=logo_file_name)
    return project_id

  def ListProjects(self, use_cache=True):
    """Return a list of project IDs that the current user may view."""
    with self.mr.profiler.Phase('list projects for %r' % self.mr.auth.user_id):
      project_ids = self.services.project.GetVisibleLiveProjects(
          self.mr.cnxn, self.mr.auth.user_pb, self.mr.auth.effective_ids,
          use_cache=use_cache)
    return project_ids

  def GetProject(self, project_id, use_cache=True):
    """Return the specified project, TODO: iff the signed in user may view it.

    Args:
      project_id: int project_id of the project to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified project.

    Raises:
      NoSuchProjectException: There is no project with that ID.
    """
    with self.mr.profiler.Phase('getting project %r' % project_id):
      project = self.services.project.GetProject(
          self.mr.cnxn, project_id, use_cache=use_cache)
    return project

  def GetProjectByName(self, project_name, use_cache=True):
    """Return the named project, TODO: iff the signed in user may view it.

    Args:
      project_name: string name of the project to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified project.

    Raises:
      NoSuchProjectException: There is no project with that name.
    """
    with self.mr.profiler.Phase('getting project %r' % project_name):
      project = self.services.project.GetProjectByName(
          self.mr.cnxn, project_name, use_cache=use_cache)
    if not project:
      raise exceptions.NoSuchProjectException()
    return project

  def UpdateProject(
      self, project_id, summary=None, description=None,
      state=None, state_reason=None, access=None, issue_notify_address=None,
      attachment_bytes_used=None, attachment_quota=None, moved_to=None,
      process_inbound_email=None, only_owners_remove_restrictions=None,
      read_only_reason=None, cached_content_timestamp=None,
      only_owners_see_contributors=None, delete_time=None,
      recent_activity=None, revision_url_format=None, home_page=None,
      docs_url=None, source_url=None, logo_gcs_id=None, logo_file_name=None):
    """Update the DB with the given project information."""
    with self.mr.profiler.Phase('updating project %r' % project_id):
      self.services.project.UpdateProject(
          self.mr.cnxn, project_id, summary=summary, description=description,
          state=state, state_reason=state_reason, access=access,
          issue_notify_address=issue_notify_address,
          attachment_bytes_used=attachment_bytes_used,
          attachment_quota=attachment_quota, moved_to=moved_to,
          process_inbound_email=process_inbound_email,
          only_owners_remove_restrictions=only_owners_remove_restrictions,
          read_only_reason=read_only_reason,
          cached_content_timestamp=cached_content_timestamp,
          only_owners_see_contributors=only_owners_see_contributors,
          delete_time=delete_time, recent_activity=recent_activity,
          revision_url_format=revision_url_format, home_page=home_page,
          docs_url=docs_url, source_url=source_url,
          logo_gcs_id=logo_gcs_id, logo_file_name=logo_file_name)


  def DeleteProject(self, project_id):
    """Mark the project as deletable.  It will be reaped by a cron job.

    Args:
      project_id: int ID of the project to delete.

    Returns:
      Nothing.

    Raises:
      NoSuchProjectException: There is no project with that ID.
    """
    with self.mr.profiler.Phase('marking deletable %r' % project_id):
      _project = self.GetProject(project_id)
      self.services.project.MarkProjectDeletable(
          self.mr.cnxn, project_id, self.services.config)

  def StarProject(self, project_id, starred):
    """Star or unstar the specified project.

    Args:
      project_id: int ID of the project to star/unstar.
      starred: true to add a star, false to remove it.

    Returns:
      Nothing.

    Raises:
      NoSuchProjectException: There is no project with that ID.
    """
    if not self.mr.auth.user_id:
      raise ValueError('Anon user cannot star')
    with self.mr.profiler.Phase('(un)starring project %r' % project_id):
      _project = self.GetProject(project_id)
      self.services.project_star.SetStar(
          self.mr.cnxn, project_id, self.mr.auth.user_id, starred)

  def IsProjectStarred(self, project_id):
    """Return True if the current user has starred the given project.

    Args:
      project_id: int ID of the project to check.

    Returns:
      True if starred.

    Raises:
      NoSuchProjectException: There is no project with that ID.
    """
    if project_id is None:
      return False
    with self.mr.profiler.Phase('checking project star %r' % project_id):
      _project = self.GetProject(project_id)
      return self.services.project_star.IsItemStarredBy(
        self.mr.cnxn, project_id, self.mr.auth.user_id)

  def ListStarredProjects(self, viewed_user_id=None):
    """Return a list of projects starred by the current or viewed user.

    Args:
      viewed_user_id: optional user ID for another user's profile page, if
          not supplied, the signed in user is used.

    Returns:
      A list of projects that were starred by current user and that they
      are currently allowed to view.
    """
    if viewed_user_id is None:
      if self.mr.auth.user_id:
        viewed_user_id = self.mr.auth.user_id
      else:
        return []  # Anon user and no viewed user specified.
    with self.mr.profiler.Phase('ListStarredProjects for %r' % viewed_user_id):
      viewable_projects = sitewide_helpers.GetViewableStarredProjects(
          self.mr.cnxn, self.services, viewed_user_id,
          self.mr.auth.effective_ids, self.mr.auth.user_pb)
    return viewable_projects

  def GetProjectConfig(self, project_id, use_cache=True):
    """Return the specifed config, TODO: iff the signed in user may view it.

    Args:
      project_name: string name of the project to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified config.

    Raises:
      NoSuchProjectException: There is no matching config.
    """
    with self.mr.profiler.Phase('getting config for %r' % project_id):
      config = self.services.config.GetProjectConfig(
          self.mr.cnxn, project_id, use_cache=use_cache)
    if not config:
      raise exceptions.NoSuchProjectException()
    return config

  # FUTURE: labels, statuses, fields, components, rules, templates, and views.
  # FUTURE: project saved queries.
  # FUTURE: GetProjectPermissionsForUser()

  ### Issue methods

  def CreateIssue(
      self, project_id, summary, status, owner_id, cc_ids, labels,
      field_values, component_ids, marked_description, blocked_on=None,
      blocking=None, attachments=None, phases=None):
    """Create and store a new issue with all the given information.

    Args:
      project_id: int ID for the current project.
      summary: one-line summary string summarizing this issue.
      status: string issue status value.  E.g., 'New'.
      owner_id: user ID of the issue owner.
      cc_ids: list of user IDs for users to be CC'd on changes.
      labels: list of label strings.  E.g., 'Priority-High'.
      field_values: list of FieldValue PBs.
      component_ids: list of int component IDs.
      marked_description: issue description with initial HTML markup.
      blocked_on: list of issue_ids that this issue is blocked on.
      blocking: list of issue_ids that this issue blocks.
      attachments: [(filename, contents, mimetype),...] attachments uploaded at
          the time the comment was made.
      phases: list of Milestone PBs.

    Returns:
      A tuple (newly created Issue, Comment PB for the description).
    """
    with self.mr.profiler.Phase('creating issue in project %r' % project_id):
      reporter_id = self.mr.auth.user_id
      new_local_id, comment = self.services.issue.CreateIssue(
          self.mr.cnxn, self.services, project_id, summary, status,
          owner_id, cc_ids, labels, field_values, component_ids, reporter_id,
          marked_description, blocked_on=blocked_on, blocking=blocking,
          attachments=attachments, index_now=False, phases=phases)
      logging.info('created issue %r in project %r', new_local_id, project_id)

    with self.mr.profiler.Phase('following up after issue creation'):
      self.services.project.UpdateRecentActivity(self.mr.cnxn, project_id)
      new_issue = self.services.issue.GetIssueByLocalID(
          self.mr.cnxn, project_id, new_local_id)
      self.services.issue.EnqueueIssuesForIndexing(self.mr.cnxn,
          [new_issue.issue_id])

    return new_issue, comment

  def ListIssues(self):
    """Do an issue search using info in mr and return a pipeline object."""
    with self.mr.profiler.Phase('searching issues'):
      pipeline = frontendsearchpipeline.FrontendSearchPipeline(
          self.mr, self.services, self.mr.num)
      if not self.mr.errors.AnyErrors():
        pipeline.SearchForIIDs()
        pipeline.MergeAndSortIssues()
        pipeline.Paginate()
      return pipeline

  def FindIssuePositionInSearch(self, issue):
    """Do an issue search and return flipper info for the given issue.

    Args:
      issue: issue that the user is currently viewing.

    Returns:
      A 4-tuple of flipper info: (prev_iid, cur_index, next_iid, total_count).
    """
    with self.mr.profiler.Phase('finding issue position in search'):
      pipeline = frontendsearchpipeline.FrontendSearchPipeline(
          self.mr, self.services, None)
      if not self.mr.errors.AnyErrors():
        # Only do the search if the user's query parsed OK.
        pipeline.SearchForIIDs()

      # Note: we never call MergeAndSortIssues() because we don't need a unified
      # sorted list, we only need to know the position on such a list of the
      # current issue.
      prev_iid, cur_index, next_iid = pipeline.DetermineIssuePosition(issue)
      return prev_iid, cur_index, next_iid, pipeline.total_count

  def GetIssue(self, issue_id, use_cache=True):
    """Return the specified issue, TODO: iff the signed in user may view it."""
    with self.mr.profiler.Phase('getting issue %r' % issue_id):
      issue = self.services.issue.GetIssue(
          self.mr.cnxn, issue_id, use_cache=use_cache)
    return issue

  def GetIssueByLocalID(self, project_id, local_id, use_cache=True):
    """Return the specified issue, TODO: iff the signed in user may view it.

    Args:
      project_id: int project ID of the project that contains the issue.
      local_id: int issue local id number.
      use_cache: set to False when doing read-modify-write operations.

    Returns:
      The specified Issue PB.

    Raises:
      exceptions.InputException: Something was not specified properly.
      exceptions.NoSuchIssueException: The issue does not exist.
    """
    if project_id is None:
      raise exceptions.InputException('No project specified')
    if local_id is None:
      raise exceptions.InputException('No issue local_id specified')

    with self.mr.profiler.Phase('getting issue %r:%r' % (project_id, local_id)):
      issue = self.services.issue.GetIssueByLocalID(
          self.mr.cnxn, project_id, local_id, use_cache=use_cache)
    return issue

  def UpdateIssueApprovalStatus(
      self, issue_id, approval_id, new_status, set_on):
    """Update an issue's approvalvalue."""
    issue, approval_value = self.services.issue.GetIssueApproval(
        self.mr.cnxn, issue_id, approval_id)

    if not permissions.CanUpdateApprovalStatus(
        self.mr.auth.effective_ids, approval_value.approver_ids,
        approval_value.status, new_status):
      raise permissions.PermissionException(
          'User not allowed to make this status update.')

    with self.mr.profiler.Phase(
        'updating approvalvalue for issue %r' % issue_id):
      self.services.issue.UpdateIssueApprovalStatus(
          self.mr.cnxn, issue_id, approval_id, new_status, self.mr.auth.user_id, set_on)
      status_amendment = tracker_bizobj.MakeApprovalStatusAmendment(new_status)
      comment = self.services.issue.CreateIssueComment(
          self.mr.cnxn, issue, self.mr.auth.user_id, '', amendments=[status_amendment],
          approval_id=approval_id)
      send_notifications.PrepareAndSendApprovalChangeNotification(
          issue.issue_id, approval_value.approval_id, self.mr.request.host, comment.id)
    logging.info('updated ApprovalValue %r for issue %r',
                 approval_id, issue_id)

  def UpdateIssueApprovalApprovers(self, issue_id, approval_id, approver_ids):
    """Update an issue's approval approvers."""
    issue, approval_value = self.services.issue.GetIssueApproval(
        self.mr.cnxn, issue_id, approval_id)

    # TODO(jojwang): monorail:3582, OR this with project admin/owners
    # or check for admin/owner perms inside CanUpdateApprovers
    if not permissions.CanUpdateApprovers(
        self.mr.auth.effective_ids, approval_value.approver_ids):
      raise permissions.PermissionException(
          'User not allowed to modify approvers of this approval.')

    with self.mr.profiler.Phase(
        'updating approvers for issue %r, approval %r' % (
            issue_id, approval_id)):
      self.services.issue.UpdateIssueApprovalApprovers(
          self.mr.cnxn, issue_id, approval_id, approver_ids)
      approver_amendment = tracker_bizobj.MakeApprovalApproversAmendment(
          approval_value.approver_ids, approver_ids)
      comment = self.services.issue.CreateIssueComment(
          self.mr.cnxn, issue, self.mr.auth.user_id, '',
          amendments=[approver_amendment], approval_id=approval_id)
      send_notifications.PrepareAndSendApprovalChangeNotification(
          issue.issue_id, approval_value.approval_id, self.mr.request.host, comment.id)
    logging.info('updated approvers to %r' % approver_ids)

  def UpdateIssue(self, issue, delta, comment_content, send_email=True):
    """Update an issue, TODO: iff the signed in user may edit it.

    Args:
      issue: Existing Issue PB for the issue to be modified.
      delta: IssueDelta object containing all the changes to be made.
      comment_content: string content of the user's comment.
      send_email: set to False to suppress email notifications.

    Returns:
      Nothing.
    """
    config = self.GetProjectConfig(issue.project_id)
    old_owner_id = tracker_bizobj.GetOwnerId(issue)

    with self.mr.profiler.Phase('Updating issue %r' % (issue.issue_id)):
      amendments, comment_pb = self.services.issue.DeltaUpdateIssue(
          self.mr.cnxn, self.services, self.mr.auth.user_id, issue.project_id,
          config, issue, delta, comment=comment_content)

    with self.mr.profiler.Phase('Following up after issue update'):
      # TODO(jrobbins): side effects of setting merged_into.
      self.services.project.UpdateRecentActivity(
          self.mr.cnxn, issue.project_id)
      self.services.issue.EnqueueIssuesForIndexing(self.mr.cnxn,
          [issue.issue_id])

    with self.mr.profiler.Phase('Generating notifications'):
      hostport = framework_helpers.GetHostPort()
      reporter_id = self.mr.auth.user_id
      send_notifications.PrepareAndSendIssueChangeNotification(
          issue.issue_id, hostport, reporter_id,
          send_email=send_email, old_owner_id=old_owner_id,
          comment_id=comment_pb.id)

  def DeleteIssue(self, issue, delete):
    """Mark or unmark the given issue as deleted."""
    with self.mr.profiler.Phase('Marking issue %r deleted' % (issue.issue_id)):
      self.services.issue.SoftDeleteIssue(
          self.mr.cnxn, issue.project_id, issue.local_id, delete,
          self.services.user)

  def GetIssuePositionInHotlist(self, current_issue, hotlist):
    """Get index info of an issue within a hotlist.

    Args:
      current_issue: the currently viewed issue.
      hotlist: the hotlist this flipper is flipping through.
    """
    issues_list = self.services.issue.GetIssues(self.mr.cnxn,
        [item.issue_id for item in hotlist.items])
    project_ids = hotlist_helpers.GetAllProjectsOfIssues(
        [issue for issue in issues_list])
    config_list = hotlist_helpers.GetAllConfigsOfProjects(
        self.mr.cnxn, project_ids, self.services)
    harmonized_config = tracker_bizobj.HarmonizeConfigs(config_list)
    (sorted_issues, _hotlist_issues_context,
     _users) = hotlist_helpers.GetSortedHotlistIssues(
         self.mr, hotlist.items, issues_list, harmonized_config,
         self.services)
    (prev_iid, cur_index,
     next_iid) = features_bizobj.DetermineHotlistIssuePosition(
         current_issue, [issue.issue_id for issue in sorted_issues])
    total_count = len(sorted_issues)
    return prev_iid, cur_index, next_iid, total_count

  # FUTURE: GetIssuePermissionsForUser()

  # FUTURE: CreateComment()

  def ListIssueComments(self, viewable_issue):
    """Return comments on the specified viewable issue."""
    with self.mr.profiler.Phase(
        'getting comments for %r' % viewable_issue.issue_id):
      comments = self.services.issue.GetCommentsForIssue(
          self.mr.cnxn, viewable_issue.issue_id)
    return comments

  # FUTURE: UpdateComment()

  def DeleteComment(self, issue, comment, delete):
    """Mark or unmark a comment as deleted by the current user."""
    with self.mr.profiler.Phase(
        'deleting issue %r comment %r' % (issue.issue_id, comment.id)):
      self.services.issue.SoftDeleteComment(
          self.mr.cnxn, issue, comment, self.mr.auth.user_id,
          self.services.user, delete=delete)

  def StarIssue(self, viewable_issue, starred):
    """Set or clear a star on the given issue for the signed in user."""
    if not self.mr.auth.user_id:
      raise exceptions.InputException('Anon cannot star issues')

    with self.mr.profiler.Phase('starring issue %r' % viewable_issue.issue_id):
      config = self.services.config.GetProjectConfig(
          self.mr.cnxn, viewable_issue.project_id)
      self.services.issue_star.SetStar(
          self.mr.cnxn, self.services, config, viewable_issue.issue_id,
          self.mr.auth.user_id, starred)

  def IsIssueStarred(self, viewable_issue, cnxn=None):
    """Return True if the given issue is starred by the signed in user."""
    with self.mr.profiler.Phase('checking star %r' % viewable_issue.issue_id):
      return self.services.issue_star.IsItemStarredBy(
          cnxn or self.mr.cnxn, viewable_issue.issue_id, self.mr.auth.user_id)

  def ListStarredIssueIDs(self):
    """Return a list of the issue IDs that the current issue has starred."""
    with self.mr.profiler.Phase('getting stars %r' % self.mr.auth.user_id):
      return self.services.issue_star.LookupStarredItemIDs(
          self.mr.cnxn, self.mr.auth.user_id)

  def SnapshotCountsQuery(self, timestamp, group_by, label_prefix=None,
                          query=None, canned_query=None):
    """Query IssueSnapshots for daily counts.

    See chart_svc.QueryIssueSnapshots for more detail on arguments.

    Args:
      timestamp (int): Will query for snapshots at this timestamp.
      group_by (str): 2nd dimension, see QueryIssueSnapshots for options.
      label_prefix (str): Required for label queries. Only returns results
        with the supplied prefix.
      query (str, optional): If supplied, will parse & apply query conditions.
      canned_query (str, optional): Value derived from the can= query parameter.

    Returns:
      1. A dict of {name: count} for each item in group_by.
      2. A list of any unsupported query conditions in query.
    """
    with self.mr.profiler.Phase('querying snapshot counts'):
      return self.services.chart.QueryIssueSnapshots(
        self.mr.cnxn, self.services, timestamp, self.mr.auth.effective_ids,
        self.mr.project, self.mr.perms, group_by=group_by,
        label_prefix=label_prefix, query=query, canned_query=canned_query)

  ### User methods

  # FUTURE: GetUser()
  # FUTURE: UpdateUser()
  # FUTURE: DeleteUser()
  # FUTURE: StarUser()
  # FUTURE: IsUserStarred()
  # FUTURE: ListStarredUsers()

  ### Group methods

  # FUTURE: CreateGroup()
  # FUTURE: ListGroups()
  # FUTURE: UpdateGroup()
  # FUTURE: DeleteGroup()

  ### Hotlist methods

  # FUTURE: CreateHotlist()
  # FUTURE: ListHotlistsByUser()
  # FUTURE: UpdateHotlist()
  # FUTURE: DeleteHotlist()
