# Copyright 2021 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import six

from collections import defaultdict
import logging
import difflib
import queue
from threading import Thread
import time

from google.appengine.ext import ndb

from common.findit_http_client import FinditHttpClient
from gae_libs.gitiles.cached_gitiles_repository import CachedGitilesRepository
from handlers.code_coverage import utils
from libs import time_util
from model.code_coverage import CoverageReportModifier
from model.code_coverage import FileCoverageData
from model.code_coverage import PostsubmitReport
from model.code_coverage import SummaryCoverageData
from services import bigquery_helper
from services.code_coverage import code_coverage_util
from services.code_coverage import diff_util
from services.code_coverage import summary_coverage_aggregator

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
    'mac-code-coverage': _CLANG_SUPPORTED_EXTENSIONS,
    'win10-code-coverage': _CLANG_SUPPORTED_EXTENSIONS,
    'android-code-coverage': ['.java'],
    'android-code-coverage-native': _CLANG_SUPPORTED_EXTENSIONS,
    'android-cronet-code-coverage-native': _CLANG_SUPPORTED_EXTENSIONS,
    'ios-simulator-code-coverage': _CLANG_SUPPORTED_EXTENSIONS,
    'linux-chromeos-code-coverage': _CLANG_SUPPORTED_EXTENSIONS,
    'linux-code-coverage_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'mac-code-coverage_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'win10-code-coverage_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'android-code-coverage_unit': ['.java'],
    'android-code-coverage-native_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'android-cronet-code-coverage-native_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'ios-simulator-code-coverage_unit': _CLANG_SUPPORTED_EXTENSIONS,
    'linux-chromeos-code-coverage_unit': _CLANG_SUPPORTED_EXTENSIONS,
}
_CHROMIUM_SERVER_HOST = 'chromium.googlesource.com'
_CHROMIUM_GERRIT_HOST = 'chromium-review.googlesource.com'
_CHROMIUM_PROJECT = 'chromium/src'
_EXPONENTIAL_BACKOFF_LIMIT_SECONDS = 2048
# Http status codes which would not be logged while requesting data from gitiles
# We do not log 404, because it is a valid scenario e.g. when a file gets added
# after reference_commit. We don't log 429, because we have exponential backoff
# mechanism to take care of it
_NO_ERROR_LOGGING_STATUSES = [404, 429]


def _GetCandidateCommits(modifier_id):
  """Returns merged commits corresponding to a gerrit filter.

  Args:
    modified_id (int): id of the CoverageReportModifier entity which would
                      determine what commits are considered interesting

  Returns:
    Yields a dict which looks like
    {
      'candidate_commit' : c1
      'parent_commit': c2
      'files': list of files affected as part of the commit
      'cl_number': Change num of the gerrit CL
    }
    where c1 is the commit_hash corresponding to a gerrit filter
    and c2 is the hash of the parent commit of c1.
  """
  modifier = CoverageReportModifier.Get(modifier_id)
  changes = code_coverage_util.FetchMergedChanges(
      host=_CHROMIUM_GERRIT_HOST,
      project=_CHROMIUM_PROJECT,
      hashtag=modifier.gerrit_hashtag,
      author=modifier.author)
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
        'candidate_commit': commit,
        'parent_commit': parent_commit,
        'files': files,
        'cl_number': cl_number
    }


def _GetInterestingLines(latest_lines, candidate_commit_lines,
                         parent_commit_lines):
  """Returns interesting lines in latest_lines corresponding to candidate commit

  interesting_lines are defined as lines, which were modified/added at candidate
  commit and have not been modified/deleted since.

  Args:
    latest_lines (list): A list of strings representing the content of a file
      in latest coverage report.
    candidate_commit_lines (list): A list of strings representing the content
        of a file right after the candidate commit was merged.
    parent_commit_lines (list): A list of strings representing the content of a
      file right before the candidate commit was merged.

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

  lines_unmodified_since_candidate_commit = _GetUnmodifiedLinesSinceCommit(
      latest_lines, candidate_commit_lines)
  lines_unmodified_since_parent_commit = _GetUnmodifiedLinesSinceCommit(
      latest_lines, parent_commit_lines)

  interesting_lines = [
      x for x in range(1,
                       len(latest_lines) + 1)
      if x in lines_unmodified_since_candidate_commit and
      x not in lines_unmodified_since_parent_commit
  ]
  return set(interesting_lines)


def _GetCoveragePerFile(postsubmit_report, interesting_lines_per_file):
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
  for file_path, interesting_lines in interesting_lines_per_file.items():
    # Only export result for files which have non zero number of interesting
    # lines. If this is not the case, it means the changes done as part of the
    # a candidate commit were later overridden by a non-candidate commit.
    # Therefore, it isn't useful to export coverage info for such files.
    if not interesting_lines:
      continue
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
    interesting_line_ranges = [{'first': -1, 'last': -1, 'count': -1}]
    for line_range in file_coverage.data['lines']:
      for line_num in range(line_range['first'], line_range['last'] + 1):
        # `total` signifies number of lines which are interesting
        # AND instrumented. This means There could be lines which are
        # interesting but are not included in `total` because they were not
        # instrumented. e.g. it's a comment line
        if line_num in interesting_lines:
          total += 1
          if (line_num == interesting_line_ranges[-1]['last'] + 1 and
              line_range['count'] == interesting_line_ranges[-1]['count']):
            # Append to the last interesting line range if line numbers are
            # continuous and they share the same execution count
            interesting_line_ranges[-1]['last'] += 1
          else:
            # Line range gets broken by an uninteresting line
            # or new line range with a different execution count is encountered
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


def _FlushEntities(entities, total, last=False):
  """Creates datastore entities in a batched manner"""
  if len(entities) < 100 and not (last and entities):
    return entities, total

  ndb.put_multi(entities)
  total += len(entities)
  logging.info('Dumped %d coverage data entities', total)

  return [], total


def _CreateModifiedFileCoverage(coverage_per_file, postsubmit_report,
                                modifier_id):
  """Creates file coverage entities corresponding to a modifier.

  Args:
    coverage_per_file (dict): The dict has filepath as key and value
          representing File proto at https://bit.ly/3yry0KR.
    postsubmit_report (PostsubmitReport): Full codebase coverage report from
          which modified reports are derived.
    modifier_id (int): Id of the CoverageReportModifier corresponding to
                          the gerrit hashtag.
  """
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
    entities, total = _FlushEntities(entities, total, last=False)
  _FlushEntities(entities, total, last=True)


def _CreateModifiedDirSummaryCoverage(directories_coverage, postsubmit_report,
                                      modifier_id):
  """Exports directory summary coverage entities to datastore.

  Args:
    directories_coverage(dict): Mapping from directory path to corresponding
                                coverage data.
    postsubmit_report(PostsubmitReport): Full codebase report for which
                                referenced entities are being exported.
    modifier_id(int): Id of the CoverageReportModifier corresponding to the
                      reference commit.
  """
  entities = []
  total = 0
  logging.info("Dumping directory coverage")
  for path, data in directories_coverage.items():
    entity = SummaryCoverageData.Create(
        postsubmit_report.gitiles_commit.server_host,
        postsubmit_report.gitiles_commit.project,
        postsubmit_report.gitiles_commit.ref,
        postsubmit_report.gitiles_commit.revision, 'dirs', path,
        postsubmit_report.bucket, postsubmit_report.builder, data, modifier_id)
    entities.append(entity)
    entities, total = _FlushEntities(entities, total)
  _FlushEntities(entities, total, last=True)


def _CreateBigqueryRows(postsubmit_report, run_id, modifier_id,
                        coverage_per_file, files_with_missing_coverage,
                        interesting_lines_per_file):
  """Create bigquery rows for files modified as part of a candiate commit.

  Args:
    postsubmit_report (PostsubmitReport): Full codebase report object containing
      metadata corresponding to the report e.g. builder, revision etc.
      modifier_id (int): Id of the CoverageReportModifier corresponding to
                          the gerrit hashtag.
      run_id (int): Unique id of this execution of the algorithm. This is
                    exported to bigquery, so that downstream reporting logic can
                    group rows added in one execution of the algorithm.
      coverage_per_file (dict): Mapping from file_path to the coverage data
                              corresponding to interesting lines in the file.
      files_with_missing_coverage(set): A set of files for which coverage info
                                        was not found.
      interesting_lines_per_file (dict): A dict mapping file name to interesting
                                        lines

  Returns:
    A list of dict objects whose keys are column names and values are column
    values corresponding to the schema of the bigquery table.
  """
  modifier = CoverageReportModifier.Get(modifier_id)
  bq_rows = []
  for file_path in coverage_per_file.keys():
    bq_rows.append({
        'project':
            postsubmit_report.gitiles_commit.project,
        'revision':
            postsubmit_report.gitiles_commit.revision,
        'builder':
            postsubmit_report.builder,
        'run_id':
            run_id,
        'gerrit_hashtag':
            modifier.gerrit_hashtag,
        'author':
            modifier.author,
        'modifier_id':
            modifier_id,
        'path':
            file_path[2:],
        'total_lines':
            coverage_per_file[file_path]['summaries'][0]['total'],
        'covered_lines':
            coverage_per_file[file_path]['summaries'][0]['covered'],
        'interesting_lines':
            len(interesting_lines_per_file[file_path]),
        'commit_timestamp':
            postsubmit_report.commit_timestamp.isoformat(),
        'insert_timestamp':
            time_util.GetUTCNow().isoformat()
    })
  for file_path in files_with_missing_coverage:
    bq_rows.append({
        'project': postsubmit_report.gitiles_commit.project,
        'revision': postsubmit_report.gitiles_commit.revision,
        'run_id': run_id,
        'builder': postsubmit_report.builder,
        'gerrit_hashtag': modifier.gerrit_hashtag,
        'author': modifier.author,
        'modifier_id': modifier_id,
        'path': file_path[2:],
        'total_lines': None,
        'covered_lines': None,
        'interesting_lines': len(interesting_lines_per_file[file_path]),
        'commit_timestamp': postsubmit_report.commit_timestamp.isoformat(),
        'insert_timestamp': time_util.GetUTCNow().isoformat()
    })
  return bq_rows


def _FetchFileContentAtCommit(file_path, revision, manifest,
                              file_content_queue):
  """Fetches lines in a file at the specified revision.

  Args:
    file_path (string): chromium/src relative path to file whose content is to
      be fetched. Must start with '//'.
    revision (string): commit hash of the revision.
    manifest (DependencyRepository): Entity containing mapping from path prefix
                                     to corresponding repo.
    file_content_queue (Queue): Queue which holds the output.
  """
  assert file_path.startswith('//'), 'All file path should start with "//".'
  assert revision, 'A valid revision is required'
  dependency = utils.GetMatchedDependencyRepository(manifest, file_path)
  assert dependency, ('%s file does not belong to any dependency repository' %
                      file_path)
  # Calculate the relative path to the root of the dependency repository itself.
  relative_file_path = file_path[len(dependency.path):]
  repo = CachedGitilesRepository(
      FinditHttpClient(no_error_logging_statuses=_NO_ERROR_LOGGING_STATUSES),
      dependency.project_url)
  content, status = repo.GetSourceAndStatus(relative_file_path, revision)
  wait_sec = 1
  # short term qps exceeded, retry with exponential backoff
  while status == 429 and wait_sec < _EXPONENTIAL_BACKOFF_LIMIT_SECONDS:
    wait_sec *= 2
    time.sleep(wait_sec)
    content, status = repo.GetSourceAndStatus(relative_file_path, revision)
  if wait_sec >= _EXPONENTIAL_BACKOFF_LIMIT_SECONDS:
    logging.warning(
        "Couldn't fetch content for %s at revision %s due to exceeding qps",
        file_path, revision)
  file_content_queue.put((revision, content.split('\n') if content else []))


def _GetAllowedBuilders():
  return _SOURCE_BUILDERS


def ExportCoverage(modifier_id, run_id):
  """Exports coverage metrics to Datastore and Bigquery for gerrit based filter.

  Args:
    modifier_id(int): Id of the CoverageReportModifier corresponding
                      to the gerrit filter
    run_id(int): Unique id corresponding to the current execution of
                this algorithm.
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

  modifier = CoverageReportModifier.Get(modifier_id)
  assert modifier.gerrit_hashtag or modifier.author, (
      'One of gerrit_hashtag or author must be present')
  builder_to_latest_report = {}
  for builder in _GetAllowedBuilders().keys():
    # Fetch latest full codebase coverage report for the builder
    query = PostsubmitReport.query(
        PostsubmitReport.gitiles_commit.server_host == _CHROMIUM_SERVER_HOST,
        PostsubmitReport.gitiles_commit.project == _CHROMIUM_PROJECT,
        PostsubmitReport.bucket == 'ci', PostsubmitReport.builder == builder,
        PostsubmitReport.visible == True, PostsubmitReport.modifier_id ==
        0).order(-PostsubmitReport.commit_timestamp)
    report = query.fetch(limit=1)[0]
    builder_to_latest_report[builder] = report
  # manifest is supposed to be same across all chromium coverage builders
  manifest = list(builder_to_latest_report.values())[0].manifest
  file_content_queue = queue.Queue()
  files_deleted_at_latest = defaultdict(list)
  interesting_lines_per_builder_per_file = defaultdict(lambda: defaultdict(set))
  commits = _GetCandidateCommits(modifier_id)
  for commit in commits:
    for file_path in commit['files']:
      file_path = '//' + file_path

      # Fetch content at latest coverage report commits
      builder_to_latest_content_thread = {}
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
        builder_to_latest_content_thread[builder] = Thread(
            target=_FetchFileContentAtCommit,
            args=(file_path, report.gitiles_commit.revision, manifest,
                  file_content_queue))
        builder_to_latest_content_thread[builder].start()

      # Skip processing if file type is not supported by any builder.
      # This is done particulartly because there are large configuration files
      # like .xml/.pyl etc, fetching whose content result in oom errors
      if not file_type_supported_by_builders:
        continue

      # Fetch content at candidate and parent commit
      candidate_commit_thread = Thread(
          target=_FetchFileContentAtCommit,
          args=(file_path, commit['candidate_commit'], manifest,
                file_content_queue))
      parent_commit_thread = Thread(
          target=_FetchFileContentAtCommit,
          args=(file_path, commit['parent_commit'], manifest,
                file_content_queue))
      candidate_commit_thread.start()
      parent_commit_thread.start()

      # Wait for all threads to finish
      for thread in builder_to_latest_content_thread.values():
        thread.join()
      candidate_commit_thread.join()
      parent_commit_thread.join()

      # Consume content from all threads
      contents = defaultdict(list)
      while not file_content_queue.empty():
        # It's correct to do block=False as all threads have been joined before.
        k, v = file_content_queue.get(block=False)
        contents[k] = v

      # File content must be there at candidate commit
      assert contents[commit['candidate_commit']], (
          "File Content not found for path %s at commit %s (CL: %s)" %
          (file_path, commit['candidate_commit'], commit['cl_number']))

      # Calculate interesting lines for file corresponding to each builder
      for builder, report in builder_to_latest_report.items():
        # Check again if the file type is supported by the builder. This is
        # needed because content_at_latest could still be available if two
        # builders have PostsubmitReports generated at the same commit
        if not any([
            file_path.endswith(extension)
            for extension in _SOURCE_BUILDERS[builder]
        ]):
          continue
        content_at_latest = contents[report.gitiles_commit.revision]
        if content_at_latest:
          interesting_lines = _GetInterestingLines(
              content_at_latest, contents[commit['candidate_commit']],
              contents[commit['parent_commit']])
          interesting_lines_per_builder_per_file[builder][file_path] = (
              interesting_lines_per_builder_per_file[builder][file_path]
              | interesting_lines)
        else:
          files_deleted_at_latest[builder].append(file_path)
    logging.info("process commit %s", commit['candidate_commit'])
  # Export coverage data to Datastore and Bigquery
  for builder, report in builder_to_latest_report.items():
    coverage_per_file, files_with_missing_coverage = _GetCoveragePerFile(
        report, interesting_lines_per_builder_per_file[builder])
    _CreateModifiedFileCoverage(coverage_per_file, report, modifier_id)
    aggregator = summary_coverage_aggregator.SummaryCoverageAggregator(
        metrics=['line'])
    for file_coverage in coverage_per_file.values():
      aggregator.consume_file_coverage(file_coverage)
    directory_coverage = aggregator.produce_summary_coverage()
    _CreateModifiedDirSummaryCoverage(directory_coverage, report, modifier_id)
    # Create a top level PostsubmitReport entity with visible = True
    if directory_coverage:
      modified_report = PostsubmitReport.Create(
          server_host=report.gitiles_commit.server_host,
          project=report.gitiles_commit.project,
          ref=report.gitiles_commit.ref,
          revision=report.gitiles_commit.revision,
          bucket=report.bucket,
          builder=report.builder,
          commit_timestamp=report.commit_timestamp,
          manifest=report.manifest,
          summary_metrics=directory_coverage['//']['summaries'],
          build_id=report.build_id,
          visible=True,
          modifier_id=modifier_id)
      modified_report.put()

    bq_rows = _CreateBigqueryRows(
        report, run_id, modifier_id, coverage_per_file,
        files_with_missing_coverage,
        interesting_lines_per_builder_per_file[builder])
    if bq_rows:
      bigquery_helper.ReportRowsToBigquery(bq_rows, 'findit-for-me',
                                           'code_coverage_summaries',
                                           'gerrit_filter_coverage')
      logging.info('Rows added for modifier %d and builder %s = %d',
                   modifier_id, builder, len(bq_rows))
