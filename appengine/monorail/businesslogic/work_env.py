# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""WorkEnv is a context manager and API for high-level operations.

A work environment is used by request handlers for the legacy UI, v1
API, and v2 API.  The WorkEnvironment operations are a common code
path that does permission checking, input validation, coordination of
service-level calls, follow-up tasks (e.g., triggering
notifications after certain operations) and other systemic
functionality so that that code is not duplicated in multiple request
handlers.

Responsibilities of request handers (legacy UI and external API) and associated
frameworks:
+ API: check oauth client whitelist or XSRF token
+ Rate-limiting
+ Create a MonorailContext (or MonorailRequest) object:
  - Parse the request, including syntactic validation, e.g, non-negative ints
  - Authenticate the requesting user
+ Call the WorkEnvironment to perform the requested action
  - Catch exceptions and generate error messages
+ UI: Decide screen flow, and on-page online-help
+ Render the result business objects as UI HTML or API response protobufs

Responsibilities of WorkEnv:
+ Most monitoring, profiling, and logging
+ Apply business rules:
  - Check permissions
  - Detailed validation of request parameters
  - Raise exceptions to indicate problems
+ Make coordinated calls to the services layer to make DB changes
  - E.g., calls may need to be made in a specific order
+ Enqueue tasks for background follow-up work:
  - E.g., email notifications

Responsibilities of the Services layer:
+ Individual CRUD operations on objects in the database
  - Each services class should be independent of others
+ App-specific interface around external services:
  - E.g., GAE search, GCS, monorail-predict
+ Business object caches
+ Breaking large operations into batches as appropriate for the underlying
  data storage service, e.g., DB shards and search engine indexing.
"""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import collections
import logging
import time

import settings
from features import features_constants
from features import filterrules_helpers
from features import send_notifications
from features import features_bizobj
from features import hotlist_helpers
from framework import exceptions
from framework import framework_bizobj
from framework import framework_constants
from framework import framework_helpers
from framework import framework_views
from framework import permissions
from search import frontendsearchpipeline
from services import features_svc
from services import tracker_fulltext
from sitewide import sitewide_helpers
from tracker import rerank_helpers
from tracker import tracker_bizobj
from tracker import tracker_constants
from tracker import tracker_helpers
from proto import project_pb2
from proto import tracker_pb2
from proto import user_pb2


# TODO(jrobbins): break this file into one facade plus ~5
# implementation parts that roughly correspond to services files.


class WorkEnv(object):

  def __init__(self, mc, services, phase=None):
    self.mc = mc
    self.services = services
    self.phase = phase

  def __enter__(self):
    if self.mc.profiler and self.phase:
      self.mc.profiler.StartPhase(name=self.phase)
    return self  # The instance of this class is the context object.

  def __exit__(self, exception_type, value, traceback):
    if self.mc.profiler and self.phase:
      self.mc.profiler.EndPhase()
    return False  # Re-raise any exception in the with-block.

  def _UserCanViewProject(self, project):
    """Test if the user may view the given project."""
    return permissions.UserCanViewProject(
        self.mc.auth.user_pb, self.mc.auth.effective_ids, project)

  def _FilterVisibleProjectsDict(self, projects):
    """Filter out projects the user doesn't have permission to view."""
    return {
        key: proj
        for key, proj in projects.items()
        if self._UserCanViewProject(proj)}

  def _AssertPermInProject(self, perm, project):
    """Make sure the user may use perm in the given project."""
    project_perms = permissions.GetPermissions(
        self.mc.auth.user_pb, self.mc.auth.effective_ids, project)
    permitted = project_perms.CanUsePerm(
        perm, self.mc.auth.effective_ids, project, [])
    if not permitted:
      raise permissions.PermissionException(
        'User lacks permission %r in project %s' % (perm, project.project_name))

  def _UserCanViewIssue(self, issue, allow_viewing_deleted=False):
    """Test if user may view an issue according to perms in issue's project."""
    project = self.GetProject(issue.project_id)
    config = self.GetProjectConfig(issue.project_id)
    granted_perms = tracker_bizobj.GetGrantedPerms(
        issue, self.mc.auth.effective_ids, config)
    project_perms = permissions.GetPermissions(
        self.mc.auth.user_pb, self.mc.auth.effective_ids, project)
    issue_perms = permissions.UpdateIssuePermissions(
        project_perms, project, issue, self.mc.auth.effective_ids,
        granted_perms=granted_perms)
    permit_view = permissions.CanViewIssue(
        self.mc.auth.effective_ids, issue_perms, project, issue,
        allow_viewing_deleted=allow_viewing_deleted,
        granted_perms=granted_perms)
    return issue_perms, permit_view

  def _AssertUserCanViewIssue(self, issue, allow_viewing_deleted=False):
    """Make sure the user may view the issue."""
    issue_perms, permit_view = self._UserCanViewIssue(
        issue, allow_viewing_deleted)
    if not permit_view:
      raise permissions.PermissionException(
          'User is not allowed to view this issue')
    return issue_perms

  def _UserCanUsePermInIssue(self, issue, perm):
    """Test if the user may use perm on the given issue."""
    issue_perms = self._AssertUserCanViewIssue(
        issue, allow_viewing_deleted=True)
    return issue_perms.HasPerm(perm, None, None, [])

  def _AssertPermInIssue(self, issue, perm):
    """Make sure the user may use perm on the given issue."""
    permitted = self._UserCanUsePermInIssue(issue, perm)
    if not permitted:
      raise permissions.PermissionException(
        'User lacks permission %r in issue' % perm)

  def _AssertUserCanDeleteComment(self, issue, comment):
    issue_perms = self._AssertUserCanViewIssue(
       issue, allow_viewing_deleted=True)
    commenter = self.services.user.GetUser(self.mc.cnxn, comment.user_id)
    permitted = permissions.CanDeleteComment(
        comment, commenter, self.mc.auth.user_id, issue_perms)
    if not permitted:
      raise permissions.PermissionException('Cannot delete comment')

  def _AssertUserCanViewHotlist(self, hotlist):
    """Make sure the user may view the hotlist."""
    if not permissions.CanViewHotlist(
        self.mc.auth.effective_ids, self.mc.perms, hotlist):
      raise permissions.PermissionException(
          'User is not allowed to view this hotlist')

  def _AssertUserCanEditHotlist(self, hotlist):
    if not permissions.CanEditHotlist(
        self.mc.auth.effective_ids, self.mc.perms, hotlist):
      raise permissions.PermissionException(
          'User is not allowed to edit this hotlist')

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
    if not permissions.CanCreateProject(self.mc.perms):
      raise permissions.PermissionException(
          'User is not allowed to create a project')

    with self.mc.profiler.Phase('creating project %r' % project_name):
      project_id = self.services.project.CreateProject(
          self.mc.cnxn, project_name, owner_ids, committer_ids, contributor_ids,
          summary, description, state=state, access=access,
          read_only_reason=read_only_reason, home_page=home_page,
          docs_url=docs_url, source_url=source_url, logo_gcs_id=logo_gcs_id,
          logo_file_name=logo_file_name)
      self.services.template.CreateDefaultProjectTemplates(self.mc.cnxn,
          project_id)
    return project_id

  def ListProjects(self, domain=None, use_cache=True):
    """Return a list of project IDs that the current user may view."""
    # Note: No permission checks because anyone can list projects, but
    # the results are filtered by permission to view each project.

    with self.mc.profiler.Phase('list projects for %r' % self.mc.auth.user_id):
      project_ids = self.services.project.GetVisibleLiveProjects(
          self.mc.cnxn, self.mc.auth.user_pb, self.mc.auth.effective_ids,
          domain=domain, use_cache=use_cache)

    project_ids = sorted(project_ids)
    return project_ids

  def CheckProjectName(self, project_name):
    """Check that a project name is valid and not already in use.

    Args:
      project_name: str the project name to check.

    Returns:
      None if the user can create a project with that name, or a string with the
      reason the name can't be used.

    Raises:
      PermissionException: The user is not allowed to create a project.
    """
    # We check that the user can create a project so we don't leak information
    # about project names.
    if not permissions.CanCreateProject(self.mc.perms):
      raise permissions.PermissionException(
          'User is not allowed to create a project')

    with self.mc.profiler.Phase('checking project name %s' % project_name):
      if not framework_bizobj.IsValidProjectName(project_name):
        return '"%s" is not a valid project name.' % project_name
      if self.services.project.LookupProjectIDs(self.mc.cnxn, [project_name]):
        return 'There is already a project with that name.'
    return None

  def CheckComponentName(self, project_id, parent_path, component_name):
    """Check that the component name is valid and not already in use.

    Args:
      project_id: int with the id of the project where we want to create the
          component.
      parent_path: optional str with the path of the parent component.
      component_name: str with the name of the proposed component.

    Return:
      None if the user can create a component with that name, or a string with
      the reason the name can't be used.
    """
    # Check that the project exists and the user can view it.
    self.GetProject(project_id)
    # If a parent component is given, make sure it exists.
    config = self.GetProjectConfig(project_id)
    if parent_path and not tracker_bizobj.FindComponentDef(parent_path, config):
      raise exceptions.NoSuchComponentException(
          'Component %r not found' % parent_path)
    with self.mc.profiler.Phase(
        'checking component name %r %r' % (parent_path, component_name)):
      if not tracker_constants.COMPONENT_NAME_RE.match(component_name):
        return '"%s" is not a valid component name.' % component_name
      if parent_path:
        component_name = '%s>%s' % (parent_path, component_name)
      if tracker_bizobj.FindComponentDef(component_name, config):
        return 'There is already a component with that name.'
    return None

  def CheckFieldName(self, project_id, field_name):
    """Check that the field name is valid and not already in use.

    Args:
      project_id: int with the id of the project where we want to create the
          field.
      field_name: str with the name of the proposed field.

    Return:
      None if the user can create a field with that name, or a string with
      the reason the name can't be used.
    """
    # Check that the project exists and the user can view it.
    self.GetProject(project_id)
    config = self.GetProjectConfig(project_id)

    field_name = field_name.lower()
    with self.mc.profiler.Phase('checking field name %r' % field_name):
      if not tracker_constants.FIELD_NAME_RE.match(field_name):
        return '"%s" is not a valid field name.' % field_name
      if field_name in tracker_constants.RESERVED_PREFIXES:
        return 'That name is reserved'
      if field_name.endswith(
          tuple(tracker_constants.RESERVED_COL_NAME_SUFFIXES)):
        return 'That suffix is reserved'
      for fd in config.field_defs:
        fn = fd.field_name.lower()
        if field_name == fn:
          return 'There is already a field with that name.'
        if field_name.startswith(fn + '-'):
          return 'An existing field is a prefix of that name.'
        if fn.startswith(field_name + '-'):
          return 'That name is a prefix of an existing field name.'

    return None

  def GetProjects(self, project_ids, use_cache=True):
    """Return the specified projects.

    Args:
      project_ids: int project_ids of the projects to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified projects.

    Raises:
      NoSuchProjectException: There is no project with that ID.
    """
    with self.mc.profiler.Phase('getting projects %r' % project_ids):
      projects = self.services.project.GetProjects(
          self.mc.cnxn, project_ids, use_cache=use_cache)

    projects = self._FilterVisibleProjectsDict(projects)
    return projects

  def GetProject(self, project_id, use_cache=True):
    """Return the specified project.

    Args:
      project_id: int project_id of the project to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified project.

    Raises:
      NoSuchProjectException: There is no project with that ID.
    """
    projects = self.GetProjects([project_id], use_cache=use_cache)
    if project_id not in projects:
      raise permissions.PermissionException(
          'User is not allowed to view this project')
    return projects[project_id]

  def GetProjectsByName(self, project_names, use_cache=True):
    """Return the named project.

    Args:
      project_names: string names of the projects to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified projects.
    """
    with self.mc.profiler.Phase('getting projects %r' % project_names):
      projects = self.services.project.GetProjectsByName(
          self.mc.cnxn, project_names, use_cache=use_cache)

    for pn in project_names:
      if pn not in projects:
        raise exceptions.NoSuchProjectException('Project %r not found.' % pn)

    projects = self._FilterVisibleProjectsDict(projects)
    return projects

  def GetProjectByName(self, project_name, use_cache=True):
    """Return the named project.

    Args:
      project_name: string name of the project to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified project.

    Raises:
      NoSuchProjectException: There is no project with that name.
    """
    projects = self.GetProjectsByName([project_name], use_cache)
    if not projects:
      raise permissions.PermissionException(
          'User is not allowed to view this project')

    return projects[project_name]

  def GetUserRolesInAllProjects(self, viewed_user_effective_ids):
    """Return the projects where the user has a role.

    Args:
      viewed_user_effective_ids: list of IDs of the user whose projects we want
          to see.

    Returns:
      A triple with projects where the user is an owner, a member or a
      contributor.
    """
    with self.mc.profiler.Phase(
        'Finding roles in all projects for %r' % viewed_user_effective_ids):
      project_ids = self.services.project.GetUserRolesInAllProjects(
          self.mc.cnxn, viewed_user_effective_ids)

    owner_projects = self.GetProjects(project_ids[0])
    member_projects = self.GetProjects(project_ids[1])
    contrib_projects = self.GetProjects(project_ids[2])

    return owner_projects, member_projects, contrib_projects

  def GetUserProjects(self, viewed_user_effective_ids):
    """Get the projects to display in the user's profile.

    Args:
      viewed_user_effective_ids: set of int user IDs of the user being viewed.

    Returns:
      A 4-tuple of lists of PBs:
        - live projects the viewed user owns
        - archived projects the viewed user owns
        - live projects the viewed user is a member of
        - live projects the viewed user is a contributor to

      Any projects the viewing user should not be able to see are filtered out.
      Admins can see everything, while other users can see all non-locked
      projects they own or are a member of, as well as all live projects.
    """
    # Permissions are checked in we.GetUserRolesInAllProjects()
    owner_projects, member_projects, contrib_projects = (
        self.GetUserRolesInAllProjects(viewed_user_effective_ids))

    # We filter out DELETABLE projects, and keep a project where the user has a
    # highest role, e.g. if the user is both an owner and a member, the project
    # is listed under owner projects, not under member_projects.
    archived_projects = [
        project
        for project in owner_projects.values()
        if project.state == project_pb2.ProjectState.ARCHIVED]

    contrib_projects = [
        project
        for pid, project in contrib_projects.items()
        if pid not in owner_projects
        and pid not in member_projects
        and project.state != project_pb2.ProjectState.DELETABLE
        and project.state != project_pb2.ProjectState.ARCHIVED]

    member_projects = [
        project
        for pid, project in member_projects.items()
        if pid not in owner_projects
        and project.state != project_pb2.ProjectState.DELETABLE
        and project.state != project_pb2.ProjectState.ARCHIVED]

    owner_projects = [
        project
        for pid, project in owner_projects.items()
        if project.state != project_pb2.ProjectState.DELETABLE
        and project.state != project_pb2.ProjectState.ARCHIVED]

    by_name = lambda project: project.project_name
    owner_projects = sorted(owner_projects, key=by_name)
    archived_projects = sorted(archived_projects, key=by_name)
    member_projects = sorted(member_projects, key=by_name)
    contrib_projects = sorted(contrib_projects, key=by_name)

    return owner_projects, archived_projects, member_projects, contrib_projects

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
    project = self.GetProject(project_id)
    self._AssertPermInProject(permissions.EDIT_PROJECT, project)

    with self.mc.profiler.Phase('updating project %r' % project_id):
      self.services.project.UpdateProject(
          self.mc.cnxn, project_id, summary=summary, description=description,
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
    project = self.GetProject(project_id)
    self._AssertPermInProject(permissions.EDIT_PROJECT, project)

    with self.mc.profiler.Phase('marking deletable %r' % project_id):
      _project = self.GetProject(project_id)
      self.services.project.MarkProjectDeletable(
          self.mc.cnxn, project_id, self.services.config)

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
    project = self.GetProject(project_id)
    self._AssertPermInProject(permissions.SET_STAR, project)

    with self.mc.profiler.Phase('(un)starring project %r' % project_id):
      self.services.project_star.SetStar(
          self.mc.cnxn, project_id, self.mc.auth.user_id, starred)

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
      raise exceptions.InputException('No project specified')

    if not self.mc.auth.user_id:
      return False

    with self.mc.profiler.Phase('checking project star %r' % project_id):
      # Make sure the project exists and user has permission to see it.
      _project = self.GetProject(project_id)
      return self.services.project_star.IsItemStarredBy(
        self.mc.cnxn, project_id, self.mc.auth.user_id)

  def GetProjectStarCount(self, project_id):
    """Return the number of times the project has been starred.

    Args:
      project_id: int ID of the project to check.

    Returns:
      The number of times the project has been starred.

    Raises:
      NoSuchProjectException: There is no project with that ID.
    """
    if project_id is None:
      raise exceptions.InputException('No project specified')

    with self.mc.profiler.Phase('counting stars for project %r' % project_id):
      # Make sure the project exists and user has permission to see it.
      _project = self.GetProject(project_id)
      return self.services.project_star.CountItemStars(self.mc.cnxn, project_id)

  def ListStarredProjects(self, viewed_user_id=None):
    """Return a list of projects starred by the current or viewed user.

    Args:
      viewed_user_id: optional user ID for another user's profile page, if
          not supplied, the signed in user is used.

    Returns:
      A list of projects that were starred by current user and that they
      are currently allowed to view.
    """
    # Note: No permission checks for this call, but the list of starred
    # projects is filtered based on permission to view.

    if viewed_user_id is None:
      if self.mc.auth.user_id:
        viewed_user_id = self.mc.auth.user_id
      else:
        return []  # Anon user and no viewed user specified.
    with self.mc.profiler.Phase('ListStarredProjects for %r' % viewed_user_id):
      viewable_projects = sitewide_helpers.GetViewableStarredProjects(
          self.mc.cnxn, self.services, viewed_user_id,
          self.mc.auth.effective_ids, self.mc.auth.user_pb)
    return viewable_projects

  def GetProjectConfigs(self, project_ids, use_cache=True):
    """Return the specifed configs.

    Args:
      project_ids: int IDs of the projects to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified configs.
    """
    with self.mc.profiler.Phase('getting configs for %r' % project_ids):
      configs = self.services.config.GetProjectConfigs(
          self.mc.cnxn, project_ids, use_cache=use_cache)

    projects = self._FilterVisibleProjectsDict(
        self.GetProjects(list(configs.keys())))
    configs = {project_id: configs[project_id] for project_id in projects}

    return configs

  def GetProjectConfig(self, project_id, use_cache=True):
    """Return the specifed config.

    Args:
      project_name: string name of the project to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified config.

    Raises:
      NoSuchProjectException: There is no matching config.
    """
    configs = self.GetProjectConfigs([project_id], use_cache)
    if not configs:
      raise exceptions.NoSuchProjectException()
    return configs[project_id]

  def ListProjectTemplates(self, project):
    templates = self.services.template.GetProjectTemplates(
        self.mc.cnxn, project.project_id)
    # Filter non-viewable templates
    if framework_bizobj.UserIsInProject(project, self.mc.auth.effective_ids):
      return templates
    return [template for template in templates if not template.members_only]

  # FUTURE: labels, statuses, fields, components, rules, templates, and views.
  # FUTURE: project saved queries.
  # FUTURE: GetProjectPermissionsForUser()

  ### Issue methods

  def CreateIssue(
      self, project_id, summary, status, owner_id, cc_ids, labels,
      field_values, component_ids, marked_description, blocked_on=None,
      blocking=None, attachments=None, phases=None, approval_values=None,
      send_email=True, reporter_id=None, timestamp=None,
      dangling_blocked_on=None, dangling_blocking=None):
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
      phases: list of Phase PBs.
      approval_values: list of ApprovalValue PBs.
      send_email: set to False to avoid email notifications.
      reporter_id: optional user ID of a different user to attribute this
          issue report to.  The requester must have the ImportComment perm.
      timestamp: optional int timestamp of an imported issue.
      dangling_blocked_on: a list of DanglingIssueRefs this issue is blocked on.
      dangling_blocking: a list of DanglingIssueRefs that this issue blocks.

    Returns:
      A tuple (newly created Issue, Comment PB for the description).
    """
    project = self.GetProject(project_id)
    self._AssertPermInProject(permissions.CREATE_ISSUE, project)

    if reporter_id and reporter_id != self.mc.auth.user_id:
      self._AssertPermInProject(permissions.IMPORT_COMMENT, project)
      importer_id = self.mc.auth.user_id
    else:
      reporter_id = self.mc.auth.user_id
      importer_id = None

    with self.mc.profiler.Phase('creating issue in project %r' % project_id):
      new_local_id, comment = self.services.issue.CreateIssue(
          self.mc.cnxn, self.services, project_id, summary, status,
          owner_id, cc_ids, labels, field_values, component_ids, reporter_id,
          marked_description, blocked_on=blocked_on, blocking=blocking,
          attachments=attachments, index_now=False, phases=phases,
          approval_values=approval_values, timestamp=timestamp,
          importer_id=importer_id, dangling_blocked_on=dangling_blocked_on,
          dangling_blocking=dangling_blocking)
      logging.info('created issue %r in project %r', new_local_id, project_id)

    with self.mc.profiler.Phase('following up after issue creation'):
      self.services.project.UpdateRecentActivity(self.mc.cnxn, project_id)
      new_issue = self.services.issue.GetIssueByLocalID(
          self.mc.cnxn, project_id, new_local_id)

    if send_email:
      with self.mc.profiler.Phase('queueing notification tasks'):
        hostport = framework_helpers.GetHostPort(
            project_name=project.project_name)
        send_notifications.PrepareAndSendIssueChangeNotification(
            new_issue.issue_id, hostport, reporter_id, comment_id=comment.id)
        send_notifications.PrepareAndSendIssueBlockingNotification(
            new_issue.issue_id, hostport, new_issue.blocked_on_iids,
            reporter_id)

    return new_issue, comment

  def MoveIssue(self, issue, target_project):
    """Move issue to the target_project.

    The current user needs to have permission to delete the current issue, and
    to edit issues on the target project.

    Args:
      issue: the issue PB.
      target_project: the project PB where the issue should be moved to.
    Returns:
      The issue PB of the new issue on the target project.
    """
    self._AssertPermInIssue(issue, permissions.DELETE_ISSUE)
    self._AssertPermInProject(permissions.EDIT_ISSUE, target_project)

    if permissions.GetRestrictions(issue):
      raise exceptions.InputException(
          'Issues with Restrict labels are not allowed to be moved')

    with self.mc.profiler.Phase('Moving Issue'):
      tracker_fulltext.UnindexIssues([issue.issue_id])

      # issue is modified by MoveIssues
      old_text_ref = 'issue %s:%s' % (issue.project_name, issue.local_id)
      moved_back_iids = self.services.issue.MoveIssues(
          self.mc.cnxn, target_project, [issue], self.services.user)
      new_text_ref = 'issue %s:%s' % (issue.project_name, issue.local_id)

      if issue.issue_id in moved_back_iids:
        content = 'Moved %s back to %s again.' % (old_text_ref, new_text_ref)
      else:
        content = 'Moved %s to now be %s.' % (old_text_ref, new_text_ref)
      self.services.issue.CreateIssueComment(
          self.mc.cnxn, issue, self.mc.auth.user_id, content,
          amendments=[
              tracker_bizobj.MakeProjectAmendment(target_project.project_name)])

      tracker_fulltext.IndexIssues(
          self.mc.cnxn, [issue], self.services.user, self.services.issue,
          self.services.config)

    return issue

  def CopyIssue(self, issue, target_project):
    """Copy issue to the target_project.

    The current user needs to have permission to delete the current issue, and
    to edit issues on the target project.

    Args:
      issue: the issue PB.
      target_project: the project PB where the issue should be copied to.
    Returns:
      The issue PB of the new issue on the target project.
    """
    self._AssertPermInIssue(issue, permissions.DELETE_ISSUE)
    self._AssertPermInProject(permissions.EDIT_ISSUE, target_project)

    if permissions.GetRestrictions(issue):
      raise exceptions.InputException(
          'Issues with Restrict labels are not allowed to be copied')

    with self.mc.profiler.Phase('Copying Issue'):
      copied_issue = self.services.issue.CopyIssues(
          self.mc.cnxn, target_project, [issue], self.services.user,
          self.mc.auth.user_id)[0]

      issue_ref = 'issue %s:%s' % (issue.project_name, issue.local_id)
      copied_issue_ref = 'issue %s:%s' % (
          copied_issue.project_name, copied_issue.local_id)

      # Add comment to the original issue.
      content = 'Copied %s to %s' % (issue_ref, copied_issue_ref)
      self.services.issue.CreateIssueComment(
          self.mc.cnxn, issue, self.mc.auth.user_id, content)

      # Add comment to the newly created issue.
      # Add project amendment only if the project changed.
      amendments = []
      if issue.project_id != copied_issue.project_id:
        amendments.append(
            tracker_bizobj.MakeProjectAmendment(target_project.project_name))
      new_issue_content = 'Copied %s from %s' % (copied_issue_ref, issue_ref)
      self.services.issue.CreateIssueComment(
          self.mc.cnxn, copied_issue, self.mc.auth.user_id, new_issue_content,
          amendments=amendments)

      tracker_fulltext.IndexIssues(
          self.mc.cnxn, [copied_issue], self.services.user, self.services.issue,
          self.services.config)

    return copied_issue

  def _MergeLinkedAccounts(self, me_user_id):
    """Return a list of the given user ID and any linked accounts."""
    if not me_user_id:
      return []

    result = [me_user_id]
    me_user = self.services.user.GetUser(self.mc.cnxn, me_user_id)
    if me_user:
      if me_user.linked_parent_id:
        result.append(me_user.linked_parent_id)
      result.extend(me_user.linked_child_ids)
    return result

  def ListIssues(self, query_string, query_project_names, me_user_id,
                 items_per_page, paginate_start, url_params, can,
                 group_by_spec, sort_spec, use_cached_searches,
                 display_mode=None, project=None):
    """Do an issue search w/ mc + passed in args to return a pipeline object."""
    # Permission to view a project is checked in Frontendsearchpipeline().
    # Individual results are filtered by permissions in SearchForIIDs().

    with self.mc.profiler.Phase('searching issues'):
      me_user_ids = self._MergeLinkedAccounts(me_user_id)
      pipeline = frontendsearchpipeline.FrontendSearchPipeline(
          self.mc.cnxn, self.services, self.mc.auth, me_user_ids,
          query_string, query_project_names, items_per_page, paginate_start,
          url_params, can, group_by_spec, sort_spec, self.mc.warnings,
          self.mc.errors, use_cached_searches, self.mc.profiler,
          display_mode=display_mode, project=project)
      if not self.mc.errors.AnyErrors():
        pipeline.SearchForIIDs()
        pipeline.MergeAndSortIssues()
        pipeline.Paginate()
      # TODO(jojwang): raise InvalidQueryException.
      return pipeline

  # TODO(jrobbins): This method also requires self.mc to be a MonorailRequest.
  def FindIssuePositionInSearch(self, issue):
    """Do an issue search and return flipper info for the given issue.

    Args:
      issue: issue that the user is currently viewing.

    Returns:
      A 4-tuple of flipper info: (prev_iid, cur_index, next_iid, total_count).
    """
    # Permission to view a project is checked in Frontendsearchpipeline().
    # Individual results are filtered by permissions in SearchForIIDs().

    with self.mc.profiler.Phase('finding issue position in search'):
      url_params = [(name, self.mc.GetParam(name)) for name in
                    framework_helpers.RECOGNIZED_PARAMS]
      me_user_ids = self._MergeLinkedAccounts(self.mc.me_user_id)
      pipeline = frontendsearchpipeline.FrontendSearchPipeline(
           self.mc.cnxn, self.services, self.mc.auth, me_user_ids,
           self.mc.query, self.mc.query_project_names, self.mc.num,
           self.mc.start, url_params, self.mc.can, self.mc.group_by_spec,
           self.mc.sort_spec, self.mc.warnings, self.mc.errors,
           self.mc.use_cached_searches, self.mc.profiler,
           display_mode=self.mc.mode, project=self.mc.project)
      if not self.mc.errors.AnyErrors():
        # Only do the search if the user's query parsed OK.
        pipeline.SearchForIIDs()

      # Note: we never call MergeAndSortIssues() because we don't need a unified
      # sorted list, we only need to know the position on such a list of the
      # current issue.
      prev_iid, cur_index, next_iid = pipeline.DetermineIssuePosition(issue)
      return prev_iid, cur_index, next_iid, pipeline.total_count

  def GetIssuesDict(self, issue_ids, use_cache=True,
                    allow_viewing_deleted=False):
    """Return a dict {iid: issue} with the specified issues, if allowed.

    Args:
      issue_ids: int global issue IDs.
      use_cache: set to false to ensure fresh issues.
      allow_viewing_deleted: set to true to allow user to view deleted issues.

    Returns:
      A dict {issue_id: issue} for only those issues that the user is allowed
      to view.
    """
    with self.mc.profiler.Phase('getting issues %r' % issue_ids):
      issues = self.services.issue.GetIssuesDict(
          self.mc.cnxn, issue_ids, use_cache=use_cache)

    if len(issues) != len(set(issue_ids)):
      raise exceptions.NoSuchIssueException()

    issues = {
        issue_id: issue
        for issue_id, issue in issues.items()
        if self._UserCanViewIssue(issue, allow_viewing_deleted)[-1]}
    return issues

  def GetIssue(self, issue_id, use_cache=True, allow_viewing_deleted=False):
    """Return the specified issue.

    Args:
      issue_id: int global issue ID.
      use_cache: set to false to ensure fresh issue.
      allow_viewing_deleted: set to true to allow user to view a deleted issue.

    Returns:
      The requested Issue PB.
    """
    if issue_id is None:
      raise exceptions.InputException('No issue issue_id specified')

    with self.mc.profiler.Phase('getting issue %r' % issue_id):
      issue = self.services.issue.GetIssue(
          self.mc.cnxn, issue_id, use_cache=use_cache)

    self._AssertUserCanViewIssue(
        issue, allow_viewing_deleted=allow_viewing_deleted)
    return issue

  def ListReferencedIssues(self, ref_tuples, default_project_name):
    """Return the specified issues."""
    # Make sure ref_tuples are unique, preserving order.
    ref_tuples = list(collections.OrderedDict(
        list(zip(ref_tuples, ref_tuples))))
    ref_projects = self.services.project.GetProjectsByName(
        self.mc.cnxn,
        [(ref_pn or default_project_name) for ref_pn, _ in ref_tuples])
    issue_ids, _misses = self.services.issue.ResolveIssueRefs(
        self.mc.cnxn, ref_projects, default_project_name, ref_tuples)
    open_issues, closed_issues = (
        tracker_helpers.GetAllowedOpenedAndClosedIssues(
            self.mc, issue_ids, self.services))
    return open_issues, closed_issues

  def ListApplicableFieldDefs(self, issue_ids, config):
    """Return the applicable FieldDefs for the given issue_ids."""
    issues_dict = self.GetIssuesDict(issue_ids)
    issue_labels = []
    issue_approval_ids = []
    for issue in issues_dict.values():
      issue_labels.extend(issue.labels)
      issue_approval_ids.extend(
          [approval.approval_id for approval in issue.approval_values])
    labels_by_prefix = tracker_bizobj.LabelsByPrefix(
        list(set(issue_labels)), [])
    types = set(labels_by_prefix.get('type', []))
    types_lower = [t.lower() for t in types]
    applicable_fds = []
    for fd in config.field_defs:
      if fd.field_id in issue_approval_ids:
        applicable_fds.append(fd)
      elif fd.field_type != tracker_pb2.FieldTypes.APPROVAL_TYPE and (
          not fd.applicable_type or fd.applicable_type.lower() in types_lower):
        applicable_fds.append(fd)
    return applicable_fds

  def GetIssueByLocalID(
      self, project_id, local_id, use_cache=True,
      allow_viewing_deleted=False):
    """Return the specified issue, TODO: iff the signed in user may view it.

    Args:
      project_id: int project ID of the project that contains the issue.
      local_id: int issue local id number.
      use_cache: set to False when doing read-modify-write operations.
      allow_viewing_deleted: set to True to return a deleted issue so that
          an authorized user may undelete it.

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

    with self.mc.profiler.Phase('getting issue %r:%r' % (project_id, local_id)):
      issue = self.services.issue.GetIssueByLocalID(
          self.mc.cnxn, project_id, local_id, use_cache=use_cache)

    self._AssertUserCanViewIssue(
        issue, allow_viewing_deleted=allow_viewing_deleted)
    return issue

  def GetRelatedIssueRefs(self, issues):
    """Return a dict {iid: (project_name, local_id)} for all related issues."""
    related_iids = set()
    with self.mc.profiler.Phase('getting related issue refs'):
      for issue in issues:
        related_iids.update(issue.blocked_on_iids)
        related_iids.update(issue.blocking_iids)
        if issue.merged_into:
          related_iids.add(issue.merged_into)
      logging.info('related_iids is %r', related_iids)
      return self.services.issue.LookupIssueRefs(self.mc.cnxn, related_iids)

  def GetIssueRefs(self, issue_ids):
    """Return a dict {iid: (project_name, local_id)} for all issue_ids."""
    return self.services.issue.LookupIssueRefs(self.mc.cnxn, issue_ids)

  def BulkUpdateIssueApprovals(self, issue_ids, approval_id, project,
                               approval_delta, comment_content,
                               send_email):
    """Update all given issues' specified approval."""
    # Anon users and users with no permission to view the project
    # will get permission denied. Missing permissions to update
    # individual issues will not throw exceptions. Issues will just not be
    # updated.
    if not self.mc.auth.user_id:
      raise permissions.PermissionException('Anon cannot make changes')
    if not self._UserCanViewProject(project):
      raise permissions.PermissionException('User cannot view project')
    updated_issue_ids = []
    for issue_id in issue_ids:
      try:
        self.UpdateIssueApproval(
            issue_id, approval_id, approval_delta, comment_content, False,
            send_email=False)
        updated_issue_ids.append(issue_id)
      except exceptions.NoSuchIssueApprovalException as e:
        logging.info('Skipping issue %s, no approval: %s', issue_id, e)
      except permissions.PermissionException as e:
        logging.info('Skipping issue %s, update not allowed: %s', issue_id, e)
    # TODO(jojwang): send bulk approval update email if send_email
    if send_email:
      pass
    return updated_issue_ids

  def UpdateIssueApproval(self, issue_id, approval_id, approval_delta,
                          comment_content, is_description, attachments=None,
                          send_email=True, kept_attachments=None):
    """Update an issue's approval."""

    issue, approval_value = self.services.issue.GetIssueApproval(
        self.mc.cnxn, issue_id, approval_id, use_cache=False)

    self._AssertPermInIssue(issue, permissions.EDIT_ISSUE)
    project = self.GetProject(issue.project_id)
    config = self.GetProjectConfig(issue.project_id)

    if attachments:
      with self.mc.profiler.Phase('Accounting for quota'):
        new_bytes_used = tracker_helpers.ComputeNewQuotaBytesUsed(
          project, attachments)
        self.services.project.UpdateProject(
          self.mc.cnxn, issue.project_id, attachment_bytes_used=new_bytes_used)

    if kept_attachments:
      with self.mc.profiler.Phase('Filtering kept attachments'):
        kept_attachments = tracker_helpers.FilterKeptAttachments(
            is_description, kept_attachments, self.ListIssueComments(issue),
            approval_id)

    if approval_delta.status:
      if not permissions.CanUpdateApprovalStatus(
          self.mc.auth.effective_ids, self.mc.perms, project,
          approval_value.approver_ids, approval_delta.status):
        raise permissions.PermissionException(
            'User not allowed to make this status update.')

    if approval_delta.approver_ids_remove or approval_delta.approver_ids_add:
      if not permissions.CanUpdateApprovers(
          self.mc.auth.effective_ids, self.mc.perms, project,
          approval_value.approver_ids):
        raise permissions.PermissionException(
            'User not allowed to modify approvers of this approval.')

    with self.mc.profiler.Phase(
        'updating approval for issue %r, aprpoval %r' % (
            issue_id, approval_id)):
      comment_pb = self.services.issue.DeltaUpdateIssueApproval(
          self.mc.cnxn, self.mc.auth.user_id, config, issue, approval_value,
          approval_delta, comment_content=comment_content,
          is_description=is_description, attachments=attachments,
          kept_attachments=kept_attachments)
      hostport = framework_helpers.GetHostPort(
          project_name=project.project_name)
      send_notifications.PrepareAndSendApprovalChangeNotification(
          issue_id, approval_id, hostport, comment_pb.id,
          send_email=send_email)

    return approval_value, comment_pb

  def ConvertIssueApprovalsTemplate(
      self, config, issue, template_name, comment_content, send_email=True):
    """Convert an issue's existing approvals structure to match the one of
       the given template."""
    self._AssertPermInIssue(issue, permissions.EDIT_ISSUE)

    template = self.services.template.GetTemplateByName(
        self.mc.cnxn, template_name, issue.project_id)
    if not template:
      raise exceptions.NoSuchTemplateException(
          'Template %s is not found' % template_name)

    with self.mc.profiler.Phase('updating issue %r' % issue):
      comment_pb = self.services.issue.UpdateIssueStructure(
          self.mc.cnxn, config, issue, template, self.mc.auth.user_id,
          comment_content)
      hostport = framework_helpers.GetHostPort(project_name=issue.project_name)
      send_notifications.PrepareAndSendIssueChangeNotification(
          issue.issue_id, hostport, self.mc.auth.user_id,
          send_email=send_email, comment_id=comment_pb.id)

  def UpdateIssue(
      self, issue, delta, comment_content, attachments=None, send_email=True,
      is_description=False, kept_attachments=None, inbound_message=None):
    """Update an issue with a set of changes and add a comment.

    Args:
      issue: Existing Issue PB for the issue to be modified.
      delta: IssueDelta object containing all the changes to be made.
      comment_content: string content of the user's comment.
      attachments: List [(filename, contents, mimetype),...] of attachments.
      send_email: set to False to suppress email notifications.
      is_description: True if this adds a new issue description.
      kept_attachments: This should be a list of int attachment ids for
          attachments kept from previous descriptions, if the comment is
          a change to the issue description.
      inbound_message: optional string full text of an email that caused
          this comment to be added.

    Returns:
      Nothing.
    """
    if not self._UserCanUsePermInIssue(issue, permissions.EDIT_ISSUE):
      # We're editing the issue description. Only users with EditIssue
      # permission can edit the description.
      if is_description:
        raise permissions.PermissionException(
            'Users lack permission EditIssue in issue')
      # If we're adding a comment, we must have AddIssueComment permission.
      if comment_content:
        self._AssertPermInIssue(issue, permissions.ADD_ISSUE_COMMENT)
      # If we're modifying the issue, check that we only modify the fields we're
      # allowed to edit.
      if delta != tracker_pb2.IssueDelta():
        allowed_delta = tracker_pb2.IssueDelta()
        if self._UserCanUsePermInIssue(issue, permissions.EDIT_ISSUE_STATUS):
          allowed_delta.status = delta.status
        if self._UserCanUsePermInIssue(issue, permissions.EDIT_ISSUE_SUMMARY):
          allowed_delta.summary = delta.summary
        if self._UserCanUsePermInIssue(issue, permissions.EDIT_ISSUE_OWNER):
          allowed_delta.owner_id = delta.owner_id
        if self._UserCanUsePermInIssue(issue, permissions.EDIT_ISSUE_CC):
          allowed_delta.cc_ids_add = delta.cc_ids_add
          allowed_delta.cc_ids_remove = delta.cc_ids_remove
        if delta != allowed_delta:
          raise permissions.PermissionException(
              'Users lack permission EditIssue in issue')

    if delta.merged_into:
      # Reject attempts to merge an issue into an issue we cannot view and edit.
      merged_into_issue = self.GetIssue(
          delta.merged_into, use_cache=False, allow_viewing_deleted=True)
      self._AssertPermInIssue(issue, permissions.EDIT_ISSUE)
      # Reject attempts to merge an issue into itself.
      if issue.issue_id == delta.merged_into:
        raise exceptions.InputException(
          'Cannot merge an issue into itself.')

    # Reject attempts to block on issue on itself.
    if (issue.issue_id in delta.blocked_on_add
        or issue.issue_id in delta.blocking_add):
      raise exceptions.InputException(
        'Cannot block an issue on itself.')

    project = self.GetProject(issue.project_id)
    config = self.GetProjectConfig(issue.project_id)
    old_owner_id = tracker_bizobj.GetOwnerId(issue)

    if attachments:
      with self.mc.profiler.Phase('Accounting for quota'):
        new_bytes_used = tracker_helpers.ComputeNewQuotaBytesUsed(
            project, attachments)
        self.services.project.UpdateProject(
            self.mc.cnxn, issue.project_id,
            attachment_bytes_used=new_bytes_used)

    with self.mc.profiler.Phase('Validating the issue change'):
      # If the owner changed, it must be a project member.
      if (delta.owner_id is not None and delta.owner_id != issue.owner_id):
        parsed_owner_valid, msg = tracker_helpers.IsValidIssueOwner(
          self.mc.cnxn, project, delta.owner_id, self.services)
        if not parsed_owner_valid:
          raise exceptions.InputException(msg)

    if kept_attachments:
      with self.mc.profiler.Phase('Filtering kept attachments'):
        kept_attachments = tracker_helpers.FilterKeptAttachments(
            is_description, kept_attachments, self.ListIssueComments(issue),
            None)

    with self.mc.profiler.Phase('Updating issue %r' % (issue.issue_id)):
      _amendments, comment_pb = self.services.issue.DeltaUpdateIssue(
          self.mc.cnxn, self.services, self.mc.auth.user_id, issue.project_id,
          config, issue, delta, comment=comment_content,
          attachments=attachments, is_description=is_description,
          kept_attachments=kept_attachments, inbound_message=inbound_message)

    with self.mc.profiler.Phase('Following up after issue update'):
      if delta.merged_into:
        new_starrers = tracker_helpers.GetNewIssueStarrers(
            self.mc.cnxn, self.services, issue.issue_id,
            delta.merged_into)
        merged_into_project = self.GetProject(merged_into_issue.project_id)
        tracker_helpers.AddIssueStarrers(
            self.mc.cnxn, self.services, self.mc,
            delta.merged_into, merged_into_project, new_starrers)
        # Load target issue again to get the updated star count.
        merged_into_issue = self.GetIssue(
            merged_into_issue.issue_id, use_cache=False)
        tracker_helpers.MergeCCsAndAddComment(
            self.services, self.mc, issue, merged_into_issue)
      self.services.project.UpdateRecentActivity(
          self.mc.cnxn, issue.project_id)

    with self.mc.profiler.Phase('Generating notifications'):
      if comment_pb:
        hostport = framework_helpers.GetHostPort(
            project_name=project.project_name)
        reporter_id = self.mc.auth.user_id
        send_notifications.PrepareAndSendIssueChangeNotification(
            issue.issue_id, hostport, reporter_id,
            send_email=send_email, old_owner_id=old_owner_id,
            comment_id=comment_pb.id)
        delta_blocked_on_iids = delta.blocked_on_add + delta.blocked_on_remove
        send_notifications.PrepareAndSendIssueBlockingNotification(
            issue.issue_id, hostport, delta_blocked_on_iids,
            reporter_id, send_email=send_email)

  def DeleteIssue(self, issue, delete):
    """Mark or unmark the given issue as deleted."""
    self._AssertPermInIssue(issue, permissions.DELETE_ISSUE)

    with self.mc.profiler.Phase('Marking issue %r deleted' % (issue.issue_id)):
      self.services.issue.SoftDeleteIssue(
          self.mc.cnxn, issue.project_id, issue.local_id, delete,
          self.services.user)

  def FlagIssues(self, issues, flag):
    """Flag or unflag the given issues as spam."""
    for issue in issues:
      self._AssertPermInIssue(issue, permissions.FLAG_SPAM)

    issue_ids = [issue.issue_id for issue in issues]
    with self.mc.profiler.Phase('Marking issues %r as spam' % issue_ids):
      self.services.spam.FlagIssues(
          self.mc.cnxn, self.services.issue, issues, self.mc.auth.user_id,
          flag)
      if self._UserCanUsePermInIssue(issue, permissions.VERDICT_SPAM):
        self.services.spam.RecordManualIssueVerdicts(
            self.mc.cnxn, self.services.issue, issues, self.mc.auth.user_id,
            flag)

  def LookupIssuesFlaggers(self, issues):
    """Returns users who've reported the issue or its comments as spam.

    Args:
      issues: the list of issues to query.
    Returns:
      A dictionary
        {issue_id: ([issue_reporters], {comment_id: [comment_reporters]})}
      For each issue id, a tuple with the users who have flagged the issue;
      and a dictionary of users who have flagged a comment for each comment id.
    """
    for issue in issues:
      self._AssertUserCanViewIssue(issue)

    issue_ids = [issue.issue_id for issue in issues]
    with self.mc.profiler.Phase('Looking up flaggers for %s' % issue_ids):
      reporters = self.services.spam.LookupIssuesFlaggers(
          self.mc.cnxn, issue_ids)

    return reporters

  def LookupIssueFlaggers(self, issue):
    """Returns users who've reported the issue or its comments as spam.

    Args:
      issue: the issue to query.
    Returns:
      A tuple
        ([issue_reporters], {comment_id: [comment_reporters]})
      With the users who have flagged the issue; and a dictionary of users who
      have flagged a comment for each comment id.
    """
    return self.LookupIssuesFlaggers([issue])[issue.issue_id]

  def GetIssuePositionInHotlist(
      self, current_issue, hotlist, can, sort_spec, group_by_spec):
    # type: (Issue, Hotlist, int, str, str) -> (int, int, int, int)
    """Get index info of an issue within a hotlist.

    Args:
      current_issue: the currently viewed issue.
      hotlist: the hotlist this flipper is flipping through.
      can: int "canned query" number to scope the visible issues.
      sort_spec: string that lists the sort order.
      group_by_spec: string that lists the grouping order.
    """
    issues_list = self.services.issue.GetIssues(self.mc.cnxn,
        [item.issue_id for item in hotlist.items])
    project_ids = hotlist_helpers.GetAllProjectsOfIssues(issues_list)
    config_list = hotlist_helpers.GetAllConfigsOfProjects(
        self.mc.cnxn, project_ids, self.services)
    harmonized_config = tracker_bizobj.HarmonizeConfigs(config_list)
    (sorted_issues, _hotlist_issues_context,
     _users) = hotlist_helpers.GetSortedHotlistIssues(
         self.mc.cnxn, hotlist.items, issues_list, self.mc.auth,
         can, sort_spec, group_by_spec, harmonized_config, self.services,
         self.mc.profiler)
    (prev_iid, cur_index,
     next_iid) = features_bizobj.DetermineHotlistIssuePosition(
         current_issue, [issue.issue_id for issue in sorted_issues])
    total_count = len(sorted_issues)
    return prev_iid, cur_index, next_iid, total_count

  def RerankBlockedOnIssues(self, issue, moved_id, target_id, split_above):
    """Rerank the blocked on issues for issue_id.

    Args:
      issue: The issue to modify.
      moved_id: The id of the issue to move.
      target_id: The id of the issue to move |moved_issue| to.
      split_above: Whether to move |moved_issue| before or after |target_issue|.
    """
    # Make sure the user has permission to edit the issue.
    self._AssertPermInIssue(issue, permissions.EDIT_ISSUE)
    # Make sure the moved and target issues are in the blocked-on list.
    if moved_id not in issue.blocked_on_iids:
      raise exceptions.InputException(
          'The issue to move is not in the blocked-on list.')
    if target_id not in issue.blocked_on_iids:
      raise exceptions.InputException(
          'The target issue is not in the blocked-on list.')

    phase_name = 'Moving issue %r %s issue %d.' % (
        moved_id, 'above' if split_above else 'below', target_id)
    with self.mc.profiler.Phase(phase_name):
      lower, higher = tracker_bizobj.SplitBlockedOnRanks(
          issue, target_id, split_above,
          [iid for iid in issue.blocked_on_iids if iid != moved_id])
      rank_changes = rerank_helpers.GetInsertRankings(
          lower, higher, [moved_id])
      if rank_changes:
        self.services.issue.ApplyIssueRerank(
            self.mc.cnxn, issue.issue_id, rank_changes)

  # FUTURE: GetIssuePermissionsForUser()

  # FUTURE: CreateComment()

  def ListIssueComments(self, issue):
    """Return comments on the specified viewable issue."""
    self._AssertUserCanViewIssue(issue)

    with self.mc.profiler.Phase('getting comments for %r' % issue.issue_id):
      comments = self.services.issue.GetCommentsForIssue(
          self.mc.cnxn, issue.issue_id)

    return comments

  # FUTURE: UpdateComment()

  def DeleteComment(self, issue, comment, delete):
    """Mark or unmark a comment as deleted by the current user."""
    self._AssertUserCanDeleteComment(issue, comment)
    if comment.is_spam and self.mc.auth.user_id == comment.user_id:
      raise permissions.PermissionException('Cannot delete comment.')

    with self.mc.profiler.Phase(
        'deleting issue %r comment %r' % (issue.issue_id, comment.id)):
      self.services.issue.SoftDeleteComment(
          self.mc.cnxn, issue, comment, self.mc.auth.user_id,
          self.services.user, delete=delete)

  def DeleteAttachment(self, issue, comment, attachment_id, delete):
    """Mark or unmark a comment attachment as deleted by the current user."""
    # A user can delete an attachment iff they can delete a comment.
    self._AssertUserCanDeleteComment(issue, comment)

    phase_message = 'deleting issue %r comment %r attachment %r' % (
        issue.issue_id, comment.id, attachment_id)
    with self.mc.profiler.Phase(phase_message):
      self.services.issue.SoftDeleteAttachment(
          self.mc.cnxn, issue, comment, attachment_id, self.services.user,
          delete=delete)

  def FlagComment(self, issue, comment, flag):
    """Mark or unmark a comment as spam."""
    self._AssertPermInIssue(issue, permissions.FLAG_SPAM)
    with self.mc.profiler.Phase(
        'flagging issue %r comment %r' % (issue.issue_id, comment.id)):
      self.services.spam.FlagComment(
          self.mc.cnxn, issue.issue_id, comment.id, comment.user_id,
          self.mc.auth.user_id, flag)
      if self._UserCanUsePermInIssue(issue, permissions.VERDICT_SPAM):
        self.services.spam.RecordManualCommentVerdict(
            self.mc.cnxn, self.services.issue, self.services.user, comment.id,
            self.mc.auth.user_id, flag)

  def StarIssue(self, issue, starred):
    """Set or clear a star on the given issue for the signed in user."""
    if not self.mc.auth.user_id:
      raise permissions.PermissionException('Anon cannot star issues')
    self._AssertPermInIssue(issue, permissions.SET_STAR)

    with self.mc.profiler.Phase('starring issue %r' % issue.issue_id):
      config = self.services.config.GetProjectConfig(
          self.mc.cnxn, issue.project_id)
      self.services.issue_star.SetStar(
          self.mc.cnxn, self.services, config, issue.issue_id,
          self.mc.auth.user_id, starred)

  def IsIssueStarred(self, issue, cnxn=None):
    """Return True if the given issue is starred by the signed in user."""
    self._AssertUserCanViewIssue(issue)

    with self.mc.profiler.Phase('checking star %r' % issue.issue_id):
      return self.services.issue_star.IsItemStarredBy(
          cnxn or self.mc.cnxn, issue.issue_id, self.mc.auth.user_id)

  def ListStarredIssueIDs(self):
    """Return a list of the issue IDs that the current issue has starred."""
    # This returns an unfiltered list of issue_ids.  Permissions will be
    # applied if and when the caller attempts to load each issue.

    with self.mc.profiler.Phase('getting stars %r' % self.mc.auth.user_id):
      return self.services.issue_star.LookupStarredItemIDs(
          self.mc.cnxn, self.mc.auth.user_id)

  def SnapshotCountsQuery(self, project, timestamp, group_by, label_prefix=None,
                          query=None, canned_query=None, hotlist=None):
    """Query IssueSnapshots for daily counts.

    See chart_svc.QueryIssueSnapshots for more detail on arguments.

    Args:
      project (Project): Project to search.
      timestamp (int): Will query for snapshots at this timestamp.
      group_by (str): 2nd dimension, see QueryIssueSnapshots for options.
      label_prefix (str): Required for label queries. Only returns results
        with the supplied prefix.
      query (str, optional): If supplied, will parse & apply query conditions.
      canned_query (str, optional): Parsed canned query.
      hotlist (Hotlist, optional): Hotlist to search under (in lieu of project).

    Returns:
      1. A dict of {name: count} for each item in group_by.
      2. A list of any unsupported query conditions in query.
    """
    # This returns counts of viewable issues.
    with self.mc.profiler.Phase('querying snapshot counts'):
      return self.services.chart.QueryIssueSnapshots(
        self.mc.cnxn, self.services, timestamp, self.mc.auth.effective_ids,
        project, self.mc.perms, group_by=group_by, label_prefix=label_prefix,
        query=query, canned_query=canned_query, hotlist=hotlist)

  ### User methods

  def GetUser(self, user_id):
    """Return the user with the given ID."""
    # Make sure the requested user exists.
    with self.mc.profiler.Phase('getting user %s' % user_id):
      self.services.user.LookupUserEmail(self.mc.cnxn, user_id)
      return self.services.user.GetUser(self.mc.cnxn, user_id)

  def GetMemberships(self, user_id):
    """Return the user group ids for the given user visible to the requester."""
    group_ids = self.services.usergroup.LookupMemberships(self.mc.cnxn, user_id)
    if user_id == self.mc.auth.user_id:
      return group_ids
    (member_ids_by_ids, owner_ids_by_ids
    ) = self.services.usergroup.LookupAllMembers(
        self.mc.cnxn, group_ids)
    settings_by_id = self.services.usergroup.GetAllGroupSettings(
        self.mc.cnxn, group_ids)

    (owned_project_ids, membered_project_ids,
     contrib_project_ids) = self.services.project.GetUserRolesInAllProjects(
         self.mc.cnxn, self.mc.auth.effective_ids)
    project_ids = owned_project_ids.union(
        membered_project_ids).union(contrib_project_ids)

    visible_group_ids = []
    for group_id, group_settings in settings_by_id.items():
      member_ids = member_ids_by_ids.get(group_id)
      owner_ids = owner_ids_by_ids.get(group_id)
      if permissions.CanViewGroupMembers(
          self.mc.perms, self.mc.auth.effective_ids, group_settings,
          member_ids, owner_ids, project_ids):
        visible_group_ids.append(group_id)

    return visible_group_ids

  def ListReferencedUsers(self, emails):
    """Return a list of the given emails' User PBs, plus linked account ids.

    Args:
      emails: list of emails of users to look up.

    Returns:
      A pair (users, linked_users_ids) where users is an unsorted list of
      User PBs and linked_user_ids is a list of user IDs of any linked accounts.
    """
    with self.mc.profiler.Phase('getting existing users'):
      user_id_dict = self.services.user.LookupExistingUserIDs(
          self.mc.cnxn, emails)
      users_by_id = self.services.user.GetUsersByIDs(
          self.mc.cnxn, list(user_id_dict.values()))
      user_list = list(users_by_id.values())

      linked_user_ids = []
      for user in user_list:
        if user.linked_parent_id:
          linked_user_ids.append(user.linked_parent_id)
        linked_user_ids.extend(user.linked_child_ids)

    return user_list, linked_user_ids

  def StarUser(self, user_id, starred):
    """Star or unstar the specified user.

    Args:
      user_id: int ID of the user to star/unstar.
      starred: true to add a star, false to remove it.

    Returns:
      Nothing.

    Raises:
      NoSuchUserException: There is no user with that ID.
    """
    if not self.mc.auth.user_id:
      raise exceptions.InputException('No current user specified')

    with self.mc.profiler.Phase('(un)starring user %r' % user_id):
      # Make sure the user exists and user has permission to see it.
      self.services.user.LookupUserEmail(self.mc.cnxn, user_id)
      self.services.user_star.SetStar(
          self.mc.cnxn, user_id, self.mc.auth.user_id, starred)

  def IsUserStarred(self, user_id):
    """Return True if the current user has starred the given user.

    Args:
      user_id: int ID of the user to check.

    Returns:
      True if starred.

    Raises:
      NoSuchUserException: There is no user with that ID.
    """
    if user_id is None:
      raise exceptions.InputException('No user specified')

    if not self.mc.auth.user_id:
      return False

    with self.mc.profiler.Phase('checking user star %r' % user_id):
      # Make sure the user exists.
      self.services.user.LookupUserEmail(self.mc.cnxn, user_id)
      return self.services.user_star.IsItemStarredBy(
        self.mc.cnxn, user_id, self.mc.auth.user_id)

  def GetUserStarCount(self, user_id):
    """Return the number of times the user has been starred.

    Args:
      user_id: int ID of the user to check.

    Returns:
      The number of times the user has been starred.

    Raises:
      NoSuchUserException: There is no user with that ID.
    """
    if user_id is None:
      raise exceptions.InputException('No user specified')

    with self.mc.profiler.Phase('counting stars for user %r' % user_id):
      # Make sure the user exists.
      self.services.user.LookupUserEmail(self.mc.cnxn, user_id)
      return self.services.user_star.CountItemStars(self.mc.cnxn, user_id)

  def GetPendingLinkedInvites(self, user_id=None):
    """Return info about a user's linked account invites."""
    with self.mc.profiler.Phase('checking linked account invites'):
      result = self.services.user.GetPendingLinkedInvites(
          self.mc.cnxn, user_id or self.mc.auth.user_id)
      return result

  def InviteLinkedParent(self, parent_email):
    """Invite a matching account to be my parent."""
    if not parent_email:
      raise exceptions.InputException('No parent account specified')
    if not self.mc.auth.user_id:
      raise permissions.PermissionException('Anon cannot link accounts')
    with self.mc.profiler.Phase('Validating proposed parent'):
      # We only offer self-serve account linking to matching usernames.
      (p_username, p_domain,
       _obs_username, _obs_email) = framework_views.ParseAndObscureAddress(
          parent_email)
      c_view = self.mc.auth.user_view
      if p_username != c_view.username:
        logging.info('Username %r != %r', p_username, c_view.username)
        raise exceptions.InputException('Linked account names must match')
      allowed_domains = settings.linkable_domains.get(c_view.domain, [])
      if p_domain not in allowed_domains:
        logging.info('parent domain %r is not in list for %r: %r',
                     p_domain, c_view.domain, allowed_domains)
        raise exceptions.InputException('Linked account unsupported domain')
      parent_id = self.services.user.LookupUserID(self.mc.cnxn, parent_email)
    with self.mc.profiler.Phase('Creating linked account invite'):
      self.services.user.InviteLinkedParent(
          self.mc.cnxn, parent_id, self.mc.auth.user_id)

  def AcceptLinkedChild(self, child_id):
    """Accept an invitation from a child account."""
    with self.mc.profiler.Phase('Accept linked account invite'):
      self.services.user.AcceptLinkedChild(
          self.mc.cnxn, self.mc.auth.user_id, child_id)

  def UnlinkAccounts(self, parent_id, child_id):
    """Delete a linked-account relationship."""
    if (self.mc.auth.user_id != parent_id and
        self.mc.auth.user_id != child_id):
      permitted = self.mc.perms.CanUsePerm(
        permissions.EDIT_OTHER_USERS, self.mc.auth.effective_ids, None, [])
      if not permitted:
        raise permissions.PermissionException(
          'User lacks permission to unlink accounts')

    with self.mc.profiler.Phase('Unlink accounts'):
      self.services.user.UnlinkAccounts(self.mc.cnxn, parent_id, child_id)

  def UpdateUserSettings(self, user, **kwargs):
    """Update the preferences of the specified user.

    Args:
      user: User PB for the user to update.
      keyword_args: dictionary of setting names mapped to new values.
    """
    if not user or not user.user_id:
      raise exceptions.InputException('Cannot update user settings for anon.')

    with self.mc.profiler.Phase(
        'updating settings for %s with %s' % (self.mc.auth.user_id, kwargs)):
      self.services.user.UpdateUserSettings(
          self.mc.cnxn, user.user_id, user, **kwargs)

  def GetUserPrefs(self, user_id):
    """Get the UserPrefs for the specified user."""
    # Anon user always has default prefs.
    if not user_id:
      return user_pb2.UserPrefs(user_id=0)
    if user_id != self.mc.auth.user_id:
      if not self.mc.perms.HasPerm(permissions.EDIT_OTHER_USERS, None, None):
        raise permissions.PermissionException(
            'Only site admins may see other users\' preferences')
    with self.mc.profiler.Phase('Getting prefs for %s' % user_id):
      userprefs = self.services.user.GetUserPrefs(self.mc.cnxn, user_id)

    # Hard-coded user prefs for at-risk users that should use "corp mode".
    # TODO(jrobbins): Remove this when user group preferences are implemented.
    if framework_bizobj.IsCorpUser(self.mc.cnxn, self.services, user_id):
      # Copy so that cached version is not modified.
      userprefs = user_pb2.UserPrefs(user_id=user_id, prefs=userprefs.prefs)
      pref_names = {pref.name for pref in userprefs.prefs}
      if 'restrict_new_issues' not in pref_names:
        userprefs.prefs.append(user_pb2.UserPrefValue(
            name='restrict_new_issues', value='true'))
      if 'public_issue_notice' not in pref_names:
        userprefs.prefs.append(user_pb2.UserPrefValue(
            name='public_issue_notice', value='true'))

    return userprefs

  def SetUserPrefs(self, user_id, prefs):
    """Set zero or more UserPrefValue for the specified user."""
    # Anon user always has default prefs.
    if not user_id:
      raise exceptions.InputException('Anon cannot have prefs')
    if user_id != self.mc.auth.user_id:
      if not self.mc.perms.HasPerm(permissions.EDIT_OTHER_USERS, None, None):
        raise permissions.PermissionException(
            'Only site admins may set other users\' preferences')
    for pref in prefs:
      error_msg = framework_bizobj.ValidatePref(pref.name, pref.value)
      if error_msg:
        raise exceptions.InputException(error_msg)
    with self.mc.profiler.Phase(
        'setting prefs for %s' % (self.mc.auth.user_id)):
      self.services.user.SetUserPrefs(self.mc.cnxn, user_id, prefs)

  # FUTURE: GetUser()
  # FUTURE: UpdateUser()
  # FUTURE: DeleteUser()
  # FUTURE: ListStarredUsers()

  def ExpungeUsers(self, emails, check_perms=True, commit=True):
    """Permanently deletes user data and removes remaining user references
       for all listed users.

      To avoid any executions that might take too long and make the site hang,
      a limit clause will be added to some operations. If any user references
      are left behind due to the cut-off, the final services.user.ExpungeUsers
      will fail because we cannot delete User rows that are still referenced
      in other tables. work_env.ExpungeUsers can be called again until all user
      references are removed and the final services.user.ExpungeUsers succeeds.
      The limit clause will not be applied in operations for tables that contain
      user_id or email columns but do not officially Reference the User table.
      E.g. SpamVerdict and SpamReport. These user references must all be removed
      before the attempt to delete rows from User is made. The limit will also
      not be applied for sets of operations where values removed in earlier
      operations would have to be known in order for later operations to
      succeed.  E.g. ExpungeUsersIngroups().
    """
    if check_perms:
      if not permissions.CanExpungeUsers(self.mc):
        raise permissions.PermissionException(
            'User is not allowed to delete users.')

    limit = 10000
    user_ids_by_email = self.services.user.LookupExistingUserIDs(
        self.mc.cnxn, emails)
    user_ids = list(user_ids_by_email.values())
    if framework_constants.DELETED_USER_ID in user_ids:
      raise exceptions.InputException(
          'Reserved deleted_user_id found in deletion request and'
          'should not be deleted')
    if not user_ids:
      logging.info('Emails %r not found in DB. No users deleted', emails)
      return

    # The operations made in the methods below can be limited.
    # We can adjust 'limit' as necessary to avoid timing out.
    self.services.issue_star.ExpungeStarsByUsers(
        self.mc.cnxn, user_ids, limit=limit)
    self.services.project_star.ExpungeStarsByUsers(
        self.mc.cnxn, user_ids, limit=limit)
    self.services.hotlist_star.ExpungeStarsByUsers(
        self.mc.cnxn, user_ids, limit=limit)
    self.services.user_star.ExpungeStarsByUsers(
        self.mc.cnxn, user_ids, limit=limit)
    for user_id in user_ids:
      self.services.user_star.ExpungeStars(
          self.mc.cnxn, user_id, commit=False, limit=limit)

    self.services.features.ExpungeQuickEditsByUsers(
        self.mc.cnxn, user_ids, limit=limit)
    self.services.features.ExpungeSavedQueriesByUsers(
        self.mc.cnxn, user_ids, limit=limit)

    self.services.template.ExpungeUsersInTemplates(
        self.mc.cnxn, user_ids, limit=limit)
    self.services.config.ExpungeUsersInConfigs(
        self.mc.cnxn, user_ids, limit=limit)

    self.services.project.ExpungeUsersInProjects(
        self.mc.cnxn, user_ids, limit=limit)

    # The upcoming operations cannot be limited with 'limit'.
    # So it's possible that these operations below may lead to timing out
    # and ExpungeUsers will have to run again to fully delete all users.
    # We commit the above operations here, so if a failure does happen
    # below, the second run of ExpungeUsers will have less work to do.
    if commit:
      self.mc.cnxn.Commit()

    affected_issue_ids = self.services.issue.ExpungeUsersInIssues(
        self.mc.cnxn, user_ids_by_email, limit=limit)
    # Commit ExpungeUsersInIssues here, as it has many operations
    # and at least one operation that cannot be limited.
    if commit:
      self.mc.cnxn.Commit()
      self.services.issue.EnqueueIssuesForIndexing(
          self.mc.cnxn, affected_issue_ids)

    # Spam verdict and report tables have user_id columns that do not
    # reference User. No limit will be applied.
    self.services.spam.ExpungeUsersInSpam(self.mc.cnxn, user_ids)
    if commit:
      self.mc.cnxn.Commit()

    # No limit will be applied for expunging in hotlists.
    self.services.features.ExpungeUsersInHotlists(
        self.mc.cnxn, user_ids, self.services.hotlist_star, self.services.user,
        self.services.chart)
    if commit:
      self.mc.cnxn.Commit()

    # No limit will be applied for expunging in UserGroups.
    self.services.usergroup.ExpungeUsersInGroups(
        self.mc.cnxn, user_ids)
    if commit:
      self.mc.cnxn.Commit()

    # No limit will be applied for expunging in FilterRules.
    deleted_rules_by_project = self.services.features.ExpungeFilterRulesByUser(
        self.mc.cnxn, user_ids_by_email)
    rule_strs_by_project = filterrules_helpers.BuildRedactedFilterRuleStrings(
        self.mc.cnxn, deleted_rules_by_project, self.services.user, emails)
    if commit:
      self.mc.cnxn.Commit()

    # We will attempt to expunge all given users here. Limiting the users we
    # delete should be done before work_env.ExpungeUsers is called.
    self.services.user.ExpungeUsers(self.mc.cnxn, user_ids)
    if commit:
      self.mc.cnxn.Commit()
      self.services.usergroup.group_dag.MarkObsolete()

    for project_id, filter_rule_strs in rule_strs_by_project.items():
      project = self.services.project.GetProject(self.mc.cnxn, project_id)
      hostport = framework_helpers.GetHostPort(
          project_name=project.project_name)
      send_notifications.PrepareAndSendDeletedFilterRulesNotification(
          project_id, hostport, filter_rule_strs)

  def TotalUsersCount(self):
    """Returns the total number of Users in Monorail."""
    return self.services.user.TotalUsersCount(self.mc.cnxn)

  def GetAllUserEmailsBatch(self, limit=1000, offset=0):
    """Returns a list emails that belong to Users in Monorail.

    Returns:
      A list of emails for Users within Monorail ordered by the user.user_ids.
      The list will hold at most [limit] emails and will start at the given
      [offset].
    """
    return self.services.user.GetAllUserEmailsBatch(
        self.mc.cnxn, limit=limit, offset=offset)

  ### Group methods

  # FUTURE: CreateGroup()
  # FUTURE: ListGroups()
  # FUTURE: UpdateGroup()
  # FUTURE: DeleteGroup()

  ### Hotlist methods

  def CreateHotlist(
      self, name, summary, description, editor_ids, issue_ids, is_private):
    """Create a hotlist.

    Args:
      name: a valid hotlist name.
      summary: one-line explanation of the hotlist.
      description: one-page explanation of the hotlist.
      editor_ids: a list of user IDs for the hotlist editors.
      issue_ids: a list of issue IDs for the hotlist issues.
      is_private: True if the hotlist can only be viewed by owners and editors.

    Returns:
      The newly created hotlist.

    Raises:
      HotlistAlreadyExists: A hotlist with the given name already exists.
      InputException: No user is signed in or the proposed name is invalid.
    """
    if not self.mc.auth.user_id:
      raise exceptions.InputException('Anon cannot create hotlists.')

    # Prevent users from adding issues they can't view to hotlists.
    issues_dict = self.services.issue.GetIssuesDict(self.mc.cnxn, issue_ids)
    for issue in issues_dict.values():
      self._AssertUserCanViewIssue(issue)

    with self.mc.profiler.Phase('creating hotlist %s' % name):
      hotlist = self.services.features.CreateHotlist(
          self.mc.cnxn, name, summary, description, [self.mc.auth.user_id],
          editor_ids, issue_ids, is_private, ts=int(time.time()))

    return hotlist

  def UpdateHotlistSettings(
      self, hotlist_id, name=None, summary=None,
      description=None, is_private=None, default_col_spec=None):
    # type: (int, str, str, str, bool, str) -> None
    """Update the given hotlist's settings.

    If a new value is None, the value in the hotlist does not get
    changed to None, it just does not get updated.

    Args:
      hotlist_id: int hotlist_id of the hotlist to update.
      name: str proposed new name for the hotlist.
      summary: str new summary for the hotlist.
      description: str new description for the hotlist.
      is_private: bool true if hotlist should be updated to private.
      default_col_spec: str new default columns for hotlist list view.

    Raises:
      InputException: The given hotlist_id is None or proposed new name is not
        a valid hotlist name.
      NoSuchHotlistException: There is no hotlist with the given ID.
      PermissionException: The logged-in user is not allowed to update
        this hotlist's settings.
      HotlistAlreadyExists: The hotlist owner already owns a hotlist
        with the proposed new name.
    """
    hotlist = self.GetHotlist(hotlist_id, use_cache=False)
    update_permitted = permissions.CanAdministerHotlist(
        self.mc.auth.effective_ids, self.mc.perms, hotlist)
    if not update_permitted:
      raise permissions.PermissionException(
          'User is not allowed to update hotlist settings.')

    if name is not None:
      if hotlist.name == name:
        name = None
      else:
        name = name.strip()
        if not framework_bizobj.IsValidHotlistName(name):
          raise exceptions.InputException(
              '"%s" is not a valid hotlist name' % name)
        if self.services.features.LookupHotlistIDs(
            self.mc.cnxn, [name], hotlist.owner_ids):
          raise features_svc.HotlistAlreadyExists(
              'Hotlist already owns a hotlist with name %s' % name)

    if ([name, summary, description, is_private, default_col_spec] ==
        [None, None, None, None, None]):
      logging.info('No updates given')
      return

    # TODO(jojwang): monorail:6564, should check description and summary are
    # not empty strings before calling UpdateHotlist.
    self.services.features.UpdateHotlist(
        self.mc.cnxn, hotlist_id, name=name, summary=summary,
        description=description, is_private=is_private,
        default_col_spec=default_col_spec)

  def GetHotlist(self, hotlist_id, use_cache=True):
    # int, Optional[bool] -> Hotlist
    """Return the specified hotlist.

    Args:
      hotlist_id: int hotlist_id of the hotlist to retrieve.
      use_cache: set to false when doing read-modify-write.

    Returns:
      The specified hotlist.

    Raises:
      NoSuchHotlistException: There is no hotlist with that ID.
      PermissionException: The user is not allowed to view the hotlist.
    """
    if hotlist_id is None:
      raise exceptions.InputException('No hotlist specified')

    with self.mc.profiler.Phase('getting hotlist %r' % hotlist_id):
      hotlist = self.services.features.GetHotlist(
          self.mc.cnxn, hotlist_id, use_cache=use_cache)
    self._AssertUserCanViewHotlist(hotlist)
    return hotlist

  def ListHotlistItems(self, hotlist_id, max_items, start, can, sort_spec,
                       group_by_spec, use_cache=True):
    # (int, int, int, int, str, str, bool) -> (
    #     List(HotlistItem), ProjectIssueConfig, Dict{user_id: UserView, ...})
    """Return a list of HotlistItems for the given hotlist that
       are visible by the user.

    Args:
      hotlist_id: int hotlist_id of the hotlist.
      max_items: int the maximum number of HotlistItems we want to return.
      start: int start position in the total sorted items.
      can: int "canned_query" number to scope the visible issues.
      sort_spec: string that lists the sort order.
      group_by_spec: string that lists the grouping order.
      use_cache: set to false when doing read-modify-write.

    Returns:
      A tuple of (visible_hotlist_items, harmonized_config) where:

      visible_hotlist_items: a list of sorted HotlistItems with
        issues filtered out based on start, max_items, can, and the users's
        permisisons.
      harmonized_config: ProjectIssueConfig for all projects involved in
        with the hotlist.

    Raises:
      NoSuchHotlistException: There is no hotlist with that ID.
    """
    hotlist = self.GetHotlist(hotlist_id, use_cache=use_cache)
    hotlist_issues = self.services.issue.GetIssues(
        self.mc.cnxn, [item.issue_id for item in hotlist.items])
    project_ids = hotlist_helpers.GetAllProjectsOfIssues(hotlist_issues)
    config_list = hotlist_helpers.GetAllConfigsOfProjects(
        self.mc.cnxn, project_ids, self.services)
    harmonized_config = tracker_bizobj.HarmonizeConfigs(config_list)

    (sorted_issues, _hotlist_items_context,
     _users_by_id) = hotlist_helpers.GetSortedHotlistIssues(
        self.mc.cnxn, hotlist.items, hotlist_issues, self.mc.auth, can,
        sort_spec, group_by_spec, harmonized_config, self.services,
        self.mc.profiler)

    if not start or start < 1:
      range_start = 0
    else:
      range_start = start - 1
    if not max_items or max_items < 1:
      max_items = features_constants.DEFAULT_RESULTS_PER_PAGE
    range_end = range_start + max_items
    visible_issues = sorted_issues[range_start: range_end]
    hotlist_items_dict = {item.issue_id: item for item in hotlist.items}
    visible_hotlist_items = [hotlist_items_dict.get(issue.issue_id) for
                            issue in visible_issues]
    return (visible_hotlist_items, harmonized_config)

  def TransferHotlistOwnership(self, hotlist_id, new_owner_id, remain_editor,
                               use_cache=True, commit=True):
    """Transfer ownership of hotlist from current owner to new_owner.

    Args:
      hotlist_id: int hotlist_id of the hotlist we want to transfer
      new_owner_id: user_id of the new owner
      remain_editor: True if the old owner should remain on the hotlist as
        editor.
      use_cache: set to false when doing read-modify-write.
      commit: True, if changes should be committed.

    Raises:
      NoSuchHotlistException: There is not hotlist with the given ID.
      PermissionException: The logged-in user is not allowed to change ownership
        of the hotlist.
      InputException: The proposed new owner already owns a hotlist with the
        same name.
    """
    hotlist = self.services.features.GetHotlist(
        self.mc.cnxn, hotlist_id, use_cache=use_cache)
    edit_permitted = permissions.CanAdministerHotlist(
        self.mc.auth.effective_ids, self.mc.perms, hotlist)
    if not edit_permitted:
      raise permissions.PermissionException(
          'User is not allowed to update hotlist members.')

    if self.services.features.LookupHotlistIDs(
        self.mc.cnxn, [hotlist.name], [new_owner_id]):
      raise exceptions.InputException(
          'Proposed new owner already owns a hotlist with this name.')

    self.services.features.TransferHotlistOwnership(
        self.mc.cnxn, hotlist, new_owner_id, remain_editor, commit=commit)

  def DeltaUpdateHotlistRoles(
      self, hotlist_id, new_owner_id=None, add_editor_ids=None,
      add_follower_ids=None, remove_user_ids=None, use_cache=True,
      commit=True):
    """Update the members of a hotlist.

    A user can only hold one role in a hotlist. If an existing hotlist
    member is in any add_ list, they will be reassigned to the new role.
    If a user appears in an add_ list and the remove_user_ids list,
    the user will be added as a hotlist member.

    Args:
      hotlist_id: the id of the hotlist we want to update
      new_owner_id: user_id of new owner for hotlist
      add_editor_ids: list of editor_ids to add to hotlist editors
      add_follower_ids: list of follower_ids to add to hotlistfollower
      remove_user_ids: list of user_ids to remove from hotlist members

    Raises:
      NoSuchHotlistException: There is not hotlist with the given ID.
      PermissionException: The logged-in user is not allowed to make the hotlist
        roles change.
      InputException: The add_ lists or new_owner_id have overlapping user_ids,
        or the changes would result in an unowned hotlist.
    """
    hotlist = self.services.features.GetHotlist(
        self.mc.cnxn, hotlist_id, use_cache=use_cache)
    edit_permitted = permissions.CanAdministerHotlist(
        self.mc.auth.effective_ids, self.mc.perms, hotlist)
    # check if user is only removing themselves from the hotlist.
    # removing linked accounts is allowed but users cannot remove groups
    # they are part of from hotlists.
    user_or_linked_ids = (self.mc.auth.user_pb.linked_child_ids +
                          [self.mc.auth.user_id])
    if self.mc.auth.user_pb.linked_parent_id:
      user_or_linked_ids.append(self.mc.auth.user_pb.linked_parent_id)
    removing_self_only = (not new_owner_id and not add_editor_ids and
                     not add_follower_ids and
                     set(remove_user_ids).issubset(set(user_or_linked_ids)))
    if not (removing_self_only or edit_permitted):
      raise permissions.PermissionException(
          'User is not allowed to make this hotlist members update.')
    if (new_owner_id in (add_editor_ids + add_follower_ids) or
        set(add_editor_ids) & set(add_follower_ids)):
      raise exceptions.InputException(
          'User cannot have multiple roles in hotlist.')

    new_owner_ids = set(hotlist.owner_ids)
    new_editor_ids = set(hotlist.editor_ids)
    new_follower_ids = set(hotlist.follower_ids)
    if remove_user_ids:
      new_owner_ids -= set(remove_user_ids)
      new_editor_ids -= set(remove_user_ids)
      new_follower_ids -= set(remove_user_ids)
    if add_follower_ids:
      new_follower_ids.update(set(add_follower_ids))
      # remove new followers from other roles
      new_editor_ids -= new_follower_ids
      new_owner_ids -= new_follower_ids
    if add_editor_ids:
      new_editor_ids.update(set(add_editor_ids))
      # remove new editors from other roles
      new_follower_ids -= new_editor_ids
      new_owner_ids -= new_follower_ids
    if new_owner_id:
      new_owner_ids = {new_owner_id}
      # remove new owner from other roles
      new_editor_ids -= new_owner_ids
      new_follower_ids -= new_owner_ids

    if not new_owner_ids:
      raise exceptions.InputException('Hotlist must have owner.')

    self.services.features.UpdateHotlistRoles(
        self.mc.cnxn, hotlist_id, list(new_owner_ids), list(new_editor_ids),
        list(new_follower_ids), commit=commit)

  def DeleteHotlist(self, hotlist_id):
    """Delete the given hotlist from the DB.

    Args:
      hotlist_id (int): The id of the hotlist to delete.

    Raises:
      NoSuchHotlistException: There is not hotlist with the given ID.
      PermissionException: The logged-in user is not allowed to
        delete the hotlist.
    """
    hotlist = self.services.features.GetHotlist(
        self.mc.cnxn, hotlist_id, use_cache=False)
    edit_permitted = permissions.CanAdministerHotlist(
        self.mc.auth.effective_ids, self.mc.perms, hotlist)
    if not edit_permitted:
      raise permissions.PermissionException(
          'User is not allowed to delete hotlist')

    self.services.features.ExpungeHotlists(
        self.mc.cnxn, [hotlist.hotlist_id], self.services.hotlist_star,
        self.services.user,  self.services.chart)


  def ListHotlistsByUser(self, user_id):
    """Return the hotlists for the given user.

    Args:
      user_id (int): The id of the user to query.

    Returns:
      The hotlists for the given user.
    """
    if user_id is None:
      raise exceptions.InputException('No user specified')

    with self.mc.profiler.Phase('querying hotlists for user %r' % user_id):
      hotlists = self.services.features.GetHotlistsByUserID(
          self.mc.cnxn, user_id)

    # Filter the hotlists that the currently authenticated user cannot see.
    result = [
        hotlist
        for hotlist in hotlists
        if permissions.CanViewHotlist(
            self.mc.auth.effective_ids, self.mc.perms, hotlist)]
    return result

  def ListHotlistsByIssue(self, issue_id):
    """Return the hotlists the given issue is part of.

    Args:
      issue_id (int): The id of the issue to query.

    Returns:
      The hotlists the given issue is part of.
    """
    # Check that the issue exists and the user has permission to see it.
    self.GetIssue(issue_id)

    with self.mc.profiler.Phase('querying hotlists for issue %r' % issue_id):
      hotlists = self.services.features.GetHotlistsByIssueID(
          self.mc.cnxn, issue_id)

    # Filter the hotlists that the currently authenticated user cannot see.
    result = [
        hotlist
        for hotlist in hotlists
        if permissions.CanViewHotlist(
            self.mc.auth.effective_ids, self.mc.perms, hotlist)]
    return result

  def ListRecentlyVisitedHotlists(self):
    """Return the recently visited hotlists for the logged in user.

    Returns:
      The recently visited hotlists for the given user, or an empty list if no
      user is logged in.
    """
    if not self.mc.auth.user_id:
      return []

    with self.mc.profiler.Phase(
        'get recently visited hotlists for user %r' % self.mc.auth.user_id):
      hotlist_ids = self.services.user.GetRecentlyVisitedHotlists(
          self.mc.cnxn, self.mc.auth.user_id)
      hotlists_by_id = self.services.features.GetHotlists(
          self.mc.cnxn, hotlist_ids)
      hotlists = [hotlists_by_id[hotlist_id] for hotlist_id in hotlist_ids]

    # Filter the hotlists that the currently authenticated user cannot see.
    # It might be that some of the hotlists have become private since the user
    # last visited them, or the user has lost access for other reasons.
    result = [
        hotlist
        for hotlist in hotlists
        if permissions.CanViewHotlist(
            self.mc.auth.effective_ids, self.mc.perms, hotlist)]
    return result

  def ListStarredHotlists(self):
    """Return the starred hotlists for the logged in user.

    Returns:
      The starred hotlists for the logged in user.
    """
    if not self.mc.auth.user_id:
      return []

    with self.mc.profiler.Phase(
        'get starred hotlists for user %r' % self.mc.auth.user_id):
      hotlist_ids = self.services.hotlist_star.LookupStarredItemIDs(
          self.mc.cnxn, self.mc.auth.user_id)
      hotlists_by_id, _ = self.services.features.GetHotlistsByID(
          self.mc.cnxn, hotlist_ids)
      hotlists = [hotlists_by_id[hotlist_id] for hotlist_id in hotlist_ids]

    # Filter the hotlists that the currently authenticated user cannot see.
    # It might be that some of the hotlists have become private since the user
    # starred them, or the user has lost access for other reasons.
    result = [
        hotlist
        for hotlist in hotlists
        if permissions.CanViewHotlist(
            self.mc.auth.effective_ids, self.mc.perms, hotlist)]
    return result

  def StarHotlist(self, hotlist_id, starred):
    """Star or unstar the specified hotlist.

    Args:
      hotlist_id: int ID of the hotlist to star/unstar.
      starred: true to add a star, false to remove it.

    Returns:
      Nothing.

    Raises:
      NoSuchHotlistException: There is no hotlist with that ID.
    """
    if hotlist_id is None:
      raise exceptions.InputException('No hotlist specified')

    if not self.mc.auth.user_id:
      raise exceptions.InputException('No current user specified')

    with self.mc.profiler.Phase('(un)starring hotlist %r' % hotlist_id):
      # Make sure the hotlist exists and user has permission to see it.
      self.GetHotlist(hotlist_id)
      self.services.hotlist_star.SetStar(
          self.mc.cnxn, hotlist_id, self.mc.auth.user_id, starred)

  def IsHotlistStarred(self, hotlist_id):
    """Return True if the current hotlist has starred the given hotlist.

    Args:
      hotlist_id: int ID of the hotlist to check.

    Returns:
      True if starred.

    Raises:
      NoSuchHotlistException: There is no hotlist with that ID.
    """
    if hotlist_id is None:
      raise exceptions.InputException('No hotlist specified')

    if not self.mc.auth.user_id:
      return False

    with self.mc.profiler.Phase('checking hotlist star %r' % hotlist_id):
      # Make sure the hotlist exists and user has permission to see it.
      self.GetHotlist(hotlist_id)
      return self.services.hotlist_star.IsItemStarredBy(
        self.mc.cnxn, hotlist_id, self.mc.auth.user_id)

  def GetHotlistStarCount(self, hotlist_id):
    """Return the number of times the hotlist has been starred.

    Args:
      hotlist_id: int ID of the hotlist to check.

    Returns:
      The number of times the hotlist has been starred.

    Raises:
      NoSuchHotlistException: There is no hotlist with that ID.
    """
    if hotlist_id is None:
      raise exceptions.InputException('No hotlist specified')

    with self.mc.profiler.Phase('counting stars for hotlist %r' % hotlist_id):
      # Make sure the hotlist exists and user has permission to see it.
      self.GetHotlist(hotlist_id)
      return self.services.hotlist_star.CountItemStars(self.mc.cnxn, hotlist_id)

  def CheckHotlistName(self, name):
    """Check that a hotlist name is valid and not already in use.

    Args:
      name: str the hotlist name to check.

    Returns:
      None if the user can create a hotlist with that name, or a string with the
      reason the name can't be used.

    Raises:
      InputException: The user is not signed in.
    """
    if not self.mc.auth.user_id:
      raise exceptions.InputException('No current user specified')

    with self.mc.profiler.Phase('checking hotlist name: %r' % name):
      if not framework_bizobj.IsValidHotlistName(name):
        return '"%s" is not a valid hotlist name.' % name
      if self.services.features.LookupHotlistIDs(
          self.mc.cnxn, [name], [self.mc.auth.user_id]):
        return 'There is already a hotlist with that name.'

    return None

  def RemoveIssuesFromHotlists(self, hotlist_ids, issue_ids):
    """Remove the issues given in issue_ids from the given hotlists.

    Args:
      hotlist_ids: a list of hotlist ids to remove the issues from.
      issue_ids: a list of issue_ids to be removed.

    Raises:
      PermissionException: The user has no permission to edit the hotlist.
      NoSuchHotlistException: One of the hotlist ids was not found.
    """
    for hotlist_id in hotlist_ids:
      self._AssertUserCanEditHotlist(self.GetHotlist(hotlist_id))

    with self.mc.profiler.Phase(
        'Removing issues %r from hotlists %r' % (issue_ids, hotlist_ids)):
      self.services.features.RemoveIssuesFromHotlists(
          self.mc.cnxn, hotlist_ids, issue_ids, self.services.issue,
          self.services.chart)

  def AddIssuesToHotlists(self, hotlist_ids, issue_ids, note):
    """Add the issues given in issue_ids to the given hotlists.

    Args:
      hotlist_ids: a list of hotlist ids to add the issues to.
      issue_ids: a list of issue_ids to be added.
      note: a string with a message to record along with the issues.

    Raises:
      PermissionException: The user has no permission to edit the hotlist.
      NoSuchHotlistException: One of the hotlist ids was not found.
    """
    for hotlist_id in hotlist_ids:
      self._AssertUserCanEditHotlist(self.GetHotlist(hotlist_id))

    # Even though we check permissions when viewing the issues in a hotlist,
    # we also check permissions when adding issues to a hotlist to prevent
    # some clever ways that attackers could find some issue details.
    issues_dict = self.services.issue.GetIssuesDict(self.mc.cnxn, issue_ids)
    for issue in issues_dict.values():
      self._AssertUserCanViewIssue(issue)

    added_tuples = [
        (issue_id, self.mc.auth.user_id, int(time.time()), note)
        for issue_id in issue_ids]

    with self.mc.profiler.Phase(
        'Removing issues %r from hotlists %r' % (issue_ids, hotlist_ids)):
      self.services.features.AddIssuesToHotlists(
          self.mc.cnxn, hotlist_ids, added_tuples, self.services.issue,
          self.services.chart)

  def RerankHotlistItems(self, hotlist_id, moved_issue_ids, target_position):
    # type: (int, list(int), int) -> Hotlist
    """Rerank the moved items for a hotlist.
        e.g. For a hotlist with items (a, b, c, d, e), if moved_issue_ids were
        [e.issue_id, c.issue_id] and target_position were 0, the hotlist items
        would be reranked as (e, c, a, b, d).

    Args:
      hotlist_id: A hotlist ID of the hotlist to rerank.
      moved_issue_ids: A list of issue IDs to be moved together, in the order
        they should have after the reranking.
      target_position: The index, starting at 0, of the new position the
        first issue in moved_issue_ids should have. This value cannot be greater
        than (len(hotlist.items) - len(moved_issue_ids)).

    Returns:
      The updated hotlist.

    Raises:
      PermissionException: If the user lacks permissions to rerank the hotlist.
      NoSuchHotlistException: If the hotlist is not found.
      InputException: If the target_position or moved_issue_ids are not valid.
    """
    hotlist = self.GetHotlist(hotlist_id)
    self._AssertUserCanEditHotlist(hotlist)
    if not moved_issue_ids:
      raise exceptions.InputException('`moved_issue_ids` empty.')

    rank_changes = rerank_helpers.GetHotlistRerankChanges(
        hotlist.items, moved_issue_ids, target_position)
    if rank_changes:
      relations_to_change = dict(rank_changes)
      self.services.features.UpdateHotlistItemsFields(
          self.mc.cnxn, hotlist_id, new_ranks=relations_to_change)

    return self.GetHotlist(hotlist.hotlist_id)

  # TODO(crbug/monorail/7031): Remove this method
  # and corresponding v0 prpc method.
  def RerankHotlistIssues(self, hotlist_id, moved_ids, target_id, split_above):
    """Rerank the moved issues for the hotlist.

    Args:
      hotlist_id: an int with the id of the hotlist.
      moved_ids: The id of the issues to move.
      target_id: the id of the issue to move the issues to.
      split_above: True if moved issues should be moved before the target issue.
    """
    hotlist = self.GetHotlist(hotlist_id)
    self._AssertUserCanEditHotlist(hotlist)
    hotlist_issue_ids = [item.issue_id for item in hotlist.items]
    if not set(moved_ids).issubset(set(hotlist_issue_ids)):
      raise exceptions.InputException('The issue to move is not in the hotlist')
    if target_id not in hotlist_issue_ids:
      raise exceptions.InputException('The target issue is not in the hotlist.')

    phase_name = 'Moving issues %r %s issue %d.' % (
        moved_ids, 'above' if split_above else 'below', target_id)
    with self.mc.profiler.Phase(phase_name):
      lower, higher = features_bizobj.SplitHotlistIssueRanks(
          target_id, split_above,
          [(item.issue_id, item.rank) for item in hotlist.items if
           item.issue_id not in moved_ids])
      rank_changes = rerank_helpers.GetInsertRankings(lower, higher, moved_ids)
      if rank_changes:
        relations_to_change = {
            issue_id: rank for issue_id, rank in rank_changes}
        self.services.features.UpdateHotlistItemsFields(
            self.mc.cnxn, hotlist_id, new_ranks=relations_to_change)

  def UpdateHotlistIssueNote(self, hotlist_id, issue_id, note):
    """Update the given issue of the given hotlist with the given note.

    Args:
      hotlist_id: an int with the id of the hotlist.
      issue_id: an int with the id of the issue.
      note: a string with a message to record for the given issue.
    Raises:
      PermissionException: The user has no permission to edit the hotlist.
      NoSuchHotlistException: The hotlist id was not found.
      InputException: The issue is not part of the hotlist.
    """
    # Make sure the hotlist exists and we have permission to see and edit it.
    hotlist = self.GetHotlist(hotlist_id)
    self._AssertUserCanEditHotlist(hotlist)

    # Make sure the issue exists and we have permission to see it.
    self.GetIssue(issue_id)

    # Make sure the issue belongs to the hotlist.
    if not any(item.issue_id == issue_id for item in hotlist.items):
      raise exceptions.InputException('The issue is not part of the hotlist.')

    with self.mc.profiler.Phase(
        'Editing note for issue %s in hotlist %s' % (issue_id, hotlist_id)):
      new_notes = {issue_id: note}
      self.services.features.UpdateHotlistItemsFields(
          self.mc.cnxn, hotlist_id, new_notes=new_notes)

  def ListHotlistPermissions(self, hotlist_id):
    # type: (int) -> List(str)
    """Return the list of permissions the current user has for the given
       hotlist.

    This method only allows for fetching the permisisons of the current anon
    or logged-in user. Users will only be able to look up their own permissions.
    """
    hotlist = self.services.features.GetHotlist(self.mc.cnxn, hotlist_id)
    if permissions.CanAdministerHotlist(
        self.mc.auth.effective_ids, self.mc.perms, hotlist):
      return permissions.HOTLIST_OWNER_PERMISSIONS
    if permissions.CanEditHotlist(
        self.mc.auth.effective_ids, self.mc.perms, hotlist):
      return permissions.HOTLIST_EDITOR_PERMISSIONS
    if permissions.CanViewHotlist(
        self.mc.auth.effective_ids, self.mc.perms, hotlist):
      return permissions.STANDARD_HOTLIST_PERMISSIONS
    return []

  # FUTURE: DeleteHotlist()

  def expungeUsersFromStars(self, user_ids):
    """Wipes any starred user or user's stars from all star services.

    This method will not commit the operation. This method will not
    make changes to in-memory data.
    """

    self.services.project_star.ExpungeStarsByUsers(self.mc.cnxn, user_ids)
    self.services.issue_star.ExpungeStarsByUsers(self.mc.cnxn, user_ids)
    self.services.hotlist_star.ExpungeStarsByUsers(self.mc.cnxn, user_ids)
    self.services.user_star.ExpungeStarsByUsers(self.mc.cnxn, user_ids)
    for user_id in user_ids:
      self.services.user_star.ExpungeStars(self.mc.cnxn, user_id, commit=False)
