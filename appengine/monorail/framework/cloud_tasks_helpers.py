# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""A helper module for interfacing with google cloud tasks.

This module wraps Gooogle Cloud Tasks, link to its documentation:
https://googleapis.dev/python/cloudtasks/1.3.0/gapic/v2/api.html
"""

from __future__ import absolute_import
from __future__ import division
from __future__ import print_function

import logging

from google.api_core import exceptions

import settings

if not settings.unit_test_mode:
  import grpc
  from google.cloud import tasks

_client = None


def _get_client():
  # type: () -> tasks.CloudTasksClient
  """Returns a cloud tasks client."""
  global _client
  if not _client:
    if settings.local_mode:
      _client = tasks.CloudTasksClient(
          channel=grpc.insecure_channel(settings.CLOUD_TASKS_EMULATOR_ADDRESS))
    else:
      _client = tasks.CloudTasksClient()
  return _client


def create_task(task, queue='default', **kwargs):
  # type: (Union[dict, tasks.types.Task], str, **Any) ->
  #     tasks.types.Task
  """Tries and catches creating a cloud task.

  This exposes a simplied task creation interface by wrapping
  tasks.CloudTasksClient.create_task; see its documentation:
  https://googleapis.dev/python/cloudtasks/1.5.0/gapic/v2/api.html#google.cloud.tasks_v2.CloudTasksClient.create_task

  To allow for local dev that does not run a cloud tasks emulator this catches
  GoogleAPICallErrors.

  Args:
    task: A dict or Task describing the task to add.
    queue: A string indicating name of the queue to add task to.
    kwargs: Additional arguments to pass to cloud task client's create_task

  Returns:
    Successfully created Task object.

  Raises:
    AttributeError: If input task is malformed or missing attributes.
    google.api_core.exceptions.GoogleAPICallError: If the request failed for any
        reason.
    google.api_core.exceptions.RetryError: If the request failed due to a
        retryable error and retry attempts failed.
    ValueError: If the parameters are invalid.
  """
  client = _get_client()

  parent = client.queue_path(
      settings.app_id, settings.CLOUD_TASKS_REGION, queue)
  target = task.get('app_engine_http_request').get('relative_uri')
  logging.info('Enqueueing %s task to %s', target, parent)
  try:
    return client.create_task(parent, task, **kwargs)
  except exceptions.ServiceUnavailable as err:
    # TODO(crbug/monorail/8360): Remove try catch after formalizing local dev.
    # We catch this exception to allow local dev that does not run a local
    # cloud tasks emulator
    logging.exception(err)
