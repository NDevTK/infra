# Copyright 2015 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from __future__ import absolute_import
import datetime
import unittest
import os

import gae_ts_mon
import mock

from google.appengine.api.runtime import runtime
from .test_support import test_case

from infra_libs.ts_mon import config
from infra_libs.ts_mon import handlers
from infra_libs.ts_mon import shared
from infra_libs.ts_mon.common import interface
from infra_libs.ts_mon.common import targets


class HelperFunctionsTest(unittest.TestCase):
  def test_find_gaps(self):
    self.assertEqual(
        list(zip(range(5), handlers.find_gaps([1, 3, 5]))),
        list(enumerate([0, 2, 4, 6, 7])))
    self.assertEqual(
        list(zip(range(5), handlers.find_gaps([0, 1, 2, 3, 5]))),
        list(enumerate([4, 6, 7, 8, 9])))
    self.assertEqual(
        list(zip(range(3), handlers.find_gaps([]))), list(enumerate([0, 1, 2])))
    self.assertEqual(
        list(zip(range(3), handlers.find_gaps([2]))), list(
            enumerate([0, 1, 3])))


class HandlersTest(test_case.TestCase):
  def setUp(self):
    super(HandlersTest, self).setUp()

    config.reset_for_unittest()
    target = targets.TaskTarget(
        'test_service', 'test_job', 'test_region', 'test_host')
    self.mock_state = interface.State(target=target)
    mock.patch('infra_libs.ts_mon.common.interface.state',
        new=self.mock_state).start()

  def tearDown(self):
    mock.patch.stopall()
    config.reset_for_unittest()
    super(HandlersTest, self).tearDown()

  def test_assign_task_num(self):
    time_now = datetime.datetime(2016, 2, 8, 1, 0, 0)
    time_current = time_now - datetime.timedelta(
        seconds=shared.INSTANCE_EXPIRE_SEC-1)
    time_expired = time_now - datetime.timedelta(
        seconds=shared.INSTANCE_EXPIRE_SEC+1)

    with shared.instance_namespace_context():
      shared.Instance(id='expired', task_num=0, last_updated=time_expired).put()
      shared.Instance(
          id='inactive', task_num=-1, last_updated=time_expired).put()
      shared.Instance(id='new', task_num=-1, last_updated=time_current).put()
      shared.Instance(id='current', task_num=2, last_updated=time_current).put()

      handlers._assign_task_num(time_fn=lambda: time_now)

      expired = shared.Instance.get_by_id('expired')
      inactive = shared.Instance.get_by_id('inactive')
      new = shared.Instance.get_by_id('new')
      current = shared.Instance.get_by_id('current')

    self.assertIsNone(expired)
    self.assertIsNone(inactive)
    self.assertIsNotNone(new)
    self.assertIsNotNone(current)
    self.assertEqual(2, current.task_num)
    self.assertEqual(1, new.task_num)


class ReportMemoryTest(test_case.TestCase):

  def test_report_skipped_if_development(self):
    env = os.environ.copy()
    env['SERVER_SOFTWARE'] = 'Development'
    with mock.patch.object(os, 'environ', env):

      @handlers.report_memory
      def handler(request):
        return request

      with mock.patch.object(runtime, 'memory_usage') as m:
        handler('input params')
        m.assert_not_called()

  def test_report_with_production(self):
    env = os.environ.copy()
    env['SERVER_SOFTWARE'] = 'Production'
    with mock.patch.object(os, 'environ', env):
      # a test handler should be added in the scope of the mock for os.environ.
      @handlers.report_memory
      def handler(request):
        return request

      with mock.patch.object(runtime, 'memory_usage') as m:
        handler('input params')
        m.assert_called()
