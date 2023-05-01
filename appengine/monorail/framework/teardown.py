# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
"""A handler run on Flask request teardown."""

import logging
import os

from google.appengine.api import runtime
from googleapiclient.discovery import build
from oauth2client.client import GoogleCredentials

import settings

_MAXIMUM_MEMORY_USAGE = 2000


def Teardown(_exc):
  if settings.local_mode:
    return

  # Check memory usage and delete it if it's too high.
  memory_usage = runtime.memory_usage().average10m
  if memory_usage >= _MAXIMUM_MEMORY_USAGE:
    logging.critical('Deleting instance using %d MB of memory.' % memory_usage)
    credentials = GoogleCredentials.get_application_default()
    appengine = build('appengine', 'v1', credentials=credentials)
    appengine.apps().services().versions().instances().delete(
        appsId=os.environ.get('GAE_APPLICATION').split('~')[-1],
        servicesId=os.environ.get('GAE_SERVICE'),
        versionsId=os.environ.get('GAE_VERSION'),
        instancesId=os.environ.get('GAE_INSTANCE')).execute()
