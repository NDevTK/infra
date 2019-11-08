# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Utility functions for code coverage.

The code coverage data format is defined at:
https://chromium.googlesource.com/infra/infra/+/refs/heads/master/appengine/findit/model/proto/code_coverage.proto.
"""

import base64
import difflib
import json
import urllib2

from google.appengine.ext import ndb

from common.findit_http_client import FinditHttpClient
from components import gitiles
from gae_libs.caches import PickledMemCache
from libs.cache_decorator import Cached
from model.code_coverage import CoveragePercentage
from services.code_coverage import diff_util

# Mapping from metric names to detailed explanations, and one use case is to use
# as the tooltips.
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

# List of patchset kinds that are applicable for sharing coverage data between
# patchsets, and the list of possible kinds is defined in:
# https://gerrit-review.googlesource.com/Documentation/json.html
_NON_CONFLICT_CHANGE_KIND = [
    # Conflict-free merge between the new parent and the prior patch set.
    'TRIVIAL_REBASE',
    # Conflict-free change of first (left) parent of a merge commit.
    'MERGE_FIRST_PARENT_UPDATE',
    # No code changed; same tree and same parent tree.
    'NO_CODE_CHANGE',
    # No changes; same commit message, same tree and same parent tree.
    'NO_CHANGE'
]


class MissingChangeDataException(Exception):
  pass


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


@Cached(PickledMemCache(), namespace='coverage_equivalent_patchsets')
def GetEquivalentPatchsets(host, project, change, patchset):
  """Gets equivalent patchsets that are applicable for sharing coverage data.

  The reason why this is not just the current patchset number is because there
  may have been a succession of "trivial" changes before the current patchset.

  Args:
    host (str): The host name.
    project (str): The project name.
    change (int): The change number.
    patchset (int): The patchset number.

  Returns:
    A list of equivalent patchset numbers in descending order.
  """
  assert isinstance(change, int), 'Change is expected to be an integer'
  assert isinstance(patchset, int), 'Patchset is expected to be an integer'

  change_details = _FetchChangeDetails(host, project, change)
  revisions = change_details['revisions'].values()
  revisions.sort(key=lambda r: r['_number'], reverse=True)
  patchsets = []
  for i, r in enumerate(revisions):
    if i == 0 and change_details['status'] == 'MERGED':
      # Depending on the submit strategy, the last patchset of submitted CLs
      # might be autogenerated and whose kind is labeled as 'REWORK' even though
      # it's actually trivial rebase.
      #
      # This function assumes that the submit strategy is 'Rebase Always' (such
      # as Chromium project), and it may break for projects with other submit
      # strategies, for example: crbug.com/809182.
      #
      # TODO(crbug.com/809182): Make the equivalent patchsets logic generic
      # across all projects. Note that the bug specifically refers to
      # buildbucket, but the same reasonings apply here.
      continue

    current_patchset = r['_number']
    if current_patchset > patchset:
      continue

    patchsets.append(current_patchset)
    if r['kind'] not in _NON_CONFLICT_CHANGE_KIND:
      # If this revision was a non-trivial change, don't consider patchsets
      # prior to it.
      break

  return patchsets


def _FetchChangeDetails(host, project, change):
  """Fetches change detail for a given change.

  Args:
    host (str): The url of the host.
    project (str): The project name.
    change (int): The change number.

  Returns:
    A dict whose format conforms to the ChangeInfo object:
    https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#change-info
  """
  # Uses the Get Change API to get and parse the details of this change.
  # https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#get-change.
  template_to_get_change = (
      'https://%s/changes/%s?o=ALL_REVISIONS&o=SKIP_MERGEABLE')
  url = template_to_get_change % (host, _GetChangeId(project, change))
  status_code, response, _ = FinditHttpClient().Get(url)
  _CheckChangeDetailsResponseCode(status_code, response)

  # Remove XSSI magic prefix
  if response.startswith(')]}\''):
    response = response[4:]

  return json.loads(response)


def _CheckChangeDetailsResponseCode(status_code, response):
  if status_code == 200:
    return
  error_message = (
      'Failed to get change details with status code: %d, response: %s' %
      (status_code, response))
  if status_code == 404:
    raise MissingChangeDataException(error_message)
  raise RuntimeError(error_message)


def _GetChangeId(project, change):
  """Gets the change id for a given change.

  Args:
    project (str): The project name.
    change (int): The change number.

  Returns:
    A string representing a change id according to:
    https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#change-id
  """
  project_quoted = urllib2.quote(project, safe='')
  return '%s~%d' % (project_quoted, change)


def DecompressLineRanges(line_ranges):
  """Decompress the lines ranges data to a flat format.

  For example:
  [
    {
      "count": 1,
      "first": 165, // inclusive
      "last": 166 // inclusive
    }
  ]

  After decompressing, it becomes:
  [
    {
      "line": 165,
      "count": 1
    },
    {
      "line": 166,
      "count": 1
    }
  ]

  Args:
    line_ranges: A list of dict, with format
                 [{"first": int, "last": int, "count": int}, ...], and note that
                 the [first, last] are both inclusive.

  Returns:
    A list of dict, with format: [{"line": int, "count": int}].
  """
  decompressed_lines = []
  for line_range in line_ranges:
    for line_num in xrange(line_range['first'], line_range['last'] + 1):
      decompressed_lines.append({
          'line': line_num,
          'count': line_range['count'],
      })

  return decompressed_lines


def CompressLines(lines):
  """Compress the lines data to ranges.

  This is a reverse-operation of DecompressLines.

  Args:
    lines: A list of dict, with format: [{"line": int, "count": int}].

  Returns:
    A list of dict, with format: {"first": int, "last": int, "count": int}, and
    note that the [first, last] are both inclusive.
  """
  range_start_index = 0
  line_ranges = []
  for i in xrange(1, len(lines) + 1):
    is_continous_line = (
        i < len(lines) and lines[i]['line'] == lines[i - 1]['line'] + 1)
    has_same_count = (
        i < len(lines) and lines[i]['count'] == lines[i - 1]['count'])
    if is_continous_line and has_same_count:
      continue

    line_ranges.append({
        'first': lines[range_start_index]['line'],
        'last': lines[i - 1]['line'],
        'count': lines[range_start_index]['count'],
    })
    range_start_index = i

  return line_ranges


def RebasePresubmitCoverageDataBetweenPatchsets(
    host, project, change, patchset_src, patchset_dest, coverage_data_src):
  """Gets line-number rebased coverage data for a patchset based on another one.

  This function assumes that the two patchsets are a sequence of trivial-rebase
  or commit-message-edit away, for more details, please see
  |GetEquivalentPatchsets|.

  Args:
    host (str): The url of the host.
    project (str): The project name.
    change (int): The change number.
    patchset_src (int): The patchset number to rebase coverage data from.
    patchset_dest (int): The patchset number to rebase coverage data for.
    coverage_data_src (list): A list of File in coverage proto.

  Returns:
    A list of File in coverage proto.
  """
  change_details = _FetchChangeDetails(host, project, change)

  # Cannot directly use the list of files of |patchset_src| as the files to
  # rebase because two patchsets could have different list of changed files even
  # though they're trivial-rebase away. This could happen when for example,
  # patchset1 of a CL modifies file1 and file2, and the changes to file2 are
  # extracted to another CL and landed, then this CL is rebased to patchset2 so
  # that only file1 is modified now.
  patchset_dest_files = _FetchPatchsetFiles(
      host, project, change, _GetPatchsetRevision(patchset_dest,
                                                  change_details))
  coverage_data = [
      d for d in coverage_data_src if d['path'][2:] in patchset_dest_files
  ]

  files_content = {}
  for d in coverage_data:
    f = d['path'][2:]
    files_content[f] = [
        # In luci/components, all paths start with '/'.
        gitiles.get_file_content_async(
            host.replace('-review', ''), project,
            _GetPatchsetRevision(ps, change_details), '/' + f)
        for ps in (patchset_src, patchset_dest)
    ]

  coverage_data_dest = []
  for file_data_src in coverage_data:
    content_src, content_dest = [
        fut.get_result() for fut in files_content[file_data_src['path'][2:]]
    ]
    diff_lines = list(
        difflib.unified_diff(content_src.splitlines(),
                             content_dest.splitlines()))
    mapping = diff_util.generate_line_number_mapping(diff_lines,
                                                     content_src.splitlines(),
                                                     content_dest.splitlines())

    lines_src = DecompressLineRanges(file_data_src['lines'])
    lines_dest = []
    for line in lines_src:
      if line['line'] not in mapping:
        continue

      lines_dest.append({
          'line': mapping[line['line']][0],
          'count': line['count']
      })

    blocks_src = file_data_src.get('uncovered_blocks', [])
    blocks_dest = []
    for block in blocks_src:
      if block['line'] not in mapping:
        continue

      blocks_dest.append({
          'line': mapping[block['line']][0],
          'ranges': block['ranges'],
      })

    line_data_dest = {
        'path': file_data_src['path'],
        'lines': CompressLines(lines_dest),
    }
    if blocks_dest:
      line_data_dest['uncovered_blocks'] = blocks_dest

    coverage_data_dest.append(line_data_dest)

  return coverage_data_dest


def _FetchPatchsetFiles(host, project, change, patchset_revision):
  """Fetches the list of files modified, added or deleted by a patchset.

  Args:
    host (str): The url of the host.
    project (str): The project name.
    change (int): The change number.
    patchset_revision (str): The commit id of the patchset.

  Returns:
    A list of dict that conforms to the FileInfo object:
    https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#file-info
  """
  # https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#list-files
  template_to_get_change = 'https://%s/changes/%s/revisions/%s/files'
  url = template_to_get_change % (host, _GetChangeId(project, change),
                                  patchset_revision)
  status_code, response, _ = FinditHttpClient().Get(url)
  _CheckChangeDetailsResponseCode(status_code, response)

  # Remove XSSI magic prefix
  if response.startswith(')]}\''):
    response = response[4:]

  return json.loads(response)


def CalculateAbsolutePercentages(coverage_data):
  """Calculates absolute coverage percentages for the given coverage data.

  Args:
    coverage_data (list): A list of File in coverage proto.

  Returns:
    A list of CoveragePercentage model entities.
  """
  results = []
  for per_file_data in coverage_data:
    covered_lines = 0
    total_lines = 0
    for range_data in per_file_data['lines']:
      num_lines = range_data['last'] - range_data['first'] + 1
      total_lines += num_lines
      covered_lines += num_lines if range_data['count'] > 0 else 0

    assert total_lines > 0, (
        'Valid coverage data is expected to have at least one executable file.')

    results.append(
        CoveragePercentage(
            path=per_file_data['path'],
            total_lines=total_lines,
            covered_lines=covered_lines))

  return results


def CalculateIncrementalPercentages(host, project, change, patchset,
                                    coverage_data):
  """Calculates incremental coverage percentages for the given coverage data.

  Here incremental means that ONLY lines added or modified by the patchset are
  taken into consideration.

  Args:
    host (str): The url of the host.
    project (str): The project name.
    change (int): The change number.
    patchset (int): The patchset number.
    coverage_data (list): A list of File in coverage proto.

  Returns:
    A list of CoveragePercentage model entities.
  """
  change_details = _FetchChangeDetails(host, project, change)
  diff = _FetchDiffForPatchset(host, project, change,
                               _GetPatchsetRevision(patchset, change_details))
  added_lines = diff_util.parse_added_line_num_from_unified_diff(
      diff.splitlines())

  results = []
  for per_file_data in coverage_data:
    path = per_file_data['path'][2:]
    if path not in added_lines:
      continue

    covered_lines = 0
    total_lines = 0
    for range_data in per_file_data['lines']:
      for line_num in xrange(range_data['first'], range_data['last'] + 1):
        if line_num not in added_lines[path]:
          continue

        total_lines += 1
        covered_lines += 1 if range_data['count'] > 0 else 0

    if total_lines == 0:
      continue

    results.append(
        CoveragePercentage(
            path=per_file_data['path'],
            total_lines=total_lines,
            covered_lines=covered_lines))

  return results


def _FetchDiffForPatchset(host, project, change, patchset_revision):
  """Fetches unified diff for a given patchset.

  Args:
    host (str): The url of the host.
    project (str): The project name.
    change (int): The change number.
    patchset (int): The patchset number.
    patchset_revision (str): The commit id of the patchset.

  Returns:
    A string representing the unified diff for all the files in the patchset.
  """
  # https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#get-patch
  url_template = 'https://%s/changes/%s/revisions/%s/patch'
  url = url_template % (host, _GetChangeId(project, change), patchset_revision)
  status_code, response, _ = FinditHttpClient().Get(url)
  _CheckChangeDetailsResponseCode(status_code, response)

  return base64.b64decode(response)


def _GetPatchsetRevision(patchset, change_details):
  """Gets the corresponding revision of a given patchset.

  Args:
    patchset (int): The patchset number.
    change_details (dict): The format conforms to the ChangeInfo object:
                           https://gerrit-review.googlesource.com/Documentation/rest-api-changes.html#change-info

  Returns:
    Returns a string representing the revision if found in the change details,
    otherwise, a runtime error is raised.
  """
  for revision, value in change_details['revisions'].iteritems():
    if patchset == value['_number']:
      return revision

  raise MissingChangeDataException(
      'Patchset %d is not found in the returned change details: %s' %
      (patchset, json.dumps(change_details)))


def MergeFilesCoverageDataForPerCL(a, b):
  """Merges coverage data files for per-cl coverage data.

  The original data includes 'uncovered blocks', but this merge function drops
  it because it's non-trivial to merge them, and it won't be useful at all
  unless Gerrit decides to support displaying them, and it's still unclear
  whether it will happen or not.

  Args:
    a (list): A list of File in coverage proto.
    b (list): A list of File in coverage proto.

  Returns:
    A list of File in coverage proto.
  """
  merged = []
  a_dict = {i['path']: i for i in a}
  b_dict = {i['path']: i for i in b}
  for path in set(a_dict.keys() + b_dict.keys()):
    if path not in a_dict:
      merged.append(b_dict[path])
      continue

    if path not in b_dict:
      merged.append(a_dict[path])
      continue

    a_lines_dict = {
        i['line']: i['count']
        for i in DecompressLineRanges(a_dict[path]['lines'])
    }
    b_lines_dict = {
        i['line']: i['count']
        for i in DecompressLineRanges(b_dict[path]['lines'])
    }
    merged_lines_dict = {}
    for k in set(a_lines_dict.keys() + b_lines_dict.keys()):
      merged_lines_dict[k] = a_lines_dict.get(k, 0) + b_lines_dict.get(k, 0)

    merged_lines = [{
        'line': l,
        'count': c,
    } for l, c in sorted(merged_lines_dict.iteritems())]
    merged.append({
        'path': path,
        'lines': CompressLines(merged_lines),
    })

  return sorted(merged, key=lambda x: x['path'])
