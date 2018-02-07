# Copyright 2016 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import base64
import logging
import math

from google.appengine.ext import ndb

from dto.int_range import IntRange
from gae_libs.model.versioned_model import VersionedModel
from libs import analysis_status
from model import result_status
from model import triage_status
from model.base_analysis import BaseAnalysis
from model.base_build_model import BaseBuildModel
from model.base_triaged_model import TriagedModel
from model.flake.flake_swarming_task import FlakeSwarmingTaskData
from waterfall.build_info import BuildInfo


class DataPoint(ndb.Model):
  # TODO(crbug.com/809218): Deprecate fields that refer to 'build.'

  # The build number corresponding to this data point. Only relevant for
  # analysis at the build level.
  build_number = ndb.IntegerProperty(indexed=False)

  # The pass rate of the test when run against this commit.
  # -1 means that the test doesn't exist at this commit/build.
  pass_rate = ndb.FloatProperty(indexed=False)

  # The ID of the swarming task responsible for generating this data.
  task_ids = ndb.StringProperty(indexed=False, repeated=True)

  # The commit position of this data point.
  commit_position = ndb.IntegerProperty(indexed=False)

  # The git hash of this data point.
  git_hash = ndb.StringProperty(indexed=False)

  # The commit position of the build preceding this one. Only relevant if this
  # data point is generated at the build level.
  previous_build_commit_position = ndb.IntegerProperty(indexed=False)

  # The git hash of the data point 1 build before this one. Only relevant if
  # this data point is generated as the result of a flake swarming task.
  previous_build_git_hash = ndb.StringProperty(indexed=False)

  # The list of revisions between this build and the previous build. Only
  # relevant if this data point is generated as the result of a flake swarming
  # task.
  blame_list = ndb.StringProperty(repeated=True)

  # The URL to the try job that generated this data point, if any.
  try_job_url = ndb.StringProperty(indexed=False)

  # A flag indicates whether the checked build has valid artifact.
  # This flag is only for build level data points.
  has_valid_artifact = ndb.BooleanProperty(indexed=False, default=True)

  # The number of iterations run to determine this data point's pass rate.
  iterations = ndb.IntegerProperty(indexed=False)

  # The total seconds that these iterations took to compute.
  elapsed_seconds = ndb.IntegerProperty(indexed=False, default=0)

  # The number of times a swarming task had an error while generating this
  # data point.
  failed_swarming_task_attempts = ndb.IntegerProperty(indexed=False, default=0)

  @staticmethod
  def Create(build_number=None,
             pass_rate=None,
             task_ids=None,
             commit_position=None,
             git_hash=None,
             previous_build_commit_position=None,
             previous_build_git_hash=None,
             blame_list=None,
             try_job_url=None,
             has_valid_artifact=True,
             iterations=None,
             elapsed_seconds=0,
             failed_swarming_task_attempts=0):
    data_point = DataPoint()
    data_point.build_number = build_number
    data_point.pass_rate = pass_rate
    task_ids = task_ids or []
    data_point.task_ids = task_ids
    data_point.commit_position = commit_position
    data_point.git_hash = git_hash
    data_point.previous_build_commit_position = previous_build_commit_position
    data_point.previous_build_git_hash = previous_build_git_hash
    data_point.blame_list = blame_list or []
    data_point.try_job_url = try_job_url
    data_point.has_valid_artifact = has_valid_artifact
    data_point.iterations = iterations
    data_point.elapsed_seconds = elapsed_seconds
    data_point.failed_swarming_task_attempts = failed_swarming_task_attempts
    return data_point

  def GetSwarmingTaskId(self):
    """Returns the last element in the list.

    Guaranteed to be within flake_constants.CONVERGENCE_PERCENT of the true pass
    rate if the pass rate converged.
    """
    return self.task_ids[-1] if self.task_ids else None

  def GetCommitPosition(self, revision):
    """Gets the commit position of a revision within blame_list.

    Args:
      revision (str): The revision to search for.

    Returns:
      commit_position (int): The calculated commit position of revision.
    """
    assert revision in self.blame_list

    for i in range(0, len(self.blame_list)):  # pragma: no branch
      if revision == self.blame_list[i]:
        return i + self.previous_build_commit_position + 1

  def GetRevisionAtCommitPosition(self, commit_position):
    """Gets the corresponding revision to commit_position.

    Args:
      commit_position (int): The commit position for which to find the
          corresponding revision within self.blame_list.

    Returns:
      revision (str): The git revision corresponding to commit_position.
    """
    length = len(self.blame_list)
    assert (commit_position > self.commit_position - length and
            commit_position <= self.commit_position)
    return self.blame_list[length - (self.commit_position - commit_position) -
                           1]

  def GetDictOfCommitPositionAndRevision(self):
    """Gets a dict of commit_position:revision items for this data_point."""
    blamed_cls = {}
    commit_position = self.commit_position
    for i in xrange(len(self.blame_list) - 1, -1, -1):
      blamed_cls[commit_position] = self.blame_list[i]
      commit_position -= 1

    return blamed_cls


class MasterFlakeAnalysis(BaseAnalysis, BaseBuildModel, VersionedModel,
                          TriagedModel):
  """Represents an analysis of a flaky test on a Waterfall test cycle."""

  @ndb.ComputedProperty
  def step_name(self):
    return self.key.pairs()[0][1].split('/')[3]

  @ndb.ComputedProperty
  def canonical_step_name(self):
    return self.step_name.split(' on ')[0]

  @ndb.ComputedProperty
  def test_name(self):
    return base64.urlsafe_b64decode(self.key.pairs()[0][1].split('/')[4])

  @property
  def error_message(self):
    if not self.error:
      return None
    return self.error.get('message')

  @property
  def iterations_to_rerun(self):
    if not self.algorithm_parameters:
      return -1
    return (self.algorithm_parameters.get('swarming_rerun',
                                          {}).get('iterations_to_rerun') or
            self.algorithm_parameters.get('iterations_to_rerun'))

  @staticmethod
  def _CreateAnalysisId(master_name, builder_name, build_number, step_name,
                        test_name):
    encoded_test_name = base64.urlsafe_b64encode(test_name)
    return '%s/%s/%s/%s/%s' % (master_name, builder_name, build_number,
                               step_name, encoded_test_name)

  @staticmethod
  def GetBuildConfigurationFromKey(master_flake_analysis_key):
    """Extracts master_name and builder_name from key."""
    if not master_flake_analysis_key:
      return None, None

    components = master_flake_analysis_key.pairs()[0][1].split('/')
    master_name = components[0]
    builder_name = components[1]
    return master_name, builder_name

  # Arguments number differs from overridden method - pylint: disable=W0221
  @classmethod
  def Create(cls, master_name, builder_name, build_number, step_name,
             test_name):  # pragma: no cover.
    # TODO(wylieb): Populate original_* fields with these, add test case for
    # Create.
    return super(MasterFlakeAnalysis, cls).Create(
        MasterFlakeAnalysis._CreateAnalysisId(
            master_name, builder_name, build_number, step_name, test_name))

  # Arguments number differs from overridden method - pylint: disable=W0221
  @classmethod
  def GetVersion(cls,
                 master_name,
                 builder_name,
                 build_number,
                 step_name,
                 test_name,
                 version=None):  # pragma: no cover.
    return super(MasterFlakeAnalysis, cls).GetVersion(
        key=MasterFlakeAnalysis._CreateAnalysisId(
            master_name, builder_name, build_number, step_name, test_name),
        version=version)

  def LogInfo(self, message):
    logging.info('%s/%s/%s/%s/%s %s', self.master_name, self.builder_name,
                 self.build_number, self.step_name, self.test_name, message)

  def LogWarning(self, message):
    logging.warning('%s/%s/%s/%s/%s %s', self.master_name, self.builder_name,
                    self.build_number, self.step_name, self.test_name, message)

  def LogError(self, message):
    logging.error('%s/%s/%s/%s/%s %s', self.master_name, self.builder_name,
                  self.build_number, self.step_name, self.test_name, message)

  def CanRunHeuristicAnalysis(self):
    """Determies whether heuristic analysis can be attempted."""
    already_run_statuses = [
        analysis_status.SKIPPED, analysis_status.COMPLETED,
        analysis_status.ERROR
    ]

    return (self.suspected_flake_build_id is not None and
            self.heuristic_analysis_status not in already_run_statuses)

  def UpdateTriageResult(self,
                         triage_result,
                         suspect_info,
                         user_name,
                         version_number=None):
    """Updates triage result for a flake analysis.

    If there is culprit for the analysis, triage will be at CL level;
    otherwise the triage will be for suspected_flake_build.
    """
    super(MasterFlakeAnalysis, self).UpdateTriageResult(
        triage_result, suspect_info, user_name, version_number=version_number)

    if triage_result == triage_status.TRIAGED_CORRECT:
      self.result_status = result_status.FOUND_CORRECT
      if suspect_info.get('culprit_revision'):
        self.correct_culprit = True
    else:
      self.result_status = result_status.FOUND_INCORRECT
      if suspect_info.get('culprit_revision'):
        self.correct_culprit = False

  def GetDataPointOfSuspectedBuild(self):
    """Gets the corresponding data point to the suspected flake build."""
    if self.suspected_flake_build_number is not None:
      for data_point in self.data_points:
        if data_point.build_number == self.suspected_flake_build_number:
          return data_point

    return None

  def Reset(self):
    super(MasterFlakeAnalysis, self).Reset()
    self.original_master_name = None
    self.original_builder_name = None
    self.original_build_number = None
    self.original_step_name = None
    self.original_test_name = None
    self.bug_id = None
    self.error = None
    self.correct_regression_range = None
    self.correct_culprit = None
    self.algorithm_parameters = None
    self.suspected_flake_build_number = None
    self.suspect_urlsafe_keys = []
    self.culprit_urlsafe_key = None
    self.try_job_status = None
    self.data_points = []
    self.result_status = None
    self.last_attempted_build_number = None
    self.last_attempted_swarming_task_id = None
    self.last_attempted_revision = None

  def GetCommitPositionOfBuild(self, build_number):
    """Gets the commit position of a build in self.data_points if available.

      Searches self.data_points for the data point with the corresponding build
      number and returns its commit position if found, else None. Note that data
      points generated as a result of try jobs should not have build_number set.

    Args:
      build_number (int): The build number to find the matching data point.

    Returns:
      The commit position of the data point with the matching build number.
    """
    for data_point in self.data_points:
      # Skip try job data points since they should not have build_number.
      if (data_point.build_number == build_number and
          data_point.try_job_url is None):
        return data_point.commit_position
    return None

  # TODO(crbug.com/798231): Remove once build level analysis is removed.
  def GetDataPointsWithinBuildNumberRange(self, lower_bound_build_number,
                                          upper_bound_build_number):
    """Filters data_points by lower and upper bound build numbers.

      All data points within the build number range will be returned, including
      data points created by try jobs.

    Args:
      data_points (list): A list of DataPoint objects.
      lower_bound_build_number (int): The earlist build number a data point can
          have not to be filtered out. If None is passed, defaults to 0.
      upper_bound_build_number (int): The latest build number a data point can
          have not to be filtered out. If none is passed, defaults to infinity.

    Returns:
      A list of DataPoints filtered by the input build numbers.
    """
    if lower_bound_build_number is None and upper_bound_build_number is None:
      return self.data_points

    lower_bound = self.GetCommitPositionOfBuild(lower_bound_build_number) or 0
    upper_bound = self.GetCommitPositionOfBuild(upper_bound_build_number)

    return self.GetDataPointsWithinCommitPositionRange(
        IntRange(lower=lower_bound, upper=upper_bound))

  def GetDataPointsWithinCommitPositionRange(self, int_range):
    """Filters data_points by lower and upper bound commit positions.

    Args:
      int_range (IntRange): The upper and lower bound commit positions to
          include in the returned results.

    Returns:
      A list of DataPoints filtered by the input commit positions.
    """
    lower = int_range.lower
    upper = int_range.upper if int_range.upper is not None else float('inf')

    def PositionInBounds(data_point):
      return (data_point.commit_position >= lower and
              data_point.commit_position <= upper)

    return filter(PositionInBounds, self.data_points)

  def RemoveDataPointWithBuildNumber(self, build_number):
    self.data_points = filter(lambda x: x.build_number != build_number,
                              self.data_points)

  def RemoveDataPointWithCommitPosition(self, commit_position):
    self.data_points = filter(lambda x: x.commit_position != commit_position,
                              self.data_points)

  def FindMatchingDataPointWithCommitPosition(self, commit_position):
    """Finds the data point with the same commit_position as the given one.

    Args:
      commit_position (int): DataPoint with the matching commit position to
          search for in the list.

    Returns:
      A DataPoint with the matching commit_position if found, else None.
    """
    if commit_position is None:
      return None

    return next((data_point for data_point in self.data_points
                 if data_point.commit_position == commit_position), None)

  def FindMatchingDataPointWithBuildNumber(self, build_number):
    """Finds the data point with the same build_number as the given one.

    Args:
      build_number (int): DataPoint with the matching build number to search for
          in the list.

    Returns:
      A DataPoint with the matching build_number if found, else None.
    """
    if build_number is None:
      return None

    return next((data_point for data_point in self.data_points
                 if data_point.build_number == build_number), None)

  def UpdateSuspectedBuildID(self, lower_bound_build, upper_bound_build):
    """Sets the suspected build ID if appropriate.

      A suspected build cycle can be set when a regression range is identified
      and spans at most a single build cycle.

    Args:
      analysis_urlsafe_key (str): The key to the analysis to update.
      lower_bound_build (BuildInfo): The earlier build whose commit position to
          check.
      upper_bound_build (BuildInfo): The later build whose commit
          position to check, assumed to be 1 build cycle apartlower_bound_build.
    """
    lower_bound = lower_bound_build.commit_position
    upper_bound = upper_bound_build.commit_position
    assert upper_bound > lower_bound

    if (self.suspected_flake_build_id is None and
        self.FindMatchingDataPointWithCommitPosition(lower_bound) and
        self.FindMatchingDataPointWithCommitPosition(upper_bound)):
      self.Update(suspected_flake_build_id=str(upper_bound_build.build_number))

  def Update(self, **kwargs):
    """Updates fields according to what's specified in kwargs.

      Fields specified in kwargs will be updated accordingly, while those not
      present in kwargs will be untouched.

    Args:
      algorithm_parameters (dict): The analysis' algorithm parameters.
      confidence_in_culprit (float): Confidence score for the suspected culprit
          CL, if any.
      confidence_in_suspected_build (float): Confidence score for the suspected
          build number.
      culprit_urlsafe_key (str): The urlsafe-key coresponding to a FlakeCulprit
          that caused flakiness.
      end_time (datetime): The timestamp that the overall analysis is completed.
      error (dict): Dict containing error information.
      last_attempted_swarming_revision (str): The last attempted try job
          revision.
      last_attempted_swarming_revision (str): The ID of the last attempted
          swarming task.
      result_status (int): The triage result status of this analysis.
      status (int): The status of the regression-range identification analysis.
      start_time (datetime): The timestamp that the overall analysis started.
      suspected_build (int): The suspected build number.
      try_job_status (int): The status of try job/culprit analysis.
    """
    any_changes = False

    for arg, value in kwargs.iteritems():
      current_value = getattr(self, arg, None)
      if current_value != value:
        setattr(self, arg, value)
        any_changes = True

    if any_changes:
      self.put()

  # The original build/step/test in which a flake actually occurred.
  # A CQ trybot step has to be mapped to a Waterfall buildbot step.
  # A gtest suite.PRE_PRE_test has to be normalized to suite.test.
  original_master_name = ndb.StringProperty(indexed=True)
  original_builder_name = ndb.StringProperty(indexed=True)
  original_build_number = ndb.IntegerProperty(indexed=True)
  original_step_name = ndb.StringProperty(indexed=True)
  original_test_name = ndb.StringProperty(indexed=True)

  # The bug id in which this flake is reported.
  bug_id = ndb.IntegerProperty(indexed=True)

  # A bit to track if a bug filing has been attempted.
  has_attempted_filing = ndb.BooleanProperty(default=False)

  # Error code and message, if any.
  error = ndb.JsonProperty(indexed=False)

  # Boolean whether the suspected regression range/build number is correct.
  correct_regression_range = ndb.BooleanProperty(indexed=True)

  # Boolean whether the suspected CL for found in the regression range
  # is correct.
  correct_culprit = ndb.BooleanProperty(indexed=True)

  # The look back algorithm parameters that were used, as specified in Findit's
  # configuration. For example,
  # {
  #     'iterations_to_rerun': 100,
  #     'lower_flake_threshold': 0.02,
  #     'max_build_numbers_to_look_back': 500,
  #     'max_flake_in_a_row': 4,
  #     'max_stable_in_a_row': 4,
  #     'upper_flake_threshold': 0.98
  # }
  algorithm_parameters = ndb.JsonProperty(indexed=False)

  # The ID of the build flakiness was suspected to have been introduced in.
  suspected_flake_build_id = ndb.StringProperty(indexed=False)

  # The suspected build number to have introduced the flakiness.
  # TODO(crbug.com/799324): Remove once build numbers are deprecated in LUCI.
  suspected_flake_build_number = ndb.IntegerProperty()

  # The confidence in the suspected build to have introduced the flakiness.
  confidence_in_suspected_build = ndb.FloatProperty(indexed=False)

  # The confidence in the suspected CL to have introduced the flakiness.
  confidence_in_culprit = ndb.FloatProperty(indexed=False)

  # TODO(crbug.com/799308): Use KeyProperty instead.
  # The urlsafe key to a FlakeCulprit associated with the try job results.
  culprit_urlsafe_key = ndb.StringProperty(indexed=False)

  # TODO(crbug.com/799308): Use KeyProperty instead.
  # A list of url-safe keys to FlakeCulprits identified by heuristic analysis.
  suspect_urlsafe_keys = ndb.StringProperty(repeated=True)

  # Heuristic anlysis status. Can be PENDING if not yet ran, SKIPPED if wil not
  # be run, COMPLETED or ERROR if already ran.
  heuristic_analysis_status = ndb.IntegerProperty(indexed=False)

  # The status of try jobs, if any. None if analysis is still performing
  # swarming reruns, SKIPPED if try jobs will not be triggered, RUNNING when
  # the first try job is triggered, COMPLETED when the last one finishes, and
  # ERROR if any try job ends with error.
  try_job_status = ndb.IntegerProperty(indexed=False)

  # The data points used to plot the flakiness graph build over build.
  data_points = ndb.LocalStructuredProperty(
      DataPoint, repeated=True, compressed=True)

  # Whether the analysis was triggered by a manual request through check flake,
  # Findit's automatic analysis upon detection, or Findit API.
  triggering_source = ndb.IntegerProperty(default=None, indexed=True)

  # Who triggered the analysis. Used for differentiating between manual and
  # automatic runs, and determining the most active users to gather feedback.
  triggering_user_email = ndb.StringProperty(default=None, indexed=False)

  # Whether the user email is obscured.
  triggering_user_email_obscured = ndb.BooleanProperty(
      default=False, indexed=True)

  # Overall conclusion of analysis result for the flake. Found untriaged, Found
  # Correct, etc. used to filter what is displayed on the check flake dashboard.
  result_status = ndb.IntegerProperty(indexed=True)

  # The build number corresponding to the last attempted swarming task.
  last_attempted_build_number = ndb.IntegerProperty(indexed=False)

  # The task id of the last-attempted swarming task.
  last_attempted_swarming_task_id = ndb.StringProperty(indexed=False)

  # The revision the last-attempted try job tried to run on.
  last_attempted_revision = ndb.StringProperty(indexed=False)

  # Store the root pipeline id to look up the pipeline from the analysis.
  root_pipeline_id = ndb.StringProperty(indexed=False)

  # How many times the swarming task has been attempted (and failed) for a
  # particular build during an analysis. Reset when the build number is changed.
  swarming_task_attempts_for_build = ndb.IntegerProperty(
      default=0, indexed=False)
