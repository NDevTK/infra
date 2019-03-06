# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.ext import ndb

from common import monitoring
from dto.int_range import IntRange
from dto.step_metadata import StepMetadata
from gae_libs.pipelines import GeneratorPipeline
from gae_libs.pipelines import pipeline
from libs import analysis_status
from libs import time_util
from libs.list_of_basestring import ListOfBasestring
from libs.structured_object import StructuredObject
from model import result_status
from model.base_build_model import BaseBuildModel
from pipelines.delay_pipeline import DelayPipeline
from pipelines.flake_failure.create_and_submit_revert_pipeline import (
    CreateAndSubmitRevertInput)
from pipelines.flake_failure.create_and_submit_revert_pipeline import (
    CreateAndSubmitRevertPipeline)
from pipelines.flake_failure.determine_approximate_pass_rate_pipeline import (
    DetermineApproximatePassRateInput)
from pipelines.flake_failure.determine_approximate_pass_rate_pipeline import (
    DetermineApproximatePassRatePipeline)
from pipelines.flake_failure.get_isolate_sha_pipeline import (
    GetIsolateShaForCommitPositionParameters)
from pipelines.flake_failure.get_isolate_sha_pipeline import (
    GetIsolateShaForCommitPositionPipeline)
from pipelines.flake_failure.next_commit_position_pipeline import (
    NextCommitPositionInput)
from pipelines.flake_failure.next_commit_position_pipeline import (
    NextCommitPositionOutput)
from pipelines.flake_failure.next_commit_position_pipeline import (
    NextCommitPositionPipeline)
from pipelines.flake_failure.notify_culprit_pipeline import NotifyCulpritInput
from pipelines.flake_failure.notify_culprit_pipeline import (
    NotifyCulpritPipeline)
from pipelines.flake_failure.update_flake_analysis_data_points_pipeline import (
    UpdateFlakeAnalysisDataPointsInput)
from pipelines.flake_failure.update_flake_analysis_data_points_pipeline import (
    UpdateFlakeAnalysisDataPointsPipeline)
from pipelines.report_event_pipeline import ReportAnalysisEventPipeline
from pipelines.report_event_pipeline import ReportEventInput
from services.flake_failure import confidence_score_util
from services.actions import flake_analysis_actions
from services.flake_failure import flake_analysis_util


class AnalyzeFlakeInput(StructuredObject):
  # The urlsafe key to the MasterFlakeAnalysis in progress.
  analysis_urlsafe_key = basestring

  # Information on the exact commit position to analyze.
  analyze_commit_position_parameters = NextCommitPositionOutput

  # The lower/upper bound commit positions not to exceed.
  commit_position_range = IntRange

  # Dimensions of the bot that will be used to trigger try jobs.
  dimensions = ListOfBasestring

  # A flag indicating this pipeline was triggered by a human request.
  manually_triggered = bool

  # Whether this is an admin-triggered rerun of an existing analysis. Not to be
  # confused with manually_triggered, which can be a manual request via Findit's
  # homepage.
  rerun = bool

  # The number of times bots have been checked for availability.
  retries = int

  # Information about the test used to find bots that will run swarming tasks.
  step_metadata = StepMetadata


class AnalyzeFlakePipeline(GeneratorPipeline):
  """The main driving pipeline for flake analysis."""

  input_type = AnalyzeFlakeInput

  def OnFinalized(self, parameters):
    if not self.IsRootPipeline():
      # AnalyzeFlakePipeline is recursive. Only the root pipeline should update.
      return

    analysis_urlsafe_key = parameters.analysis_urlsafe_key
    analysis = ndb.Key(urlsafe=analysis_urlsafe_key).get()
    assert analysis, 'Cannot retrieve analysis entry from datastore'

    # Get the analysis' already-detected error, if any.
    error = analysis.error

    if self.was_aborted:
      error = analysis.GetError()  # Capture any undetected error.
      monitoring.aborted_pipelines.increment({'type': 'flake'})

    status = analysis_status.ERROR if error else analysis_status.COMPLETED
    analysis.Update(error=error, end_time=time_util.GetUTCNow(), status=status)

    # TODO(crbug.com/847644): If error is set, report to ts_mon.

    # Monitor completion of pipeline.
    monitoring.completed_pipelines.increment({'type': 'flake'})

  def RunImpl(self, parameters):
    analysis_urlsafe_key = parameters.analysis_urlsafe_key
    analysis = ndb.Key(urlsafe=analysis_urlsafe_key).get()
    assert analysis, 'Cannot retrieve analysis entry from datastore'
    if analysis.request_time:
      monitoring.pipeline_times.increment_by(
          int((time_util.GetUTCNow() - analysis.request_time).total_seconds()),
          {'type': 'flake'})

    commit_position_parameters = parameters.analyze_commit_position_parameters
    commit_position_to_analyze = (
        commit_position_parameters.next_commit_id.commit_position
        if commit_position_parameters.next_commit_id else None)
    culprit_commit_position = (
        commit_position_parameters.culprit_commit_id.commit_position
        if commit_position_parameters.culprit_commit_id else None)

    if commit_position_to_analyze is None:
      # No commit position to analyze. The analysis is finished.

      if culprit_commit_position is None:
        # No culprit was identified. No further action.
        analysis.LogInfo('Analysis completed with no findings')
        analysis.Update(result_status=result_status.NOT_FOUND_UNTRIAGED)

        # Report events to BQ.
        yield ReportAnalysisEventPipeline(
            self.CreateInputObjectInstance(
                ReportEventInput, analysis_urlsafe_key=analysis_urlsafe_key))
        return

      # Create a FlakeCulprit.
      culprit_revision = commit_position_parameters.culprit_commit_id.revision
      assert culprit_revision, 'No revision for commit {}'.format(
          culprit_commit_position)
      culprit = flake_analysis_util.UpdateCulprit(
          analysis_urlsafe_key, culprit_revision, culprit_commit_position)
      confidence_score = confidence_score_util.CalculateCulpritConfidenceScore(
          analysis, culprit_commit_position)

      # Update the analysis' culprit.
      analysis.Update(
          confidence_in_culprit=confidence_score,
          culprit_urlsafe_key=culprit.key.urlsafe(),
          result_status=result_status.FOUND_UNTRIAGED)

      # TODO(crbug.com/905754): Call auto actions as an async taskqueue task.
      flake_analysis_actions.OnCulpritIdentified(analysis_urlsafe_key)

      with pipeline.InOrder():
        if flake_analysis_util.ShouldTakeAutoAction(
            analysis, parameters.rerun):  # pragma: no branch
          # Determine the test's location for filing bugs.
          culprit_data_point = analysis.FindMatchingDataPointWithCommitPosition(
              culprit_commit_position)
          assert culprit_data_point, 'Culprit unexpectedly missing!'

          # Data needed for reverts.
          build_key = BaseBuildModel.CreateBuildKey(
              analysis.original_master_name, analysis.original_builder_name,
              analysis.original_build_number)

          # Revert culprit if applicable.
          yield CreateAndSubmitRevertPipeline(
              self.CreateInputObjectInstance(
                  CreateAndSubmitRevertInput,
                  analysis_urlsafe_key=analysis.key.urlsafe(),
                  build_key=build_key))

          # Update culprit code review.
          yield NotifyCulpritPipeline(
              self.CreateInputObjectInstance(
                  NotifyCulpritInput,
                  analysis_urlsafe_key=analysis_urlsafe_key))

        if not parameters.rerun:  # pragma: no branch
          # Report events to BQ.
          yield ReportAnalysisEventPipeline(
              self.CreateInputObjectInstance(
                  ReportEventInput, analysis_urlsafe_key=analysis_urlsafe_key))
        return

    revision_to_analyze = commit_position_parameters.next_commit_id.revision
    assert revision_to_analyze, 'No revision for commit {}'.format(
        commit_position_to_analyze)

    # Check for bot availability. If this is a user rerun or the maximum retries
    # have been reached, continue regardless of bot availability.
    if flake_analysis_util.CanStartAnalysisImmediately(
        parameters.step_metadata, parameters.retries,
        parameters.manually_triggered):

      # Set analysis status to RUNNING if not already.
      analysis.InitializeRunning()

      analysis.LogInfo(
          'Analyzing commit position {}'.format(commit_position_to_analyze))

      with pipeline.InOrder():
        # Determine isolate sha to run swarming tasks on.
        upper_bound_build_number = analysis.GetLowestUpperBoundBuildNumber(
            commit_position_to_analyze)
        get_sha_output = yield GetIsolateShaForCommitPositionPipeline(
            self.CreateInputObjectInstance(
                GetIsolateShaForCommitPositionParameters,
                analysis_urlsafe_key=analysis_urlsafe_key,
                commit_position=commit_position_to_analyze,
                dimensions=parameters.dimensions,
                step_metadata=parameters.step_metadata,
                revision=revision_to_analyze,
                upper_bound_build_number=upper_bound_build_number))

        # Determine approximate pass rate at the commit position/isolate sha.
        flakiness = yield DetermineApproximatePassRatePipeline(
            self.CreateInputObjectInstance(
                DetermineApproximatePassRateInput,
                builder_name=analysis.builder_name,
                commit_position=commit_position_to_analyze,
                flakiness_thus_far=None,
                get_isolate_sha_output=get_sha_output,
                master_name=analysis.master_name,
                previous_swarming_task_output=None,
                reference_build_number=analysis.build_number,
                revision=revision_to_analyze,
                step_name=analysis.step_name,
                test_name=analysis.test_name))

        yield UpdateFlakeAnalysisDataPointsPipeline(
            self.CreateInputObjectInstance(
                UpdateFlakeAnalysisDataPointsInput,
                analysis_urlsafe_key=analysis_urlsafe_key,
                flakiness=flakiness))

        # Determine the next commit position to analyze.
        next_commit_position_output = yield NextCommitPositionPipeline(
            self.CreateInputObjectInstance(
                NextCommitPositionInput,
                analysis_urlsafe_key=analysis_urlsafe_key,
                commit_position_range=parameters.commit_position_range,
                step_metadata=parameters.step_metadata))

        # Recurse on the new commit position.
        yield RecursiveAnalyzeFlakePipeline(
            self.CreateInputObjectInstance(
                AnalyzeFlakeInput,
                analysis_urlsafe_key=analysis_urlsafe_key,
                analyze_commit_position_parameters=next_commit_position_output,
                commit_position_range=parameters.commit_position_range,
                dimensions=parameters.dimensions,
                manually_triggered=parameters.manually_triggered,
                rerun=parameters.rerun,
                retries=0,
                step_metadata=parameters.step_metadata))
    else:
      # Can't start the analysis just yet, reschedule.
      parameters.retries += 1
      delay_seconds = flake_analysis_util.CalculateDelaySecondsBetweenRetries(
          analysis, parameters.retries, parameters.manually_triggered)
      delay = yield DelayPipeline(delay_seconds)

      with pipeline.After(delay):
        yield RecursiveAnalyzeFlakePipeline(parameters)


class RecursiveAnalyzeFlakePipeline(GeneratorPipeline):
  """A wrapper for AnalyzeFlakePipeline for testability only.

    Because AnalyzeFlakePipeline is recursive, in unit tests it is not possible
    to mock only the recursive call to validate its input independently of the
    original call.
  """
  input_type = AnalyzeFlakeInput

  def RunImpl(self, parameters):
    yield AnalyzeFlakePipeline(parameters)
