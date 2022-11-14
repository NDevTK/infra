# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import defaultdict
import json
import logging

from google.appengine.ext import ndb
from google.protobuf.field_mask_pb2 import FieldMask

from common.waterfall.buildbucket_client import GetV2Build
from handlers.code_coverage import utils

from model.code_coverage import CoverageReportModifier
from model.code_coverage import FileCoverageData
from model.code_coverage import PostsubmitReport
from model.code_coverage import SummaryCoverageData
from services.code_coverage import summary_coverage_aggregator

_CHROMIUM_TO_GOOGLER_MAPPING_PATH = '/cr2goog/cr2goog.txt'
_CHROMIUM_SERVER_HOST = 'chromium.googlesource.com'
_CHROMIUM_PROJECT = 'chromium/src'


def _GetChromiumToGooglerMapping():
  content = utils.GetFileContentFromGs(_CHROMIUM_TO_GOOGLER_MAPPING_PATH)
  assert content, ('Failed to fetch account mappings data from %s' %
                   _CHROMIUM_TO_GOOGLER_MAPPING_PATH)
  return json.loads(content)


def _GetBlameListData(gs_path):
  """Returns blame list data for given set of files.

  Returns a dict which looks like following
  {
    'abc/myfile1.cc': {...},
    'abc/myfile2.cc': {...}
  }

  The keys in the dict are file names, values is another dict,
  where keys are author emails and values are lists containing
  line numbers modified by the author in the last X weeks,
  where X is decided by recipe code.
  """
  logging.info('Fetching data from %s', gs_path)
  content = utils.GetFileContentFromGs(gs_path)
  assert content, 'Failed to fetch blamelist json data from %s' % gs_path
  data = json.loads(content)
  # According to https://developers.google.com/discovery/v1/type-format, certain
  # serialization APIs will automatically convert int64 to string when
  # serializing to JSON, and to facilitate later computations, the following for
  # loops convert them back to int64 (int in Python).
  # The following workaround should be removed when the service migrates away
  # from JSON.
  for file_name in data:
    for author in data[file_name]:
      lines = []
      for line in data[file_name][author]:
        lines.append(int(line))
      data[file_name][author] = lines
  return data


def _FlushEntities(entities, total, last=False):
  """Creates datastore entities in a batched manner"""
  if len(entities) < 100 and not (last and entities):
    return entities, total

  ndb.put_multi(entities)
  total += len(entities)
  logging.info('Dumped %d coverage data entities', total)

  return [], total


def _ExportDirSummaryCoverage(directories_coverage, postsubmit_report,
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


def _CreateAuthorCoverage(postsubmit_report):
  build = GetV2Build(
      postsubmit_report.build_id,
      fields=FieldMask(paths=['id', 'output.properties', 'input', 'builder']))
  # Convert the Struct to standard dict, to use .get, .iteritems etc.
  properties = dict(build.output.properties.items())
  gs_bucket = properties.get('coverage_gs_bucket')
  gs_metadata_dirs = properties.get('coverage_metadata_gs_paths')
  mimic_builder_names = properties.get('mimic_builder_names')
  account_mapping = _GetChromiumToGooglerMapping()

  for gs_metadata_dir, mimic_builder_name in zip(gs_metadata_dirs,
                                                 mimic_builder_names):
    if mimic_builder_name != postsubmit_report.builder:
      continue

    # Step 1. Fetch blamelist from GCS
    full_gs_metadata_dir = '/%s/%s' % (gs_bucket, gs_metadata_dir)
    # coverage_xml_path = '%s/coverage.xml' % full_gs_metadata_dir
    # line_coverage_per_file = _GetJacocoCoverageData(coverage_xml_path)
    blamelist_json_path = '%s/blame.json' % full_gs_metadata_dir
    blamelist = _GetBlameListData(blamelist_json_path)

    total_entities = 0
    modified_file_coverage_entities = defaultdict(lambda: list)
    aggregators = defaultdict(lambda: summary_coverage_aggregator.
                              SummaryCoverageAggregator(metrics=['line']))

    # Step 2. Iterate over each file in blamelist,
    for file_name in blamelist:
      # and find latest file coverage report,
      query = FileCoverageData.query(
          FileCoverageData.gitiles_commit.server_host ==
          postsubmit_report.gitiles_commit.server_host,
          FileCoverageData.gitiles_commit.project ==
          postsubmit_report.gitiles_commit.project,
          FileCoverageData.gitiles_commit.ref ==
          postsubmit_report.gitiles_commit.ref,
          FileCoverageData.gitiles_commit.revision ==
          postsubmit_report.gitiles_commit.revision,
          FileCoverageData.bucket == postsubmit_report.bucket,
          FileCoverageData.builder == postsubmit_report.builder,
          FileCoverageData.path == file_name, FileCoverageData.modifier_id == 0)
      file_coverage = query.fetch(limit=1)[0]
      # and for each author for the file in the blamelist,
      for author in blamelist[file_name]:
        # find corresponding google.com account for a chromium.org account,
        googler = account_mapping.get(author, author)
        # and create an active CoverageReportModifier corresponding
        # to the googler, if needed.
        modifier = CoverageReportModifier.InsertModifierForAuthorIfNeeded(
            googler)
        # Then determine author's contribution to file coverage
        modified_line_coverage = FileCoverageData.GetModifiedLineCoverage(
            file_coverage, blamelist[file_name][author])
        if modified_line_coverage:
          # and create corresponding FileCoverageData entity and process it towards
          # summary coverage.
          entity = FileCoverageData.Create(
              server_host=postsubmit_report.gitiles_commit.server_host,
              project=postsubmit_report.gitiles_commit.project,
              ref=postsubmit_report.gitiles_commit.ref,
              revision=postsubmit_report.gitiles_commit.revision,
              path=file_name,
              bucket=postsubmit_report.bucket,
              builder=postsubmit_report.builder,
              data=modified_line_coverage,
              modifier_id=modifier.id)
          total_entities += 1
          modified_file_coverage_entities[modifier.id].append(entity)
          aggregators[modifier.id].consume_file_coverage(modified_line_coverage)

      # Step 3. Insert all File Coverage entities in Datastore
      for entities in modified_file_coverage_entities.values:
        ndb.put_multi(entities)
        total_entities += len(entities)

    # Step 4. Insert all Summary Coverage and Postsubmit Report
    # entities in datastore
    for modifier_id, aggregator in aggregators:
      modified_directory_coverage = aggregator.produce_summary_coverage()
      _ExportDirSummaryCoverage(modified_directory_coverage, postsubmit_report,
                                modifier_id)
      # Create a top level PostsubmitReport entity with visible = True
      if modified_directory_coverage:
        referenced_report = PostsubmitReport.Create(
            server_host=postsubmit_report.gitiles_commit.server_host,
            project=postsubmit_report.gitiles_commit.project,
            ref=postsubmit_report.gitiles_commit.ref,
            revision=postsubmit_report.gitiles_commit.revision,
            bucket=postsubmit_report.bucket,
            builder=postsubmit_report.builder,
            commit_timestamp=postsubmit_report.commit_timestamp,
            manifest=postsubmit_report.manifest,
            summary_metrics=modified_directory_coverage['//']['summaries'],
            build_id=postsubmit_report.build_id,
            visible=True,
            modifier_id=modifier_id)
        referenced_report.put()


def CreateAuthorCoverage(builder):
  # NDB caches each result in the in-context cache while accessing.
  # This is problematic as due to the size of the result set,
  # cache grows beyond the memory quota. Turn this off to prevent oom errors.
  #
  # Read more at:
  # https://cloud.google.com/appengine/docs/standard/python/ndb/cache#incontext
  # https://github.com/googlecloudplatform/datastore-ndb-python/issues/156#issuecomment-110869490
  context = ndb.get_context()
  context.set_cache_policy(False)
  # Fetch latest full codebase coverage report for the builder
  query = PostsubmitReport.query(
      PostsubmitReport.gitiles_commit.server_host == _CHROMIUM_SERVER_HOST,
      PostsubmitReport.gitiles_commit.project == _CHROMIUM_PROJECT,
      PostsubmitReport.bucket == 'ci', PostsubmitReport.builder == builder,
      PostsubmitReport.visible == True, PostsubmitReport.modifier_id ==
      0).order(-PostsubmitReport.commit_timestamp)
  report = query.fetch(limit=1)[0]
  _CreateAuthorCoverage(report)
