# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""A servlet for project owners to edit/delete a template"""

import collections
import logging
import time

from third_party import ezt

from framework import authdata
from framework import framework_bizobj
from framework import framework_helpers
from framework import servlet
from framework import urls
from framework import permissions
from tracker import field_helpers
from tracker import template_helpers
from tracker import tracker_bizobj
from tracker import tracker_helpers
from tracker import tracker_views
from services import user_svc


class TemplateDetail(servlet.Servlet):
  """Servlet allowing project owners to edit/delete an issue template"""

  _MAIN_TAB_MODE = servlet.Servlet.MAIN_TAB_PROCESS
  _PAGE_TEMPLATE = 'tracker/template-detail-page.ezt'
  _PROCESS_SUBTAB = servlet.Servlet.PROCESS_TAB_TEMPLATES

  def AssertBasePermission(self, mr):
    """Check whether the user has any permission to visit this page.

    Args:
      mr: commonly used info parsed from the request.
    """
    super(TemplateDetail, self).AssertBasePermission(mr)
    config = self.services.config.GetProjectConfig(mr.cnxn, mr.project_id)

    template = tracker_bizobj.FindIssueTemplate(mr.template_name, config)

    if template:
      allow_view = permissions.CanViewTemplate(
          mr.auth.effective_ids, mr.perms, mr.project, template)
      if not allow_view:
        raise permissions.PermissionException(
            'User is not allowed to view this issue template')
    else:
      self.abort(404, 'issue template not found %s' % mr.template_name)

  def GatherPageData(self, mr):
    """Build up a dictionary of data values to use when rendering the page.

    Args:
      mr: commonly used info parsed from the request.

    Returns:
      Dict of values used by EZT for rendering the page.
    """

    config = self.services.config.GetProjectConfig(mr.cnxn, mr.project_id)
    template = tracker_bizobj.FindIssueTemplate(mr.template_name, config)
    template_view = tracker_views.IssueTemplateView(
        mr, template, self.services.user, config)
    fd_id_to_fvs = collections.defaultdict(list)
    for fv in template.field_values:
      fd_id_to_fvs[fv.field_id].append(fv)

    field_views = [
      tracker_views.MakeFieldValueView(fd, config, [], [],
                                       fd_id_to_fvs[fd.field_id], {})
      for fd in config.field_defs if not fd.is_deleted]

    allow_edit = permissions.CanEditTemplate(
        mr.auth.effective_ids, mr.perms, mr.project, template)

    return {
        'admin_tab_mode': self._PROCESS_SUBTAB,
        'allow_edit': ezt.boolean(allow_edit),
        'template': template_view,
        'fields': field_views,
        'labels': template.labels,
        'initial_owner': template_view.ownername,
        'initial_components': template_view.components,
        'initial_admins': template_view.admin_names,
        }

  def ProcessFormData(self, mr, post_data):
    """Validate and store the contents of the issues tracker admin page.

    Args:
      mr: commonly used info parsed from the request.
      post_data: HTML form data from the request.

    Returns:
      String URL to redirect the user to, or None if response was already sent.
    """

    # TODO(jojwang): add delete template section
    if 'cancel' in post_data:
      return framework_helpers.FormatAbsoluteURL(
          mr, urls.ADMIN_TEMPLATES, ts=int(time.time()))

    config = self.services.config.GetProjectConfig(mr.cnxn, mr.project_id)
    parsed = template_helpers.ParseTemplateRequest(post_data, config)
    template = tracker_bizobj.FindIssueTemplate(parsed.name, config)
    template_id = template.template_id
    allow_edit = permissions.CanEditTemplate(
        mr.auth.effective_ids, mr.perms, mr.project, template)
    if not allow_edit:
      raise permissions.PermissionException(
          'User is not allowed edit this issue template.')

    admin_ids, admin_str = tracker_helpers.ParseAdminUsers(
        mr.cnxn, post_data.get('admin_names', ''), self.services.user)

    owner_id = 0
    if parsed.owner_str:
      try:
        user_id = self.services.user.LookupUserID(mr.cnxn, parsed.owner_str)
        auth = authdata.AuthData.FromUserID(mr.cnxn, user_id, self.services)
        if framework_bizobj.UserIsInProject(mr.project, auth.effective_ids):
          owner_id = user_id
        else:
          mr.errors.owner = 'User is not a member of this project.'
      except user_svc.NoSuchUserException:
        mr.errors.owner = 'Owner not found.'

    component_ids = tracker_helpers.LookupComponentIDs(
        parsed.component_paths, config, mr.errors)

    field_values = field_helpers.ParseFieldValues(
        mr.cnxn, self.services.user, parsed.field_val_strs, config)
    for fv in field_values:
      logging.info('field_value is %r: %r',
                   fv.field_id, tracker_bizobj.GetFieldValue(fv, {}))

    updated_template = tracker_bizobj.MakeIssueTemplate(
        parsed.name, parsed.summary, parsed.status, owner_id, parsed.content,
        parsed.labels, field_values, admin_ids, component_ids,
        parsed.summary_must_be_edited, parsed.owner_defaults_to_member,
        parsed.component_required, parsed.members_only, template_id)

    if mr.errors.AnyErrors():
      fd_id_to_fvs = collections.defaultdict(list)
      for fv in field_values:
        fd_id_to_fvs[fv.field_id].append(fv)

      field_views = [
          tracker_views.MakeFieldValueView(fd, config, [], [],
                                           fd_id_to_fvs[fd.field_id], {})
          for fd in config.field_defs if not fd.is_deleted]

      template_view = tracker_views.IssueTemplateView(
        mr, updated_template, self.services.user, config)
      self.PleaseCorrect(
          mr,
          template=template_view,
          fields=field_views,
          labels=parsed.labels,
          initial_owner=parsed.owner_str,
          initial_components=', '.join(parsed.component_paths),
          initial_admins=admin_str
      )
      return

    templates = [tmpl for tmpl in config.templates if tmpl.template_id is not
                 template_id]
    templates.append(updated_template)

    self.services.config.UpdateConfig(
        mr.cnxn, mr.project, templates=templates)

    return framework_helpers.FormatAbsoluteURL(
        mr, urls.TEMPLATE_DETAIL, template=template.name,
        saved=1, ts=int(time.time()))
