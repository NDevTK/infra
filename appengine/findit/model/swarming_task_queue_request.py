# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import json

from google.appengine.ext import ndb
from google.appengine.ext.ndb import msgprop
from protorpc import messages

from libs import time_util


class SwarmingTaskQueuePriority():
  # Forced a rerun of a failure or flake.
  FORCE = 1
  # Compile or test failure.
  FAILURE = 2
  # Test flake.
  FLAKE = 5
  # A request made through findit api.
  API_CALL = 10


class SwarmingTaskQueueState(messages.Enum):
  # Task has been sent to taskqueue, but not to swarming.
  SCHEDULED = 250

  # Task has been sent to swarming, but it hasn't completed.
  PENDING = 500

  # Task has been complete, but results haven't been requested.
  COMPLETED = 750


class SwarmingTaskQueueRequest(ndb.Model):
  # State of the request, see SwarmingTaskQueueState for details.
  state = msgprop.EnumProperty(SwarmingTaskQueueState, indexed=True)

  # Priority of the request, see SwarmingTaskQueuePriority for details.
  priority = ndb.FloatProperty()

  # Timestamp to order things in the Taskqueue.
  request_time = ndb.DateTimeProperty()

  # Used to uniquely identify the machine this request needs.
  dimensions = ndb.StringProperty()

  # The actual request from waterfall/swarming_task_request's Serialize method.
  swarming_task_request = ndb.TextProperty()

  # Used to callback the pipeline back when swarming task is complete.
  callback_url = ndb.StringProperty()

  # The analysis that requested this swarming task.
  analysis_urlsafe_key = ndb.StringProperty()

  # Task id of a swarming task.
  swarming_task_id = ndb.StringProperty(indexed=True)

  @staticmethod
  def Create(state=SwarmingTaskQueueState.SCHEDULED,
             priority=SwarmingTaskQueuePriority.FORCE,
             request_time=time_util.GetUTCNow(),
             dimensions=None,
             callback_url=None,
             analysis_urlsafe_key=None,
             swarming_task_id=None,
             swarming_task_request=None):
    swarming_task_queue_request = SwarmingTaskQueueRequest()
    swarming_task_queue_request.state = state
    swarming_task_queue_request.priority = priority
    swarming_task_queue_request.request_time = request_time
    swarming_task_queue_request.dimensions = dimensions
    swarming_task_queue_request.callback_url = callback_url
    swarming_task_queue_request.analysis_urlsafe_key = (analysis_urlsafe_key)
    swarming_task_queue_request.swarming_task_id = swarming_task_id
    swarming_task_queue_request.swarming_task_request = swarming_task_request
    return swarming_task_queue_request
