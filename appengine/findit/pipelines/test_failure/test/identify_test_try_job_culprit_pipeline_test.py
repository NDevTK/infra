# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from common.waterfall import failure_type
from gae_libs.pipeline_wrapper import pipeline_handlers
from libs import analysis_status
from gae_libs.gitiles.cached_gitiles_repository import CachedGitilesRepository
from model import result_status
from model.wf_analysis import WfAnalysis
from model.wf_try_job import WfTryJob
from model.wf_try_job_data import WfTryJobData
from pipelines.pipeline_inputs_and_outputs import BuildKey
from pipelines.pipeline_inputs_and_outputs import CLKey
from pipelines.pipeline_inputs_and_outputs import DictOfCLKeys
from pipelines.pipeline_inputs_and_outputs import ListOfCLKeys
from pipelines.pipeline_inputs_and_outputs import (
    RevertAndNotifyCulpritPipelineInput)
from pipelines.test_failure.revert_and_notify_test_culprit_pipeline import (
    RevertAndNotifyTestCulpritPipeline)
from pipelines.test_failure.identify_test_try_job_culprit_pipeline import (
    IdentifyTestTryJobCulpritPipeline)
from waterfall.test import wf_testcase


class IdentifyTestTryJobCulpritPipelineTest(wf_testcase.WaterfallTestCase):
  app_module = pipeline_handlers._APP

  def _MockGetChangeLog(self, revision):

    class MockedChangeLog(object):

      def __init__(self, commit_position, code_review_url):
        self.commit_position = commit_position
        self.code_review_url = code_review_url
        self.change_id = str(commit_position)

    mock_change_logs = {}
    mock_change_logs['rev1'] = MockedChangeLog(1, 'url_1')
    mock_change_logs['rev2'] = MockedChangeLog(2, 'url_2')
    return mock_change_logs.get(revision)

  def setUp(self):
    super(IdentifyTestTryJobCulpritPipelineTest, self).setUp()

    self.mock(CachedGitilesRepository, 'GetChangeLog', self._MockGetChangeLog)

  def _CreateEntities(self,
                      master_name,
                      builder_name,
                      build_number,
                      try_job_id,
                      try_job_status,
                      test_results=None):
    try_job = WfTryJob.Create(master_name, builder_name, build_number)
    try_job.status = try_job_status
    if test_results:
      try_job.test_results = test_results
    try_job.put()

    try_job_data = WfTryJobData.Create(try_job_id)
    try_job_data.try_job_key = try_job.key
    try_job_data.put()

  def testIdentifyCulpritForTestTryJobNoTryJobResultNoHeuristicResult(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 1
    try_job_id = '1'

    self._CreateEntities(
        master_name,
        builder_name,
        build_number,
        try_job_id,
        try_job_status=analysis_status.RUNNING)

    analysis = WfAnalysis.Create(master_name, builder_name, build_number)
    analysis.put()

    self.MockGeneratorPipeline(
        pipeline_class=RevertAndNotifyTestCulpritPipeline,
        expected_input=RevertAndNotifyCulpritPipelineInput(
            build_key=BuildKey(
                master_name=master_name,
                builder_name=builder_name,
                build_number=build_number),
            culprits=DictOfCLKeys(),
            heuristic_cls=ListOfCLKeys()),
        mocked_output=False)

    pipeline = IdentifyTestTryJobCulpritPipeline(master_name, builder_name,
                                                 build_number, '1', None)
    pipeline.start()
    self.execute_queued_tasks()

    try_job_data = WfTryJobData.Get(try_job_id)
    self.assertIsNone(try_job_data.culprits)
    self.assertIsNone(analysis.result_status)
    self.assertIsNone(analysis.suspected_cls)

  def testIdentifyCulpritForTestTryJobNoTryJobResultWithHeuristicResult(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 1
    try_job_id = '1'

    suspected_cl = {
        'revision': 'rev1',
        'commit_position': 1,
        'url': 'url_1',
        'repo_name': 'chromium'
    }

    self._CreateEntities(
        master_name,
        builder_name,
        build_number,
        try_job_id,
        try_job_status=analysis_status.RUNNING)

    # Heuristic analysis already provided some results.
    analysis = WfAnalysis.Create(master_name, builder_name, build_number)
    analysis.result_status = result_status.FOUND_UNTRIAGED
    analysis.suspected_cls = [suspected_cl]
    analysis.put()

    heuristic_cls = ListOfCLKeys()
    heuristic_cls.append(CLKey(repo_name=u'chromium', revision=u'rev1'))
    self.MockGeneratorPipeline(
        pipeline_class=RevertAndNotifyTestCulpritPipeline,
        expected_input=RevertAndNotifyCulpritPipelineInput(
            build_key=BuildKey(
                master_name=master_name,
                builder_name=builder_name,
                build_number=build_number),
            culprits=DictOfCLKeys(),
            heuristic_cls=heuristic_cls),
        mocked_output=False)

    pipeline = IdentifyTestTryJobCulpritPipeline(master_name, builder_name,
                                                 build_number, '1', None)
    pipeline.start()
    self.execute_queued_tasks()

    try_job_data = WfTryJobData.Get(try_job_id)
    self.assertIsNone(try_job_data.culprits)

    # Ensure analysis results are not updated since no culprit from try job.
    self.assertEqual(analysis.result_status, result_status.FOUND_UNTRIAGED)
    self.assertEqual(analysis.suspected_cls, [suspected_cl])

  def testIdentifyCulpritForTestTryJobReturnNoneIfNoRevisionToCheck(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 1
    try_job_id = '1'

    test_result = {
        'report': {
            'result': {
                'rev1': {
                    'a_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['a_test1']
                    }
                }
            }
        },
        'url': 'url',
        'try_job_id': try_job_id
    }

    self._CreateEntities(
        master_name,
        builder_name,
        build_number,
        try_job_id,
        try_job_status=analysis_status.RUNNING)

    analysis = WfAnalysis.Create(master_name, builder_name, build_number)
    analysis.put()

    self.MockGeneratorPipeline(
        pipeline_class=RevertAndNotifyTestCulpritPipeline,
        expected_input=RevertAndNotifyCulpritPipelineInput(
            build_key=BuildKey(
                master_name=master_name,
                builder_name=builder_name,
                build_number=build_number),
            culprits=DictOfCLKeys(),
            heuristic_cls=ListOfCLKeys()),
        mocked_output=False)

    pipeline = IdentifyTestTryJobCulpritPipeline(master_name, builder_name,
                                                 build_number, '1', test_result)
    pipeline.start()
    self.execute_queued_tasks()

    try_job_data = WfTryJobData.Get(try_job_id)
    self.assertIsNone(try_job_data.culprits)

    self.assertIsNone(analysis.result_status)
    self.assertIsNone(analysis.suspected_cls)

  def testIdentifyCulpritForTestTryJobReturnRevisionIfNoCulpritInfo(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 1
    try_job_id = '1'

    test_result = {
        'report': {
            'result': {
                'rev3': {
                    'a_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['a_test1']
                    }
                }
            },
            'culprits': {
                'a_test': {
                    'a_test1': 'rev3'
                }
            }
        },
        'url': 'url',
        'try_job_id': try_job_id
    }

    self._CreateEntities(
        master_name,
        builder_name,
        build_number,
        try_job_id,
        try_job_status=analysis_status.RUNNING)

    analysis = WfAnalysis.Create(master_name, builder_name, build_number)
    analysis.put()

    expected_analysis_suspected_cls = [{
        'revision': 'rev3',
        'repo_name': 'chromium',
        'failures': {
            'a_test': ['a_test1']
        },
        'top_score': None
    }]

    culprits = DictOfCLKeys()
    culprits['rev3'] = CLKey(repo_name=u'chromium', revision=u'rev3')
    self.MockGeneratorPipeline(
        pipeline_class=RevertAndNotifyTestCulpritPipeline,
        expected_input=RevertAndNotifyCulpritPipelineInput(
            build_key=BuildKey(
                master_name=master_name,
                builder_name=builder_name,
                build_number=build_number),
            culprits=culprits,
            heuristic_cls=ListOfCLKeys()),
        mocked_output=False)

    pipeline = IdentifyTestTryJobCulpritPipeline(master_name, builder_name,
                                                 build_number, '1', test_result)
    pipeline.start()
    self.execute_queued_tasks()

    try_job_data = WfTryJobData.Get(try_job_id)
    analysis = WfAnalysis.Get(master_name, builder_name, build_number)
    expected_culprit_data = {'a_test': {'a_test1': 'rev3'}}
    self.assertEqual(expected_culprit_data, try_job_data.culprits)
    self.assertEqual(analysis.result_status, result_status.FOUND_UNTRIAGED)
    self.assertEqual(analysis.suspected_cls, expected_analysis_suspected_cls)

  def testIdentifyCulpritForTestTryJobSuccess(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 1
    try_job_id = '1'

    test_result = {
        'report': {
            'result': {
                'rev0': {
                    'a_test': {
                        'status': 'passed',
                        'valid': True,
                    },
                    'b_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['b_test1']
                    }
                },
                'rev1': {
                    'a_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['a_test1']
                    },
                    'b_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['b_test1']
                    }
                },
                'rev2': {
                    'a_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['a_test1', 'a_test2']
                    },
                    'b_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['b_test1']
                    }
                }
            },
            'culprits': {
                'a_test': {
                    'a_test1': 'rev1',
                    'a_test2': 'rev2'
                },
            },
            'flakes': {
                'b_test': ['b_test1']
            }
        },
        'url': 'url',
        'try_job_id': try_job_id
    }

    self._CreateEntities(
        master_name,
        builder_name,
        build_number,
        try_job_id,
        try_job_status=analysis_status.RUNNING,
        test_results=[test_result])

    analysis = WfAnalysis.Create(master_name, builder_name, build_number)
    analysis.put()

    a_test1_suspected_cl = {
        'revision': 'rev1',
        'commit_position': 1,
        'url': 'url_1',
        'repo_name': 'chromium'
    }
    a_test2_suspected_cl = {
        'revision': 'rev2',
        'commit_position': 2,
        'url': 'url_2',
        'repo_name': 'chromium'
    }

    expected_test_result = {
        'report': {
            'result': {
                'rev0': {
                    'a_test': {
                        'status': 'passed',
                        'valid': True,
                    },
                    'b_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['b_test1']
                    }
                },
                'rev1': {
                    'a_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['a_test1']
                    },
                    'b_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['b_test1']
                    }
                },
                'rev2': {
                    'a_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['a_test1', 'a_test2']
                    },
                    'b_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['b_test1']
                    }
                }
            },
            'culprits': {
                'a_test': {
                    'a_test1': 'rev1',
                    'a_test2': 'rev2'
                },
            },
            'flakes': {
                'b_test': ['b_test1']
            }
        },
        'url': 'url',
        'try_job_id': try_job_id,
        'culprit': {
            'a_test': {
                'tests': {
                    'a_test1': a_test1_suspected_cl,
                    'a_test2': a_test2_suspected_cl
                }
            }
        }
    }

    culprits = DictOfCLKeys()
    culprits['rev1'] = CLKey(repo_name=u'chromium', revision=u'rev1')
    culprits['rev2'] = CLKey(repo_name=u'chromium', revision=u'rev2')
    self.MockGeneratorPipeline(
        pipeline_class=RevertAndNotifyTestCulpritPipeline,
        expected_input=RevertAndNotifyCulpritPipelineInput(
            build_key=BuildKey(
                master_name=master_name,
                builder_name=builder_name,
                build_number=build_number),
            culprits=culprits,
            heuristic_cls=ListOfCLKeys()),
        mocked_output=False)

    pipeline = IdentifyTestTryJobCulpritPipeline(master_name, builder_name,
                                                 build_number, '1', test_result)
    pipeline.start()
    self.execute_queued_tasks()

    try_job = WfTryJob.Get(master_name, builder_name, build_number)
    self.assertEqual(expected_test_result, try_job.test_results[-1])
    self.assertEqual(analysis_status.COMPLETED, try_job.status)

    try_job_data = WfTryJobData.Get(try_job_id)
    analysis = WfAnalysis.Get(master_name, builder_name, build_number)
    expected_culprit_data = {
        'a_test': {
            'a_test1': 'rev1',
            'a_test2': 'rev2',
        }
    }

    expected_cls = [{
        'revision': 'rev1',
        'commit_position': 1,
        'url': 'url_1',
        'repo_name': 'chromium',
        'failures': {
            'a_test': ['a_test1'],
            'b_test': ['b_test1'],
        },
        'top_score': None
    }, {
        'revision': 'rev2',
        'commit_position': 2,
        'url': 'url_2',
        'repo_name': 'chromium',
        'failures': {
            'a_test': ['a_test1', 'a_test2'],
            'b_test': ['b_test1'],
        },
        'top_score': None
    }]
    self.assertEqual(expected_culprit_data, try_job_data.culprits)
    self.assertEqual(analysis.result_status, result_status.FOUND_UNTRIAGED)
    self.assertEqual(analysis.suspected_cls, expected_cls)

  def testReturnNoneIfNoTryJob(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 8

    WfTryJob.Create(master_name, builder_name, build_number).put()

    self.MockGeneratorPipeline(
        pipeline_class=RevertAndNotifyTestCulpritPipeline,
        expected_input=RevertAndNotifyCulpritPipelineInput(
            build_key=BuildKey(
                master_name=master_name,
                builder_name=builder_name,
                build_number=build_number),
            culprits=DictOfCLKeys(),
            heuristic_cls=ListOfCLKeys()),
        mocked_output=False)
    pipeline = IdentifyTestTryJobCulpritPipeline(master_name, builder_name,
                                                 build_number, None, None)
    pipeline.start()
    self.execute_queued_tasks()

    try_job = WfTryJob.Get(master_name, builder_name, build_number)
    self.assertEqual(try_job.test_results, [])
    self.assertEqual(try_job.status, analysis_status.COMPLETED)
