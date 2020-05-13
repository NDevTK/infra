# Copyright 2020 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import unittest

from unittest import mock

import datastore_utils


def _get_token(i, n_pages=None):
  if i == 0 or i == n_pages:
    return None
  return b'token-%d' % i


def _generate_entities(i, page_size):
  return [
      mock.Mock(key=key)
      for key in range(i * page_size, (i + 1) * page_size)]


def _generate_fetch_results(n_pages, page_size):
  fetch_results = {}
  for i in range(n_pages):
    fetch_results[_get_token(i)] = mock.Mock(
      pages=iter([_generate_entities(i, page_size)]),
      next_page_token=_get_token(i + 1, n_pages))
  return fetch_results


class DatastoreUtilsTest(unittest.TestCase):
  def setUp(self):
    super().setUp()
    self._cache = {}
    mock.patch('os.path.isdir', return_value=True).start()
    mock.patch(
        'datastore_utils._get_cache', side_effect=self._cache.get).start()
    mock.patch(
        'datastore_utils._update_cache',
        side_effect=self._cache.__setitem__).start()
    self.addCleanup(mock.patch.stopall)
    self.addCleanup(self._cache.clear)

  def _mock_fetch(self, fetch_results):
    def fn(limit, start_cursor):
      self.assertEqual(datastore_utils._PAGE_SIZE, limit)
      return fetch_results[start_cursor]
    return fn

  def test_fetch_query_empty_cache(self):
    n_pages = 5
    page_size = 10
    fetch_results = _generate_fetch_results(n_pages, page_size)
    query = mock.Mock()
    query.fetch.side_effect = self._mock_fetch(fetch_results)

    self.assertEqual(
        set(range(n_pages * page_size)),
        set(datastore_utils._fetch_query(query, 'foo')))
    self.assertEqual(n_pages, len(query.fetch.mock_calls))

  def test_fetch_query_interrupted_and_resumed(self):
    n_pages = 5
    page_size = 10
    fetch_results = _generate_fetch_results(n_pages, page_size)

    # Delete the result for the 3rd fetch call so that an exception is raised
    # and the query is interrupted.
    del fetch_results[_get_token(2)]
    query = mock.Mock()
    query.fetch.side_effect = self._mock_fetch(fetch_results)
    with self.assertRaises(KeyError):
      set(datastore_utils._fetch_query(query, 'foo'))

    # Three calls should have been made, two that succeeded and one that failed.
    self.assertEqual(3, len(query.fetch.mock_calls))

    # Run the query again without raising an exception.
    fetch_results = _generate_fetch_results(n_pages, page_size)
    query = mock.Mock()
    query.fetch.side_effect = self._mock_fetch(fetch_results)

    # We should get the full results.
    self.assertEqual(
        set(range(n_pages * page_size)),
        set(datastore_utils._fetch_query(query, 'foo')))

    # Only three calls should have been made, one to retry the 3rd call, and the
    # two calls that we didn't run the first time.
    self.assertEqual(3, len(query.fetch.mock_calls))


if __name__ == '__main__':
  unittest.main()
