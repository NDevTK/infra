# Copyright 2023 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import unittest

from google.appengine.ext import ndb
from google.appengine.ext import testbed
from redirect import redirect_custom_value


class TestRedirectCustomValue(unittest.TestCase):

  def setUp(self):
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_datastore_v3_stub()
    self.testbed.init_memcache_stub()
    ndb.get_context().clear_cache()

  def tearDown(self):
    self.testbed.deactivate()

  def testGetRedirectCustomValue(self):
    redirectCustomValue = redirect_custom_value.RedirectCustomValue
    redirectCustomValue(
        ProjectName='a',
        MonorailType='test',
        MonorailValue='a',
        RedirectType='t',
        RedirectValue='v').put()

    (t, v) = redirectCustomValue.Get('a', 'test', 'a')
    self.assertEqual(t, 't')
    self.assertEqual(v, 'v')

  def testGetRedirectCustomValueWithoutValue(self):
    redirectCustomValue = redirect_custom_value.RedirectCustomValue

    (t, v) = redirectCustomValue.Get('a', 'test', 'a')
    self.assertEqual(t, None)
    self.assertEqual(v, None)

  def testGetRedirectCustomValueOnlyReturnTheFirstMatch(self):
    # There should be only one match in db.
    # This may change if we decided to support mutiple value mapping.
    redirectCustomValue = redirect_custom_value.RedirectCustomValue
    redirectCustomValue(
        ProjectName='a',
        MonorailType='test',
        MonorailValue='a',
        RedirectType='t1',
        RedirectValue='v1').put()
    redirectCustomValue(
        ProjectName='a',
        MonorailType='test',
        MonorailValue='a',
        RedirectType='t2',
        RedirectValue='v2').put()
    (t, v) = redirectCustomValue.Get('a', 'test', 'a')
    self.assertEqual(t, 't1')
    self.assertEqual(v, 'v1')
