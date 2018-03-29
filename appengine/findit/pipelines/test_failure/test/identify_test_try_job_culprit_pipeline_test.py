# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import mock

from dto.dict_of_basestring import DictOfBasestring
from gae_libs.pipeline_wrapper import pipeline_handlers
from libs import analysis_status
from libs.list_of_basestring import ListOfBasestring
from model.wf_try_job import WfTryJob
from pipelines.test_failure.revert_and_notify_test_culprit_pipeline import (
    RevertAndNotifyTestCulpritPipeline)
from pipelines.test_failure.identify_test_try_job_culprit_pipeline import (
    IdentifyTestTryJobCulpritPipeline)
from model.wf_suspected_cl import WfSuspectedCL
from services.parameters import BuildKey
from services.parameters import CulpritActionParameters
from services.parameters import IdentifyTestTryJobCulpritParameters
from services.parameters import TestTryJobResult
from services.test_failure import test_try_job
from waterfall.test import wf_testcase


class IdentifyTestTryJobCulpritPipelineTest(wf_testcase.WaterfallTestCase):
  app_module = pipeline_handlers._APP

  @mock.patch.object(test_try_job, 'IdentifyTestTryJobCulprits')
  def testIdentifyCulpritForTestTryJobSuccess(self, mock_fn):
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
                        'pass_fail_counts': {
                            'a_test1': {
                                'pass_count': 20,
                                'fail_count': 0
                            }
                        }
                    },
                    'b_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['b_test1'],
                        'pass_fail_counts': {
                            'b_test1': {
                                'pass_count': 0,
                                'fail_count': 20
                            }
                        }
                    }
                },
                'rev1': {
                    'a_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['a_test1'],
                        'pass_fail_counts': {
                            'a_test1': {
                                'pass_count': 0,
                                'fail_count': 20
                            }
                        }
                    },
                    'b_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['b_test1'],
                        'pass_fail_counts': {
                            'b_test1': {
                                'pass_count': 0,
                                'fail_count': 20
                            }
                        }
                    }
                },
                'rev2': {
                    'a_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['a_test1', 'a_test2'],
                        'pass_fail_counts': {
                            'a_test1': {
                                'pass_count': 0,
                                'fail_count': 0
                            },
                            'b_test1': {
                                'pass_count': 0,
                                'fail_count': 0
                            }
                        }
                    },
                    'b_test': {
                        'status': 'failed',
                        'valid': True,
                        'failures': ['b_test1'],
                        'pass_fail_counts': {
                            'b_test1': {
                                'pass_count': 0,
                                'fail_count': 0
                            }
                        }
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
                'a_test1': 'rev1',
                'a_test2': 'rev2'
            },
        }
    }

    repo_name = 'chromium'
    revision = 'rev2'

    culprit = WfSuspectedCL.Create(repo_name, revision, 100)
    culprit.put()

    culprits_result = {
        'rev1': {
            'revision': 'rev1',
            'repo_name': 'chromium',
            'commit_position': 1,
            'url': 'url_1'
        },
        'rev2': {
            'revision': revision,
            'commit_position': 2,
            'url': 'url_2',
            'repo_name': repo_name
        }
    }
    mock_fn.return_value = culprits_result, ListOfBasestring()

    culprits = DictOfBasestring()
    culprits['rev2'] = culprit.key.urlsafe()
    self.MockGeneratorPipeline(
        pipeline_class=RevertAndNotifyTestCulpritPipeline,
        expected_input=CulpritActionParameters(
            build_key=BuildKey(
                master_name=master_name,
                builder_name=builder_name,
                build_number=build_number),
            culprits=culprits,
            heuristic_cls=ListOfBasestring()),
        mocked_output=False)

    parameters = IdentifyTestTryJobCulpritParameters(
        build_key=BuildKey(
            master_name=master_name,
            builder_name=builder_name,
            build_number=build_number),
        result=TestTryJobResult.FromSerializable(test_result))
    pipeline = IdentifyTestTryJobCulpritPipeline(parameters)
    pipeline.start()
    self.execute_queued_tasks()

  def testReturnNoneIfNoTryJob(self):
    master_name = 'm'
    builder_name = 'b'
    build_number = 8

    WfTryJob.Create(master_name, builder_name, build_number).put()

    self.MockGeneratorPipeline(
        pipeline_class=RevertAndNotifyTestCulpritPipeline,
        expected_input=CulpritActionParameters(
            build_key=BuildKey(
                master_name=master_name,
                builder_name=builder_name,
                build_number=build_number),
            culprits=DictOfBasestring(),
            heuristic_cls=ListOfBasestring()),
        mocked_output=False)
    parameters = IdentifyTestTryJobCulpritParameters(
        build_key=BuildKey(
            master_name=master_name,
            builder_name=builder_name,
            build_number=build_number),
        result=None)
    pipeline = IdentifyTestTryJobCulpritPipeline(parameters)
    pipeline.start()
    self.execute_queued_tasks()

    try_job = WfTryJob.Get(master_name, builder_name, build_number)
    self.assertEqual(try_job.test_results, [])
    self.assertEqual(try_job.status, analysis_status.COMPLETED)
