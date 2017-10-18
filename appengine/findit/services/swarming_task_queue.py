from queue import PriorityQueue
from protorpc import messages

from libs import time_util

class TaskPriorityQueueItem():
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

class TaskPriorityQueue():

  def __init__(self):
    self._last_priority_dequeued = None
    self._number_of_dequeues = None
    self._queue = []

  def Enqueue(self, item, priority, time=time_util.GetUTCNow()):
    self._queue.append( TaskPriorityQueueItem(item, priority) )
    self._queue.sort(key=lambda task: (task.priority, time), reverse=True)

  def Dequeue(self):
    if not self._queue:  # Tried to Dequeue from an empty queue.
      raise Exception()

    if self._last_priority_dequeued is None:
      self._number_of_dequeues = 0
      self._last_priority_dequeued = self._queue[0].priority

    [task.priority = task.priority * 2 for task in self._queue]
    return self._queue.pop(0)

def TaskQueuePriority(messages.Enum):
  # Forced a rerun of a failure or flake.
  FORCE = 100
  FAILURE = 50
  FLAKE = 25
  # A request made through findit api.
  API_CALL = 10

# Tasks that have been queued, but not sent to swarming.
scheduled_tasks = [PriorityQueue()]
# Offset into the scheduled tasks multiqueue where a task was last taken from.
scheduled_tasks_ptr = 0

# Tasks that have been sent to swarming, but no results have been recieved.
pending_tasks = []

# Tasks that have been been completed, but the results haven't been retreived.
completed_tasks = []

# Tasks that have results and need to be sent to calling pipeline.
ready_tasks = []

# Flag to indicate this is the first call to the module. If it is the first
# call to the module, some initial population is required.
has_initialized = False

def _FetchAndFillList(query, lst):
  """ Fetch everything from the query and fill the given list with it.

  This function mutates the given list rather than returning a new one.

  Args:
    query (query): SwarmingTaskRequest query that contains all the SCHEDULED tasks.
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
    [queue.Enqueue(result, priority, result.request_time) for result in results]


def _InitializeTaskQueue():
  """Fetch and fill queues with existing data."""
  scheduled_tasks = TaskPriorityQueue()
  pending_tasks = []
  completed_tasks = []
  ready_tasks = []

  query = swarming_task_request.SwarmingTaskRequest.query(
    swarming_task_request.SwarmingTaskRequest.state ==
    swarming_task_request.SwarmingTaskState.SCHEDULED)
  _FetchAndFillDimensionMultiqueue(query, scheduled_tasks)

  query = swarming_task_request.SwarmingTaskRequest.query(
    swarming_task_request.SwarmingTaskRequest.state ==
    swarming_task_request.SwarmingTaskState.PENDING)
  _FetchAndFillList(query, pending_tasks)

  query = swarming_task_request.SwarmingTaskRequest.query(
    swarming_task_request.SwarmingTaskRequest.state ==
    swarming_task_request.SwarmingTaskState.COMPLETED)
  _FetchAndFillList(query, completed_tasks)

  query = swarming_task_request.SwarmingTaskRequest.query(
    swarming_task_request.SwarmingTaskRequest.state ==
    swarming_task_request.SwarmingTaskState.READY)
  _FetchAndFillList(query, ready_tasks)

  has_initialized = True


def EnqueueTask(request, priority):
  """Enqueue a SwarmingTaskRequest to the TaskQueue."""
  if not has_initialized:
    Update()
  scheduled_tasks.Enqueue(request)
  request.state = swarming_task_request.SwarmingTaskState.SCHEDULED
  request.priority = priority
  request.put()

def Update():
  _InitializeTaskQueue()

