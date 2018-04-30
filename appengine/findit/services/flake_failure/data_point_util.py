# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""Functions to assist in operations on DataPoint objects."""

import logging

from google.appengine.ext import ndb

from model.flake.master_flake_analysis import DataPoint
from services import git
from services import step_util
from services.flake_failure import flake_analysis_util
from services.flake_failure import pass_rate_util
from waterfall import waterfall_config
from waterfall.flake import flake_constants


def _CreateNewDataPoint(commit_position, revision, swarming_task_output,
                        lower_bound_build, upper_bound_build, error):
  """Returns a new DataPoint based on swarming_task_output.

    The calling code should be responsible for handling ndb writes to retain the
    returned DataPoint.

  Args:
    commit_position (int): The commit position at which this DataPoint is to
        represent.
    revision (str): The Chromium revision corresponding to commit_position.
    swarming_task_output (FlakeSwarmingTaskOutput): The output of a flake
        swarming task with which to craft the corresponding data point.
    lower_bound_build (BuildInfo): The build preceding the commit position.
  """
  if error and not flake_analysis_util.CanFailedSwarmingTaskBeSalvaged(
      swarming_task_output):
    # TODO(crbug.com/808947): A failed swarming task's partial data can
    # sometimes still be salvaged.
    elapsed_seconds = None
    failed_swarming_task_attempts = 1
    iterations = None
    pass_rate = None
  else:
    elapsed_seconds = swarming_task_output.GetElapsedSeconds()
    failed_swarming_task_attempts = 0
    iterations = swarming_task_output.iterations
    pass_rate = pass_rate_util.GetPassRate(swarming_task_output)

  # Populate build-level information if available.
  if (lower_bound_build and upper_bound_build and
      upper_bound_build.commit_position == commit_position):
    # This data point was generated using existing build artifacts.
    blame_list = git.GetCommitsBetweenRevisionsInOrder(
        lower_bound_build.chromium_revision,
        upper_bound_build.chromium_revision,
        ascending=True)
    build_number = upper_bound_build.build_number
    previous_build_commit_position = lower_bound_build.commit_position
    previous_build_git_hash = lower_bound_build.chromium_revision
  else:
    # This data point was generated by compiling at a commit position.
    # TODO(crbug.com/832724): Set a flag in the data point indicating compile
    # was needed in order to run.
    blame_list = None
    build_number = None
    previous_build_commit_position = None
    previous_build_git_hash = None

  return DataPoint.Create(
      blame_list=blame_list,
      build_number=build_number,
      commit_position=commit_position,
      elapsed_seconds=elapsed_seconds,
      failed_swarming_task_attempts=failed_swarming_task_attempts,
      git_hash=revision,
      iterations=iterations,
      pass_rate=pass_rate,
      previous_build_commit_position=previous_build_commit_position,
      previous_build_git_hash=previous_build_git_hash,
      task_ids=[swarming_task_output.task_id])


def _UpdateExistingDataPoint(data_point, swarming_task_output):
  """Updates a data point's fields to reflect new swarming task output.

    It is up to the calling code to perform the actual ndb write to commit the
    new data.

  Args:
    data_point (DataPoint): The data point to update.
    swarming_task_output (FlakeSwarmingTaskOutput): The result of a flake
        swarming task whose data is to be used to update the data point.
  """
  assert data_point, '_UpdateExistingDataPoint must be called with a data point'
  assert swarming_task_output, (
      '_UpdateExistingDataPoint must be called with swarming task output')

  # Ensure no undetected error.
  assert swarming_task_output.iterations is not None, 'Iterations is None'
  assert swarming_task_output.pass_count is not None, 'Pass count is None'
  assert swarming_task_output.task_id is not None, 'Task ID is None'

  # The latest swarming task completed successfully. Incorporate the incoming
  # data with the existing data point.
  old_pass_rate = data_point.pass_rate
  old_iterations = data_point.iterations
  incoming_iterations = swarming_task_output.iterations
  old_pass_rate = data_point.pass_rate
  incoming_pass_rate = pass_rate_util.GetPassRate(swarming_task_output)

  # Ensure the are no discrepancies between old and new pass rates about the
  # test existing or not at the same commit position.
  assert not (pass_rate_util.TestDoesNotExist(incoming_pass_rate) and
              not pass_rate_util.TestDoesNotExist(old_pass_rate)), (
                  'Inconsistent test existence state')

  incoming_elapsed_seconds = swarming_task_output.GetElapsedSeconds()
  assert incoming_elapsed_seconds is not None, 'No elapsed seconds'

  data_point.task_ids.append(swarming_task_output.task_id)
  data_point.pass_rate = pass_rate_util.CalculateNewPassRate(
      old_pass_rate, old_iterations, incoming_pass_rate, incoming_iterations)
  data_point.iterations += incoming_iterations
  data_point.elapsed_seconds += incoming_elapsed_seconds


def GetMaximumIterationsToRunPerDataPoint():
  return waterfall_config.GetCheckFlakeSettings().get(
      'max_iterations_to_rerun',
      flake_constants.DEFAULT_MAX_ITERATIONS_TO_RERUN)


def GetMaximumSwarmingTaskRetriesPerDataPoint():
  return waterfall_config.GetCheckFlakeSettings().get(
      'maximum_swarming_task_retries_per_data_point',
      flake_constants.DEFAULT_MAX_SWARMING_TASK_RETRIES_PER_DATA_POINT)


def HasSeriesOfFullyStablePointsPrecedingCommitPosition(
    data_points, commit_position, required_number_of_stable_points):
  """Checks for a minimum number of fully-stable points before a given commit.

    Fully-stable must also be the same type of stable and for existing tests
    only. Forexample, fully-passing to fully-passing and fully-failing to fully-
    failing. This function should not be used when handling newly-added tests.

  Args:
    data_points ([DataPoint]): The list of data points of a MasterFlakeAnalysis.
        data_points is expected to be pre-sorted in ascending order by commit
        position.
    commit_position (int): The commit position to find stable points preceding.
    required_number_of_stable_points (int): The minimum number of data points
        of the same fully-stable type required in order to send a notification
        to a code review.
  """
  if required_number_of_stable_points > len(data_points):
    return False

  fully_stable_data_points_in_a_row = 0
  previous_data_point = data_points[0]

  for data_point in data_points:
    if data_point.commit_position == commit_position:
      break

    if pass_rate_util.IsFullyStable(data_point.pass_rate):
      # Only 100% passing or 100% failing can count towards fully-stable.
      if pass_rate_util.ArePassRatesEqual(data_point.pass_rate,
                                          previous_data_point.pass_rate):
        # Must be the same type of fully-stable in order to count towards the
        # series.
        fully_stable_data_points_in_a_row += 1
      else:
        # A new series of stable passing/failing began. For example, if a series
        # of passes is followed by a failure, begin counting at the failure.
        fully_stable_data_points_in_a_row = 1
    else:
      # A slightly-flaky data point was encuntered. Reset the count.
      fully_stable_data_points_in_a_row = 0

    previous_data_point = data_point

  return fully_stable_data_points_in_a_row >= required_number_of_stable_points


def MaximumSwarmingTaskRetriesReached(data_point):
  """Determines whether a data point has too many failed swarming task attempts.

  Args:
    data_point (DataPoint): The data point to check.

  Returns:
    True if the data point has had too many failed attempts at a swarming task.
  """
  max_swarming_retries = GetMaximumSwarmingTaskRetriesPerDataPoint()
  return data_point.failed_swarming_task_attempts > max_swarming_retries


def MaximumIterationsPerDataPointReached(iterations):
  max_iterations_to_run = GetMaximumIterationsToRunPerDataPoint()
  return iterations >= max_iterations_to_run


def UpdateFailedSwarmingTaskAttempts(data_point):
  assert data_point
  data_point.failed_swarming_task_attempts += 1
  data_point.put()


@ndb.transactional(xg=True)
def UpdateAnalysisDataPoints(analysis_urlsafe_key, commit_position, revision,
                             swarming_task_output, lower_bound_build,
                             upper_bound_build):
  """Updates an analysis' data points with the output of a swarming taks.

  Args:
    analysis_urlsafe_key (str): The urlsafe key to the MasterFlakeAnalysis
        whose data points are to be updated.
    commit_position (int): The commit position to insert or update a data point.
    revision (str): The corresponding revision to commit_position.
    swarming_task_output (RunFlakeSwarmingTaskOutput): The returned object
        of the last-completed flake swarming task to deterine flakiness at
        commit_position.
    lower_bound_build (BuildInfo): Info about the build prior to
        commit_position for updating data points' fields with. If None, the
        lower bound build level information won't be set.
    upper_bound_build (BuildInfo) Info about the build containing
        commit_position for updating data points' fields with. If None, the
        build level information won't be set.
  """
  analysis = ndb.Key(urlsafe=analysis_urlsafe_key).get()
  assert analysis, 'MasterFlakeAnalysis was deleted unexpectedly!'

  data_point = analysis.FindMatchingDataPointWithCommitPosition(commit_position)
  error = swarming_task_output.error
  if data_point:
    # A data point has already been created. Data should be combined.
    if error and not flake_analysis_util.CanFailedSwarmingTaskBeSalvaged(
        swarming_task_output):
      # The latest swarming task ran into an error.
      # TODO(crbug.com/808947): A failed swarming task's partial data can
      # sometimes still be salvaged.
      data_point.failed_swarming_task_attempts += 1
      if swarming_task_output.task_id:
        # Capture the task ID if available for diagnostic information.
        data_point.task_ids.append(swarming_task_output.task_id)
      analysis.put()
      return

    # Update the existing data point's data with the results of the swarming
    # task.
    _UpdateExistingDataPoint(data_point, swarming_task_output)
  else:
    # A new data point should always be created and appended, especially for
    # diagnostics in case of failures. Note if no pass rate is ultimately able
    # to be determined due to too many failed attempts, the UI should be
    # responsible not to display a data point with a pass_rate of None.
    new_data_point = _CreateNewDataPoint(
        commit_position, revision, swarming_task_output, lower_bound_build,
        upper_bound_build, error)
    analysis.data_points.append(new_data_point)

  analysis.put()
