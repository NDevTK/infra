# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock

from common import constants
from gae_libs.pipelines import pipeline_handlers
from libs import analysis_status
from model.wf_analysis import WfAnalysis
from services.parameters import TestHeuristicAnalysisOutput
from services.parameters import TestHeuristicAnalysisParameters
from pipelines.test_failure import analyze_test_failure_pipeline
from pipelines.test_failure.analyze_test_failure_pipeline import (
    AnalyzeTestFailurePipeline)
from waterfall.test import wf_testcase


class AnalyzeTestFailurePipelineTest(wf_testcase.WaterfallTestCase):
  app_module = pipeline_handlers._APP

  def _SetupAnalysis(self,
                     master_name,
                     builder_name,
                     build_number,
                     status=analysis_status.PENDING,
                     signals=None,
                     failure_info=None):
    analysis = WfAnalysis.Create(master_name, builder_name, build_number)
    analysis.status = status
    analysis.failure_info = failure_info
    analysis.signals = signals
    analysis.put()

  def testBuildFailurePipelineFlow(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 124
    current_failure_info = {}

    self._SetupAnalysis(master_name, builder_name, build_number)

    heuristic_params = TestHeuristicAnalysisParameters.FromSerializable({
        'failure_info': current_failure_info,
        'build_completed': False
    })
    heuristic_output = {'failure_info': None, 'heuristic_result': None}
    self.MockSynchronousPipeline(
        analyze_test_failure_pipeline.HeuristicAnalysisForTestPipeline,
        heuristic_params, TestHeuristicAnalysisOutput.FromSerializable({}))
    self.MockPipeline(
        analyze_test_failure_pipeline.TriggerSwarmingTasksPipeline,
        None,
        expected_args=[
            master_name, builder_name, build_number, heuristic_output, False
        ],
        expected_kwargs={})
    self.MockPipeline(
        analyze_test_failure_pipeline.ProcessSwarmingTasksResultPipeline,
        None,
        expected_args=[
            master_name, builder_name, build_number, heuristic_output, False
        ],
        expected_kwargs={})
    self.MockPipeline(
        analyze_test_failure_pipeline.StartTestTryJobPipeline,
        'try_job_result',
        expected_args=[
            master_name, builder_name, build_number, heuristic_output, False,
            False
        ],
        expected_kwargs={})
    self.MockPipeline(
        analyze_test_failure_pipeline.TriggerFlakeAnalysesPipeline,
        None,
        expected_args=[master_name, builder_name, build_number],
        expected_kwargs={})

    root_pipeline = AnalyzeTestFailurePipeline(
        master_name, builder_name, build_number, current_failure_info, False,
        False)
    root_pipeline.start(queue_name=constants.DEFAULT_QUEUE)
    self.execute_queued_tasks()
    analysis = WfAnalysis.Get(master_name, builder_name, build_number)
    self.assertEqual(analysis_status.RUNNING, analysis.status)

  def testBuildFailurePipelineStartWithNoneResultStatus(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 124

    self._SetupAnalysis(master_name, builder_name, build_number)

    root_pipeline = AnalyzeTestFailurePipeline(master_name, builder_name,
                                               build_number, None, False, False)
    root_pipeline._ResetAnalysis(master_name, builder_name, build_number)
    analysis = WfAnalysis.Get(master_name, builder_name, build_number)
    self.assertIsNotNone(analysis)
    self.assertEqual(analysis_status.RUNNING, analysis.status)
    self.assertIsNone(analysis.result_status)

  def testAnalyzeTestFailurePipelineAbortedIfWithError(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 124

    self._SetupAnalysis(
        master_name, builder_name, build_number, status=analysis_status.RUNNING)

    root_pipeline = AnalyzeTestFailurePipeline(master_name, builder_name,
                                               build_number, None, False, False)
    root_pipeline._HandleUnexpectedAborting(True)

    analysis = WfAnalysis.Get(master_name, builder_name, build_number)
    self.assertIsNotNone(analysis)
    self.assertEqual(analysis_status.ERROR, analysis.status)
    self.assertIsNone(analysis.result_status)
    self.assertTrue(analysis.aborted)

  def testAnalyzeTestFailurePipelineNotAbortedIfWithoutError(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 124

    self._SetupAnalysis(
        master_name,
        builder_name,
        build_number,
        status=analysis_status.COMPLETED)

    root_pipeline = AnalyzeTestFailurePipeline(master_name, builder_name,
                                               build_number, None, False, False)
    root_pipeline._HandleUnexpectedAborting(True)

    analysis = WfAnalysis.Get(master_name, builder_name, build_number)
    self.assertIsNotNone(analysis)
    self.assertNotEqual(analysis_status.ERROR, analysis.status)

  @mock.patch.object(analyze_test_failure_pipeline, 'StartTestTryJobPipeline')
  def testAnalyzeTestFailurePipelineStartTryJob(self, mocked_pipeline):
    master_name = 'm'
    builder_name = 'b'
    build_number = 124
    failure_info = {
        'test': {
            'last_pass': 122,
            'current_failure': 123,
            'first_failure': 123
        }
    }

    self._SetupAnalysis(
        master_name,
        builder_name,
        build_number,
        status=analysis_status.RUNNING,
        signals={},
        failure_info=failure_info)

    root_pipeline = AnalyzeTestFailurePipeline(master_name, builder_name,
                                               build_number, None, False, False)
    root_pipeline._HandleUnexpectedAborting(True)

    heuristic_result = {'failure_info': failure_info, 'heuristic_result': None}
    mocked_pipeline.assert_called_once_with(
        master_name, builder_name, build_number, heuristic_result, False, False)
    mocked_pipeline.assert_has_calls(
        [mock.call().start(queue_name=constants.WATERFALL_ANALYSIS_QUEUE)])
