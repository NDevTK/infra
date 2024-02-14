# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import logging
import os
import six

from google.appengine.api import runtime as apiruntime
from google.appengine.ext import ndb

from infra_libs.ts_mon import shared


def find_gaps(num_iter):
  """Generate integers not present in an iterable of integers.

  Caution: this is an infinite generator.
  """
  next_num = -1
  for n in num_iter:
    next_num += 1
    while next_num < n:
      yield next_num
      next_num += 1
  while True:
    next_num += 1
    yield next_num


def _assign_task_num(time_fn=datetime.datetime.utcnow):
  expired_keys = []
  unassigned = []
  used_task_nums = []
  time_now = time_fn()
  expired_time = time_now - datetime.timedelta(
      seconds=shared.INSTANCE_EXPIRE_SEC)
  for entity in shared.Instance.query():
    # Don't reassign expired task_num right away to avoid races.
    if entity.task_num >= 0:
      used_task_nums.append(entity.task_num)
    # At the same time, don't assign task_num to expired entities.
    if entity.last_updated < expired_time:
      expired_keys.append(entity.key)
      shared.expired_counter.increment()
      logging.debug(
          'Expiring %s task_num %d, inactive for %s',
          entity.key.id(), entity.task_num,
          time_now - entity.last_updated)
    elif entity.task_num < 0:
      shared.started_counter.increment()
      unassigned.append(entity)

  logging.debug('Found %d expired and %d unassigned instances',
                len(expired_keys), len(unassigned))

  used_task_nums = sorted(used_task_nums)
  for entity, task_num in zip(unassigned, find_gaps(used_task_nums)):
    entity.task_num = task_num
    logging.debug('Assigned %s task_num %d', entity.key.id(), task_num)
  futures_unassigned = ndb.put_multi_async(unassigned)
  futures_expired = ndb.delete_multi_async(expired_keys)
  ndb.Future.wait_all(futures_unassigned + futures_expired)
  logging.debug('Committed all changes')


def report_memory(handler):
  """Wraps an app so handlers log when memory usage increased by at least 0.5MB
  after the handler completed.
  """
  if os.environ.get('SERVER_SOFTWARE', '').startswith('Development'):
    # This is to detect "dev_appserver" environment and skip report_memory().
    # report_memory() fails with the following exception in dev_appserver.
    # :AssertionError: No api proxy found for service "system"
    return handler  # pragma: no cover

  min_delta = 0.5

  def dispatch_and_report(*args, **kwargs):
    if six.PY2:  # pragma: no cover
      before = apiruntime.runtime.memory_usage().current()
    else:  # pragma: no cover
      before = apiruntime.runtime.memory_usage().current
    try:
      return handler(*args, **kwargs)
    finally:
      if six.PY2:  # pragma: no cover
        after = apiruntime.runtime.memory_usage().current()
      else:  # pragma: no cover
        after = apiruntime.runtime.memory_usage().current
      if after >= before + min_delta:  # pragma: no cover
        logging.debug('Memory usage: %.1f -> %.1f MB; delta: %.1f MB', before,
                      after, after - before)

  return dispatch_and_report
