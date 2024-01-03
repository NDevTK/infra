# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# This file is a reduced copy of
# `luci/appengine/components/components/gerrit.py`.
# Ideally it shouldn't exist and the source code should be used directly, but
# we had to create a copy to simplify python3 migration. If some feature is
# missing from this file it can probably be copied over from source.
"""Gerrit functions for GAE environment."""

from six.moves import urllib

from google.appengine.ext import ndb

from common import net

AUTH_SCOPE = 'https://www.googleapis.com/auth/gerritcodereview'
RESPONSE_PREFIX = ")]}'"


@ndb.tasklet
def fetch_async(hostname, path, **kwargs):
  """Sends request to Gerrit, returns raw response.

  See 'net.request_async' for list of accepted kwargs.

  Returns:
    Response body on success.
    None on 404 response.

  Raises:
    net.Error on communication errors.
  """
  assert not path.startswith('/'), path
  assert 'scopes' not in kwargs, kwargs['scopes']
  try:
    url = urllib.parse.urljoin('https://' + hostname, 'a/' + path)
    result = yield net.request_async(url, scopes=[AUTH_SCOPE], **kwargs)
    raise ndb.Return(result)
  except net.NotFoundError:
    raise ndb.Return(None)


def fetch(*args, **kwargs):
  """Blockng version of fetch_async."""
  return fetch_async(*args, **kwargs).get_result()
