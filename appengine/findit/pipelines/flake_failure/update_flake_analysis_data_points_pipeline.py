# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Updates a flake analysis' data points with incoming pass rate information."""

from dto.flake_swarming_task_output import FlakeSwarmingTaskOutput
from gae_libs.pipelines import GeneratorPipeline
from libs.structured_object import StructuredObject
from services.flake_failure import data_point_util


class UpdateFlakeAnalysisDataPointsInput(StructuredObject):
  # The urlsafe-key to the analysis to update.
  analysis_urlsafe_key = basestring

  # The url to the build whose artifacts were used to create the data point.
  # Can be None if existing build artifacts were not used (compile was needed).
  build_url = basestring

  # The data point with matching commit position to update.
  commit_position = int

  # The revision corresponding to the data point.
  revision = basestring

  # The url to the try job that generated the build artifacts to generate the
  # data point. Can be None if existing build artifacts were used (commit
  # position mapped to a nearby valid build).
  try_job_url = basestring

  # The results of the flake swarming task to update data points with.
  swarming_task_output = FlakeSwarmingTaskOutput


class UpdateFlakeAnalysisDataPointsPipeline(GeneratorPipeline):
  """Updates a MasterFlakeAnalysis' data points with swarming task results."""

  input_type = UpdateFlakeAnalysisDataPointsInput

  def RunImpl(self, parameters):
    """Creates or updates existing data points with swarming task results."""
    data_point_util.UpdateAnalysisDataPoints(
        parameters.analysis_urlsafe_key, parameters.commit_position,
        parameters.revision, parameters.build_url, parameters.try_job_url,
        parameters.swarming_task_output)
