# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.ext import ndb

from common.waterfall import buildbucket_client
from common.waterfall import failure_type
from libs import time_util
from model.flake.flake_try_job import FlakeTryJob
from model.flake.flake_try_job_data import FlakeTryJobData
from waterfall import waterfall_config
from waterfall.schedule_try_job_pipeline import ScheduleTryJobPipeline

_DEFAULT_ITERATIONS_TO_RERUN = 100
_MAX_DELALY = 2 * 60 * 60  # Two hours in seconds.


class ScheduleFlakeTryJobPipeline(ScheduleTryJobPipeline):
  """A pipeline for scheduling a new flake try job for a flaky test."""
  async = True

  # Arguments number differs from overridden method - pylint: disable=W0221
  def _GetBuildProperties(self, master_name, builder_name, canonical_step_name,
                          test_name, git_hash, iterations_to_rerun):
    iterations = iterations_to_rerun or _DEFAULT_ITERATIONS_TO_RERUN

    return {
        'recipe': 'findit/chromium/flake',
        'target_mastername': master_name,
        'target_testername': builder_name,
        'test_revision': git_hash,
        'test_repeat_count': iterations,
        'tests': {
            canonical_step_name: [test_name]
        }
    }

  def _GetTrybot(self, master_name, builder_name, force_buildbot=False):
    """Overrides the base method to get a dedicated flake trybot instead."""
    try_master, try_builder = waterfall_config.GetFlakeTrybot(
        master_name, builder_name, force_buildbot=force_buildbot)
    return try_master, try_builder

  @ndb.transactional
  def _CreateTryJobData(self, build_id, try_job_key, urlsafe_analysis_key):
    try_job_data = FlakeTryJobData.Create(build_id)
    try_job_data.created_time = time_util.GetUTCNow()
    try_job_data.try_job_key = try_job_key
    try_job_data.analysis_key = ndb.Key(urlsafe=urlsafe_analysis_key)
    try_job_data.put()

  def _ScheduleTryJobNow(self,
                         master_name,
                         builder_name,
                         canonical_step_name,
                         test_name,
                         git_hash,
                         urlsafe_analysis_key,
                         cache_name,
                         dimensions,
                         iterations_to_rerun=None):
    properties = self._GetBuildProperties(master_name, builder_name,
                                          canonical_step_name, test_name,
                                          git_hash, iterations_to_rerun)
    build_id = self._TriggerTryJob(
        master_name, builder_name, properties, {},
        failure_type.GetDescriptionForFailureType(failure_type.FLAKY_TEST),
        cache_name, dimensions)

    try_job = FlakeTryJob.Get(master_name, builder_name, canonical_step_name,
                              test_name, git_hash)
    try_job.flake_results.append({'try_job_id': build_id})
    try_job.try_job_ids.append(build_id)
    try_job.put()

    # Create a corresponding Flake entity to capture as much metadata as early
    # as possible.
    self._CreateTryJobData(build_id, try_job.key, urlsafe_analysis_key)

    return build_id

  # Arguments number differs from overridden method - pylint: disable=W0221
  def run(self,
          master_name,
          builder_name,
          canonical_step_name,
          test_name,
          git_hash,
          urlsafe_analysis_key,
          cache_name,
          dimensions,
          iterations_to_rerun=None):
    """Triggers a flake try job.

    Args:
      master_name (str): The master name of a flaky test.
      builder_name (str): The builder name of a flaky test.
      canonical_step_name (str): The canonical name of the step the flaky test
          occurred on.
      test_name (str): The name of the flaky test.
      git_hash (str): The git hash of the revision to run the try job against.
      urlsafe_analysis_key (str): The urlsafe key of the original
          MasterFlakeAnalysis that triggered this try job.
      cache_name (str): A string to identify separate directories for different
          waterfall bots on the trybots.
      dimensions (list): A list of strings in the format
          ["key1:value1", "key2:value2"].
      iterations_to_rerun (int): The number of iterations to rerun.

    Returns:
      build_id (str): Id of the triggered try job.
    """
    is_swarmbucket = buildbucket_client.IsSwarmbucketMaster(
        self._GetTrybot(master_name, builder_name)[0])
    if not is_swarmbucket:
      self.complete(
          self._ScheduleTryJobNow(master_name, builder_name,
                                  canonical_step_name, test_name, git_hash,
                                  urlsafe_analysis_key, cache_name, dimensions,
                                  iterations_to_rerun))

    else:
      self.callback(master_name, builder_name, canonical_step_name, test_name,
                    git_hash, urlsafe_analysis_key, cache_name, dimensions,
                    iterations_to_rerun)

  def callback(self,
               master_name,
               builder_name,
               canonical_step_name,
               test_name,
               git_hash,
               urlsafe_analysis_key,
               cache_name,
               dimensions,
               iterations_to_rerun,
               delay=120):
    assert buildbucket_client.IsSwarmbucketMaster(
        _GetTrybot(mastername, builder_name)[0])

    next_off_peak_slot = swarming_util.GetETAToStartAnalysis(False)
    currently_off_peak = time_util.GettUTCNow() > next_off_peak_slot
    enough_available_bots = False
    if not currently_off_peak:
      dimensions = waterfall_config.GetTrybotDimensions(master_name,
                                                        builder_name)
      required_reserved_bots = waterfall_config.GetReservedBotCount(
          master_name, builder_name)
      available_bots = swarming_util.GetSwarmingBotCounts(dimensions)[
          'available']
      enough_available_bots = available_bots > required_reserved_bots
    if currently_off_peak or enough_available_bots:
      self.complete(
          self._ScheduleTryJobNow(master_name, builder_name,
                                  canonical_step_name, test_name, git_hash,
                                  urlsafe_analysis_key, cache_name, dimensions,
                                  iterations_to_rerun))
      return
    params = {
        'master_name': master_name,
        'builder_name': builder_name,
        'canonical_step_name': canonical_step_name,
        'test_name': test_name,
        'git_hash': git_hash,
        'urlsafe_analysis_key': urlsafe_analysis_key,
        'cache_name': cache_name,
        'dimensions': dimensions,
        'iterations_to_rerun': iterations_to_rerun,
        'delay': delay
    }
    # Exponential backoff, with a max delay of 2 hours.
    task = self.get_callback_task(
        countdown=min(2 * delay, _MAX_DELAY),
        params=params,
        target=constants.WATERFALL_BACKEND)
    task.add(queue_name=self.queue_name)
