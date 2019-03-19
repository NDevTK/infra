# Copyright 2019 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from google.appengine.api import taskqueue
from google.appengine.ext import ndb


@ndb.tasklet
def enqueue_async(
    queue_name, task_kwargs, transactional=True
):  # pragma: no cover
  """Enqueues tasks. Mocked in tests."""
  q = taskqueue.Queue(queue_name)
  # Cannot just return add_async's return value because it is
  # a non-Future object and does not play nice with `yield fut1, fut2` construct
  tasks = [taskqueue.Task(**kwargs) for kwargs in task_kwargs]
  yield q.add_async(tasks, transactional=transactional)


# # Mocked in tests.
# @ndb.tasklet
# def enqueue_push_async(
#     queue_name, tasks, transactional=True
# ):  # pragma: no cover
#   """Enqueues push tasks. Mocked in tests."""
#   tasks = [
#       taskqueue.Task(
#           url=t.url,
#           payload=t.prepare_payload(),
#           retry_options=taskqueue.TaskRetryOptions(
#               task_age_limit=t.age_limit_sec,
#           )
#       ) for t in task_defs
#   ]

#   q = taskqueue.Queue(queue_name)
#   # Cannot just return add_async's return value because it is
#   # a non-Future object and does not play nice with `yield fut1, fut2` construct
#   yield q.add_async(tasks, transactional=transactional)

# # Mocked in tests.
# @ndb.tasklet
# def enqueue_pull_async(
#     queue_name, tasks, transactional=True
# ):  # pragma: no cover
#   """Enqueues pull tasks. Mocked in tests."""
#   tasks = [
#       taskqueue.Task(
#           method='PULL',
#           payload=t.prepare_payload(),
#       ) for t in task_defs
#   ]

#   q = taskqueue.Queue(queue_name)
#   # Cannot just return add_async's return value because it is
#   # a non-Future object and does not play nice with `yield fut1, fut2` construct
#   yield q.add_async(tasks, transactional=transactional)
