# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Utility functions for code coverage."""

import json
import logging
import urllib2

from common.findit_http_client import FinditHttpClient
from gae_libs.caches import PickledMemCache
from libs.cache_decorator import Cached

# Mapping from metrics names to detailed explanations, and one use case is to
# use as the tooltips.
_METRIC_NAME_DETAIL_MAPPING = {
    'line': (
        "Line coverage is the percentage of code lines which have been "
        "executed at least once. Only executable lines within function bodies "
        "are considered to be code lines."),
    'function': (
        "Function coverage is the percentage of functions which have been "
        "executed at least once. A function is considered to be executed if "
        "any of its instantiations are executed."),
    'region': (
        "Region coverage is the percentage of code regions which have been "
        "executed at least once. A code region may span multiple lines (e.g in "
        "a large function body with no control flow). However, it's also "
        "possible for a single line to contain multiple code regions (e.g in "
        "'return x || y &amp;&amp; z')."),
    'branch': (
        "Branch coverage is the percentage of branches from each decision "
        "point is executed at least once."),
    'instruction': (
        "Java instruction coverage is the percentage of the Java byte code "
        "instructions which have been executed at least once."),
}

# List of patchset kind that are applicable for sharing coverage data between
# patchsets, and the list of possible kinds is defined in:
# https://gerrit-review.googlesource.com/Documentation/json.html
#
# REWORK
# Nontrivial content changes.
#
# TRIVIAL_REBASE
# Conflict-free merge between the new parent and the prior patch set.
#
# MERGE_FIRST_PARENT_UPDATE
# Conflict-free change of first (left) parent of a merge commit.
#
# NO_CODE_CHANGE
# No code changed; same tree and same parent tree.
#
# NO_CHANGE
# No changes; same commit message, same tree and same parent tree.
_NON_CONFLICT_CHANGE_KIND = [
    'TRIVIAL_REBASE', 'MERGE_FIRST_PARENT_UPDATE', 'NO_CODE_CHANGE', 'NO_CHANGE'
]

# Duration (in seconds) to cache equivalent patchsets for a given patchset.
_EQUIVALENT_PATCHSETS_CACHE_EXPIRE_TIME_SECONDS = 7 * 24 * 60 * 60


def GetMetricsBasedOnCoverageTool(coverage_tool):
  """Gets a list of metrics for the given coverage tool.

  Args:
    coverage_tool(str): Name of the coverage tool, such as clang and jacoco.

  Returns:
    A list of dict of following format:
    {'name': clang, 'detail': blala}, where the name is the name of the metric
    and detail is an explanation of what the metric stands for.
  """
  assert coverage_tool in ('clang', 'jacoco'), (
      'Unrecognized coverage tool: %s' % coverage_tool)

  metrics = []
  if coverage_tool == 'clang':
    metrics = ['line', 'function', 'region']
  else:
    metrics = ['line', 'branch', 'instruction']

  return [{
      'name': m,
      'detail': _METRIC_NAME_DETAIL_MAPPING.get(m, '')
  } for m in metrics]


@Cached(
    PickledMemCache(),
    namespace='coverage_equivalent_patchsets',
    expire_time=_EQUIVALENT_PATCHSETS_CACHE_EXPIRE_TIME_SECONDS)
def GetEquivalentPatchsets(host, project, change, patchset):
  """Gets equivalent patchsets that are applicable for sharing coverag data.

  The reason why this is not just the current patchset number is because there
  may have been a succession of "trivial" changes before the current patchset.

  Args:
    host (str): The host name.
    project (str): The project name.
    change (str): The change number.
    patchset (str): The patchset number.

  Returns:
    A list of patchset numbers in descending order and the current patchset is
    included. In case this method fails to communicate with Gerrit, an mepty
    list is returned.
  """
  project_quoted = urllib2.quote(project, safe='')

  # Uses the Get Change API to get and parse the details of this change.
  # https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#get-change.
  template_to_get_change = (
      'https://%s/changes/%s~%s?o=ALL_REVISIONS&o=SKIP_MERGEABLE')
  url = template_to_get_change % (host, project_quoted, change)
  status_code, content, _ = FinditHttpClient().Get(url)
  if status_code != 200:
    logging.error('Failed to get change details with status code: %d',
                  status_code)
    return []

  # Remove XSSI magic prefix
  if content.startswith(')]}\''):
    content = content[4:]
  change_details = json.loads(content)

  revisions = change_details['revisions'].values()
  revisions.sort(key=lambda r: r['_number'], reverse=True)
  patchsets = []
  for r in revisions:
    current_patchset = r['_number']
    if current_patchset > patchset:
      continue

    patchsets.append(current_patchset)
    if r['kind'] not in _NON_CONFLICT_CHANGE_KIND:
      # If this revision was a non-trivial change, don't consider patchsets
      # prior to it.
      break

  return patchsets
