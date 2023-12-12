# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import base64
from datetime import datetime
from datetime import timedelta
import hashlib
import hmac
import os

from google.appengine.ext import ndb
import six

from gae_libs.http import auth_util
from libs import time_util

_DELIMITER = b':'
_RANDOM_BYTE_LENGTH = 512


def GenerateRandomHexKey(length=_RANDOM_BYTE_LENGTH):
  """Returns a key hexed from random bytes at the given length for crypto."""
  # After encoded in hex, the length doubles.
  if six.PY2:
    return os.urandom(length).encode('hex')
  else:
    return six.ensure_binary(os.urandom(length).hex())


class SecretKey(ndb.Model):
  # Store the secret key.
  secret_key = ndb.StringProperty(indexed=False)

  @classmethod
  def GetSecretKey(cls, user_id):
    """Returns a secret key for the user and creates it on demand."""
    uid = six.ensure_binary(
        hashlib.sha256(six.ensure_binary(user_id)).hexdigest())
    entity = ndb.Key(cls, uid).get()
    if not entity:
      entity = cls(id=uid, secret_key=GenerateRandomHexKey())
      entity.put()
    return six.ensure_binary(entity.secret_key)


def GenerateAuthToken(key_name, user_id, action_id='', when=None):
  """Generates a URL-safe token based on XSRFToken but for generla purpose.

  Args:
    key_name (str): name of secret key to generate token.
    user_id (str): the user ID of the authenticated user.
    action_id (str): a string identifier of the action they requested
      authorization for.
    when (datetime): the time when the user was authorized for this action.
      If not set the current utc time is used.
  Returns:
    A string token.
  """
  key = SecretKey.GetSecretKey(key_name)
  when = when or time_util.GetUTCNow()
  when_timestamp = time_util.ConvertToTimestamp(when)
  if six.PY2:
    digester = hmac.new(key)
  else:
    digester = hmac.new(key, msg=None, digestmod='MD5')
  digester.update(six.ensure_binary(user_id))
  digester.update(_DELIMITER)
  digester.update(six.ensure_binary(action_id))
  digester.update(_DELIMITER)
  digester.update(six.ensure_binary(str(when_timestamp)))
  digest = digester.digest()

  return base64.urlsafe_b64encode(
      six.ensure_binary('%s%s%d' % (digest, _DELIMITER, when_timestamp)))


def ValidateAuthToken(key_name, token, user_id, action_id='', valid_hours=1):
  """Validates a token.

  Args:
    key_name (str): name of secret key to generate token.
    token (str): a string of the token generated by GenerateAuthToken.
    user_id (str): the user ID of the authenticated user.
    action_id (str): a string identifier of the action they requested
      authorization for.
    valid_hours (int): number of hours the token should be valid.

  Returns:
    A bool whether the token is valid.
    A bool if the token is expired.
  """

  if not token:
    return False, False
  try:
    decoded = base64.urlsafe_b64decode(str(token))
    token_time = datetime.utcfromtimestamp(int(decoded.split(_DELIMITER)[-1]))
  except (TypeError, ValueError):
    return False, False

  current_time = time_util.GetUTCNow()
  expired = False
  # If the token is too old it's not valid.
  if current_time - token_time > timedelta(hours=valid_hours):
    expired = True

  # The given token should match the generated one with the same time.
  expected_token = GenerateAuthToken(
      key_name, user_id, action_id=action_id, when=token_time)
  if len(token) != len(expected_token):
    return False, expired

  # Perform constant time comparison to avoid timing attacks.
  different = 0
  for x, y in zip(token, expected_token):
    different |= ord(x) ^ ord(y)
  if different:
    return False, expired

  return True, expired


class AddXSRFToken(object):
  """A decorator to add a XSRF token to the response for the handler."""

  def __init__(self, action_id=''):
    self._action_id = action_id

  def __call__(self, handler_method):

    def AddToken(handler, *args, **kwargs):
      result = handler_method(handler, *args, **kwargs)
      user_email = auth_util.GetUserEmail()
      if not user_email:
        return result
      xsrf_token = GenerateAuthToken('site', user_email, self._action_id)
      result = result or {}
      result['data'] = result.get('data', {})
      result['data']['xsrf_token'] = xsrf_token
      return result

    return AddToken


class VerifyXSRFToken(object):
  """A decorator to enforce that the XSRF token is validated for the handler."""

  def __init__(self, action_id=''):
    self._action_id = action_id

  def __call__(self, handler_method):

    def VerifyToken(handler, *args, **kwargs):
      user_email = auth_util.GetUserEmail()
      xsrf_token = str(handler.request.values.get('xsrf_token'))
      valid, expired = ValidateAuthToken('site', xsrf_token,
                                         user_email, self._action_id)
      if (not user_email or not valid or expired):
        pass
        return handler.CreateError(
            'Invalid XSRF token. Please log in or refresh the page first.',
            return_code=403)
      return handler_method(handler, *args, **kwargs)

    return VerifyToken
