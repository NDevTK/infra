# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import defaultdict
import json
import logging
import difflib
import time

from google.appengine.ext import ndb

from common.findit_http_client import FinditHttpClient
from gae_libs.http import auth_util
from libs import time_util
from libs.gitiles.gitiles_repository import GitilesRepository
from model.code_coverage import CoverageReportModifier
from model.code_coverage import FileCoverageData
from model.code_coverage import PostsubmitReport
from services import bigquery_helper
from services.code_coverage import code_coverage_util
from services.code_coverage import diff_util

_PAGE_SIZE = 100

# This should be in sync with allowed file types during code generation
# See https://bit.ly/37aP7Vg
_CLANG_SUPPORTED_EXTENSIONS = [
    '.mm', '.S', '.c', '.hh', '.cxx', '.hpp', '.cc', '.cpp', '.ipp', '.h', '.m',
    '.hxx'
]
# List of builders for which coverage metrics to be exported.
# These should be ci builders.
_SOURCE_BUILDERS = {
    'linux-code-coverage': _CLANG_SUPPORTED_EXTENSIONS,
    'win10-code-coverage': _CLANG_SUPPORTED_EXTENSIONS,
    'android-code-coverage': ['.java'],
    'android-code-coverage-native': _CLANG_SUPPORTED_EXTENSIONS,
    'ios-simulator-code-coverage': _CLANG_SUPPORTED_EXTENSIONS,
    'linux-chromeos-code-coverage': _CLANG_SUPPORTED_EXTENSIONS,
    'linux-code-coverage_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'win10-code-coverage_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'android-code-coverage_unit': ['.java'],
    'android-code-coverage-native_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'ios-simulator-code-coverage_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'linux-chromeos-code-coverage_unit': _CLANG_SUPPORTED_EXTENSIONS,
}
_CHROMIUM_SERVER_HOST = 'chromium.googlesource.com'
_CHROMIUM_GERRIT_HOST = 'chromium-review.googlesource.com'
_CHROMIUM_PROJECT = 'chromium/src'
_CHROMIUM_REPO = GitilesRepository(
    FinditHttpClient(
        # Don't log 404 as it is expected
        # e.g. a file can missing at a parent commit.
        no_error_logging_statuses=[404]),
    'https://%s/%s.git' % (_CHROMIUM_SERVER_HOST, _CHROMIUM_PROJECT))
_EXPONENTIAL_BACKOFF_LIMIT_SECONDS = 2048


def _GetFeatureCommits(hashtag):
  """Returns merged commits corresponding to a feature.

  Args:
    hashtag (string): Gerrit hashtag corresponding to the feature.

  Returns:
    Yields a dict which looks like
    {
      'feature_commit' : c1
      'parent_commit': c2
      'files': list of files affected as part of the commit
      'cl_number': Change num of the gerrit CL
    }
    where c1 is the commit_hash corresponding to a feature CL
    submitted as part of the feature and c2 is the hash of the parent commit of
    c1.
  """
  changes = code_coverage_util.FetchMergedChangesWithHashtag(
      _CHROMIUM_GERRIT_HOST, _CHROMIUM_PROJECT, hashtag)
  for change in changes:
    commit = change['current_revision']
    parent_commit = change['revisions'][commit]['commit']['parents'][0][
        'commit']
    files = []
    for file_name, properties in change['revisions'][commit]['files'].items():
      if properties.get('status', '') == 'D':
        # file was deleted at the CL
        continue
      files.append(file_name)
    cl_number = change['_number']
    yield {
        'feature_commit': commit,
        'parent_commit': parent_commit,
        'files': files,
        'cl_number': cl_number
    }


def _GetInterestingLines(latest_lines, feature_commit_lines,
                         parent_commit_lines):
  """Returns interesting_lines in latest_lines corresponding to a feature commit

  interesting_lines are defined as lines, which were modified/added at feature
  commit and have not been modified/deleted since.

  Args:
    latest_lines (list): A list of strings representing the content of a file
      in latest coverage report.
    feature_commit_lines (list): A list of strings representing the content of a
      file right after the feature commit was merged.
    parent_commit_lines (list): A list of strings representing the content of a
      file right before the feature commit was merged.

  Returns:
    A set of integers, representing interesting line numbers in latest_lines.
    Line numbers start from 1.
  """

  def _GetUnmodifiedLinesSinceCommit(latest_lines, commit_lines):
    if not commit_lines:
      return []
    diff_lines = [
        x
        for x in difflib.unified_diff(latest_lines, commit_lines, lineterm='')
    ]
    unchanged_lines = (
        diff_util.generate_line_number_mapping(diff_lines, latest_lines,
                                               commit_lines).keys())
    return unchanged_lines

  lines_unmodified_since_feature_commit = _GetUnmodifiedLinesSinceCommit(
      latest_lines, feature_commit_lines)
  lines_unmodified_since_parent_commit = _GetUnmodifiedLinesSinceCommit(
      latest_lines, parent_commit_lines)

  interesting_lines = [
      x for x in range(1,
                       len(latest_lines) + 1)
      if x in lines_unmodified_since_feature_commit and
      x not in lines_unmodified_since_parent_commit
  ]
  return set(interesting_lines)


def _GetFeatureCoveragePerFile(postsubmit_report, interesting_lines_per_file):
  """Returns line coverage metrics for interesting lines in a file.

  Args:
    postsubmit_report (PostsubmitReport): Full codebase report object containing
      metadata required to fetch filecoverage report e.g. builder, revision etc.
    interesting_lines_per_file (dict): A mapping from filepath to the set of
    interesting lines.

  Returns:
    A tuple of dict and a set. The dict has filepath as key and value
    representing File proto at https://bit.ly/3yry0KR, which contains line
    coverage metric limited to only interesting lines.
    The set contains file names for which no coverage was found.
  """
  coverage_per_file = {}
  files_with_missing_coverage = set()
  for file_path in interesting_lines_per_file.keys():
    file_coverage = FileCoverageData.Get(
        postsubmit_report.gitiles_commit.server_host,
        postsubmit_report.gitiles_commit.project,
        postsubmit_report.gitiles_commit.ref,
        postsubmit_report.gitiles_commit.revision, file_path,
        postsubmit_report.bucket, postsubmit_report.builder)
    if not file_coverage:
      files_with_missing_coverage.add(file_path)
      continue
    total = 0
    covered = 0
    # add a dummy range to simplify logic
    interesting_line_ranges = [{'first': -1, 'last': -1}]
    for line_range in file_coverage.data['lines']:
      for line_num in range(line_range['first'], line_range['last'] + 1):
        if line_num in interesting_lines_per_file[file_path]:
          total += 1
          if line_num == interesting_line_ranges[-1]['last'] + 1:
            interesting_line_ranges[-1]['last'] += 1
          else:
            # Line range gets broken by an uninteresting line
            interesting_line_ranges.append({
                'first': line_num,
                'last': line_num,
                'count': line_range['count']
            })
          if line_range['count'] != 0:
            covered += 1
    if total > 0:
      coverage_per_file[file_path] = {
          'path': file_path,
          'lines': interesting_line_ranges[1:],
          'summaries': [{
              'name': 'line',
              'total': total,
              'covered': covered
          }],
          'revision': postsubmit_report.gitiles_commit.revision
      }
  return coverage_per_file, files_with_missing_coverage


def _CreateModifiedFileCoverage(coverage_per_file, postsubmit_report,
                                gerrit_hashtag, modifier_id):
  """Creates file coverage entities corresponding to a modifier.

  Args:
    coverage_per_file (dict): The dict has filepath as key and value
          representing File proto at https://bit.ly/3yry0KR.
    postsubmit_report (PostsubmitReport): Full codebase coverage report from
          which modified reports are derived.
    gerrit_hashtag (string): Gerrit hashtag corresponding to the feature.
    modifier_id (int): Id of the CoverageReportModifier corresponding to
                          the gerrit hashtag.
  """

  def FlushEntities(entries, total, last=False):
    # Flush the data in a batch and release memory.
    if len(entries) < 100 and not (last and entries):
      return entries, total
    ndb.put_multi(entries)
    total += len(entries)
    logging.info('Dumped %d coverage data entries for feature %s', total,
                 gerrit_hashtag)
    return [], total

  entities = []
  total = 0
  for file_path in coverage_per_file:
    entities.append(
        FileCoverageData.Create(
            server_host=postsubmit_report.gitiles_commit.server_host,
            project=postsubmit_report.gitiles_commit.project,
            ref=postsubmit_report.gitiles_commit.ref,
            revision=postsubmit_report.gitiles_commit.revision,
            path=file_path,
            bucket=postsubmit_report.bucket,
            builder=postsubmit_report.builder,
            data=coverage_per_file[file_path],
            modifier_id=modifier_id))
    entities, total = FlushEntities(entities, total, last=False)
  FlushEntities(entities, total, last=True)


def _CreateBigqueryRows(postsubmit_report, gerrit_hashtag, modifier_id,
                        coverage_per_file, files_with_missing_coverage):
  """Create bigquery rows for files modified as part of a feature.

  Args:
    postsubmit_report (PostsubmitReport): Full codebase report object containing
      metadata corresponding to the report e.g. builder, revision etc.
      gerrit_hashtag (string): Gerrit hashtag corresponding to the feature.
      modifier_id (int): Id of the CoverageReportModifier corresponding to
                          the gerrit hashtag.
      coverage_per_file (dict): Mapping from file_path to the coverage data
                              corresponding to interesting lines in the file.
      files_with_missing_coverage(set): A set of files for which coverage info
                                        was not found.

  Returns:
    A list of dict objects whose keys are column names and values are column
    values corresponding to the schema of the bigquery table.
  """
  bq_rows = []
  for file_path in coverage_per_file.keys():
    bq_rows.append({
        'project':
            postsubmit_report.gitiles_commit.project,
        'revision':
            postsubmit_report.gitiles_commit.revision,
        'builder':
            postsubmit_report.builder,
        'gerrit_hashtag':
            gerrit_hashtag,
        'modifier_id':
            modifier_id,
        'path':
            file_path[2:],
        'total_lines':
            coverage_per_file[file_path]['summaries'][0]['total'],
        'covered_lines':
            coverage_per_file[file_path]['summaries'][0]['covered'],
        'commit_timestamp':
            postsubmit_report.commit_timestamp.isoformat(),
        'insert_timestamp':
            time_util.GetUTCNow().isoformat()
    })
  for file_path in files_with_missing_coverage:
    bq_rows.append({
        'project': postsubmit_report.gitiles_commit.project,
        'revision': postsubmit_report.gitiles_commit.revision,
        'builder': postsubmit_report.builder,
        'gerrit_hashtag': gerrit_hashtag,
        'modifier_id': modifier_id,
        'path': file_path[2:],
        'total_lines': None,
        'covered_lines': None,
        'commit_timestamp': postsubmit_report.commit_timestamp.isoformat(),
        'insert_timestamp': time_util.GetUTCNow().isoformat()
    })
  return bq_rows


def _FetchFileContentAtCommit(file_path, revision):
  """Fetches lines in a file at the specified revision.

  Args:
    file_path (string): chromium/src relative path to file whose content is to
      be fetched. Must start with '//'.
    revision (string): commit hash of the revision.
  Returns:
    A list of strings representing the file content at the revision. If file
    is not found at the revision, an empty list is returned.
  """
  assert file_path.startswith('//'), 'All file path should start with "//".'
  content, status = _CHROMIUM_REPO.GetSourceAndStatus(file_path[2:], revision)
  wait_sec = 1
  # short term qps exceeded, retry with exponential backoff
  while status == 429 and wait_sec < _EXPONENTIAL_BACKOFF_LIMIT_SECONDS:
    wait_sec *= 2
    time.sleep(wait_sec)
    content, status = _CHROMIUM_REPO.GetSourceAndStatus(file_path[2:], revision)
  return content.split('\n') if content else []


def _GetAllowedBuilders():
  return _SOURCE_BUILDERS


def ExportFeatureCoverage(modifier_id):
  """Exports coverage metrics to Datastore and Bigquery for input feature.

  Args:
    modifier_id(int): Id of the CoverageReportModifier corresponding
                      to the feature
  """
  # NDB caches each result in the in-context cache while accessing.
  # This is problematic as due to the size of the result set,
  # cache grows beyond the memory quota. Turn this off to prevent oom errors.
  #
  # Read more at:
  # https://cloud.google.com/appengine/docs/standard/python/ndb/cache#incontext
  # https://github.com/googlecloudplatform/datastore-ndb-python/issues/156#issuecomment-110869490
  context = ndb.get_context()
  context.set_cache_policy(False)
  context.set_memcache_policy(False)

  gerrit_hashtag = CoverageReportModifier.Get(modifier_id).gerrit_hashtag

  builder_to_latest_report = {}
  for builder in _GetAllowedBuilders().keys():
    # Fetch latest full codebase coverage report for the builder
    query = PostsubmitReport.query(
        PostsubmitReport.gitiles_commit.server_host == _CHROMIUM_SERVER_HOST,
        PostsubmitReport.gitiles_commit.project == _CHROMIUM_PROJECT,
        PostsubmitReport.bucket == 'ci', PostsubmitReport.builder == builder,
        PostsubmitReport.visible == True).order(
            -PostsubmitReport.commit_timestamp)
    report = query.fetch(limit=1)[0]
    builder_to_latest_report[builder] = report

  files_deleted_at_latest = defaultdict(list)
  interesting_lines_per_builder_per_file = defaultdict(lambda: defaultdict(set))
  commits = _GetFeatureCommits(gerrit_hashtag)
  for commit in commits:
    for file_path in commit['files']:
      file_path = '//' + file_path

      # Fetch content at latest coverage report commits
      builder_to_latest_content = defaultdict(list)
      file_type_supported_by_builders = False
      for builder, report in builder_to_latest_report.items():
        if not any([
            file_path.endswith(extension)
            for extension in _SOURCE_BUILDERS[builder]
        ]):
          continue
        file_type_supported_by_builders = True
        if file_path in files_deleted_at_latest[builder]:
          continue
        builder_to_latest_content[builder] = _FetchFileContentAtCommit(
            file_path, report.gitiles_commit.revision)

      # Skip processing if file type is not supported by any builder.
      # This is done particularly because there are large configuration files
      # like .xml/.pyl etc, fetching whose content result in oom errors
      if not file_type_supported_by_builders:
        continue

      content_at_feature_commit = _FetchFileContentAtCommit(
          file_path, commit['feature_commit'])
      content_at_parent_commit = _FetchFileContentAtCommit(
          file_path, commit['parent_commit'])
      # File content must be there at feature commit
      assert content_at_feature_commit, (
          "File Content not found for path %s at commit %s (CL: %s)" %
          (file_path, commit['feature_commit'], commit['cl_number']))

      # Calculate interesting lines for file corresponding to each builder
      for builder, report in builder_to_latest_report.items():
        content_at_latest = builder_to_latest_content[builder]
        if content_at_latest:
          interesting_lines = _GetInterestingLines(content_at_latest,
                                                   content_at_feature_commit,
                                                   content_at_parent_commit)
          interesting_lines_per_builder_per_file[builder][file_path] = (
              interesting_lines_per_builder_per_file[builder][file_path]
              | interesting_lines)
        else:
          files_deleted_at_latest[builder].append(file_path)

  # Export feature coverage data to Datastore and Bigquery
  for builder, report in builder_to_latest_report.items():
    coverage_per_file, files_with_missing_coverage = _GetFeatureCoveragePerFile(
        report, interesting_lines_per_builder_per_file[builder])
    _CreateModifiedFileCoverage(coverage_per_file, report, gerrit_hashtag,
                                modifier_id)
    bq_rows = _CreateBigqueryRows(report, gerrit_hashtag, modifier_id,
                                  coverage_per_file,
                                  files_with_missing_coverage)
    if bq_rows:
      bigquery_helper.ReportRowsToBigquery(bq_rows, 'findit-for-me',
                                           'code_coverage_summaries',
                                           'feature_coverage')
      logging.info('Rows added for feature %s and builder %s = %d',
                   gerrit_hashtag, builder, len(bq_rows))
