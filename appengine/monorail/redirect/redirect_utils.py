# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Utils for redirect."""
import urllib
from werkzeug.datastructures import MultiDict
from redirect import redirect_project_template

from tracker import tracker_constants
from tracker import tracker_bizobj
from redirect import redirect_custom_value

PROJECT_REDIRECT_MAP = {
    'pigweed': 'https://issues.pigweed.dev',
    'git': 'https://git.issues.gerritcodereview.com',
    'gerrit': 'https://issues.gerritcodereview.com',
    'skia': 'http://issues.skia.org',
    'fuchsia': 'https://issues.fuchsia.dev',
}

MAX_MONORAIL_ISSUE_ID = 10000000

TRACKER_SEARCH_KEY_MAP = {
    'cc': 'cc',
    'owner': 'assignee',
    'commentby': 'commenter',
    'reporter': 'reporter',
    'is': 'is',
}

VALID_IS_SEARCH_VALUE = ['open', 'starred']


def GetRedirectURL(project_name):
  return PROJECT_REDIRECT_MAP.get(project_name, None)


def GetNewIssueParams(params: MultiDict, project_name: str):
  new_issue_params = {}

  # Get component and template id.
  template_name = params.get('template', type=str, default='default')
  redirect_component_id, redirect_template_id = (
    redirect_project_template.RedirectProjectTemplate.Get(
    project_name, template_name))
  if redirect_component_id:
    new_issue_params['component'] = redirect_component_id
  if redirect_template_id:
    new_issue_params['template'] = redirect_template_id

  if params.get('summary', type=str):
    new_issue_params['title'] = params.get('summary', type=str)

  if (params.get('description', type=str) or params.get('comment', type=str)):
    new_issue_params['description'] = (
        params.get('description', type=str) or params.get('comment', type=str))

  if params.get('cc', type=str):
    new_issue_params['cc'] = params.get('cc', type=str)

  if params.get('owner', type=str):
    new_issue_params['assignee'] = params.get('owner', type=str).split('@')[0]

  # TODO(b/283983843): redirect when custom field settled. (components)
  return urllib.parse.urlencode(new_issue_params)


def GetSearchQuery(project_name, params):
  search_conds = []

  # can param is the default search query used in monorail.
  # Each project can customize the canned queries.
  # (eg.can=41013401 in Monorail is the Triage Queue.)
  # For redirect we will just support the build in can query as the first step.
  # TODO(b/283983843): support customized can query as needed.
  can_param = params.get(
      'can', type=int, default=tracker_constants.OPEN_ISSUES_CAN)
  # TODO(b/283983843): move the BuiltInQuery to redirect folder.
  default_search_string = tracker_bizobj.GetBuiltInQuery(can_param)
  for cond in default_search_string.split(' '):
    search_conds.append(cond)

  # q param is the user defined search query.
  if params.get('q', type=str):
    search_string = urllib.parse.unquote(params.get('q', type=str))
    for cond in search_string.split(' '):
      search_conds.append(cond)

  query_string = ''
  for cond in search_conds:
    condition_pair = _ConvertSearchCondition(project_name, cond)
    if condition_pair:
      (k, v) = condition_pair
      query_string += ' {0}:{1}'.format(k, v)
  return urllib.parse.urlencode({'q': query_string.strip()})


# Convert monorail search conditions to tracker search conditions.
def _ConvertSearchCondition(project_name, cond):
  cond_pair = []
  # In monorail the search condition can be either ':' or '='.
  if ':' in cond:
    cond_pair = cond.split(':')
  if '=' in cond:
    cond_pair = cond.split('=')

  if len(cond_pair) != 2:
    return None
  # '-' stand for NOT.
  pre = '-' if cond_pair[0].startswith('-') else ''
  key_val = cond_pair[0][1:] if cond_pair[0].startswith('-') else cond_pair[0]

  k, v = _GenerateTrackerSearchKeyValuePair(project_name, key_val, cond_pair[1])
  if not k or not v:
    return None

  return pre + k, v


# Convert the search value to tracker search format.
def _GenerateTrackerSearchKeyValuePair(project_name, key, value):
  if len(value) == 0:
    return None, None
  # Find the related search filter from datastore.
  new_key, new_value = redirect_custom_value.RedirectCustomValue.Get(
      project_name, key, value)
  if new_key and new_value:
    return new_key, new_value

  # If the value is not store in datastore check the general filter set.
  new_key = TRACKER_SEARCH_KEY_MAP.get(key, None)
  if not new_key:
    return None, None

  if new_key == 'is':
    return new_key, value if value in VALID_IS_SEARCH_VALUE else None

  new_value = value.replace(',', '|')
  return new_key, '({})'.format(new_value)
