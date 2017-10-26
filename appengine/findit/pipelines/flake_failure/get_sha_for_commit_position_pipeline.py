# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.ext import ndb
from common.findit_http_client import FinditHttpClient
from gae_libs.pipelines import GeneratorPipeline
from libs.structured_object import StructuredObject

from waterfall import build_util
from waterfall import swarming_util


class GetShaForCommitPositionPipelineInputObject(StructuredObject):
  analysis_urlsafe_key = unicode
  commit_position = int


class GetShaForCommitPositionPipeline(GeneratorPipeline):

  input_type = GetShaForCommitPositionPipelineInputObject

  def RunImpl(self, input_object):
    """Determines the Isolated sha to run in swarming given a commit position.

    Maps a commit position to the nearest containing build that the analysis has
    not yet analyzed and return's that build's sha. If the resulting build has
    already been analyzed, then it is assumed at this point the analysis is
    identifying an exact culprit commit, thus a try job will be triggered to
    compile at that exact commit position to generate the sha to return.

    The requested commit_position should already assumed to be within range
    """
    analysis = ndb.Key(urlsafe=input_object.analysis_urlsafe_key).get()
    assert analysis

    master_name = analysis.original_master_name
    builder_name = analysis.original_builder_name
    step_name = analysis.step_name
    commit_position = input_object.commit_position
    earliest_containing_build = build_util.GetEarliestContainingBuild(
        master_name, builder_name, None, None, commit_position)

    assert earliest_containing_build.commit_position >= commit_position

    build_number = earliest_containing_build.build_number

    isolated_sha = None

    if analysis.FindMatchingDataPointWithBuildNumber(
        build_number):  # pragma: no cover
      # There is already a data point at the ceiling build number of the
      # requested commit position, which means by the time this code is being
      # called a regression range between two adjacent builds has already been
      # identified and a try job is needed.

      # TODO(crbug.com/778382): Trigger try job to compile at that commit and
      # return the corresponding isolated sha from the recipe.
      pass
    else:
      isolated_sha = swarming_util.GetIsolatedShaForStep(
          master_name, builder_name, build_number, step_name,
          FinditHttpClient())

    return isolated_sha
