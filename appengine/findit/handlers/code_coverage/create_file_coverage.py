# Copyright 2022 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import logging
import os
from xml.etree import ElementTree

from google.appengine.ext import ndb
from google.protobuf.field_mask_pb2 import FieldMask

from common.findit_http_client import FinditHttpClient
from common.waterfall.buildbucket_client import GetV2Build
from gae_libs.handlers.base_handler import BaseHandler, Permission
from handlers.code_coverage import utils
from model.code_coverage import FileCoverageByAuthor
from model.code_coverage import CoverageReportModifier

_CHROMIUM_TO_GOOGLER_MAPPING_PATH = '/cr2goog/cr2goog.txt'


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


def _GetJacocoCoverageData(gs_path):
  """Fetches and returns jacoco coverage data from GCS.

  Returns a dict of form
  {
    'myfile1.java': {...},
    'myfile2.java': {...}
  }
  The keys are file paths and values are another dict, where keys are
  line numbers and values are booleans indicating if the line
  is covered or not.
  """
  logging.info('Fetching data from %s', gs_path)
  content = utils.GetFileContentFromGs(gs_path)
  assert content, 'Failed to fetch raw coverage json data from %s' % gs_path
  root = ElementTree.fromstring(content)
  response = {}
  for package in root.findall('package'):
    package_path = package.attrib['name']
    for source_file in package.findall('sourcefile'):
      source_file_name = source_file.attrib['name']
      file_path = os.path.join(package_path, source_file_name)
      response[file_path] = {}
      for line in source_file.findall('line'):
        line_number = int(line.attrib['nr'])
        covered = int(line.attrib['ci']) > 0
        response[file_path][line_number] = covered
  return response


def _GetChromiumToGooglerMapping():
  content = utils.GetFileContentFromGs(_CHROMIUM_TO_GOOGLER_MAPPING_PATH)
  assert content, ('Failed to fetch account mappings data from %s' %
                   _CHROMIUM_TO_GOOGLER_MAPPING_PATH)
  return json.loads(content)


def _FlushEntriesToDatastore(entries, total, last=False):
  # Flush the data in a batch and release memory.
  if len(entries) < 100 and not (last and entries):
    return entries, total

  ndb.put_multi(entries)
  total += len(entries)
  logging.info('Dumped %d coverage data entries', total)

  return [], total


def _CreateFileCoverageProto():
  pass


def _CreateLineRange():
  pass


def _CreateSummaries():
  pass


class CreateFileCoverageMetrics(BaseHandler):
  """Experimental code to test design of new code coverage tool."""
  PERMISSION_LEVEL = Permission.APP_SELF

  def HandleGet(self):
    build_id = int(self.request.get('build_id'))
    build = GetV2Build(
        build_id,
        fields=FieldMask(paths=['id', 'output.properties', 'input', 'builder']))

    # Convert the Struct to standard dict, to use .get, .iteritems etc.
    properties = dict(build.output.properties.items())
    gs_bucket = properties.get('coverage_gs_bucket')
    gs_metadata_dirs = properties.get('coverage_metadata_gs_paths')
    mimic_builder_names = properties.get('mimic_builder_names')

    account_mapping = _GetChromiumToGooglerMapping()

    for gs_metadata_dir, mimic_builder_name in zip(gs_metadata_dirs,
                                                   mimic_builder_names):
      full_gs_metadata_dir = '/%s/%s' % (gs_bucket, gs_metadata_dir)
      coverage_xml_path = '%s/coverage.xml' % full_gs_metadata_dir
      line_coverage_per_file = _GetJacocoCoverageData(coverage_xml_path)
      blamelist_json_path = '%s/blame.json' % full_gs_metadata_dir
      blamelist = _GetBlameListData(blamelist_json_path)

      entities = []
      num_entities = 0
      for file_name in blamelist:
        for author in blamelist[file_name]:
          total = 0
          covered = 0
          for line_number in blamelist[file_name][author]:
            if line_number in line_coverage_per_file[file_name]:
              total += 1
              covered += int(line_coverage_per_file[file_name][line_number])
          author = account_mapping.get(author, author)
          modifier = CoverageReportModifier.InsertModifierForAuthorIfNeeded(
              author)
          data = CreateFileCoverageProto()
          entity = FileCoverageData.Create(
              server_host=build.input.gitiles_commit.host,
              project=build.input.gitiles_commit.project,
              ref=build.input.gitiles_commit.ref,
              revision=build.input.gitiles_commit.id,
              path='//' + file_name,
              bucket=build.builder.bucket,
              builder=mimic_builder_name,
              data=FileCoverageData,
              author=author,
              total_lines=total,
              covered_lines=covered)
          entities.append(entity)
          entities, num_entities = _FlushEntriesToDatastore(
              entities, num_entities, last=False)
      _FlushEntriesToDatastore(entities, total, last=True)

    return {'return_code': 200}
