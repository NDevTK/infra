# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import mock
import datetime

from gae_libs.pipeline_wrapper import pipeline_handlers
from libs import time_util

from model import swarming_task_queue_request
from services import swarming_task_queue
from waterfall.test import wf_testcase


class SwarmingTaskQueueTest(wf_testcase.WaterfallTestCase):
  app_module = pipeline_handlers._APP

  def testTaskPriorityQueueItem(self):
    item = swarming_task_queue.TaskPriorityQueueItem('1', 100,
                                                     datetime.datetime(
                                                         2017, 1, 1))
    self.assertEqual(item.item, '1')
    self.assertEqual(item.priority, 100)
    self.assertEqual(item.request_time, datetime.datetime(2017, 1, 1))

    item.priority = 200
    self.assertEqual(item.priority, 200)

    with self.assertRaises(AttributeError):
      item.item = '2'

  def testTaskPriorityQueueItemEquals(self):
    item = swarming_task_queue.TaskPriorityQueueItem('1', 100,
                                                     datetime.datetime(
                                                         2017, 1, 1))
    self.assertTrue(
        item.__eq__(
            swarming_task_queue.TaskPriorityQueueItem(
                '1', 100, datetime.datetime(2017, 1, 1))))
    self.assertFalse(
        item.__eq__(
            swarming_task_queue.TaskPriorityQueueItem(
                '2', 100, datetime.datetime(2017, 1, 1))))
    self.assertFalse(item.__eq__({}))

  def testTaskPriorityQueueWhenEmpty(self):
    queue = swarming_task_queue.TaskPriorityQueue()

    with self.assertRaises(AssertionError):
      queue.Dequeue()

  def testTaskPriorityQueueDequeuesCorrectly(self):
    queue = swarming_task_queue.TaskPriorityQueue()

    queue.Enqueue('1_1', 3)
    queue.Enqueue('2_1', 2, request_time=datetime.datetime(2017, 1, 1))
    queue.Enqueue('2_2', 2, request_time=datetime.datetime(2017, 1, 2))
    queue.Enqueue('3_1', 1, request_time=datetime.datetime(2017, 1, 1))
    queue.Enqueue('3_2', 1, request_time=datetime.datetime(2017, 1, 2))
    queue.Enqueue('3_3', 1, request_time=datetime.datetime(2017, 1, 3))

    self.assertEqual(queue.Dequeue(), '3_1')
    self.assertEqual(queue.Dequeue(), '3_2')
    self.assertEqual(queue.Dequeue(), '3_3')

    self.assertEqual(queue.Dequeue(), '2_1')
    self.assertEqual(queue.Dequeue(), '2_2')

    self.assertEqual(queue.Dequeue(), '1_1')

  def testTaskPriorityQueueDequeuesCorrectlyWhenEnqueued(self):
    queue = swarming_task_queue.TaskPriorityQueue()

    queue.Enqueue('1_1', 2, request_time=datetime.datetime(2017, 1, 1))
    queue.Enqueue('2_1', 1)

    self.assertEqual(queue.Dequeue(), '2_1')

    queue.Enqueue('1_2', 2, request_time=datetime.datetime(2017, 1, 2))

    self.assertEqual(queue.Dequeue(), '1_1')
    self.assertEqual(queue.Dequeue(), '1_2')

  def testTaskPriorityQueuePeek(self):
    queue = swarming_task_queue.TaskPriorityQueue()
    queue.Enqueue('1_1', 1)
    self.assertEqual('1_1', queue.Peek())

  def testTaskPriorityQueuePeekWhenEmpty(self):
    queue = swarming_task_queue.TaskPriorityQueue()
    with self.assertRaises(AssertionError):
      queue.Peek()

  def testTaskPriorityQueueDumpItems(self):
    queue = swarming_task_queue.TaskPriorityQueue()

    queue.Enqueue('1', 1, request_time=datetime.datetime(2017, 1, 1))
    queue.Enqueue('2', 2, request_time=datetime.datetime(2017, 1, 2))

    item_1 = swarming_task_queue.TaskPriorityQueueItem('1', 1,
                                                       datetime.datetime(
                                                           2017, 1, 1))
    item_2 = swarming_task_queue.TaskPriorityQueueItem('2', 2,
                                                       datetime.datetime(
                                                           2017, 1, 2))
    items = queue.DumpItems()
    self.assertTrue(item_1 in items)
    self.assertTrue(item_2 in items)

  def testTaskPriorityMultiQueueDequeueWhenEmpty(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()

    with self.assertRaises(IndexError):
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

    multiqueue.Enqueue('key_1', 'item_1_1', 3)
    multiqueue.Enqueue('key_1', 'item_1_2', 2)
    multiqueue.Enqueue('key_1', 'item_1_3', 1)
    multiqueue.Enqueue(
        'key_2', 'item_2_1', 2, request_time=datetime.datetime(2017, 1, 1))
    multiqueue.Enqueue(
        'key_2', 'item_2_2', 2, request_time=datetime.datetime(2017, 1, 2))
    multiqueue.Enqueue(
        'key_2', 'item_2_3', 2, request_time=datetime.datetime(2017, 1, 3))

    self.assertEqual(multiqueue.Dequeue(), 'item_1_3')
    self.assertEqual(multiqueue.Dequeue(), 'item_2_1')
    self.assertEqual(multiqueue.Dequeue(), 'item_1_2')
    self.assertEqual(multiqueue.Dequeue(), 'item_2_2')
    self.assertEqual(multiqueue.Dequeue(), 'item_1_1')
    self.assertEqual(multiqueue.Dequeue(), 'item_2_3')

  def testTaskPriorityMultiQueuePeek(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()
    multiqueue.Enqueue('key_1', 'item_1_1', 1)
    self.assertEqual('item_1_1', multiqueue.Peek())

  def testTaskPriorityMultiQueuePeekWhenEmpty(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()
    with self.assertRaises(IndexError):
      multiqueue.Peek()

  def testTaskPriorityMultiQueueContinue(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()
    self.assertEqual(0, multiqueue._dequeue_ptr)
    multiqueue.Continue()
    self.assertEqual(1, multiqueue._dequeue_ptr)

  def testTaskPriorityMultiQueueIsEmptyWhenEmpty(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()
    self.assertTrue(multiqueue.IsEmpty())

  def testTaskPriorityMultiQueueIsEmptyWhenFull(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()
    multiqueue.Enqueue('key_1', 'item_1_1', 100)
    multiqueue.Enqueue('key_1', 'item_1_2', 200)
    multiqueue.Enqueue('key_1', 'item_1_3', 300)
    self.assertFalse(multiqueue.IsEmpty())

  def testTaskPriorityMultiQueueDumpItems(self):
    multiqueue = swarming_task_queue.TaskPriorityMultiQueue()
    multiqueue.Enqueue(
        'key_1', '1', 1, request_time=datetime.datetime(2017, 1, 1))
    multiqueue.Enqueue(
        'key_2', '2', 2, request_time=datetime.datetime(2017, 1, 2))
    multiqueue.Enqueue(
        'key_3', '3', 3, request_time=datetime.datetime(2017, 1, 3))

    item_1 = swarming_task_queue.TaskPriorityQueueItem('1', 1,
                                                       datetime.datetime(
                                                           2017, 1, 1))
    item_2 = swarming_task_queue.TaskPriorityQueueItem('2', 2,
                                                       datetime.datetime(
                                                           2017, 1, 2))
    item_3 = swarming_task_queue.TaskPriorityQueueItem('3', 3,
                                                       datetime.datetime(
                                                           2017, 1, 3))
    items = multiqueue.DumpItems()
    self.assertTrue(item_1 in items)
    self.assertTrue(item_2 in items)
    self.assertTrue(item_3 in items)

  def testGetTasks(self):
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
        taskqueue_state=swarming_task_queue_request.SwarmingTaskQueueState.
        COMPLETED)
    task.taskqueue_priority = (
        swarming_task_queue_request.SwarmingTaskQueuePriority.FAILURE)
    task.taskqueue_request_time = time_util.GetUTCNow()
    task.put()

    scheduled_tasks, pending_tasks = swarming_task_queue._FetchTasks()
    self.assertEqual(len(scheduled_tasks._queues.keys()), 2)
    self.assertEqual(len(pending_tasks), 1)

  def testTaskQueueEnqueue(self):
    task1 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_priority=(
            swarming_task_queue_request.SwarmingTaskQueuePriority.FAILURE),
        taskqueue_dimensions="{'dim1': 'dim1'}")

    task2 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_priority=(
            swarming_task_queue_request.SwarmingTaskQueuePriority.FLAKE),
        taskqueue_dimensions="{'dim2': 'dim2'}")

    task3 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_priority=(
            swarming_task_queue_request.SwarmingTaskQueuePriority.FORCE),
        taskqueue_dimensions="{'dim3': 'dim3'}")

    task4 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_priority=(
            swarming_task_queue_request.SwarmingTaskQueuePriority.API_CALL),
        taskqueue_dimensions="{'dim3': 'dim3'}")

    task5 = swarming_task_queue_request.SwarmingTaskQueueRequest.Create(
        taskqueue_priority=(
            swarming_task_queue_request.SwarmingTaskQueuePriority.FORCE),
        taskqueue_dimensions="{'dim3': 'dim3'}")

    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 1)):
      swarming_task_queue.EnqueueTask(task1)

    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 2)):
      swarming_task_queue.EnqueueTask(task2)

    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 3)):
      swarming_task_queue.EnqueueTask(task3)

    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 4)):
      swarming_task_queue.EnqueueTask(task4)

    with mock.patch.object(
        time_util, 'GetUTCNow', return_value=datetime.datetime(2017, 1, 5)):
      swarming_task_queue.EnqueueTask(task5)

    scheduled_tasks, _ = swarming_task_queue._FetchTasks()
    self.assertEqual(scheduled_tasks.Dequeue(), task1)
    self.assertEqual(scheduled_tasks.Dequeue(), task3)
    self.assertEqual(scheduled_tasks.Dequeue(), task2)
    self.assertEqual(scheduled_tasks.Dequeue(), task5)
    self.assertEqual(scheduled_tasks.Dequeue(), task4)