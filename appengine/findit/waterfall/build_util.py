# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import logging

from common import constants
from common.findit_http_client import FinditHttpClient
from common.waterfall import buildbucket_client
from common.waterfall import failure_type
from infra_api_clients import logdog_util
from libs import time_util
from model.wf_build import WfBuild
from services import swarming
from waterfall import buildbot

HTTP_CLIENT_NO_404_ERROR = FinditHttpClient(no_error_logging_statuses=[404])


def _BuildDataNeedUpdating(build):
  return (not build.data or (
      not build.completed and
      (time_util.GetUTCNow() - build.last_crawled_time).total_seconds() >= 300))


def DownloadBuildData(master_name, builder_name, build_number):
  """Downloads build data and returns a WfBuild instance."""
  build = WfBuild.Get(master_name, builder_name, build_number)
  if not build:
    build = WfBuild.Create(master_name, builder_name, build_number)

  status_code = None
  # Cache the data to avoid pulling from master again.
  if _BuildDataNeedUpdating(build):
    # Retrieve build data from milo.
    status_code, build.data = buildbot.GetBuildDataFromMilo(
        master_name, builder_name, build_number, HTTP_CLIENT_NO_404_ERROR)
    build.last_crawled_time = time_util.GetUTCNow()
    build.put()

  return status_code, build


def GetBuildInfo(master_name, builder_name, build_number):
  """Gets build info given a master, builder, and build number.

  Args:
    master_name (str): The name of the master.
    builder_name (str): The name of the builder.
    build_number (int): The build number.

  Returns:
    Build information as an instance of BuildInfo.
  """
  status_code, build = DownloadBuildData(master_name, builder_name,
                                         build_number)
  if not build.data:
    return status_code, None
  build_info = buildbot.ExtractBuildInfo(master_name, builder_name,
                                         build_number, build.data)

  if not build.completed:
    build.start_time = build_info.build_start_time
    build.completed = build_info.completed
    build.result = build_info.result
    build.put()

  return status_code, build_info


def GetBuildEndTime(master_name, builder_name, build_number):
  _, build = DownloadBuildData(master_name, builder_name, build_number)
  build_info = buildbot.ExtractBuildInfo(master_name, builder_name,
                                         build_number, build.data)
  return build_info.build_end_time


def CreateBuildId(master_name, builder_name, build_number):
  return '%s/%s/%s' % (master_name, builder_name, build_number)


def GetBuildInfoFromId(build_id):
  return build_id.split('/')


def GetFailureType(build_info):
  if not build_info.failed_steps:
    return failure_type.UNKNOWN
  # TODO(robertocn): Consider also bailing out of tests with infra failures.
  if constants.COMPILE_STEP_NAME in build_info.failed_steps:
    if build_info.result == buildbot.EXCEPTION:
      return failure_type.INFRA
    return failure_type.COMPILE
  # TODO(http://crbug.com/602733): differentiate test steps from infra ones.
  return failure_type.TEST


def GetLatestBuildNumber(master_name, builder_name):
  """Attempts to get the latest build number on master_name/builder_name."""
  recent_builds = buildbot.GetRecentCompletedBuilds(master_name, builder_name,
                                                    FinditHttpClient())

  if recent_builds is None:
    # Likely a network error.
    logging.error('Failed to detect latest build number on %s, %s', master_name,
                  builder_name)
    return None

  if not recent_builds:
    # In case the builder is new or was recently reset.
    logging.warning('No recent builds found on %s %s', master_name,
                    builder_name)
    return None

  return recent_builds[0]


def GetBoundingBuilds(master_name, builder_name, lower_bound_build_number,
                      upper_bound_build_number, requested_commit_position):
  """Finds the two builds immediately before and after a commit position.

  Args:
    master_name (str): The name of the master.
    builder_name (str): The name of the builder.
    lower_bound_build_number (int): The earliest build number to search.
    upper_bound_build_number (int): The latest build number to search.
    requested_commit_position (int): The specified commit_position to find the
        bounding build numbers.

  Returns:
    (BuildInfo, Buildinfo): The two nearest builds that bound the requested
        commit position, with the first being earlier of the two. For example,
        if build_1 has commit position 100, build_2 has commit position 110,
        and 105 is requested, returns (build_1, build_2). Returns None for
        either or both of the builds if they cannot be determined. If the
        requested commit is before the lower bound, returns (None, BuildInfo).
        If the requested commit is after the upper bound, returns
        (BuildInfo, None). The calling code should check for the returned builds
        and decide what to do accordingly.
  """
  lower_bound_build_number = lower_bound_build_number or 0
  _, earliest_build_info = GetBuildInfo(master_name, builder_name,
                                        lower_bound_build_number)
  assert earliest_build_info
  assert earliest_build_info.commit_position is not None

  if requested_commit_position <= earliest_build_info.commit_position:
    return None, earliest_build_info

  if upper_bound_build_number is None:
    upper_bound_build_number = GetLatestBuildNumber(master_name, builder_name)

  if upper_bound_build_number is None:
    logging.error('Failed to detect latest build number')
    return None, None

  _, latest_build_info = GetBuildInfo(master_name, builder_name,
                                      upper_bound_build_number)
  assert latest_build_info
  assert latest_build_info.commit_position is not None

  if requested_commit_position >= latest_build_info.commit_position:
    return latest_build_info, None

  # Bisect the build number range and search for the earliest build whose
  # commit position >= requested_commit_position.
  upper_bound = upper_bound_build_number
  lower_bound = lower_bound_build_number

  while upper_bound - lower_bound > 1:
    candidate_build_number = (upper_bound - lower_bound) / 2 + lower_bound
    _, candidate_build = GetBuildInfo(master_name, builder_name,
                                      candidate_build_number)
    assert candidate_build

    if candidate_build.commit_position == requested_commit_position:
      # Exact match.
      _, lower_bound_build = GetBuildInfo(master_name, builder_name,
                                          candidate_build_number - 1)
      assert lower_bound_build
      return lower_bound_build, candidate_build

    if candidate_build.commit_position > requested_commit_position:
      # Go left.
      upper_bound = candidate_build_number
    else:
      # Go right.
      lower_bound = candidate_build_number

  _, lower_bound_build = GetBuildInfo(master_name, builder_name, lower_bound)
  _, upper_bound_build = GetBuildInfo(master_name, builder_name, upper_bound)
  assert lower_bound_build
  assert upper_bound_build

  return lower_bound_build, upper_bound_build


def FindValidBuildNumberForStepNearby(master_name,
                                      builder_name,
                                      step_name,
                                      build_number,
                                      exclude_list=None,
                                      search_distance=3):
  """Finds a valid nearby build number for a step.

  Looks around the given build number for builds that have a reference task
  on swarming. We use this reference swarming task to create a task request,
  and it's required to run the test. If no reference swarming task can be
  found, it's likely that the build failed and the artifact doesn't exist.

  Args:
    master_name (str): Name of the master for this test.
    builder_name (str): Name of the builder for this test.
    step_name (str): Name of the builder for this test.
    build_number (int): Build number to look around.
    exclude_list (lst): Build numbers to exclude from the search.
    search_distance (int): Distance to search on either side of the build.

  Returns:
    (int) Valid nearby build if any, else None.
  """
  builds_to_look_at = [build_number]
  for x in range(1, search_distance + 1):
    builds_to_look_at.append(build_number + x)
    builds_to_look_at.append(build_number - x)

  logging.info('Examining build numbers %r for a valid build',
               builds_to_look_at)

  http_client = FinditHttpClient()
  for build in builds_to_look_at:
    if exclude_list and build in exclude_list:
      continue
    swarming_task_items = swarming.ListSwarmingTasksDataByTags(
        http_client, master_name, builder_name, build, step_name)
    if swarming_task_items:
      return build

  return None


def _ReturnStepLog(data, log_type):
  if not data:
    return None

  if log_type.lower() == 'json.output[ninja_info]':
    # Check if data is malformatted.
    try:
      json.loads(data)
    except ValueError:
      logging.error('json.output[ninja_info] is malformatted')
      return None

  if log_type.lower() not in ['stdout', 'json.output[ninja_info]']:
    try:
      return json.loads(data) if data else None
    except ValueError:
      logging.error('Failed to json load data for %s. Data is: %s.' % (log_type,
                                                                       data))

  return data


def GetTryJobStepLog(try_job_id, full_step_name, http_client,
                     log_type='stdout'):
  """Returns specific log of the specified step."""

  error, build = buildbucket_client.GetTryJobs([try_job_id])[0]
  if error:
    logging.exception('Error retrieving buildbucket build id: %s' % try_job_id)
    return None

  # 1. Get log.
  data = logdog_util.GetStepLogForBuild(build.response, full_step_name,
                                        log_type, http_client)

  return _ReturnStepLog(data, log_type)


def GetWaterfallBuildStepLog(master_name,
                             builder_name,
                             build_number,
                             full_step_name,
                             http_client,
                             log_type='stdout'):
  """Returns sepcific log of the specified step."""

  data = logdog_util.GetStepLogLegacy(master_name, builder_name, build_number,
                                      full_step_name, log_type, http_client)

  return _ReturnStepLog(data, log_type)


# TODO(crbug/804617): Modify this function to use new LUCI API when it's ready.
def IteratePreviousBuildsFrom(master_name, builder_name, build_number,
                              entry_limit):

  n = build_number - 1
  entry_number = 0
  while n >= 0 and entry_number <= entry_limit:  # pragma: no branch.
    status_code, build_info = GetBuildInfo(master_name, builder_name, n)
    n -= 1
    if build_info:
      entry_number += 1
      yield build_info
    elif status_code == 404:
      continue
    else:
      # 404 means we hit a gap. Otherwise there is something wrong.
      raise Exception('Failed to download build data for build %s/%s/%d' %
                      (master_name, builder_name, n))
