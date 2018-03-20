# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# TODO(crbug.com/810912): Refacor into services.

from google.appengine.ext import ndb

from common.findit_http_client import FinditHttpClient
from dto.test_location import TestLocation
from gae_libs import pipelines
from gae_libs.pipelines import pipeline
from libs.structured_object import StructuredObject
from model.flake.flake_analysis_request import FlakeAnalysisRequest
from services import issue_tracking_service
from services import swarmed_test_util
from services import swarming
from waterfall import build_util
from waterfall.flake import triggering_sources
from waterfall.flake.analyze_flake_for_build_number_pipeline import (
    AnalyzeFlakeForBuildNumberPipeline)
from waterfall.flake.lookback_algorithm import IsFullyStable

_SUBJECT_TEMPLATE = '{} is Flaky'
_BODY_TEMPLATE = ('Findit has detected a flake at test {}.\n\n'
                  'Culprit ({} confidence): {}\n'
                  'Regression range: {}\n\n'
                  'Analysis: {}\n\n'
                  'If this result was incorrect, apply the label '
                  'Findit-Incorrect-Result, mark the bug as Untriaged and the '
                  'component Tools>Test>Findit>Flakiness.')

# TODO(crbug.com/783335): Allow these values to be configurable.
_ITERATIONS_TO_CONFIRM_FLAKE = 30  # 30 iterations.
_ITERATIONS_TO_CONFIRM_FLAKE_TIMEOUT = 60 * 60  # One hour.


class CreateBugForFlakePipelineInputObject(StructuredObject):
  analysis_urlsafe_key = unicode
  test_location = TestLocation


class CreateBugForFlakePipeline(pipelines.GeneratorPipeline):
  input_type = CreateBugForFlakePipelineInputObject

  def RunImpl(self, input_object):
    """Creates a bug for a flake analysis.

    Creates a bug if certain conditions are satisfied. These conditions are
    logically unordered, and the ordering you see in the pipeline is to
    favor local operations over network requests. This pipeline shouldn't
    be retried since it files a bug with monorail. Instead a bit is set
    in MasterFlakeAnalysis before a filing is attempted `has_attemped_filing`
    in the event in a retry this pipeline will be abandoned entirely.
    """
    analysis = ndb.Key(urlsafe=input_object.analysis_urlsafe_key).get()
    assert analysis

    if not issue_tracking_service.ShouldFileBugForAnalysis(analysis):
      existing_test_bug_id = (
          issue_tracking_service.GetExistingBugIdForCustomizedField(
              analysis.test_name))
      if existing_test_bug_id and not analysis.bug_id:
        analysis.Update(bug_id=existing_test_bug_id)
      return

    most_recent_build_number = build_util.GetLatestBuildNumber(
        analysis.master_name, analysis.builder_name)
    if not most_recent_build_number:
      analysis.LogInfo('Bug not failed because latest build number not found.')
      return

    tasks = swarming.ListSwarmingTasksDataByTags(
        FinditHttpClient(), analysis.master_name, analysis.builder_name,
        most_recent_build_number, analysis.step_name)
    if not tasks:
      analysis.LogInfo('Bug not filed because no recent runs found.')
      return

    task = tasks[0]
    if not swarmed_test_util.IsTestEnabled(analysis.test_name, task.task_id):
      analysis.LogInfo('Bug not filed because test was fixed or disabled.')
      return

    analysis_pipeline = yield AnalyzeFlakeForBuildNumberPipeline(
        input_object.analysis_urlsafe_key, most_recent_build_number,
        _ITERATIONS_TO_CONFIRM_FLAKE, _ITERATIONS_TO_CONFIRM_FLAKE_TIMEOUT,
        True)
    with pipeline.After(analysis_pipeline):
      next_input_object = pipelines.CreateInputObjectInstance(
          _CreateBugIfStillFlakyInputObject,
          analysis_urlsafe_key=input_object.analysis_urlsafe_key,
          most_recent_build_number=most_recent_build_number)
      yield _CreateBugIfStillFlaky(next_input_object)


def _GenerateSubjectAndBodyForBug(analysis):
  culprit_url = 'None'
  culprit_confidence = 'None'
  if analysis.culprit_urlsafe_key:
    culprit = ndb.Key(urlsafe=analysis.culprit_urlsafe_key).get()
    assert culprit

    culprit_url = culprit.url
    culprit_confidence = "{0:0.1f}%".format(
        analysis.confidence_in_culprit * 100)

  # Find the regression range of the suspected data point.
  suspected_data_point = analysis.GetDataPointOfSuspectedBuild()

  regression_range_url = 'None'
  if suspected_data_point:
    lower_git_hash = suspected_data_point.previous_build_git_hash
    upper_git_hash = suspected_data_point.git_hash

    regression_range_url = ('https://crrev.com/%s..%s?pretty=fuller' %
                            (lower_git_hash, upper_git_hash))

  subject = _SUBJECT_TEMPLATE.format(analysis.test_name)
  analysis_link = ('https://findit-for-me.appspot.com/waterfall/flake?key=%s' %
                   analysis.key.urlsafe())
  body = _BODY_TEMPLATE.format(analysis.test_name, culprit_confidence,
                               culprit_url, regression_range_url, analysis_link)
  return subject, body


class _CreateBugIfStillFlakyInputObject(StructuredObject):
  analysis_urlsafe_key = unicode
  most_recent_build_number = int


class _CreateBugIfStillFlaky(pipelines.GeneratorPipeline):
  input_type = _CreateBugIfStillFlakyInputObject

  def RunImpl(self, input_object):
    analysis = ndb.Key(urlsafe=input_object.analysis_urlsafe_key).get()
    assert analysis

    data_point = analysis.FindMatchingDataPointWithBuildNumber(
        input_object.most_recent_build_number)

    # If we're out of bounds of the lower or upper flake threshold, this test
    # is stable (either passing or failing consistently).
    if not data_point or IsFullyStable(data_point.pass_rate):
      analysis.LogInfo('Bug not filed because test is stable in latest build.')
      return

    subject, body = _GenerateSubjectAndBodyForBug(analysis)
    priority_label = issue_tracking_service.GetPriorityLabelForConfidence(
        analysis.confidence_in_culprit)

    # Log our attempt in analysis so we don't retry perpetually.
    analysis.Update(has_attempted_filing=True)
    bug_id = issue_tracking_service.CreateBugForFlakeAnalyzer(
        analysis.test_name, subject, body, priority_label)
    if not bug_id:
      analysis.LogError('Couldn\'t create bug!')
      return

    analysis.Update(bug_id=bug_id, has_filed_bug=True)
    analysis.LogInfo('Filed bug with id %d' % bug_id)

    flake_analysis_request = FlakeAnalysisRequest.GetVersion(
        key=analysis.test_name)
    assert flake_analysis_request
    flake_analysis_request.Update(
        bug_reported_by=triggering_sources.FINDIT_PIPELINE, bug_id=bug_id)
