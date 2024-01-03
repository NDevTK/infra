# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

# This file is a reduced copy of
# `luci/appengine/components/components/gitiles.py`.
# Ideally it shouldn't exist and the source code should be used directly, but
# we had to create a copy to simplify python3 migration. If some feature is
# missing from this file it can probably be copied over from source.
"""Gitiles functions for GAE environment."""

import base64
import six
from six.moves import urllib

from google.appengine.ext import ndb

from common import gerrit


@ndb.tasklet
def get_file_content_async(hostname,
                           project,
                           treeish,
                           path,
                           cmd='',
                           **fetch_kwargs):
  """Gets file contents.

  Returns:
    Raw contents of the file or None if not found.
  """
  _validate_args(hostname, project, treeish, path, path_required=True)
  data = yield gerrit.fetch_async(
      hostname,
      '%s/+%s/%s%s' % _quote_all(project, cmd, treeish, path),
      headers={'Accept': 'text/plain'},
      **fetch_kwargs)
  raise ndb.Return(base64.b64decode(data) if data is not None else None)


def get_file_content(*args, **kwargs):
  """Blocking version of get_file_content_async."""
  return get_file_content_async(*args, **kwargs).get_result()


def _validate_args(hostname,
                   project,
                   treeish='HEAD',
                   path=None,
                   path_required=False):
  assert_non_empty_string(hostname)
  assert_non_empty_string(project)
  _validate_treeish(treeish)
  if path_required:
    assert path is not None
  if path is not None:
    assert_non_empty_string(path)
    assert path.startswith(path), path


def _validate_treeish(treeish):
  assert_non_empty_string(treeish)
  assert not treeish.startswith('/'), treeish
  assert not treeish.endswith('/'), treeish


def assert_non_empty_string(value):
  assert isinstance(value, six.string_types)
  assert value


def _quote_all(*args):
  return tuple(map(urllib.parse.quote, args))
