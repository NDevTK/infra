# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import random

from google.appengine.ext import ndb
from google.appengine.ext.ndb import msgprop
from protorpc import messages


class BuildStatus(messages.Enum):
  SCHEDULED = 1
  BUILDING = 2
  COMPLETE = 3


class Callback(ndb.Model):
  """Parameters for a callack push task."""
  url = ndb.StringProperty(required=True, indexed=False)
  headers = ndb.JsonProperty()
  method = ndb.StringProperty(indexed=False)
  queue_name = ndb.StringProperty(indexed=False)


class Build(ndb.Model):
  """Describes a build.

  Build key:
    Build keys are autogenerated integers. Has no parent.

  Attributes:
    owner (string): opaque indexed optional string that identifies the owner of
      the build. For example, this might be a buildset or Gerrit revision.
    namespace (string): a generic way to distinguish builds. Different build
      namespaces have different permissions.
    parameters (dict): immutable arbitrary build parameters.
    state (dict): mutable build state, a dictionary with arbitrary keys.
      Build state is reset to {} when an expired build is transitioned from
      BUILDING back to SCHEDULED status by a cleanup cron job.
    status (BuildStatus): status of the build.
    available_since (datetime): the earliest time the build can be leased.
      The moment the build is leased, |available_since| is set to
      (utcnow + lease_duration). On build creation, is set to utcnow.
    lease_key (int): a random value, changes every time a build is leased.
      Can be used to verify that a client is the leaseholder.
    callback (Callback): parameters for a push task creation on build status
      changes.
  """

  owner = ndb.StringProperty()
  namespace = ndb.StringProperty(required=True)
  parameters = ndb.JsonProperty()
  state = ndb.JsonProperty()
  status = msgprop.EnumProperty(BuildStatus, default=BuildStatus.SCHEDULED)
  available_since = ndb.DateTimeProperty(required=True, auto_now_add=True)
  lease_key = ndb.IntegerProperty(indexed=False)
  callback = ndb.StructuredProperty(Callback, indexed=False)

  def is_leasable(self):
    return (self.available_since <= datetime.datetime.utcnow() and
            self.status == BuildStatus.SCHEDULED)

  def regenerate_lease_key(self):
    """Changes lease key to a different random integer."""
    while True:
      new_key = random.randint(0, 1 << 31)
      if new_key != self.lease_key:  # pragma: no branch
        self.lease_key = new_key
        break
