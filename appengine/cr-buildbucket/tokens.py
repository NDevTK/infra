# Copyright 2018 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Provides functions for handling tokens."""

from components import auth
import model


class BuildToken(auth.TokenKind):
  """Used for generating tokens to validate build messages."""
  expiration_sec = model.BUILD_TIMEOUT.total_seconds()
  secret_key = auth.SecretKey('build_id')


def _token_message(build_id):
  assert isinstance(build_id, (int, long)), build_id
  return str(build_id)


def generate_build_token(build_id):
  """Returns a token associated with the build."""
  build_id_str = _token_message(build_id)
  return str(build_id) + '/' + BuildToken.generate(build_id_str)


def validate_build_token(token, build_id):
  """Raises auth.InvalidTokenError if the token is invalid."""
  parts = token.split('/')
  if len(parts) == 2:
    token_without_build_id = parts[1]
  else:  # pragma: no cover
    # TODO(chanli@): Remove after all build tokens have build_id as prefix.
    token_without_build_id = token
  return BuildToken.validate(token_without_build_id, _token_message(build_id))
