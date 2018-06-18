# Copyright 2014 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import datetime
import mock

from google.appengine.ext import ndb

from components import utils
from testing_utils import testing

import model
import v2


class BuildTest(testing.AppengineTestCase):

  def test_regenerate_lease_key(self):
    build = model.Build(bucket='chromium')
    build.put()
    orig_lease_key = 0
    build.regenerate_lease_key()
    self.assertNotEqual(build.lease_key, orig_lease_key)

  def test_put_with_bad_tags(self):
    build = model.Build(bucket='1', tags=['x'])
    with self.assertRaises(AssertionError):
      build.put()

  def test_create_build_id_generates_monotonically_decreasing_ids(self):
    now = datetime.datetime(2015, 2, 24)
    ids = []
    for i in xrange(1000):
      now += datetime.timedelta(seconds=i)
      ids.extend(model.create_build_ids(now, 5))
    self.assertEqual(ids, sorted(ids, reverse=True))

  def test_build_id_range(self):
    time_low = datetime.datetime(2015, 1, 1)
    time_high = time_low + datetime.timedelta(seconds=10)
    id_low, id_high = model.build_id_range(time_low, time_high)
    unit = model._TIME_RESOLUTION

    ones = (1 << model._BUILD_ID_SUFFIX_LEN) - 1
    for suffix in (0, ones):

      def in_range(t):
        build_id = model._id_time_segment(t) | suffix
        return id_low <= build_id < id_high

      self.assertFalse(in_range(time_low - unit))
      self.assertTrue(in_range(time_low))
      self.assertTrue(in_range(time_low + unit))

      self.assertTrue(in_range(time_high - unit))
      self.assertFalse(in_range(time_high))
      self.assertFalse(in_range(time_high + unit))


class TagIndexTest(testing.AppengineTestCase):

  def test_zeroth_shard(self):
    self.assertEqual(
        model.TagIndex.make_key(0, 'a:b'),
        ndb.Key(model.TagIndex, 'a:b'),
    )

  def test_positive_shard_index(self):
    self.assertEqual(
        model.TagIndex.make_key(1, 'a:b'),
        ndb.Key(model.TagIndex, ':1:a:b'),
    )

  def test_random_shard_key(self):
    with mock.patch('model.TagIndex.random_shard_index', return_value=2):
      self.assertEqual(
          model.TagIndex.random_shard_key('a:b'),
          ndb.Key(model.TagIndex, ':2:a:b'),
      )
