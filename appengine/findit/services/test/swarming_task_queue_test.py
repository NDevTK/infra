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
from services import swarming_task_queue
from waterfall import swarming_util
from waterfall.test import wf_testcase

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

MOCK_RESULT_DATA_RUNNING = {
    'task_id': 'task_id',
    'name': 'name',
    'tags': ['cpu:x86-64', 'findit:1'],
    'started_ts': '2017-10-24T22:19:20.104240',
    'created_ts': '2017-10-24T22:19:17.794970',
    'state': 'RUNNING'
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


class SwarmingTaskQueueTest(wf_testcase.WaterfallTestCase):
  app_module = pipeline_handlers._APP

  def testTaskPriorityQueueItem(self):
    item = swarming_task_queue.TaskPriorityQueueItem('1', 100)
    self.assertEqual(item.item, '1')
    self.assertEqual(item.priority, 100)

    item.priority = 200
    self.assertEqual(item.priority, 200)

    with self.assertRaises(AttributeError):
      item.item = '2'

  def testTaskPriorityQueueWhenEmpty(self):
    queue = swarming_task_queue.TaskPriorityQueue()

    with self.assertRaises(AssertionError):
      queue.Dequeue()

  def testTaskPriorityQueueDequeuesCorrectly(self):
    queue = swarming_task_queue.TaskPriorityQueue()

    queue.Enqueue('1_1', 100)
    queue.Enqueue('2_1', 200, request_time=datetime.datetime(2017, 1, 1))
    queue.Enqueue('2_2', 200, request_time=datetime.datetime(2017, 1, 2))
    queue.Enqueue('3_1', 300, request_time=datetime.datetime(2017, 1, 1))
    queue.Enqueue('3_2', 300, request_time=datetime.datetime(2017, 1, 2))
    queue.Enqueue('3_3', 300, request_time=datetime.datetime(2017, 1, 3))

    self.assertEqual(queue.Dequeue(), '3_1')
    self.assertEqual(queue.Dequeue(), '3_2')
    self.assertEqual(queue.Dequeue(), '3_3')

    self.assertEqual(queue.Dequeue(), '2_1')
    self.assertEqual(queue.Dequeue(), '2_2')

    self.assertEqual(queue.Dequeue(), '1_1')

  def testTaskPriorityQueueDequeuesCorrectlyWhenEnqueued(self):
    queue = swarming_task_queue.TaskPriorityQueue()

    queue.Enqueue('1_1', 100, request_time=datetime.datetime(2017, 1, 1))
    queue.Enqueue('2_1', 200)

    self.assertEqual(queue.Dequeue(), '2_1')

    queue.Enqueue('2_2', 200, request_time=datetime.datetime(2017, 1, 2))

    self.assertEqual(queue.Dequeue(), '1_1')
    self.assertEqual(queue.Dequeue(), '2_2')

  def testTaskPriorityMultiQueueWhenEmpty(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()

    with self.assertRaises(AssertionError):
      multiqueue.Dequeue()

  def testTaskPriorityMultiQueueWithInvalidEnqueueParameters(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()

    with self.assertRaises(AssertionError):
      multiqueue.Enqueue(None, 'a', 100)

    with self.assertRaises(AssertionError):
      multiqueue.Enqueue('key', None, 100)

    with self.assertRaises(AssertionError):
      multiqueue.Enqueue('key', 'a', None)

    with self.assertRaises(AssertionError):
      multiqueue.Enqueue(None, 'a', 100)

  def testTaskPriorityMultiQueueEnqueue(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()

    multiqueue.Enqueue('key_1', 'item_1', 100)

    self.assertTrue('key_1' in multiqueue._queues)

  def testTaskPriorityMultiQueueEnqueueDequeue(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()

    multiqueue.Enqueue('key_1', 'item_1_1', 100)
    multiqueue.Enqueue('key_1', 'item_1_2', 200)
    multiqueue.Enqueue('key_1', 'item_1_3', 300)
    multiqueue.Enqueue(
        'key_2', 'item_2_1', 200, request_time=datetime.datetime(2017, 1, 1))
    multiqueue.Enqueue(
        'key_2', 'item_2_2', 200, request_time=datetime.datetime(2017, 1, 2))
    multiqueue.Enqueue(
        'key_2', 'item_2_3', 200, request_time=datetime.datetime(2017, 1, 3))

    self.assertEqual(multiqueue.Dequeue(), 'item_1_3')
    self.assertEqual(multiqueue.Dequeue(), 'item_2_1')
    self.assertEqual(multiqueue.Dequeue(), 'item_1_2')
    self.assertEqual(multiqueue.Dequeue(), 'item_2_2')
    self.assertEqual(multiqueue.Dequeue(), 'item_1_1')
    self.assertEqual(multiqueue.Dequeue(), 'item_2_3')

  def testInitializeTaskQueue(self):
    task = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=(
            swarming_task_queue_request.SwarmingTaskQueueState.SCHEDULED),
        taskqueue_dimensions="{'dim1: 'dim1'}")
    task.taskqueue_priority = (
        swarming_task_queue_request.SwarmingTaskQueuePriority.FAILURE)
    task.taskqueue_request_time = time_util.GetUTCNow()
    task.put()

    task = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=(
            swarming_task_queue_request.SwarmingTaskQueueState.SCHEDULED),
        taskqueue_dimensions="{'dim2': 'dim2'}")
    task.taskqueue_priority = (
        swarming_task_queue_request.SwarmingTaskQueuePriority.FAILURE)
    task.taskqueue_request_time = time_util.GetUTCNow()
    task.put()

    task = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        PENDING)
    task.taskqueue_priority = (
        swarming_task_queue_request.SwarmingTaskQueuePriority.FAILURE)
    task.taskqueue_request_time = time_util.GetUTCNow()
    task.put()

    task = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=(
            swarming_task_queue_request.SwarmingTaskQueueState.COMPLETED))
    task.taskqueue_priority = (
        swarming_task_queue_request.SwarmingTaskQueuePriority.FAILURE)
    task.taskqueue_request_time = time_util.GetUTCNow()
    task.put()

    task = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.READY
    )
    task.taskqueue_priority = (
        swarming_task_queue_request.SwarmingTaskQueuePriority.FAILURE)
    task.taskqueue_request_time = time_util.GetUTCNow()
    task.put()

    swarming_task_queue._InitializeTaskQueue()
    self.assertEqual(len(swarming_task_queue.scheduled_tasks._queues.keys()), 2)
    self.assertEqual(len(swarming_task_queue.pending_tasks), 1)
    self.assertEqual(len(swarming_task_queue.completed_tasks), 1)
    self.assertEqual(len(swarming_task_queue.ready_tasks), 1)

  @mock.patch.object(swarming_task_queue, 'Update')
  def testTaskQueueEnqueue(self, mock_update):
    swarming_task_queue._TASKQUEUE_UPDATE_DELAY_SECONDS = 0

    def MockUpdate():
      swarming_task_queue._InitializeTaskQueue()

    mock_update.side_effect = MockUpdate

    task1 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        SCHEDULED,
        taskqueue_dimensions="{'dim1': 'dim1'}")

    task2 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        SCHEDULED,
        taskqueue_dimensions="{'dim2': 'dim2'}")

    task3 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        PENDING,
        taskqueue_dimensions="{'dim3': 'dim3'}")

    task4 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        COMPLETED,
        taskqueue_dimensions="{'dim3': 'dim3'}")

    task5 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        READY,
        taskqueue_dimensions="{'dim3': 'dim3'}")

    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 1)):
      swarming_task_queue.EnqueueTask(
          task1, swarming_task_queue_request.SwarmingTaskQueuePriority.FAILURE)
      MockUpdate()
    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 2)):
      swarming_task_queue.EnqueueTask(
          task2, swarming_task_queue_request.SwarmingTaskQueuePriority.FLAKE)
      MockUpdate()
    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 3)):
      swarming_task_queue.EnqueueTask(
          task3, swarming_task_queue_request.SwarmingTaskQueuePriority.FORCE)
      MockUpdate()
    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 4)):
      swarming_task_queue.EnqueueTask(
          task4, swarming_task_queue_request.SwarmingTaskQueuePriority.API_CALL)
      MockUpdate()
    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 5)):
      swarming_task_queue.EnqueueTask(
          task5, swarming_task_queue_request.SwarmingTaskQueuePriority.FORCE)
      MockUpdate()

    self.assertEqual(swarming_task_queue.scheduled_tasks.Dequeue(), task1)
    self.assertEqual(swarming_task_queue.scheduled_tasks.Dequeue(), task3)
    self.assertEqual(swarming_task_queue.scheduled_tasks.Dequeue(), task2)
    self.assertEqual(swarming_task_queue.scheduled_tasks.Dequeue(), task5)
    self.assertEqual(swarming_task_queue.scheduled_tasks.Dequeue(), task4)

  def testTaskQueueEnqueueAlreadyInitialized(self):
    swarming_task_queue._InitializeTaskQueue()

    task = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        SCHEDULED,
        taskqueue_dimensions="{'dim1': 'dim1'}")
    task.taskqueue_request_time = datetime.datetime(2017, 1, 1)
    swarming_task_queue.EnqueueTask(
        task, swarming_task_queue_request.SwarmingTaskQueuePriority.FAILURE)

  def testGetSwarmingTaskError(self):
    # There's an error passed by the swarming results.
    expected_error = None
    result = swarming_task_queue.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_COMPLETE, None, MOCK_LOG_DATA_COMPLETE, None)
    self.assertEqual(expected_error, result)

  def testGetSwarmingTaskErrorWithNoResultData(self):
    # There's an error passed by the swarming results.
    expected_error = {
        'code': swarming_util.NO_OUTPUT_JSON,
        'message': 'No data was found even though task was completed.'
    }
    result = swarming_task_queue.GetSwarmingTaskError(1, None, None, None, None)
    self.assertEqual(expected_error, result)

  def testGetSwarmingTaskErrorWithResultError(self):
    # There's an error passed by the swarming results.
    error = {'code': '-1', 'message': 'msg'}
    result = swarming_task_queue.GetSwarmingTaskError(1, None, error, None,
                                                      None)
    self.assertEqual(error, result)

  def testGetSwarmingTaskErrorWithMissingOutputRefs(self):
    # There's an error passed by the swarming results.
    expected_error = {
        'code': swarming_util.NO_TASK_OUTPUTS,
        'message': 'outputs_ref is None'
    }
    result = swarming_task_queue.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_NO_OUTPUTS_REF, None, None, None)
    self.assertEqual(expected_error, result)

  def testGetSwarmingTaskErrorWithoutLogData(self):
    # There's an error passed by the swarming results.
    expected_error = {
        'code': swarming_util.NO_OUTPUT_JSON,
        'message': 'output_json is None',
    }
    result = swarming_task_queue.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_COMPLETE, None, None, None)
    self.assertEqual(expected_error, result)

  def testGetSwarmingTaskErrorWithLogError(self):
    # There's an error passed by the swarming results.
    error = {'code': '-1', 'message': 'msg'}
    result = swarming_task_queue.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_COMPLETE, None, MOCK_LOG_DATA_COMPLETE, error)
    self.assertEqual(error, result)

  def testGetSwarmingTaskErrorNoPerIterationData(self):
    # There's an error passed by the swarming results.
    expected_error = {
        'code': swarming_util.NO_PER_ITERATION_DATA,
        'message': 'per_iteration_data is empty or missing'
    }
    result = swarming_task_queue.GetSwarmingTaskError(
        1, MOCK_RESULT_DATA_COMPLETE, None, MOCK_LOG_DATA_NO_PER_ITERATION_DATA,
        None)
    self.assertEqual(expected_error, result)

  def testIsTaskCompleteWhenComplete(self):
    data = {'state': swarming_util.STATE_COMPLETED}
    self.assertTrue(swarming_task_queue.IsTaskComplete(data))

  def testIsTaskCompleteWhenIncomplete(self):
    data = {'state': swarming_util.STATES_RUNNING}
    self.assertFalse(swarming_task_queue.IsTaskComplete(data))

  def testConvertDateTime(self):
    fmt = '%Y-%m-%dT%H:%M:%S.%f'
    time_string = '2016-02-10T18:32:06.538220'
    test_time = swarming_task_queue._ConvertDateTime(time_string)
    time = datetime.datetime.strptime(time_string, fmt)
    self.assertEqual(test_time, time)

  def testConvertDateTimeNone(self):
    time_string = ''
    test_time = swarming_task_queue._ConvertDateTime(time_string)
    self.assertIsNone(test_time)

  def testConvertDateTimefailure(self):
    with self.assertRaises(ValueError):
      swarming_task_queue._ConvertDateTime('abc')

  def testCheckTestsRunStatusesNoOutPutJson(self):
    self.assertEqual({}, swarming_task_queue._CheckTestsRunStatuses(None, ()))

  def testCheckTestsRunStatuses(self):
    tests_statuses = (
        swarming_task_queue._CheckTestsRunStatuses(_SAMPLE_FAILURE_LOG))
    self.assertEqual(_EXPECTED_TESTS_STATUS, tests_statuses)

  def testGetTaskRequest(self):
    task_id = 'task_id'
    task = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        PENDING,
        taskqueue_dimensions="{'dim3': 'dim3'}",
        swarming_task_id=task_id)
    task.put()
    swarming_task_queue._InitializeTaskQueue()
    self.assertEqual(task, swarming_task_queue.GetTaskRequest(task_id))

  def testGetTaskRequestNoneFound(self):
    task_id = 'task_id'
    task = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        PENDING,
        taskqueue_dimensions="{'dim3': 'dim3'}",
        swarming_task_id='not my task id')
    task.put()
    swarming_task_queue._InitializeTaskQueue()
    self.assertEqual(None, swarming_task_queue.GetTaskRequest(task_id))

  @mock.patch.object(swarming_util, 'GetSwarmingTaskFailureLog')
  @mock.patch.object(swarming_util, 'GetSwarmingTaskResultById')
  @mock.patch.object(swarming_task_queue, 'GetTaskRequest')
  def testCompleteTask(self, get_task_fn, result_fn, log_fn):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    task_id = 'task_id'

    result_fn.return_value = (MOCK_RESULT_DATA_COMPLETE, None)
    log_fn.return_value = (MOCK_LOG_DATA_COMPLETE, None)

    analysis = master_flake_analysis.MasterFlakeAnalysis.Create(
        master_name, builder_name, build_number, step_name, test_name)
    analysis.put()

    task = flake_swarming_task.FlakeSwarmingTask.Create(
        master_name, builder_name, build_number, step_name, test_name)
    task.put()

    task_queue_request = (
        swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
            taskqueue_state=(
                swarming_task_queue_request.SwarmingTaskQueueState.SCHEDULED),
            taskqueue_priority=(
                swarming_task_queue_request.SwarmingTaskQueuePriority.FORCE),
            taskqueue_analysis_urlsafe_key=analysis.key.urlsafe(),
            swarming_task_id=task_id))
    task_queue_request.put()

    get_task_fn.return_value = task_queue_request

    swarming_task_queue.CompleteTask(task_id)

  @mock.patch.object(swarming_util, 'GetSwarmingTaskFailureLog')
  @mock.patch.object(swarming_util, 'GetSwarmingTaskResultById')
  @mock.patch.object(swarming_task_queue, 'GetTaskRequest')
  @mock.patch.object(swarming_task_queue, 'GetSwarmingTaskError')
  def testCompleteTaskWithError(self, error_fn, get_task_fn, result_fn, log_fn):
    master_name = 'm'
    builder_name = 'b'
    build_number = 100
    step_name = 's'
    test_name = 't'
    task_id = 'task_id'

    result_fn.return_value = (MOCK_RESULT_DATA_COMPLETE, None)
    log_fn.return_value = (MOCK_LOG_DATA_COMPLETE, None)
    error_fn.return_value = {'error': 'error!'}

    analysis = master_flake_analysis.MasterFlakeAnalysis.Create(
        master_name, builder_name, build_number, step_name, test_name)
    analysis.put()

    task = flake_swarming_task.FlakeSwarmingTask.Create(
        master_name, builder_name, build_number, step_name, test_name)
    task.put()

    task_queue_request = (
        swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
            taskqueue_state=(
                swarming_task_queue_request.SwarmingTaskQueueState.SCHEDULED),
            taskqueue_priority=(
                swarming_task_queue_request.SwarmingTaskQueuePriority.FORCE),
            taskqueue_analysis_urlsafe_key=analysis.key.urlsafe(),
            swarming_task_id=task_id))
    task_queue_request.put()

    get_task_fn.return_value = task_queue_request

    swarming_task_queue.CompleteTask(task_id)

  @mock.patch.object(swarming_task_queue, 'IsTaskComplete', return_value=False)
  @mock.patch.object(swarming_util, 'GetSwarmingTaskFailureLog')
  @mock.patch.object(swarming_util, 'GetSwarmingTaskResultById')
  def testCompleteTaskIncomplete(self, result_fn, log_fn, _):
    task_id = 'task_id'

    result_fn.return_value = (MOCK_RESULT_DATA_COMPLETE, None)
    log_fn.return_value = (MOCK_LOG_DATA_COMPLETE, None)

    swarming_task_queue.CompleteTask(task_id)