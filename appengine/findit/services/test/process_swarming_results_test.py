# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import json
import datetime
import mock
import datetime

from gae_libs.pipeline_wrapper import pipeline_handlers
from libs import time_util
from libs import analysis_status

from model import swarming_task_queue_request
from model.flake import master_flake_analysis
from model.flake import flake_swarming_task
from services import process_swarming_results
from waterfall import swarming_util
from waterfall.test import wf_testcase

_SAMPLE_FAILURE_LOG = {
    'per_iteration_data': [{
        'TestSuite1.test1': [{
            'status': 'SUCCESS',
            'other_info': 'N/A'
        }],
        'TestSuite1.test2': [{
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'SUCCESS',
            'other_info': 'N/A'
        }],
        'TestSuite1.test3': [{
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }]
    }, {
        'TestSuite1.test1': [{
            'status': 'SUCCESS',
            'other_info': 'N/A'
        }],
        'TestSuite1.test2': [{
            'status': 'SUCCESS',
            'other_info': 'N/A'
        }],
        'TestSuite1.test3': [{
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }]
    }]
}

_EXPECTED_TESTS_STATUS = {
    'TestSuite1.test1': {
        'total_run': 2,
        'SUCCESS': 2
    },
    'TestSuite1.test2': {
        'total_run': 4,
        'SUCCESS': 2,
        'FAILURE': 2
    },
    'TestSuite1.test3': {
        'total_run': 6,
        'FAILURE': 6
    },
}

MOCK_LOG_DATA_NO_PER_ITERATION_DATA = {
    'disabled_tests': [
        'AdbClientSocketTest.TestFlushWithData',
        'AdbClientSocketTest.TestFlushWithSize',
        'AdbClientSocketTest.TestFlushWithoutSize'
    ],
    'global_tags': ['CPU_64_BITS', 'MODE_RELEASE', 'OS_MACOSX', 'OS_POSIX'],
    'all_tests': [
        'AcceleratorsCocoaBrowserTest.MainMenuAcceleratorsInMapping',
        'AcceleratorsCocoaBrowserTest.MappingAcceleratorsInMainMenu',
        'ActivityLogApiTest.TriggerEvent',
        'ActivityLogPrerenderTest.TestScriptInjected',
        'AdbClientSocketTest.TestFlushWithData'
    ],
    'per_iteration_data':
        None,
    'test_locations': {
        'PolicyToolUITest.Editing': {
            'line':
                271,
            'file':
                '../../chrome/browser/ui/webui/policy_tool_ui_browsertest.cc'
        }
    },
    'error':
        None,
    'step_name_no_platform':
        'browser_tests',
    'task_state':
        'COMPLETED'
}

MOCK_RESULT_DATA_RUNNING = {
    'task_id': 'task_id',
    'name': 'name',
    'tags': ['cpu:x86-64', 'findit:1'],
    'started_ts': '2017-10-24T22:19:20.104240',
    'created_ts': '2017-10-24T22:19:17.794970',
    'state': 'RUNNING'
}

MOCK_RESULT_DATA_NO_OUTPUTS_REF = {
    'task_id': 'task_id',
    'name': 'name',
    'tags': ['cpu:x86-64', 'findit:1'],
    'outputs_ref': None,
    'duration': 50.1,
    'completed_ts': '2017-10-24T22:20:32.361620',
    'started_ts': '2017-10-24T22:19:20.104240',
    'created_ts': '2017-10-24T22:19:17.794970',
    'state': 'COMPLETED'
}

MOCK_RESULT_DATA_COMPLETE = {
    'task_id': 'task_id',
    'name': 'name',
    'tags': ['cpu:x86-64', 'findit:1'],
    'outputs_ref': {
        'isolatedserver': 'https://isolateserver.appspot.com',
        'namespace': 'default-gzip',
        'isolated': 'eebfbfe99b391ad869871868369602e561ab9b52'
    },
    'duration': 50.1,
    'completed_ts': '2017-10-24T22:20:32.361620',
    'started_ts': '2017-10-24T22:19:20.104240',
    'created_ts': '2017-10-24T22:19:17.794970',
    'state': 'COMPLETED'
}

MOCK_LOG_DATA_COMPLETE = {
    'disabled_tests': [
        'AdbClientSocketTest.TestFlushWithData',
        'AdbClientSocketTest.TestFlushWithSize',
        'AdbClientSocketTest.TestFlushWithoutSize'
    ],
    'global_tags': ['CPU_64_BITS', 'MODE_RELEASE', 'OS_MACOSX', 'OS_POSIX'],
    'all_tests': [
        'AcceleratorsCocoaBrowserTest.MainMenuAcceleratorsInMapping',
        'AcceleratorsCocoaBrowserTest.MappingAcceleratorsInMainMenu',
        'ActivityLogApiTest.TriggerEvent',
        'ActivityLogPrerenderTest.TestScriptInjected',
        'AdbClientSocketTest.TestFlushWithData'
    ],
    'per_iteration_data': [{
        'TestSuite1.test1': [{
            'status': 'SUCCESS',
            'other_info': 'N/A'
        }],
        'TestSuite1.test2': [{
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'SUCCESS',
            'other_info': 'N/A'
        }],
        'TestSuite1.test3': [{
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }]
    }, {
        'TestSuite1.test1': [{
            'status': 'SUCCESS',
            'other_info': 'N/A'
        }],
        'TestSuite1.test2': [{
            'status': 'SUCCESS',
            'other_info': 'N/A'
        }],
        'TestSuite1.test3': [{
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }, {
            'status': 'FAILURE',
            'other_info': 'N/A'
        }]
    }],
    'test_locations': {
        'PolicyToolUITest.Editing': {
            'line':
                271,
            'file':
                '../../chrome/browser/ui/webui/policy_tool_ui_browsertest.cc'
        }
    },
    'error':
        None,
    'step_name_no_platform':
        'browser_tests',
    'task_state':
        'COMPLETED'
}


class SwarmingTaskQueueTest(wf_testcase.WaterfallTestCase):

  def testGetSwarmingTaskError(self):
    # There's an error passed by the swarming results.
    expected_error = None
    result = process_swarming_results.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_COMPLETE, None, MOCK_LOG_DATA_COMPLETE, None)
    self.assertEqual(expected_error, result)

  def testGetSwarmingTaskErrorWithNoResultData(self):
    # There's an error passed by the swarming results.
    expected_error = {
        'code': swarming_util.NO_OUTPUT_JSON,
        'message': 'No data was found even though task was completed.'
    }
    result = process_swarming_results.GetSwarmingTaskError(
        1, None, None, None, None)
    self.assertEqual(expected_error, result)

  def testGetSwarmingTaskErrorWithResultError(self):
    # There's an error passed by the swarming results.
    error = {'code': '-1', 'message': 'msg'}
    result = process_swarming_results.GetSwarmingTaskError(
        1, None, error, None, None)
    self.assertEqual(error, result)

  def testGetSwarmingTaskErrorWithMissingOutputRefs(self):
    # There's an error passed by the swarming results.
    expected_error = {
        'code': swarming_util.NO_TASK_OUTPUTS,
        'message': 'outputs_ref is None'
    }
    result = process_swarming_results.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_NO_OUTPUTS_REF, None, None, None)
    self.assertEqual(expected_error, result)

  def testGetSwarmingTaskErrorWithoutLogData(self):
    # There's an error passed by the swarming results.
    expected_error = {
        'code': swarming_util.NO_OUTPUT_JSON,
        'message': 'output_json is None',
    }
    result = process_swarming_results.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_COMPLETE, None, None, None)
    self.assertEqual(expected_error, result)

  def testGetSwarmingTaskErrorWithLogError(self):
    # There's an error passed by the swarming results.
    error = {'code': '-1', 'message': 'msg'}
    result = process_swarming_results.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_COMPLETE, None, MOCK_LOG_DATA_COMPLETE, error)
    self.assertEqual(error, result)

  def testGetSwarmingTaskErrorNoPerIterationData(self):
    # There's an error passed by the swarming results.
    expected_error = {
        'code': swarming_util.NO_PER_ITERATION_DATA,
        'message': 'per_iteration_data is empty or missing'
    }
    result = process_swarming_results.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_COMPLETE, None, MOCK_LOG_DATA_NO_PER_ITERATION_DATA,
        None)
    self.assertEqual(expected_error, result)

  def testIsTaskCompleteWhenComplete(self):
    data = {'state': swarming_util.STATE_COMPLETED}
    self.assertTrue(process_swarming_results.IsTaskComplete(data))

  def testIsTaskCompleteWhenIncomplete(self):
    data = {'state': swarming_util.STATES_RUNNING}
    self.assertFalse(process_swarming_results.IsTaskComplete(data))

  def testCheckTestsRunStatusesNoOutPutJson(self):
    self.assertEqual({},
                     process_swarming_results.CheckTestsRunStatuses(None, ()))

  def testCheckTestsRunStatuses(self):
    tests_statuses = (
        process_swarming_results.CheckTestsRunStatuses(_SAMPLE_FAILURE_LOG))
    self.assertEqual(_EXPECTED_TESTS_STATUS, tests_statuses)
