# Copyright 2016 The Chromium Authors. All rights reserved.

# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock

from monorail_api import Issue

from libs import analysis_status
from model.flake.flake_culprit import FlakeCulprit
from model.flake.master_flake_analysis import DataPoint
from model.flake.master_flake_analysis import MasterFlakeAnalysis
from waterfall.flake import update_flake_bug_pipeline
from waterfall.test import wf_testcase


class UpdateFlakeToBugPipelineTest(wf_testcase.WaterfallTestCase):

  def testGenerateCommentUponError(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.ERROR
    comment = update_flake_bug_pipeline._GenerateComment(analysis)
    self.assertTrue('due to an error' in comment, comment)

  def testGenerateCommentWithCulprit(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    culprit = FlakeCulprit.Create('c', 'r', 123, 'http://')
    culprit.flake_analysis_urlsafe_keys.append(analysis.key.urlsafe())
    culprit.put()
    analysis.culprit_urlsafe_key = culprit.key.urlsafe()
    analysis.confidence_in_culprit = 0.6713
    comment = update_flake_bug_pipeline._GenerateComment(analysis)
    self.assertTrue('culprit r123 with confidence 67.1%' in comment, comment)

  def testGenerateCommentWithSuspectedBuildHighConfidence(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.suspected_flake_build_number = 120
    analysis.confidence_in_suspected_build = 0.6641
    comment = update_flake_bug_pipeline._GenerateComment(analysis)
    self.assertTrue('started in build 120' in comment, comment)

  def testGenerateCommentWithSuspectedBuildLowConfidence(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.suspected_flake_build_number = 120
    analysis.confidence_in_suspected_build = 0.3641
    comment = update_flake_bug_pipeline._GenerateComment(analysis)
    self.assertTrue('low flakiness' in comment, comment)

  def testGenerateCommentForLongstandingFlake(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    comment = update_flake_bug_pipeline._GenerateComment(analysis)
    self.assertTrue('longstanding one' in comment, comment)

  def testShouldUpdateBugForAnalysisError(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.error = {'error': 'error', 'message': 'message'}
    self.assertFalse(
        update_flake_bug_pipeline._ShouldUpdateBugForAnalysis(analysis))

  def testShouldUpdateBugForAnalysisFalse(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.bug_id = 123
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    analysis.suspected_flake_build_number = 1
    analysis.algorithm_parameters = {'update_monorail_bug': False}
    self.assertFalse(
        update_flake_bug_pipeline._ShouldUpdateBugForAnalysis(analysis))

  def testShouldUpdateBugForAnalysisInsufficientConfidence(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.bug_id = 123
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    analysis.suspected_flake_build_number = 1
    analysis.algorithm_parameters = {
        'update_monorail_bug': True,
        'minimum_confidence_score_to_run_tryjobs': 0.6
    }
    analysis.confidence_in_suspected_build = 0.5
    self.assertFalse(
        update_flake_bug_pipeline._ShouldUpdateBugForAnalysis(analysis))

  def testShouldUpdateBugForAnalysis(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.bug_id = 123
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    analysis.suspected_flake_build_number = 1
    analysis.algorithm_parameters = {
        'update_monorail_bug': True,
        'minimum_confidence_score_to_run_tryjobs': 0.6
    }
    analysis.confidence_in_suspected_build = 0.7
    self.assertTrue(
        update_flake_bug_pipeline._ShouldUpdateBugForAnalysis(analysis))

  @mock.patch('waterfall.flake.update_flake_bug_pipeline.IssueTrackerAPI')
  def testNotUpdateBug(self, issue_tracker):
    analysis_not_completed = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis_not_completed.status = analysis_status.RUNNING
    analysis_without_bug = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis_without_bug.status = analysis_status.COMPLETED
    analysis_config_not_to_update = MasterFlakeAnalysis.Create(
        'm', 'b', 1, 's', 't')
    analysis_config_not_to_update.status = analysis_status.COMPLETED
    analysis_config_not_to_update.bug_id = 123
    analysis_config_not_to_update.data_points = [
        DataPoint(), DataPoint(), DataPoint()
    ]
    analysis_config_not_to_update.algorithm_parameters = {
        'update_monorail_bug': False,
    }
    analysis_without_enough_data_points = MasterFlakeAnalysis.Create(
        'm', 'b', 1, 's', 't')
    analysis_without_enough_data_points.status = analysis_status.COMPLETED
    analysis_without_enough_data_points.bug_id = 123
    analysis_without_enough_data_points.algorithm_parameters = {
        'update_monorail_bug': True,
    }
    analysis_without_enough_data_points.data_points = [DataPoint()]

    analyses = [
        analysis_not_completed, analysis_without_bug,
        analysis_config_not_to_update, analysis_without_enough_data_points
    ]
    for analysis in analyses:
      analysis.put()
      pipeline = update_flake_bug_pipeline.UpdateFlakeBugPipeline()
      self.assertFalse(pipeline.run(analysis.key.urlsafe()))
    self.assertFalse(issue_tracker.called)

  @mock.patch('google.appengine.api.app_identity.get_application_id',
              lambda: 'findit-for-me')
  @mock.patch('waterfall.flake.update_flake_bug_pipeline.IssueTrackerAPI')
  def testNoUpdateIfBugDeleted(self, issue_tracker):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.suspected_flake_build_number = 1
    analysis.confidence_in_suspected_build = 0.7
    analysis.bug_id = 123
    analysis.algorithm_parameters = {'update_monorail_bug': True}
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    analysis.put()

    mocked_instance = mock.Mock()
    mocked_instance.getIssue.return_value = None
    issue_tracker.return_value = mocked_instance
    pipeline = update_flake_bug_pipeline.UpdateFlakeBugPipeline()
    self.assertFalse(pipeline.run(analysis.key.urlsafe()))
    issue_tracker.assert_has_calls(
        [mock.call('chromium', use_staging=False),
         mock.call().getIssue(123)])

  @mock.patch('google.appengine.api.app_identity.get_application_id',
              lambda: 'findit-for-me-staging')
  @mock.patch('waterfall.flake.update_flake_bug_pipeline.IssueTrackerAPI')
  def testUsesStagingIssueTrackerAPI(self, issue_tracker):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.suspected_flake_build_number = 1
    analysis.confidence_in_suspected_build = 0.7
    analysis.bug_id = 123
    analysis.algorithm_parameters = {'update_monorail_bug': True}
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    analysis.put()

    mocked_instance = mock.Mock()
    mocked_instance.getIssue.return_value = None
    issue_tracker.return_value = mocked_instance
    pipeline = update_flake_bug_pipeline.UpdateFlakeBugPipeline()
    self.assertFalse(pipeline.run(analysis.key.urlsafe()))
    issue_tracker.assert_has_calls(
        [mock.call('chromium', use_staging=True),
         mock.call().getIssue(123)])

  @mock.patch('google.appengine.api.app_identity.get_application_id',
              lambda: 'findit-for-me')
  @mock.patch('waterfall.flake.update_flake_bug_pipeline.IssueTrackerAPI')
  def testBugUpdated(self, issue_tracker):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.bug_id = 123
    analysis.suspected_flake_build_number = 1
    analysis.confidence_in_suspected_build = 0.7
    analysis.algorithm_parameters = {'update_monorail_bug': True}
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    analysis.put()

    dummy_issue = Issue({})
    mocked_instance = mock.Mock()
    mocked_instance.getIssue.return_value = dummy_issue
    issue_tracker.return_value = mocked_instance
    pipeline = update_flake_bug_pipeline.UpdateFlakeBugPipeline()
    self.assertTrue(pipeline.run(analysis.key.urlsafe()))
    issue_tracker.assert_has_calls([
        mock.call('chromium', use_staging=False),
        mock.call().getIssue(123),
        mock.call().update(dummy_issue, mock.ANY, send_email=True)
    ])

  @mock.patch('google.appengine.api.app_identity.get_application_id',
              lambda: 'findit-for-me')
  @mock.patch('waterfall.flake.update_flake_bug_pipeline.IssueTrackerAPI')
  def testBugUpdatedWithCulpritFound(self, issue_tracker):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.bug_id = 123
    analysis.suspected_flake_build_number = 1
    analysis.confidence_in_culprit = 0.6713
    analysis.algorithm_parameters = {'update_monorail_bug': True}
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    culprit = FlakeCulprit.Create('c', 'r', 123, 'http://')
    culprit.flake_analysis_urlsafe_keys.append(analysis.key.urlsafe())
    culprit.put()
    analysis.culprit_urlsafe_key = culprit.key.urlsafe()
    analysis.put()

    dummy_issue = Issue({})
    mocked_instance = mock.Mock()
    mocked_instance.getIssue.return_value = dummy_issue
    issue_tracker.return_value = mocked_instance
    pipeline = update_flake_bug_pipeline.UpdateFlakeBugPipeline()
    self.assertTrue(pipeline.run(analysis.key.urlsafe()))
    issue_tracker.assert_has_calls([
        mock.call('chromium', use_staging=False),
        mock.call().getIssue(123),
        mock.call().update(dummy_issue, mock.ANY, send_email=True)
    ])
    self.assertTrue(analysis.has_commented_on_bug)

  @mock.patch('waterfall.flake.update_flake_bug_pipeline.IssueTrackerAPI')
  def testBugNotUpdatedWithCulpritFoundIfBUpdateDisabled(self, issue_tracker):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.bug_id = 123
    analysis.suspected_flake_build_number = 1
    analysis.confidence_in_culprit = 0.6713
    analysis.algorithm_parameters = {'update_monorail_bug': False}
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    culprit = FlakeCulprit.Create('c', 'r', 123, 'http://')
    culprit.flake_analysis_urlsafe_keys.append(analysis.key.urlsafe())
    culprit.put()
    analysis.culprit_urlsafe_key = culprit.key.urlsafe()
    analysis.put()

    dummy_issue = Issue({})
    pipeline = update_flake_bug_pipeline.UpdateFlakeBugPipeline()
    self.assertFalse(pipeline.run(analysis.key.urlsafe()))
    self.assertFalse(issue_tracker.called)

  @mock.patch('google.appengine.api.app_identity.get_application_id',
              lambda: 'findit-for-me')
  @mock.patch('waterfall.flake.update_flake_bug_pipeline.IssueTrackerAPI')
  def testNoUpdateWithCulpritIfBugDeleted(self, issue_tracker):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.suspected_flake_build_number = 1
    analysis.confidence_in_suspected_build = 0.7
    analysis.bug_id = 123
    analysis.algorithm_parameters = {'update_monorail_bug': True}
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    culprit = FlakeCulprit.Create('c', 'r', 123, 'http://')
    culprit.flake_analysis_urlsafe_keys.append(analysis.key.urlsafe())
    culprit.put()
    analysis.culprit_urlsafe_key = culprit.key.urlsafe()
    analysis.put()

    mocked_instance = mock.Mock()
    mocked_instance.getIssue.return_value = None
    issue_tracker.return_value = mocked_instance
    pipeline = update_flake_bug_pipeline.UpdateFlakeBugPipeline()
    self.assertFalse(pipeline.run(analysis.key.urlsafe()))
    issue_tracker.assert_has_calls(
        [mock.call('chromium', use_staging=False),
         mock.call().getIssue(123)])

  def testNoUpdateWithCulpritIfNoBugId(self):
    analysis = MasterFlakeAnalysis.Create('m', 'b', 1, 's', 't')
    analysis.status = analysis_status.COMPLETED
    analysis.suspected_flake_build_number = 1
    analysis.confidence_in_suspected_build = 0.7
    analysis.algorithm_parameters = {'update_monorail_bug': True}
    analysis.data_points = [DataPoint(), DataPoint(), DataPoint()]
    culprit = FlakeCulprit.Create('c', 'r', 123, 'http://')
    culprit.flake_analysis_urlsafe_keys.append(analysis.key.urlsafe())
    culprit.put()
    analysis.culprit_urlsafe_key = culprit.key.urlsafe()
    analysis.put()

    pipeline = update_flake_bug_pipeline.UpdateFlakeBugPipeline()
    self.assertFalse(pipeline.run(analysis.key.urlsafe()))
