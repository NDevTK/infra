# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""A handler run on Flask request teardown."""

import logging
import os

from google.appengine.api import runtime
from googleapiclient import discovery
from googleapiclient import errors
from oauth2client import client

import settings

_MAXIMUM_MEMORY_USAGE = 2000


def Teardown(_exc):
  if settings.local_mode:
    return

  # Stop the instance if it's using too much memory.
  memory_usage = runtime.memory_usage().average10m
  if memory_usage < _MAXIMUM_MEMORY_USAGE:
    return

  credentials = client.GoogleCredentials.get_application_default()
  appengine = discovery.build('appengine', 'v1', credentials=credentials)
  delete = appengine.apps().services().versions().instances().delete(
      appsId=os.environ.get('GAE_APPLICATION').split('~')[-1],
      servicesId=os.environ.get('GAE_SERVICE'),
      versionsId=os.environ.get('GAE_VERSION'),
      instancesId=os.environ.get('GAE_INSTANCE'))
  try:
    delete.execute()
  except errors.HttpError as e:
    if e.status_code != 404:
      raise
  else:
    logging.critical('Deleted instance using %d MB of memory.' % memory_usage)
