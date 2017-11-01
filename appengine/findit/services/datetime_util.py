# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.
import datetime


def ConvertDateTime(time_string):
  """Convert UTC time string to datetime.datetime."""
  if not time_string:
    return None
  for fmt in ('%Y-%m-%dT%H:%M:%S.%f', '%Y-%m-%dT%H:%M:%S'):
    # When microseconds are 0, the '.123456' suffix is elided.
    try:
      return datetime.datetime.strptime(time_string, fmt)
    except ValueError:
      pass
  raise ValueError('Failed to parse %s' % time_string)  # pragma: no cover