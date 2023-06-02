# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Utils for redirect."""
import urllib
from werkzeug.datastructures import MultiDict

PROJECT_REDIRECT_MAP = {
    'pigweed': 'https://issues.pigweed.dev',
    'git': 'https://git.issues.gerritcodereview.com',
}

MAX_MONORAIL_ISSUE_ID = 10000000


def GetRedirectURL(project_name):
  return PROJECT_REDIRECT_MAP.get(project_name, None)


def GetNewIssueParams(params: MultiDict):
  new_issue_params = {}

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
