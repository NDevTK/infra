# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import json
import logging

from google.appengine.api import urlfetch
from google.appengine.api import urlfetch_errors

from common.http import auth_util
from common.http.retry_http_client import RetryHttpClient

GAE_RETRIABLE_EXCEPTIONS = [
    urlfetch_errors.DownloadError,
    urlfetch_errors.InternalTransientError,
]


class HttpClientAppengine(RetryHttpClient):
  """A http client for running on appengine."""

  def __init__(self, follow_redirects=True, interceptor=None, *args, **kwargs):
    """Create a new client suitable for use within the app engine app."""
    interceptor = interceptor or auth_util.AuthenticatingInterceptor(
        retriable_exceptions=GAE_RETRIABLE_EXCEPTIONS)
    super(HttpClientAppengine, self).__init__(
        interceptor=interceptor, *args, **kwargs)
    self.follow_redirects = follow_redirects

  def _SendRequest(self, url, method, data, timeout, headers=None):
    headers = headers or {}
    if "isolateserver" in url:
      logging.debug("Sending isolateserver request with url = %s", url)
    result = urlfetch.fetch(
        url,
        payload=data,
        method=method,
        headers=headers,
        deadline=timeout,
        follow_redirects=self.follow_redirects,
        validate_certificate=True)

    return result.status_code, result.content, result.headers

  def _Get(self, url, timeout, headers):
    return self._SendRequest(url, urlfetch.GET, None, timeout, headers)

  def _Post(self, url, data, timeout, headers):
    return self._SendRequest(url, urlfetch.POST, data, timeout, headers)

  def _Put(self, url, data, timeout, headers):
    return self._SendRequest(url, urlfetch.PUT, data, timeout, headers)