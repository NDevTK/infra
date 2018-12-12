# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Utility functions.

Has "bb" prefix to avoid confusion with components.utils.
"""

from google.protobuf import json_format
from google.protobuf import struct_pb2


def message_to_jsonpb(msg):  # pragma: no cover
  """Converts msg to JSONPB deterministically."""
  return json_format.MessageToJson(msg, indent=None, sort_keys=True)


def dict_to_struct(d):  # pragma: no cover
  """Converts a dict to google.protobuf.Struct."""
  s = struct_pb2.Struct()
  s.update(d)
  return s


def dict_to_jsonpb(d):  # pragma: no cover
  """Converts a dict to JSONPB."""
  return message_to_jsonpb(dict_to_struct(d))


def update_struct(dest, src):  # pragma: no cover
  """Updates dest struct with values from src.

  Like dict.update, but for google.protobuf.Struct.
  """
  for key, value in src.fields.iteritems():
    # This will create a new struct_pb2.Value if one does not exist.
    dest.fields[key].CopyFrom(value)


def struct_to_dict(s):  # pragma: no cover
  """Converts a google.protobuf.Struct to dict."""
  return json_format.MessageToDict(s)
