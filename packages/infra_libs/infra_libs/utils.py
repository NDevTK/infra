# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Miscellaneous utility functions."""


from __future__ import print_function

import contextlib
import datetime
import json
import shutil
import sys
import tempfile

from six import text_type, PY2


# UTC datetime corresponding to zero Unix timestamp.
EPOCH = datetime.datetime.utcfromtimestamp(0)

def parse_rfc3339_epoch(value):
  """Parses RFC 3339 datetime string as epoch
  (as used in Timestamp proto JSON encoding).

  Keeps only second precision (dropping micro- and nanoseconds).

  Examples of the input:
    2017-08-17T04:21:32.722952943Z
    1972-01-01T10:00:20.021-05:00

  Returns:
    epoch timestamp

  Raises:
    ValueError on errors.
  """
  # Adapted from protobuf/internal/well_known_types.py Timestamp.FromJsonString.
  # We can't use the original, since it's marked as internal. Also instantiating
  # proto messages here to parse a string would been odd.
  timezone_offset = value.find('Z')
  if timezone_offset == -1:
    timezone_offset = value.find('+')
  if timezone_offset == -1:
    timezone_offset = value.rfind('-')
  if timezone_offset == -1:
    raise ValueError('Failed to parse timestamp: missing valid timezone offset')
  time_value = value[0:timezone_offset]
  # Parse datetime and nanos.
  point_position = time_value.find('.')
  if point_position == -1:
    second_value = time_value
    nano_value = ''
  else:
    second_value = time_value[:point_position]
    nano_value = time_value[point_position + 1:]
  date_object = datetime.datetime.strptime(second_value, '%Y-%m-%dT%H:%M:%S')
  td = date_object - EPOCH
  seconds = td.seconds + td.days * 86400
  if len(nano_value) > 9:
    raise ValueError(
        'Failed to parse timestamp: nanos %r more than 9 fractional digits'
        % nano_value)
  # Parse timezone offsets.
  if value[timezone_offset] == 'Z':
    if len(value) != timezone_offset + 1:
      raise ValueError('Failed to parse timestamp: invalid trailing data %r'
                       % value)
  else:
    timezone = value[timezone_offset:]
    pos = timezone.find(':')
    if pos == -1:
      raise ValueError('Invalid timezone offset value: %r' % timezone)
    if timezone[0] == '+':
      seconds -= (int(timezone[1:pos])*60+int(timezone[pos+1:]))*60
    else:
      seconds += (int(timezone[1:pos])*60+int(timezone[pos+1:]))*60
  return seconds


def read_json_as_utf8(filename=None, text=None):
  """Read and deserialize a json file or string.

  This function is different from json.load and json.loads in that it
  returns utf8-encoded string for keys and values instead of unicode.

  On python3 this doesn't do any special re-encoding.

  Args:
    filename (str): path of a file to parse
    text (str): json string to parse

  ``filename`` and ``text`` are mutually exclusive. ValueError is raised if
  both are provided.
  """

  if filename is not None and text is not None:
    raise ValueError('Only one of "filename" and "text" can be provided at '
                     'the same time')

  if filename is None and text is None:
    raise ValueError('One of "filename" and "text" must be provided')

  def to_utf8(obj):
    if isinstance(obj, dict):
      return {to_utf8(key): to_utf8(value) for key, value in list(obj.items())}
    if isinstance(obj, list):
      return [to_utf8(item) for item in obj]
    if isinstance(obj, text_type):
      return obj.encode('utf-8')
    return obj

  if filename:
    with open(filename, 'rb') as f:
      obj = json.load(f)
  else:
    obj = json.loads(text)

  return to_utf8(obj) if PY2 else obj


# We're trying to be compatible with Python3 tempfile.TemporaryDirectory
# context manager here. And they used 'dir' as a keyword argument.
# pylint: disable=redefined-builtin
@contextlib.contextmanager
def temporary_directory(suffix="", prefix="tmp", dir=None,
                        keep_directory=False):
  """Create and return a temporary directory.  This has the same
  behavior as mkdtemp but can be used as a context manager.  For
  example:

    with temporary_directory() as tmpdir:
      ...

  Upon exiting the context, the directory and everything contained
  in it are removed.

  Args:
    suffix, prefix, dir: same arguments as for tempfile.mkdtemp.
    keep_directory (bool): if True, do not delete the temporary directory
      when exiting. Useful for debugging.

  Returns:
    tempdir (str): full path to the temporary directory.
  """
  tempdir = None  # Handle mkdtemp raising an exception
  try:
    tempdir = tempfile.mkdtemp(suffix, prefix, dir)
    yield tempdir

  finally:
    if tempdir and not keep_directory:  # pragma: no branch
      try:
        # TODO(pgervais,496347) Make this work reliably on Windows.
        shutil.rmtree(tempdir, ignore_errors=True)
      except OSError as ex:  # pragma: no cover
        print(("ERROR: {!r} while cleaning up {!r}".format(ex, tempdir)),
              file=sys.stderr)
