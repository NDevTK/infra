# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# This file is a reduced copy of `luci/appengine/components/components/net.py`
# Ideally it shouldn't exist and the source code should be used directly, but
# we had to create a copy to simplify python3 migration. If some feature is
# missing from this file it can probably be copied over from source.

import logging
import os

from six.moves import urllib

from google.appengine.api import urlfetch
from google.appengine.api import app_identity
from google.appengine.ext import ndb
from google.appengine.runtime import apiproxy_errors

EMAIL_SCOPE = 'https://www.googleapis.com/auth/userinfo.email'


class Error(Exception):
  """Raised on non-transient errors.

  Attribute response is response body.
  """

  def __init__(self, msg, status_code, response, headers=None):
    super(Error, self).__init__(msg)
    self.status_code = status_code
    self.headers = headers or {}
    self.response = response


class NotFoundError(Error):
  """Raised if endpoint returns 404."""


class AuthError(Error):
  """Raised if endpoint returns 401 or 403."""


class MethodNotAllowed(Error):
  """Raised if endpoint returns 405."""


# Do not log Error exception raised from a tasklet, it is expected to happen.
ndb.add_flow_exception(Error)


def urlfetch_async(**kwargs):
  """To be mocked in tests."""
  return ndb.get_context().urlfetch(**kwargs)


def _is_transient_error(response, url):
  """Returns True to retry the request."""
  if response.status_code >= 500 or response.status_code == 408:
    return True
  # Retry 404 iff it is a Cloud Endpoints API call *and* the
  # result is not JSON. This assumes that we only use JSON encoding.
  if response.status_code == 404:
    content_type = response.headers.get('Content-Type', '')
    return (urllib.parse.urlparse(url).path.startswith('/_ah/api/') and
            not content_type.startswith('application/json'))
  return False


def _is_fatal_error(response, expected_codes):
  """Returns True if the HTTP response represents a non-transient error."""
  if expected_codes and response.status_code in expected_codes:
    return False
  return 300 <= response.status_code < 500


def _error_class_for_status(status_code):
  if status_code == 404:
    return NotFoundError
  if status_code in (401, 403):
    return AuthError
  if status_code == 405:
    return MethodNotAllowed
  return Error


def is_local_dev_server():
  """Returns True if running on local development server or in unit tests.

  This function is safe to run outside the scope of a HTTP request.
  """
  return os.environ.get('SERVER_SOFTWARE', '').startswith('Development')


@ndb.tasklet
def request_async(
    url,
    method='GET',
    payload=None,
    headers=None,
    scopes=None,
    deadline=None,
    max_attempts=None,
    response_headers=None,
    expected_codes=None,
):
  """Sends a REST API request, returns raw unparsed response.

  Retries the request on transient errors for up to |max_attempts| times.

  Args:
    url: url to send the request to.
    method: HTTP method to use, e.g. GET, POST, PUT.
    payload: raw data to put in the request body.
    headers: additional request headers.
    scopes: OAuth2 scopes for the access token (ok skip auth if None).
    deadline: deadline for a single attempt (10 sec by default).
    max_attempts: how many times to retry on errors (4 times by default).
    response_headers: a dict to populate with the response headers.
    expected_codes: a set of HTTP status codes to consider successful in
                    addition to the default set.

  Returns:
    Buffer with raw response.

  Raises:
    NotFoundError on 404 response.
    AuthError on 401 or 403 response.
    Error on any other non-transient error.
  """
  deadline = 10 if deadline is None else deadline
  max_attempts = 4 if max_attempts is None else max_attempts

  if is_local_dev_server():
    protocols = ('http://', 'https://')
  else:
    protocols = ('https://',)
  assert url.startswith(protocols), url
  headers = (headers or {}).copy()
  tok, _ = app_identity.get_access_token(scopes)
  if tok:
    headers['Authorization'] = 'Bearer ' + tok
  if payload is not None:
    assert isinstance(payload, str), type(payload)
    assert method in ('CREATE', 'POST', 'PUT'), method

  attempt = 0
  response = None
  # 901 CLIENT STATUS_ERROR. See gae_ts_mon/common/http_metrics.py
  last_status_code = 901
  while attempt < max_attempts:
    if attempt:
      logging.info('Retrying...')
    attempt += 1
    logging.info('%s %s', method, url)
    try:
      response = yield urlfetch_async(
          url=url,
          payload=payload,
          method=method,
          headers=headers,
          follow_redirects=False,
          deadline=deadline,
          validate_certificate=True)
    except (apiproxy_errors.DeadlineExceededError, urlfetch.Error) as e:
      # Transient network error or URL fetch service RPC deadline.
      logging.warning('%s %s failed: %s', method, url, e)
      # 902 CLIENT STATUS_TIMEOUT. See gae_ts_mon/common/http_metrics.py
      last_status_code = 902
      continue
    last_status_code = response.status_code

    # Transient error on the other side.
    if _is_transient_error(response, url):
      logging.warning('%s %s failed with HTTP %d\nHeaders: %r\nBody: %r',
                      method, url, response.status_code, response.headers,
                      response.content)
      continue

    # Non-transient error.
    if _is_fatal_error(response, expected_codes):
      logging.warning('%s %s failed with HTTP %d\nHeaders: %r\nBody: %r',
                      method, url, response.status_code, response.headers,
                      response.content)
      raise _error_class_for_status(response.status_code)(
          'Failed to call %s: HTTP %d' % (url, response.status_code),
          response.status_code,
          response.content,
          headers=response.headers)

    # Success. Beware of large responses.
    if response_headers is not None:
      response_headers.update(response.headers)
    if len(response.content) > 1024 * 1024:
      logging.warning('Response size: %.1f KiB', len(response.content) / 1024.0)
    raise ndb.Return(response.content)

  raise _error_class_for_status(last_status_code)(
      'Failed to call %s after %d attempts' % (url, max_attempts),
      response.status_code if response else last_status_code,
      response.content if response else None,
      headers=response.headers if response else None)
