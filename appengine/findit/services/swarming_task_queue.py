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
from common.findit_http_client import FinditHttpClient
from libs import time_util
from libs import analysis_status
from model import swarming_task_queue_request
from model import base_swarming_task
from model.flake import master_flake_analysis
from model.flake.flake_swarming_task import FlakeSwarmingTask
from services import datetime_util
from services import process_swarming_results
from waterfall import swarming_util

_TASKQUEUE_UPDATE_DELAY_SECONDS = 120

_HTTP_CLIENT = FinditHttpClient()


class TaskPriorityQueueItem(object):

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

  def __init__(self):
    self._queue = []

  def Enqueue(self, item, priority, request_time=time_util.GetUTCNow()):
    self._queue.append(TaskPriorityQueueItem(item, priority))
    self._queue.sort(
        key=lambda task: (task.priority, request_time), reverse=True)

  def Dequeue(self):
    assert self._queue

    self._queue = [
        TaskPriorityQueueItem(item.item, item.priority * 2)
        for item in self._queue
    ]

    return self._queue.pop(0).item

  def IsEmpty(self):
    return len(self._queue) == 0


class TaskPriorityMultiQueue(object):

  def __init__(self):
    self._queues = {}
    self._dequeue_ptr = 0

  def Enqueue(self, key, item, priority, request_time=time_util.GetUTCNow()):
    assert key
    assert item
    assert priority
    assert request_time

    if key not in self._queues:
      self._queues[key] = TaskPriorityQueue()

    self._queues[key].Enqueue(item, priority, request_time)

  def Dequeue(self):
    keys = sorted(self._queues.keys())
    assert keys

    current_key = keys[self._dequeue_ptr % len(keys)]
    self._dequeue_ptr += 1

    val = self._queues[current_key].Dequeue()
    if self._queues[current_key].IsEmpty():
      del self._queues[current_key]

    return val


# Tasks that have been queued, but not sent to swarming.
scheduled_tasks = TaskPriorityMultiQueue()

# Tasks that have been sent to swarming, but no results have been recieved.
pending_tasks = []

# Tasks that have been been completed, but the results haven't been retreived.
completed_tasks = []

# Tasks that have results and need to be sent to calling pipeline.
ready_tasks = []

# Flag to indicate this is the first call to the module. If it is the first
# call to the module, some initial population is required.
has_been_initialized = False


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
      _ = [
          queue.Enqueue(
              json.dumps(result.taskqueue_dimensions),
              result,
              result.taskqueue_priority,
              request_time=result.taskqueue_request_time) for result in results
      ]

  global scheduled_tasks
  global pending_tasks
  global completed_tasks
  global ready_tasks
  global has_been_initialized

  scheduled_tasks = TaskPriorityMultiQueue()
  pending_tasks = []
  completed_tasks = []
  ready_tasks = []

  query = swarming_task_queue_request.SwarmingTaskQueueRequest.query(
      swarming_task_queue_request.SwarmingTaskQueueRequest.taskqueue_state ==
      swarming_task_queue_request.SwarmingTaskQueueState.SCHEDULED)
  _FetchAndFillDimensionMultiqueue(query, scheduled_tasks)

  query = swarming_task_queue_request.SwarmingTaskQueueRequest.query(
      swarming_task_queue_request.SwarmingTaskQueueRequest.taskqueue_state ==
      swarming_task_queue_request.SwarmingTaskQueueState.PENDING)
  _FetchAndFillList(query, pending_tasks)

  query = swarming_task_queue_request.SwarmingTaskQueueRequest.query(
      swarming_task_queue_request.SwarmingTaskQueueRequest.taskqueue_state ==
      swarming_task_queue_request.SwarmingTaskQueueState.COMPLETED)
  _FetchAndFillList(query, completed_tasks)

  query = swarming_task_queue_request.SwarmingTaskQueueRequest.query(
      swarming_task_queue_request.SwarmingTaskQueueRequest.taskqueue_state ==
      swarming_task_queue_request.SwarmingTaskQueueState.READY)
  _FetchAndFillList(query, ready_tasks)

  has_been_initialized = True


def EnqueueTask(request, priority):
  """Enqueue a SwarmingTaskRequest to the TaskQueue."""

  request.taskqueue_state = (
      swarming_task_queue_request.SwarmingTaskQueueState.SCHEDULED)
  request.taskqueue_priority = priority
  request.taskqueue_request_time = time_util.GetUTCNow()
  request.put()

  # pylint: disable=global-variable-not-assigned
  global has_been_initialized
  if not has_been_initialized:
    Update()


def Update():  # pragma: no cover
  """Update method that queries for tasks, and dispatches them if necessary."""
  _InitializeTaskQueue()

  time.sleep(_TASKQUEUE_UPDATE_DELAY_SECONDS)
  Update()


def GetTaskRequest(task_id):
  for task_request in pending_tasks:
    if task_request.swarming_task_id == task_id:
      return task_request
  return None


def CompleteTask(task_id):
  """Called by pubsub push endpoint once this task is complete."""

  result_data, result_error = swarming_util.GetSwarmingTaskResultById(
      task_id, _HTTP_CLIENT)
  log_data, log_error = swarming_util.GetSwarmingTaskFailureLog(
      result_data['outputs_ref'],
      _HTTP_CLIENT) if result_data['outputs_ref'] else (None, None)

  if not process_swarming_results.IsTaskComplete(result_data):
    logging.info('Callback notified for %s, but task is incomplete' % task_id)
    return

  assert result_data
  assert log_data

  taskqueue_task = GetTaskRequest(task_id)
  analysis = ndb.Key(
      urlsafe=taskqueue_task.taskqueue_analysis_urlsafe_key).get()
  assert analysis

  task = FlakeSwarmingTask.Get(analysis.master_name, analysis.builder_name,
                               analysis.build_number, analysis.step_name,
                               analysis.test_name)
  assert task

  task.task_id = task_id
  task.created_time = datetime_util.ConvertDateTime(
      result_data.get('created_ts'))
  task.started_time = datetime_util.ConvertDateTime(
      result_data.get('started_ts'))
  task.completed_time = datetime_util.ConvertDateTime(
      result_data.get('completed_ts'))

  error = process_swarming_results.GetSwarmingTaskError(
      task_id, result_data, result_error, log_data, log_error)
  if error:
    task.status = analysis_status.ERROR
    task.error = error
    task.put()
    # TODO(crbug.com/779643): Callback pipeline.
    return

  step_name_no_platform = swarming_util.GetTagValue(
      result_data.get('tags', {}), 'ref_name')
  task.status = analysis_status.COMPLETED
  task.tests_statuses = process_swarming_results.CheckTestsRunStatuses(log_data)
  task.canonical_step_name = step_name_no_platform

  tries = task.tests_statuses.get(analysis.test_name, {}).get('total_run', 0)
  successes = task.tests_statuses.get(analysis.test_name, {}).get('SUCCESS', 0)

  task.tries = tries
  task.successes = successes
  task.put()

  # We've found all the results, callback the pipeline.
  # TODO(crbug.com/779643): Callback pipeline.
  return


# TODO(crbug.com/776440): Dispatch requests to swarming.

# TODO(crbug.com/775192): Get results and send the results to pipeline.

# TODO(crbug.com/780066): Cull or retry neglected tasks.
