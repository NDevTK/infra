# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import json
import time
import logging
import datetime
from collections import defaultdict
from protorpc import messages

from google.appengine.ext import ndb

from common import constants
from libs import time_util
from libs import analysis_status
from model import swarming_task_queue_request
from model import base_swarming_task
from model.flake import master_flake_analysis
from model.flake.flake_swarming_task import FlakeSwarmingTask
from waterfall import swarming_util


class TaskPriorityQueueItem(object):
  """Represents one item in a TaskPriorityQueue, stores priority and payload."""

  def __init__(self, item, priority):
    self._item = item
    self._priority = priority

  @property
  def item(self):
    return self._item

  @property
  def priority(self):
    return self._priority

  @priority.setter
  def priority(self, value):
    assert value > 0
    self._priority = value


class TaskPriorityQueue(object):
  """Stores multiple TaskPriorityQueue items and orders them by priority."""

  def __init__(self):
    self._queue = []

  def Enqueue(self, item, priority, request_time=None):
    """Enqueues the given item at the given priority."""
    if not request_time:
      request_time = time_util.GetUTCNow()

    self._queue.append(TaskPriorityQueueItem(item, priority))
    self._queue.sort(
        key=lambda task: (task.priority, request_time), reverse=True)

  def Dequeue(self):
    """Dequeues an element from the queue."""
    assert self._queue

    self._queue = [
        # Bump the priority up so lower priority tasks aren't starved.
        TaskPriorityQueueItem(item.item, item.priority * 2)
        for item in self._queue
    ]

    return self._queue.pop(0).item

  def Peek(self):
    """Peeks an element from the queue."""
    assert self._queue
    return self._queue[0].item

  def IsEmpty(self):
    return len(self._queue) == 0


class TaskPriorityMultiQueue(object):
  """Manages key <--> TaskPriorityQueues associations."""

  def __init__(self):
    self._queues = defaultdict(TaskPriorityQueue)
    self._dequeue_ptr = 0

  def Enqueue(self, key, item, priority, request_time=None):
    """Enqueue the given item under the queue that corresponds to key.

    Enqueues the item with the given priority for the queue that corresponds
    to the given key. If no such queue exists, one will be created.

    Args:
      key (str): Key uniquely identifies the queue.
      item (SwarmingTaskQueueRequest): Items to be enqueued.
      priority (SwarmingTaskQueuePriority): Priority of the enqueued item.
    """
    if not request_time:
      request_time = time_util.GetUTCNow()

    if not key:
      raise ValueError('Enqueue key must be not None')

    if not item:
      raise ValueError('Enqueue item must be not None')

    if not priority:
      raise ValueError('Enqueue priority must be not None')

    self._queues[key].Enqueue(item, priority, request_time)

  def Dequeue(self):
    """Dequeues an item from the queue.

    Returns:
      (SwarmingTaskQueueRequest) Item that was dequeued.
    Raises:
      (IndexError) Raised if you dequeue from an empty queue.
    """
    keys = sorted(self._queues.keys())
    if not keys:
      raise IndexError('Dequeue was called on an empty queue')

    # Pick a key from the dequeue_ptr.
    current_key = keys[self._dequeue_ptr % len(keys)]
    # Increment the dequeue_ptr so we dequeue from the next queue.
    self._dequeue_ptr += 1

    val = self._queues[current_key].Dequeue()
    if self._queues[current_key].IsEmpty():
      del self._queues[current_key]

    return val

  def Peek(self):
    """Returns an element without dequeuing it.

    Returns:
      (SwarmingTaskQueueRequest) Item that would be dequeued.
    Raises:
      (IndexError) Raised if you peek an empty queue.
    """
    keys = sorted(self._queues.keys())
    if not keys:
      raise IndexError('Peek was called on an empty queue')

    # Pick a key from the dequeue_ptr.
    current_key = keys[self._dequeue_ptr % len(keys)]

    return self._queues[current_key].Peek()

  def Continue(self):
    """Continues through the multiqueue without dequeuing anything."""
    self._dequeue_ptr += 1

  def IsEmpty(self):
    """Returns True if empty, False otherwise."""
    return not bool(self._queues.keys())


def _InitializeTaskQueue():
  """Fetch and fill queues with existing data."""

  def _FetchAndFillList(query, lst):
    """ Fetch everything from the query and fill the given list with it.

    This function mutates the given list rather than returning a new one.

    Args:
      query (query): SwarmingTaskRequest query that contains all
          the SCHEDULED tasks.
      lst (list): List to populate with query data.
    """
    more = True
    cursor = None
    while more:
      results, cursor, more = query.fetch_page(100, start_cursor=cursor)
      lst.extend(results)

  def _FetchAndFillDimensionMultiqueue(query, queue):
    more = True
    cursor = None
    while more:
      results, cursor, more = query.fetch_page(100, start_cursor=cursor)
      for result in results:
        queue.Enqueue(
            json.dumps(result.taskqueue_dimensions),
            result,
            result.taskqueue_priority,
            request_time=result.taskqueue_request_time)

  scheduled_tasks = TaskPriorityMultiQueue()
  pending_tasks = []

  query = swarming_task_queue_request.SwarmingTaskQueueRequest.query(
      swarming_task_queue_request.SwarmingTaskQueueRequest.taskqueue_state ==
      swarming_task_queue_request.SwarmingTaskQueueState.SCHEDULED)
  _FetchAndFillDimensionMultiqueue(query, scheduled_tasks)

  query = swarming_task_queue_request.SwarmingTaskQueueRequest.query(
      swarming_task_queue_request.SwarmingTaskQueueRequest.taskqueue_state ==
      swarming_task_queue_request.SwarmingTaskQueueState.PENDING)
  _FetchAndFillList(query, pending_tasks)

  return scheduled_tasks, pending_tasks


def EnqueueTask(request, priority):
  """Enqueue a SwarmingTaskRequest to the TaskQueue.

  Args:
    request (SwarmingTaskQueueRequest): The request to this taskqueue.
    priority (SwarmingTaskQueuePriorty): What priority this request is.
  """

  request.taskqueue_state = (
      swarming_task_queue_request.SwarmingTaskQueueState.SCHEDULED)
  request.taskqueue_priority = priority
  request.taskqueue_request_time = time_util.GetUTCNow()
  request.put()


# TODO(crbug.com/776440): Dispatch requests to swarming.

# TODO(crbug.com/775192): Get results and send the results to pipeline.

# TODO(crbug.com/780066): Cull or retry neglected tasks.