# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Pulls historical try job metadata from Findit and prints a report."""

import argparse
from collections import defaultdict
import datetime
import json
import numpy
import os
import sys

try:
  from matplotlib import pyplot
except ImportError:
  pyplot = None

_FINDIT_DIR = os.path.join(
    os.path.dirname(__file__), os.path.pardir, os.path.pardir)
sys.path.insert(1, _FINDIT_DIR)
from local_libs import remote_api

from model.wf_try_job_data import WfTryJobData

NOT_AVAILABLE = 'N/A'


def GetBuildInfo(master_name, builder_name, build_number):
  if build_number > 5:
    return 200, str(build_number)

  return 404, None


def GetEarliestBuild(master_name, builder_name, max_retries=10):
  """Gets the earliest build available on a master/builder.

    Assumes builds are contiguous.

  Args:
    master_name (str): The name of the master to check.
    builder_name (str): The name of the builder to check.
    max_retries (int): The maximum number of builds to try before giving up.

  Returns:
    (int, BuildInfo): The status code and earliest BuildInfo available.
  """
  assert max_retries > 0

  lower_bound_build_number = 0
  upper_bound_build_number = 2**max_retries

  status_code, lower_bound_build = GetBuildInfo(master_name, builder_name,
                                                lower_bound_build_number)

  if lower_bound_build:
    # The specified lower bound is the earliest build available within range.
    return status_code, lower_bound_build

  status_code, upper_bound_build = GetBuildInfo(master_name, builder_name,
                                                upper_bound_build_number)

  if not upper_bound_build:
    # Earliest build number cannot be determined within the specified range.
    return status_code, None

  status_code = None
  lower_bound_build = None
  upper_bound_build = None
  status_code_found = None
  tries = 0

  # Binary search for the earliest available build.
  while upper_bound_build_number > lower_bound_build_number + 1:
    candidate_build_number = (
        lower_bound_build_number +
        ((upper_bound_build_number - lower_bound_build_number) / 2))
    status_code, candidate_build = GetBuildInfo(master_name, builder_name,
                                                candidate_build_number)

    if candidate_build:
      upper_bound_build_number = candidate_build_number
      upper_bound_build = candidate_build
      status_code_found = status_code
    else:
      lower_bound_build_number = candidate_build_number
      lower_bound_build = candidate_build
    tries += 1

  found_build = lower_bound_build or upper_bound_build
  return (status_cound_found, found_build
          if found_build else status_code, found_build)


#from infra_api_clients.swarming import swarming_util
#from gae_libs.http.http_client_appengine import HttpClientAppengine
#from common.findit_http_client import FinditHttpClient
#from services import step_util
#from waterfall import build_util
#from dto.swarming_task_request import SwarmingTaskRequest
from model.flake.master_flake_analysis import MasterFlakeAnalysis
#from services import swarmed_test_util

from google.appengine.ext import ndb

if __name__ == '__main__':
  # Set up the Remote API to use services on the live App Engine.
  remote_api.EnableRemoteApi(app_id='findit-for-me')
  """START_DATE = datetime.datetime(2018, 1, 1)
  END_DATE = datetime.datetime(2018, 3, 31)

  flake_analysis_query = MasterFlakeAnalysis.query(
      MasterFlakeAnalysis.request_time >= START_DATE,
      MasterFlakeAnalysis.request_time < END_DATE)

  all_analyses = []
  cursor = None
  more = True

  while more:
    analyses, cursor, more = flake_analysis_query.fetch_page(
        page_size=1000, start_cursor=cursor)
    all_analyses.extend(analyses)

  count = 0
  error = 0
  suspected_builds = 0
  for analysis in all_analyses:
    if analysis.error:
      continue

    if not analysis.suspected_flake_build_number:
      continue

    suspected_builds += 1
    if analysis.culprit_urlsafe_key:
      count += 1
    if analysis.error:
      error += 1

  print 'total analyses %d' % len(all_analyses)
  print 'analyses with regression ranges %d' % suspected_builds
  print 'errors: %d' % error
  print 'analyses with culprits %d' % count"""

  #ref_request = swarming_util.GetSwarmingTaskRequestJson('chromium-swarm.appspot.com', '3c3f4fc761712010', HttpClientAppengine())
  #ref_request = swarming_util.GetSwarmingTaskRequest('chromium-swarm.appspot.com', '3cb03a1e4bb28810', FinditHttpClient())

  #print SwarmingTaskRequest.FromSerializable(ref_request)

  #print GetEarliestBuild('m', 'b')
  #build_util.GetEarliestBuild('chromium.win', 'Win7 Tests (dbg)(1)', max_retries=20)
  """task_id = '3ca50cf70c45b910'
  #data, output_json, error = swarmed_test_util.GetSwarmingTaskDataAndResult(task_id, FinditHttpClient())
  from services import swarming
  master_name = 'chromium.chromiumos'
  builder_name = 'linux-chromeos-rel'
  build_number = 6549
  step_name = 'SettingsAccessibilityTest.MANAGE_ACCESSIBILITY_aria_valid_attr_value'
 
  task_id, ref_request = swarming.GetReferredSwarmingTaskRequestInfo(
      master_name, builder_name, build_number, step_name, HttpClientAppengine())


  analysis_urlsafe_key = "ag9zfmZpbmRpdC1mb3ItbWVyvAELEhdNYXN0ZXJGbGFrZUFuYWx5c2lzUm9vdCKFAWNocm9taXVtLmNocm9taXVtb3MvbGludXgtY2hyb21lb3MtcmVsLzYzOTEvaW50ZXJhY3RpdmVfdWlfdGVzdHMvVTJWc1pXTjBWRzlUY0dWaGExUmxjM1F1VTIxdmIzUm9iSGxTWldGa2MwRmpjbTl6YzBadmNtMWhkSFJsWkZSbGVIUT0MCxITTWFzdGVyRmxha2VBbmFseXNpcxgCDA"

  commit_position = 546790
  isolate_sha = "8e854b221a22777a58cb5f2cd6b6ccd91a200e9d"
  iterations = 30
  timeout_seconds = 3600

  flake_swarming.TriggerSwarmingTask(analysis_urlsafe_key, isolate_sha, iterations, timeout_seconds, 'lijeffrey-test')

  import pdb
  pdb.set_trace()"""

  analysis = MasterFlakeAnalysis.GetVersion(
      'chromium.memory', 'Linux MSan Tests', 9392, 'browser_tests',
      'SessionRestoreTest.RestorePinnedSelectedTab')
  print analysis
