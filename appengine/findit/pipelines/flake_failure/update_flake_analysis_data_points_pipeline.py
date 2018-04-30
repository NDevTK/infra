# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Updates a flake analysis' data points with incoming pass rate information."""

from google.appengine.ext import ndb

from dto.flake_swarming_task_output import FlakeSwarmingTaskOutput
from gae_libs.pipelines import GeneratorPipeline
from libs.structured_object import StructuredObject
from services import step_util
from services.flake_failure import data_point_util


class UpdateFlakeAnalysisDataPointsInput(StructuredObject):
  # The urlsafe-key to the analysis to update.
  analysis_urlsafe_key = basestring

  # The data point with matching commit position to update.
  commit_position = int

  # The revision corresponding to the data point.
  revision = basestring

  # The results of the flake swarming task to update data points with.
  swarming_task_output = FlakeSwarmingTaskOutput


class UpdateFlakeAnalysisDataPointsPipeline(GeneratorPipeline):
  """Updates a MasterFlakeAnalysis' data points with swarming task results."""

  input_type = UpdateFlakeAnalysisDataPointsInput

  def RunImpl(self, parameters):
    """Creates or updates existing data points with swarming task results."""
    analysis = ndb.Key(urlsafe=parameters.analysis_urlsafe_key).get()
    assert analysis

    # Get the bounding builds' info to help craft data points with.
    (lower_bound_build,
     upper_bound_build) = step_util.GetValidBoundingBuildsForStep(
         analysis.master_name, analysis.builder_name, analysis.step_name, None,
         analysis.build_number, parameters.commit_position)

    data_point_util.UpdateAnalysisDataPoints(
        parameters.analysis_urlsafe_key, parameters.commit_position,
        parameters.revision, parameters.swarming_task_output, lower_bound_build,
        upper_bound_build)
