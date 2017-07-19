# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from collections import defaultdict
import logging

from common.waterfall import failure_type
from libs import time_util
from model.wf_try_job import WfTryJob
from model.wf_try_job_data import WfTryJobData
from waterfall.schedule_try_job_pipeline import ScheduleTryJobPipeline


class ScheduleTestTryJobPipeline(ScheduleTryJobPipeline):
  """A pipeline for scheduling a new try job for failed test build."""

  def _GetBuildProperties(self, master_name, builder_name, build_number,
                          good_revision, bad_revision, try_job_type,
                          suspected_revisions):
    properties = super(ScheduleTestTryJobPipeline, self)._GetBuildProperties(
        master_name, builder_name, build_number, good_revision, bad_revision,
        try_job_type, suspected_revisions)
    properties['target_testername'] = builder_name

    return properties

  def _CreateTryJobData(self, build_id, try_job_key, has_heuristic_results):
    try_job_data = WfTryJobData.Create(build_id)
    try_job_data.created_time = time_util.GetUTCNow()
    try_job_data.has_compile_targets = False
    try_job_data.has_heuristic_results = has_heuristic_results
    try_job_data.try_job_key = try_job_key
    try_job_data.try_job_type = failure_type.GetDescriptionForFailureType(
        failure_type.TEST)
    try_job_data.put()

  # Arguments number differs from overridden method - pylint: disable=W0221
  def run(self,
          master_name,
          builder_name,
          build_number,
          good_revision,
          bad_revision,
          try_job_type,
          suspected_revisions,
          cache_name,
          dimensions,
          targeted_tests,
          force_buildbot=False):
    """
    Args:
      master_name (str): the master name of a build.
      builder_name (str): the builder name of a build.
      build_number (int): the build number of a build.
      good_revision (str): the revision of the last passed build.
      bad__revision (str): the revision of the first failed build.
      try_job_type (int): type of the try job: TEST in this case.
      suspected_revisions (list): a list of suspected revisions from heuristic.
      cache_name (str): A string to identify separate directories for different
          waterfall bots on the trybots.
      dimensions (list): A list of strings in the format
          ["key1:value1", "key2:value2"].
      targeted_tests (dict): a dict of reliable failed tests.
      force_buildbot (bool): Whether to force a run on buildbot slaves, ignoring
          swarmbucket configuration.

    Returns:
      build_id (str): id of the triggered try job.
    """

    properties = self._GetBuildProperties(
        master_name, builder_name, build_number, good_revision, bad_revision,
        try_job_type, suspected_revisions)

    if not targeted_tests:  # pragma: no cover
      logging.info('All tests are flaky, no try job will be triggered.')
      return

    additional_parameters = {'tests': targeted_tests}

    build_id = self._TriggerTryJob(
        master_name, builder_name, properties, additional_parameters,
        failure_type.GetDescriptionForFailureType(failure_type.TEST),
        cache_name, dimensions, force_buildbot)

    try_job = WfTryJob.Get(master_name, builder_name, build_number)
    try_job.test_results.append({'try_job_id': build_id})
    try_job.try_job_ids.append(build_id)
    try_job.put()

    # Create a corresponding WfTryJobData entity to capture as much metadata as
    # early as possible.
    self._CreateTryJobData(build_id, try_job.key, bool(suspected_revisions))

    return build_id
