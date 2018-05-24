# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import base64
from datetime import datetime
import gzip
import io
import inspect
import logging
import json
import re
import urllib

from common import rpc_util
from common.waterfall import buildbucket_client
from gae_libs.caches import PickledMemCache
from libs.cache_decorator import Cached
from waterfall.build_info import BuildInfo

# TODO(crbug.com/787676): Use an api rather than parse urls to get the relevant
# data out of a build/tryjob url.
_HOST_NAME_PATTERN = (
    r'https?://(?:(?:build|ci)\.chromium\.org/p|\w+\.\w+\.google\.com/i)')

_MASTER_URL_PATTERN = re.compile(r'^%s/([^/]+)(?:/.*)?$' % _HOST_NAME_PATTERN)

_MILO_MASTER_URL_PATTERN = re.compile(
    r'^https?://luci-milo\.appspot\.com/buildbot/([^/]+)(?:/.*)?$')

_CI_MASTER_URL_PATTERN = re.compile(
    r'^https?://ci\.chromium\.org/buildbot/([^/]+)(?:/.*)?$')

_CI_LONG_MASTER_URL_PATTERN = re.compile(
    r'^https?://ci\.chromium\.org/p/chromium/([^/]+)(?:/.*)?$')

_MASTER_URL_PATTERNS = [  # yapf: disable
    _MASTER_URL_PATTERN,
    _MILO_MASTER_URL_PATTERN,
    _CI_MASTER_URL_PATTERN,
    _CI_LONG_MASTER_URL_PATTERN,
]

_MILO_SWARMING_TASK_URL_PATTERN = re.compile(
    r'^https?://luci-milo\.appspot\.com/swarming/task/([^/]+)(?:/.*)?$')

_CI_SWARMING_TASK_URL_PATTERN = re.compile(
    r'^https?://ci\.chromium\.org/swarming/task/([^/]+)(?:/.*)?$')

_SWARMING_TASK_URL_PATTERNS = [  # yapf: disable
    _MILO_SWARMING_TASK_URL_PATTERN,
    _CI_SWARMING_TASK_URL_PATTERN,
]

_BUILD_URL_PATTERN = re.compile(
    r'^%s/([^/]+)/builders/([^/]+)/builds/(\d+)(?:/.*)?$' % _HOST_NAME_PATTERN)

_MILO_BUILD_URL_PATTERN = re.compile(
    r'^https?://luci-milo\.appspot\.com/buildbot/([^/]+)/([^/]+)/(\d+)'
    '(?:/.*)?$')

_CI_BUILD_URL_PATTERN = re.compile(
    r'^https?://ci\.chromium\.org/buildbot/([^/]+)/([^/]+)/(\d+)'
    '(?:/.*)?$')

_CI_BUILD_LONG_URL_PATTERN = re.compile(
    r'^https?://ci\.chromium\.org/p/chromium/builders/([^/]+)/([^/]+)/(\d+)')

_BUILD_URL_PATTERNS = [  # yapf: disable
    _BUILD_URL_PATTERN,
    _MILO_BUILD_URL_PATTERN,
    _CI_BUILD_URL_PATTERN,
]

_MILO_BUILDINFO_ENDPOINT = ('https://luci-milo.appspot.com/'
                            'prpc/milo.BuildInfo/Get')
_MILO_ENDPOINT = 'https://luci-milo.appspot.com/prpc/milo.Buildbot'
_MILO_ENDPOINT_BUILD = '%s/GetBuildbotBuildJSON' % _MILO_ENDPOINT
_MILO_ENDPOINT_MASTER = '%s/GetCompressedMasterJSON' % _MILO_ENDPOINT

_STEP_URL_PATTERN = re.compile(
    r'^%s/([^/]+)/builders/([^/]+)/builds/(\d+)/steps/([^/]+)(/.*)?$' %
    _HOST_NAME_PATTERN)

_COMMIT_POSITION_PATTERN = re.compile(r'refs/heads/master@{#(\d+)}$',
                                      re.IGNORECASE)

# These values are buildbot constants used for Build and BuildStep.
# This line was copied from buildbot/master/buildbot/status/results.py.
SUCCESS, WARNINGS, FAILURE, SKIPPED, EXCEPTION, RETRY, CANCELLED = range(7)


def _ProcessMiloData(response_json, master_name, builder_name, build_number=''):
  if not response_json:
    return None
  try:
    response_data = json.loads(response_json)
  except Exception:  # pragma: no cover
    logging.error('Failed to load json data for %s-%s-%s' %
                  (master_name, builder_name, build_number))
    return None
  try:
    decoded_data = base64.b64decode(response_data.get('data'))
  except Exception:  # pragma: no cover
    logging.error('Failed to b64decode data for %s-%s-%s' %
                  (master_name, builder_name, build_number))
    return None

  if build_number:
    # Build data is not compressed.
    return decoded_data

  try:
    with io.BytesIO(decoded_data) as compressed_file:
      with gzip.GzipFile(fileobj=compressed_file) as decompressed_file:
        data_json = decompressed_file.read()
  except Exception:  # pragma: no cover
    logging.error('Failed to decompress data for %s-%s-%s' %
                  (master_name, builder_name, build_number))
    return None

  return data_json


def _GetMasterJsonData(http_client,
                       master_name,
                       builder_name='',
                       build_number=''):
  req = {
      'name': master_name,
      'exclude_deprecated': True,
  }
  _, response_json = rpc_util.DownloadJsonData(_MILO_ENDPOINT_MASTER, req,
                                               http_client)

  return _ProcessMiloData(response_json, master_name, builder_name,
                          build_number)


def ListBuildersOnMaster(master_name, http_client):
  master_data_json = _GetMasterJsonData(http_client, master_name)
  if not master_data_json:
    return []
  data = json.loads(master_data_json)
  return [bot for bot in data.get('builders', {}).keys()]


def GetRecentCompletedBuilds(master_name, builder_name, http_client):
  """Returns a sorted list of recent completed builds for the given builder.

  Sorted by completed time, newer builds at beginning of the returned list.
  """
  master_data_json = _GetMasterJsonData(http_client, master_name, builder_name)
  if not master_data_json:
    return []

  master_data = json.loads(master_data_json)
  meta_data = master_data.get('builders', {}).get(builder_name, {})
  cached_builds = meta_data.get('cachedBuilds', [])
  current_builds = meta_data.get('currentBuilds', [])
  return sorted(set(cached_builds) - set(current_builds), reverse=True)


def GetMasterNameFromUrl(url):
  """Parses the given url and returns the master name."""
  if not url:
    return None

  match = None
  for pattern in _MASTER_URL_PATTERNS:
    match = pattern.match(url)
    if match:
      break
  if not match:
    return None
  return match.group(1)


def _ComputeCacheKeyForLuciBuilder(func, args, kwargs, namespace):
  """Returns a key for the Luci builder passed over to _GetBuildbotMasterName"""
  params = inspect.getcallargs(func, *args, **kwargs)
  return '%s-%s::%s' % (namespace, params['bucket_name'],
                        params['builder_name'])


# TODO(crbug/802940): Remove this when the API of getting LUCI build is ready.
@Cached(
    PickledMemCache(),
    namespace='luci-builder-to-master',
    key_generator=_ComputeCacheKeyForLuciBuilder)
def _GetBuildbotMasterName(bucket_name, builder_name, build_number):
  """Gets buildbot master name based on build_address."""
  build_address = '%s/%s/%d' % (bucket_name, builder_name, build_number)
  res = buildbucket_client.SearchBuilds(tags=[(
      'tag', 'build_address:%s' % build_address)])
  if not res or len(res.get('builds', [])) < 1:
    return None

  parameters_json = res['builds'][0].get('result_details_json')
  try:
    properties = json.loads(parameters_json).get('properties', {})
    return properties.get('mastername') or properties.get('parent_mastername')
  except (ValueError, TypeError):
    logging.exception('Failed to get buildbot master name for luci build %s.',
                      build_address)
    return None


# TODO(crbug/802940): Remove this when the API of getting LUCI build is ready.
def _ParseCIBuildLongUrl(url):
  """Parses urls in _CI_BUILD_LONG_URL_PATTERN pattern."""
  match = _CI_BUILD_LONG_URL_PATTERN.match(url)
  if not match:
    return None

  bucket_name, builder_name, build_number = match.groups()
  builder_name = urllib.unquote(builder_name)
  master_name = _GetBuildbotMasterName(bucket_name, builder_name,
                                       int(build_number))
  if not master_name:
    return None

  return master_name, builder_name, int(build_number)


def ParseBuildUrl(url):
  """Parses the given build url.

  Return:
    (master_name, builder_name, build_number)
  """
  if not url:
    return None

  match = None
  for pattern in _BUILD_URL_PATTERNS:
    match = pattern.match(url)
    if match:
      break
  if not match:
    return _ParseCIBuildLongUrl(url)

  master_name, builder_name, build_number = match.groups()
  builder_name = urllib.unquote(builder_name)
  return master_name, builder_name, int(build_number)


def ParseStepUrl(url):
  """Parses the given step url.

  Return:
    (master_name, builder_name, build_number, step_name)
  """
  if not url:
    return None

  match = _STEP_URL_PATTERN.match(url)
  if not match:
    return None

  master_name, builder_name, build_number, step_name, _ = match.groups()
  builder_name = urllib.unquote(builder_name)
  return master_name, builder_name, int(build_number), step_name


def CreateBuildUrl(master_name, builder_name, build_number):
  """Creates the url for the given build."""
  builder_name = urllib.quote(builder_name)
  return 'https://ci.chromium.org/buildbot/%s/%s/%s' % (master_name,
                                                        builder_name,
                                                        build_number)


def GetBuildDataFromMilo(master_name, builder_name, build_number, http_client):
  """Returns the json-format data of the build."""
  data = {
      'master': master_name,
      'builder': builder_name,
      'buildNum': build_number,
      'exclude_deprecated': True,
  }
  status_code, response_json = rpc_util.DownloadJsonData(
      _MILO_ENDPOINT_BUILD, data, http_client)
  return status_code, _ProcessMiloData(response_json, master_name, builder_name,
                                       str(build_number))


def GetStepResult(step_data_json):
  """Returns the result of a step."""
  result = step_data_json.get('results')
  if result is None and step_data_json.get('isFinished'):
    # Without parameter filter=0 in the http request to the buildbot json api,
    # the value of the result of a passed step won't be present.
    return SUCCESS

  while isinstance(result, list):
    result = result[0]
  return result


def GetBuildProperty(properties, property_name):
  """Returns the property value from the given build properties."""
  for item in properties:
    if item[0] == property_name:
      return item[1]
  return None


def GetBuildStartTime(build_data_json):
  times = build_data_json.get('times')
  if not times or not times[0]:
    # For a build with infra failure, the start time might be set to None.
    return None
  return datetime.utcfromtimestamp(times[0])


def GetBuildEndTime(build_data_json):
  times = build_data_json.get('times')
  if not times or len(times) < 2 or not times[1]:
    return None
  return datetime.utcfromtimestamp(times[1])


def GetBuildResult(build_data_json):
  return build_data_json.get('results')


def _GetCommitPosition(commit_position_line):
  if commit_position_line:
    match = _COMMIT_POSITION_PATTERN.match(commit_position_line)
    if match:
      return int(match.group(1))
  return None


def ExtractBuildInfo(master_name, builder_name, build_number, build_data):
  """Extracts and returns build information as an instance of BuildInfo."""
  build_info = BuildInfo(master_name, builder_name, build_number)
  data_json = json.loads(build_data)

  properties = data_json.get('properties') or []
  chromium_revision = GetBuildProperty(properties, 'got_revision')
  commit_position_line = GetBuildProperty(properties, 'got_revision_cp')
  parent_buildername = GetBuildProperty(properties, 'parent_buildername')
  parent_mastername = GetBuildProperty(properties, 'parent_mastername')

  build_info.build_start_time = GetBuildStartTime(data_json)
  build_info.build_end_time = GetBuildEndTime(data_json)
  build_info.chromium_revision = chromium_revision
  build_info.commit_position = _GetCommitPosition(commit_position_line)
  build_info.completed = data_json.get('currentStep') is None
  build_info.result = GetBuildResult(data_json)
  build_info.parent_buildername = parent_buildername
  build_info.parent_mastername = parent_mastername

  changes = (data_json.get('sourceStamp') or {}).get('changes') or []
  for change in changes:
    if (change.get('revision') and
        change['revision'] not in build_info.blame_list):
      build_info.blame_list.append(change['revision'])

  # Step categories:
  # 1. A step is passed if it is in SUCCESS or WARNINGS status.
  # 2. A step is failed if it is in FAILED status.
  # 3. A step is not passed if it is not in SUCCESS or WARNINGS status. This
  #    category includes steps in statuses: FAILED, SKIPPED, EXCEPTION, RETRY,
  #    CANCELLED, etc.
  steps = data_json.get('steps') or []
  for step_data in steps:
    step_name = step_data.get('name')

    if not step_name:
      continue

    if not step_data.get('isFinished'):
      # Skip steps that haven't started yet or are still running.
      continue

    step_result = GetStepResult(step_data)
    if step_result not in (SUCCESS, WARNINGS):
      build_info.not_passed_steps.append(step_name)

    if step_name == 'Failure reason':
      # 'Failure reason' is always red when the build breaks or has exception,
      # but it is not a failed step.
      continue

    step_logs = step_data.get('logs')
    if step_logs and 'preamble' == step_logs[0][0]:
      # Skip a annotating step like "steps" or "slave_steps", which wraps other
      # steps. A failed annotated step like "content_browsertests" will make
      # the annotating step like "steps" fail too. Such annotating steps have a
      # log with name "preamble".
      continue

    if step_result in (SUCCESS, WARNINGS):
      build_info.passed_steps.append(step_name)
    elif step_result == FAILURE:
      build_info.failed_steps.append(step_name)

  return build_info


def ValidateBuildUrl(url):
  return bool(
      _MILO_MASTER_URL_PATTERN.match(url) or
      _MILO_SWARMING_TASK_URL_PATTERN.match(url) or
      _BUILD_URL_PATTERN.match(url))


def GetBuildInfo(build, http_client):
  master, builder, build_number = ParseBuildUrl(build)
  request = {
      'buildbot': {
          'masterName': master,
          'builderName': builder,
          'buildNumber': build_number
      }
  }
  _, build_info = rpc_util.DownloadJsonData(_MILO_BUILDINFO_ENDPOINT, request,
                                            http_client)
  return build_info
