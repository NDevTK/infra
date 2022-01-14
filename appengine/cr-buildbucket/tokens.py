# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Provides functions for handling tokens."""

from components.auth import b64
from go.chromium.org.luci.buildbucket.proto import token_pb2


def generate_build_token(build_id):
  """Returns a token associated with the build."""
  body = token_pb2.TokenBody(
      build_id=build_id,
      purpose=token_pb2.TokenBody.BUILD,
      state=str(build_id),
  )
  env = token_pb2.TokenEnvelope(version=0)
  env.payload = body.SerializeToString()
  return b64.encode(env.SerializeToString())
