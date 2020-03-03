# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

# For developing purpose only
import random
import string
import time

import logging

from google.protobuf import timestamp_pb2

from api import resource_name_converters as rnc
from api.v1.api_proto import feature_objects_pb2
from api.v1.api_proto import issue_objects_pb2
from api.v1.api_proto import project_objects_pb2
from api.v1.api_proto import user_objects_pb2


def ConvertHotlist(hotlist):
  # proto.features_pb2.Hotlist -> api_proto.feature_objects_pb2.Hotlist
  """Convert a protorpc Hotlist into a protoc Hotlist.

  Args:
    hotlist: Hotlist protorpc object.

  Returns:
    The equivalent Hotlist protoc Hotlist.

  """
  hotlist_resource_name = rnc.ConvertHotlistName(hotlist.hotlist_id)
  user_resource_names_dict = rnc.ConvertUserNames(
      hotlist.owner_ids + hotlist.editor_ids)
  default_columns  = [issue_objects_pb2.IssuesListColumn(column=col)
                      for col in hotlist.default_col_spec.split()]
  api_hotlist = feature_objects_pb2.Hotlist(
      name=hotlist_resource_name,
      display_name=hotlist.name,
      owner=user_resource_names_dict.get(hotlist.owner_ids[0]),
      editors=[
          user_resource_names_dict.get(user_id)
          for user_id in hotlist.editor_ids
      ],
      summary=hotlist.summary,
      description=hotlist.description,
      default_columns=default_columns)
  if not hotlist.is_private:
    api_hotlist.hotlist_privacy = (
        feature_objects_pb2.Hotlist.HotlistPrivacy.Value('PUBLIC'))
  return api_hotlist


def ConvertHotlistItems(cnxn, hotlist_id, items, services):
  # MonorailConnection, int, Sequence[proto.features_pb2.HotlistItem],
  #     Services -> Sequence[api_proto.feature_objects_pb2.Hotlist]
  """Convert a Sequence of protorpc HotlistItems into a Sequence of protoc
     HotlistItems.

  Args:
    cnxn: MonorailConnection object.
    hotlist_id: ID of the Hotlist the items belong to.
    items: Sequence of HotlistItem protorpc objects.
    services: Services object for connections to backend services.

  Returns:
    Sequence of protoc HotlistItems in the same order they are given in `items`.
    In the rare event that any issues in `items` are not found, they will be
    omitted from the result.
  """
  issue_ids = [item.issue_id for item in items]
  # Converting HotlistItemNames and IssueNames both require looking up the
  # issues in the hotlist. However, we want to keep the code clean and readable
  # so we keep the two processes separate.
  resource_names_dict = rnc.ConvertHotlistItemNames(
      cnxn, hotlist_id, issue_ids, services)
  issue_names_dict = rnc.ConvertIssueNames(cnxn, issue_ids, services)
  adder_names_dict = rnc.ConvertUserNames([item.adder_id for item in items])

  # Filter out items whose issues were not found.
  found_items = [
      item for item in items if resource_names_dict.get(item.issue_id) and
      issue_names_dict.get(item.issue_id)
  ]
  if len(items) != len(found_items):
    found_ids = [item.issue_id for item in found_items]
    missing_ids = [iid for iid in issue_ids if iid not in found_ids]
    logging.info('HotlistItem issues %r not found' % missing_ids)

  # Generate user friendly ranks (0, 1, 2, 3,...) that are exposed to API
  # clients, instead of using padded ranks (1, 11, 21, 31,...).
  sorted_ranks = sorted(item.rank for item in found_items)
  friendly_ranks_dict = {
      rank: friendly_rank for friendly_rank, rank in enumerate(sorted_ranks)
  }

  api_items = []
  for item in found_items:
      api_item = feature_objects_pb2.HotlistItem(
          name=resource_names_dict.get(item.issue_id),
          issue=issue_names_dict.get(item.issue_id),
          rank=friendly_ranks_dict[item.rank],
          adder=adder_names_dict.get(item.adder_id),
          note=item.note)
      if item.date_added:
        api_item.create_time.FromSeconds(item.date_added)
      api_items.append(api_item)

  return api_items


def ConvertIssueTemplates(templates, project_name, admins, users_by_id):
  # List(protorpc.TemplateDef), string, Dict(int, protorpc.User),
  # Dict(int: UserView) -> List(IssueTemplate)
  """Convert a project's protorpc TemplateDefs into list of IssueTemplate."""
  converted_templates = []
  i = 0
  while i < len(templates):
    template = templates[i]
    name = rnc.GetTemplateResourceName(project_name, template.name)
    summary_must_be_edited = template.summary_must_be_edited
    template_privacy = getTemplatePrivacy(template)
    default_owner = getDefaultOwner(template)
    component_required = template.component_required
    template_admins = {k:v for k,v in admins.iteritems()
      if k in template.admin_ids}
    converted_admins = ConvertUsers(template_admins, users_by_id)
    # Test data
    issue = generateIssue()

    converted_templates.append(
        project_objects_pb2.IssueTemplate(
          name=name,
          issue=issue,
          summary_must_be_edited=summary_must_be_edited,
          template_privacy=template_privacy,
          default_owner=default_owner,
          component_required=component_required,
          admins=converted_admins))
    i += 1
  return converted_templates


def getTemplatePrivacy(template):
  """Given a tracker_pb2.TemplateDef
     return project_object_pb2.IssueTemplate.TemplatePrivacy
  """
  if template.members_only:
    return 1
  elif template.members_only == False:
    return 2
  else:
    return 0


def getDefaultOwner(template):
  """Given a tracker_pb2.TemplateDef
     return project_object_pb2.IssueTemplate.DefaultOwner
  """
  if template.owner_defaults_to_member:
    return 1
  else:
    return 0


def generateIssue():
  """Generate a random issue of issues_object.Issue format"""
  name = 'project/monorail/issues/{}'.format(random.randrange(10000))
  summary = generateString(stringLength=random.randrange(10,40))
  state = random.randint(0,3) # enum between 0, 1, 2, and 3
  status = issue_objects_pb2.Issue.StatusValue(
    status='projects/monorail/statusDefs/open',
    derivation=0)
  # status = ['projects/monorail/statusDefs/open', 0]
  description = generateString(stringLength=random.randrange(40,100))
  reporter = 'users/{}'.format(random.randrange(100000))
  owner = issue_objects_pb2.Issue.UserValue(
    user='users/{}'.format(random.randrange(100000)),
    derivation=1)
  cc_users = []
  labels = []
  components = []
  field_values = []
  # merged_into_issue_ref = # Not set
  blocked_on_issue_refs = []
  blocking_issue_refs = []
  # create_time = int(time.time() - (3600 * random.randrange(1,2400)))*1000000
  # create_time = int(time.time())
  # close_time = # Not set, because open
  # modify_time = create_time
  # component_modify_time = # Not set
  # status_modify_time = # Not set
  # owner_modify_time = # Not set
  attachment_count = 0
  star_count = 0
  approval_values = []
  phases = []

  generatedIssue = issue_objects_pb2.Issue(
    name=name,
    summary=summary,
    state=state,
    status=status,
    description=description,
    reporter=reporter,
    owner=owner,
    cc_users=cc_users,
    labels=labels,
    components=components,
    field_values=field_values,
    blocked_on_issue_refs=blocked_on_issue_refs,
    blocking_issue_refs=blocking_issue_refs,
    # create_time=create_time,
    # modify_time=modify_time,
    attachment_count=attachment_count,
    star_count=star_count,
    approval_values=approval_values,
    phases=phases)
  return generatedIssue


def generateUser():
  name = 'users/{}'.format(random.randrange(100000))
  display_name = '{}@chromium.org'.format(
      generateString(random.randrange(3,10)))
  site_role = random.randint(0,2) # enum between 0, 1, and 2
  availability_message = generateString(rand.randrange(0,10))

  generatedUser = user_objects_pb2.User(
    name=name,
    display_name=display_name,
    site_role=site_role,
    availability_message=availability_message)
  return generatedUser


def generateString(stringLength=10):
    """Generate a random string of fixed length """
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(stringLength))


def ConvertUsers(userpbs_by_id, userviews_by_id):
  # Dict(int: protorpc.User), Dict(int: UserView) ->
  # api_proto.user_objects_pb2.User
  """Convert dict of users and UserViews into list of protoc Users.

  Args:
    userpbs_by_id: Dictionary of protorpc.Users
    userviews_by_id: Dictionary of UserViews 

  Returns:
    List of equivalent protoc Users.

  """
  output = []

  # Aggregate list of user ids
  all_user_ids = []
  for user_id, user in userpbs_by_id.items():
    all_user_ids.append(user_id)
    if user.linked_parent_id is not None:
      all_user_ids.append(user.linked_parent_id)
    # all_user_ids += [user_id, user.linked_parent_id]
  # Convert all of them to resource names
  user_resource_names_dict= rnc.ConvertUserNames(all_user_ids)

  for user_id, user in userpbs_by_id.items():
    user_view = userviews_by_id.get(user_id)

    name = user_resource_names_dict.get(user_id)
    display_name = user_view.display_name

    if user.is_site_admin:
      site_role = 2
    elif user.is_site_admin == False:
      site_role = 1
    else:
      site_role = 0

    availability_message = user_view.avail_message
    linked_primary_user = None
    if user.linked_parent_id is not None:
      linked_primary_user = user_resource_names_dict.get(user.linked_parent_id)

    output.append(user_objects_pb2.User(
        name=name,
        display_name=display_name,
        site_role=site_role,
        availability_message=availability_message,
        linked_primary_user=linked_primary_user))

  return output
