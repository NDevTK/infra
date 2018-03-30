# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is govered by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""An endpoint for performing IssueSnapshot queries for charts."""

from businesslogic import work_env
from framework import jsonfeed


# TODO(jeffcarp): Transition this handler to APIv2.
class SnapshotCounts(jsonfeed.InternalTask):
  """Handles IssueSnapshot queries.

  URL params:
    bucketby (str): One of (label, component). Defines the second dimension
      for bucketing IssueSnapshot counts. Defaults to 'label'.
    timestamp (int): The point in time at which snapshots will be counted.
    label_prefix (str): Required if bucketby=label. Returns only labels
      with this prefix, e.g. 'Pri'.
    q (str): Optional query string.

  Output:
    A JSON response with the following structure:
    {
      results: { name: count } for item in 2nd dimension.
      unsupported_fields: a list of strings for each unsupported field in query.
    }
  """

  def HandleRequest(self, mr):
    bucketby = mr.GetParam('bucketby') or 'label'
    label_prefix = mr.GetParam('label_prefix')
    query = mr.GetParam('q')
    timestamp = mr.GetParam('timestamp')
    if timestamp:
      timestamp = int(timestamp)
    else:
      return { 'error': 'Param `timestamp` required.' }
    if bucketby == 'label' and not label_prefix:
      return { 'error': 'Param `label_prefix` required.' }

    with work_env.WorkEnv(mr, self.services) as we:
      results, unsupported_fields = we.SnapshotCountsQuery(timestamp, bucketby,
          label_prefix, query)

    print 'unsupported_fields', unsupported_fields
    unsupported_field_names = [
        field.field_name
        for cond in unsupported_fields
        for field in cond.field_defs
    ]

    return {
      'results': results,
      'unsupported_fields': unsupported_field_names,
    }
