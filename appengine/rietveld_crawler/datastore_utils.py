# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import pickle

from google.cloud import datastore


_CACHE_DIR = '.datastore_result_cache'
_QUERY_END = 'query end'
_PAGE_SIZE = 1000


def _get_cache(name, default):
  cache = os.path.join(_CACHE_DIR, name)
  if not os.path.isfile(cache):
    return default
  with open(cache, 'rb') as f:
    return pickle.load(f), cache


def _update_cache(name, value):
  cache = os.path.join(_CACHE_DIR, name)
  with open(cache, 'wb') as f:
    pickle.dump(value, f)


def _fetch_query(query, cache_name):
  if not os.path.isdir(_CACHE_DIR):
    os.path.makedirs(_CACHE_DIR)

  cursor_cache_name = f'{cache_name}-cursor'
  cursor = _get_cache(cursor_cache_name, None)
  result = _get_cache(cache_name, set())

  yield from result

  while cursor != _QUERY_END:
    query_iter = query.fetch(limit=_PAGE_SIZE, start_cursor=cursor)
    entities = set(entity.key for entity in next(query_iter.pages))
    result.update(entities)
    cursor = query_iter.next_page_token or _QUERY_END
    _update_cache(cache_name, result)
    _update_cache(cursor_cache_name, cursor)
    yield from entities

  return result


def get_issues():
  client = datastore.Client(project=config.rietveld_project)

  private_issues_query = client.query(kind='Issue')
  private_issues_query.add_filter('private', '=', True)
  private_issues = set(_fetch_query(private_issues_query, 'private_issues'))

  all_issues_query = client.query(kind='Issue')
  for issue in _fetch_query(all_issues_query, 'all_issues'):
    yield issue, issue in private_issues
