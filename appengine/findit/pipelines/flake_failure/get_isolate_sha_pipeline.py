# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.ext import ndb

from common.findit_http_client import FinditHttpClient
from dto.list_of_basestring import ListOfBasestring
from gae_libs.pipelines import GeneratorPipeline
from gae_libs.pipelines import SynchronousPipeline
from libs.structured_object import StructuredObject
from pipelines.flake_failure.run_flake_try_job_pipeline import (
    RunFlakeTryJobParameters)
from pipelines.flake_failure.run_flake_try_job_pipeline import (
    RunFlakeTryJobPipeline)
from services import swarming
from services import swarmbot_util
from services.flake_failure import flake_try_job
from waterfall import build_util
from waterfall import waterfall_config
from waterfall.flake import flake_constants


class GetIsolateShaForCommitPositionParameters(StructuredObject):
  # The urlsafe key to the MasterFlakeAnalysis this pipeline is assisting in
  # analyyzing.
  analysis_urlsafe_key = basestring

  # The exact commit position being requested for analysis.
  commit_position = int

  # The exact revision corresponding to commit_position being requested.
  revision = basestring


class GetIsolateShaForBuildParameters(StructuredObject):
  # The name of the master to query for a pre-determined sha.
  master_name = basestring

  # The name of the builder to query for a pre-determined sha.
  builder_name = basestring

  # The build number whose to query for a pre-detrermined sha.
  build_number = int

  # The name of the step to query for a pre-determined sha.
  step_name = basestring


class GetIsolateShaForBuildPipeline(SynchronousPipeline):
  input_type = GetIsolateShaForBuildParameters
  output_type = basestring

  def RunImpl(self, parameters):
    return swarming.GetIsolatedShaForStep(
        parameters.master_name, parameters.builder_name,
        parameters.build_number, parameters.step_name, FinditHttpClient())


class GetIsolateShaForCommitPositionPipeline(GeneratorPipeline):

  input_type = GetIsolateShaForCommitPositionParameters
  output_type = basestring

  def RunImpl(self, parameters):
    """Determines the Isolated sha to run in swarming given a commit position.

    If the requested commit position maps directly to a  build, simply get that
    existing build's isolated sha. Otherwise, trigger a try job to compile and
    isolate at that revision and return the resulting sha.
    """
    analysis = ndb.Key(urlsafe=parameters.analysis_urlsafe_key).get()
    assert analysis

    master_name = analysis.master_name
    builder_name = analysis.builder_name
    test_name = analysis.test_name
    step_name = analysis.step_name
    commit_position = parameters.commit_position

    _, earliest_containing_build = build_util.GetBoundingBuilds(
        master_name, builder_name, None, None, commit_position)

    assert earliest_containing_build
    assert earliest_containing_build.commit_position >= commit_position

    if earliest_containing_build.commit_position == commit_position:
      # The requested commit position is that of an existing build.
      get_build_sha_parameters = self.CreateInputObjectInstance(
          GetIsolateShaForBuildParameters,
          master_name=master_name,
          builder_name=builder_name,
          build_number=earliest_containing_build.build_number,
          step_name=step_name)
      yield GetIsolateShaForBuildPipeline(get_build_sha_parameters)
    else:
      # The requested commit position needs to be compiled.
      _, reference_build_info = build_util.GetBuildInfo(
          master_name, builder_name, analysis.build_number)
      parent_mastername = reference_build_info.parent_mastername or master_name
      parent_buildername = (
          reference_build_info.parent_buildername or builder_name)
      cache_name = swarmbot_util.GetCacheName(
          parent_mastername,
          parent_buildername,
          suffix=flake_constants.FLAKE_CACHE_SUFFIX)
      dimensions = waterfall_config.GetTrybotDimensions(parent_mastername,
                                                        parent_buildername)
      try_job = flake_try_job.GetTryJob(master_name, builder_name, step_name,
                                        test_name, parameters.revision)
      run_flake_try_job_parameters = self.CreateInputObjectInstance(
          RunFlakeTryJobParameters,
          analysis_urlsafe_key=parameters.analysis_urlsafe_key,
          revision=parameters.revision,
          flake_cache_name=cache_name,
          dimensions=ListOfBasestring.FromSerializable(dimensions),
          urlsafe_try_job_key=try_job.key.urlsafe())
      yield RunFlakeTryJobPipeline(run_flake_try_job_parameters)
