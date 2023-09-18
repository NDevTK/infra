# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Tests for the star service."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import six
import unittest

try:
  from mox3 import mox
except ImportError:
  import mox
import mock
import time

from google.appengine.ext import testbed

from mock import Mock
from framework import sql
from services import service_manager
from services import star_svc
from testing import fake


class AbstractStarServiceTest(unittest.TestCase):

  def setUp(self):
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_memcache_stub()

    self.mox = mox.Mox()
    self.mock_tbl = self.mox.CreateMock(sql.SQLTableManager)
    self.cnxn = 'fake connection'
    self.cache_manager = fake.CacheManager()
    self.star_service = star_svc.AbstractStarService(
        self.cache_manager, self.mock_tbl, 'item_id', 'user_id', 'project')
    self.mock_tbl.Delete = Mock()

  def tearDown(self):
    self.testbed.deactivate()
    self.mox.UnsetStubs()
    self.mox.ResetAll()

  def SetUpExpungeStars(self):
    self.mock_tbl.Delete(self.cnxn, item_id=123, commit=True)

  def testExpungeStars(self):
    self.SetUpExpungeStars()
    self.mox.ReplayAll()
    self.star_service.ExpungeStars(self.cnxn, 123)
    self.mox.VerifyAll()

  def testExpungeStars_Limit(self):
    self.star_service.ExpungeStars(self.cnxn, 123, limit=50)
    self.mock_tbl.Delete.assert_called_once_with(
        self.cnxn, commit=True, limit=50, item_id=123)

  def testExpungeStarsByUsers(self):
    user_ids = [2, 3, 4]
    self.star_service.ExpungeStarsByUsers(self.cnxn, user_ids, limit=40)
    self.mock_tbl.Delete.assert_called_once_with(
        self.cnxn, user_id=user_ids, commit=False, limit=40)

  def SetUpLookupItemsStarrers(self):
    self.mock_tbl.Select(
        self.cnxn, cols=['item_id', 'user_id'],
        item_id=[234]).AndReturn([(234, 111), (234, 222)])

  def testLookupItemsStarrers(self):
    self.star_service.starrer_cache.CacheItem(123, [111, 333])
    self.SetUpLookupItemsStarrers()
    self.mox.ReplayAll()
    starrer_list_dict = self.star_service.LookupItemsStarrers(
        self.cnxn, [123, 234])
    self.mox.VerifyAll()
    six.assertCountEqual(self, [123, 234], list(starrer_list_dict.keys()))
    six.assertCountEqual(self, [111, 333], starrer_list_dict[123])
    six.assertCountEqual(self, [111, 222], starrer_list_dict[234])
    six.assertCountEqual(
        self, [111, 333], self.star_service.starrer_cache.GetItem(123))
    six.assertCountEqual(
        self, [111, 222], self.star_service.starrer_cache.GetItem(234))

  def SetUpLookupStarredItemIDs(self):
    self.mock_tbl.Select(
        self.cnxn, cols=['item_id'], user_id=111).AndReturn(
            [(123,), (234,)])

  def testLookupStarredItemIDs(self):
    self.SetUpLookupStarredItemIDs()
    self.mox.ReplayAll()
    item_ids = self.star_service.LookupStarredItemIDs(self.cnxn, 111)
    self.mox.VerifyAll()
    six.assertCountEqual(self, [123, 234], item_ids)
    six.assertCountEqual(
        self, [123, 234], self.star_service.star_cache.GetItem(111))

  def testIsItemStarredBy(self):
    self.SetUpLookupStarredItemIDs()
    self.mox.ReplayAll()
    self.assertTrue(self.star_service.IsItemStarredBy(self.cnxn, 123, 111))
    self.assertTrue(self.star_service.IsItemStarredBy(self.cnxn, 234, 111))
    self.assertFalse(
        self.star_service.IsItemStarredBy(self.cnxn, 435, 111))
    self.mox.VerifyAll()

  def SetUpCountItemStars(self):
    self.mock_tbl.Select(
        self.cnxn, cols=['item_id', 'COUNT(user_id)'], item_id=[234],
        group_by=['item_id']).AndReturn([(234, 2)])

  def testCountItemStars(self):
    self.star_service.star_count_cache.CacheItem(123, 3)
    self.SetUpCountItemStars()
    self.mox.ReplayAll()
    self.assertEqual(3, self.star_service.CountItemStars(self.cnxn, 123))
    self.assertEqual(2, self.star_service.CountItemStars(self.cnxn, 234))
    self.mox.VerifyAll()

  def testCountItemsStars(self):
    self.star_service.star_count_cache.CacheItem(123, 3)
    self.SetUpCountItemStars()
    self.mox.ReplayAll()
    count_dict = self.star_service.CountItemsStars(
        self.cnxn, [123, 234])
    self.mox.VerifyAll()
    six.assertCountEqual(self, [123, 234], list(count_dict.keys()))
    self.assertEqual(3, count_dict[123])
    self.assertEqual(2, count_dict[234])

  def SetUpSetStar_Add(self):
    self.mock_tbl.InsertRows(
        self.cnxn, ['item_id', 'user_id'], [(123, 111)], ignore=True,
        commit=True)

  def testSetStar_Add(self):
    self.SetUpSetStar_Add()
    self.mox.ReplayAll()
    self.star_service.SetStar(self.cnxn, 123, 111, True)
    self.mox.VerifyAll()
    self.assertFalse(self.star_service.star_cache.HasItem(123))
    self.assertFalse(self.star_service.starrer_cache.HasItem(123))
    self.assertFalse(self.star_service.star_count_cache.HasItem(123))

  def SetUpSetStar_Remove(self):
    self.mock_tbl.Delete(self.cnxn, item_id=123, user_id=[111])

  def testSetStar_Remove(self):
    self.SetUpSetStar_Remove()
    self.mox.ReplayAll()
    self.star_service.SetStar(self.cnxn, 123, 111, False)
    self.mox.VerifyAll()
    self.assertFalse(self.star_service.star_cache.HasItem(123))
    self.assertFalse(self.star_service.starrer_cache.HasItem(123))
    self.assertFalse(self.star_service.star_count_cache.HasItem(123))

  def SetUpSetStarsBatch_Add(self):
    self.mock_tbl.InsertRows(
        self.cnxn, ['item_id', 'user_id'], [(123, 111), (123, 222)],
        ignore=True, commit=True)

  def testSetStarsBatch_Add(self):
    self.SetUpSetStarsBatch_Add()
    self.mox.ReplayAll()
    self.star_service.SetStarsBatch(self.cnxn, 123, [111, 222], True)
    self.mox.VerifyAll()
    self.assertFalse(self.star_service.star_cache.HasItem(123))
    self.assertFalse(self.star_service.starrer_cache.HasItem(123))
    self.assertFalse(self.star_service.star_count_cache.HasItem(123))

  def SetUpSetStarsBatch_Remove(self):
    self.mock_tbl.Delete(self.cnxn, item_id=123, user_id=[111, 222])

  def testSetStarsBatch_Remove(self):
    self.SetUpSetStarsBatch_Remove()
    self.mox.ReplayAll()
    self.star_service.SetStarsBatch(self.cnxn, 123, [111, 222], False)
    self.mox.VerifyAll()
    self.assertFalse(self.star_service.star_cache.HasItem(123))
    self.assertFalse(self.star_service.starrer_cache.HasItem(123))
    self.assertFalse(self.star_service.star_count_cache.HasItem(123))


class IssueStarServiceTest(unittest.TestCase):

  def setUp(self):
    self.mox = mox.Mox()
    self.mock_tbl = self.mox.CreateMock(sql.SQLTableManager)
    self.mock_tbl.Delete = mock.Mock()
    self.mock_tbl.InsertRows = mock.Mock()

    self.services = service_manager.Services()
    self.services.issue = fake.IssueService()
    self.services.config = fake.ConfigService()
    self.services.features = fake.FeaturesService()

    self.cache_manager = fake.CacheManager()
    with mock.patch(
        'framework.sql.SQLTableManager', return_value=self.mock_tbl):
      self.issue_star = star_svc.IssueStarService(
          self.cache_manager)

    self.cnxn = 'fake connection'
    self.now = int(time.time())

  def testSetStarsBatch_Add(self):
    issue = fake.MakeTestIssue(
        project_id=789,
        local_id=1,
        reporter_id=111,
        owner_id=456,
        summary='sum',
        status='Live',
        issue_id=78901,
        project_name='proj',
        migration_modified_timestamp=1234567)
    self.services.issue.TestAddIssue(issue)
    config = self.services.config.GetProjectConfig(self.cnxn, 789)

    # Set up mock for getting counts.
    self.mock_tbl.Select(
        self.cnxn,
        cols=['issue_id', 'COUNT(user_id)'],
        group_by=['issue_id'],
        issue_id=[78901]).AndReturn([(78901, 2)])
    self.mox.ReplayAll()

    self.issue_star.SetStarsBatch(
        self.cnxn, self.services, config, 78901, [111, 222], True)

    self.mox.VerifyAll()
    self.mock_tbl.InsertRows.assert_called_once_with(
        self.cnxn, ['issue_id', 'user_id'], [(78901, 111), (78901, 222)],
        ignore=True,
        commit=True)

    self.assertIn(issue, self.services.issue.updated_issues)
    self.assertEqual(issue.migration_modified_timestamp, self.now)
    self.assertEqual(issue.star_count, 2)

  def testSetStarsBatch_SkipIssueUpdate_Remove(self):
    self.issue_star.SetStarsBatch_SkipIssueUpdate(
        self.cnxn, 78901, [111, 222], False)
    self.mock_tbl.Delete.assert_called_once_with(
        self.cnxn, issue_id=78901, user_id=[111, 222], commit=True)

  def testSetStarsBatch_SkipIssueUpdate_Remove_NoCommit(self):
    self.issue_star.SetStarsBatch_SkipIssueUpdate(
        self.cnxn, 78901, [111, 222], False, commit=False)
    self.mock_tbl.Delete.assert_called_once_with(
        self.cnxn, issue_id=78901, user_id=[111, 222], commit=False)

  def testSetStarsBatch_SkipIssueUpdate_Add(self):
    self.issue_star.SetStarsBatch_SkipIssueUpdate(
        self.cnxn, 78901, [111, 222], True)
    self.mock_tbl.InsertRows.assert_called_once_with(
        self.cnxn, ['issue_id', 'user_id'], [(78901, 111), (78901, 222)],
        ignore=True, commit=True)

  def testSetStarsBatch_SkipIssueUpdate_Add_NoCommit(self):
    self.issue_star.SetStarsBatch_SkipIssueUpdate(
        self.cnxn, 78901, [111, 222], True, commit=False)
    self.mock_tbl.InsertRows.assert_called_once_with(
        self.cnxn, ['issue_id', 'user_id'], [(78901, 111), (78901, 222)],
        ignore=True, commit=False)
