# Copyright 2016 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Unit tests for jsonfeed module."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

from six.moves import http_client
import unittest

import flask
from google.appengine.api import app_identity

from framework import jsonfeed
from framework import xsrf
from services import service_manager
from testing import testing_helpers


class JsonFeedTest(unittest.TestCase):

  def setUp(self):
    self.cnxn = 'fake cnxn'

  def testGet(self):
    """Tests handling of GET requests."""
    feed = _TestableJsonFeed()

    # all expected args are present + a bonus arg that should be ignored
    feed.mr = testing_helpers.MakeMonorailRequest(
        path='/foo/bar/wee?sna=foo', method='POST',
        params={'a': '123', 'z': 'zebra'})
    feed.get()

    self.assertEqual(True, feed.handle_request_called)
    self.assertEqual(1, len(feed.json_data))

  def testPost(self):
    """Tests handling of POST requests."""
    feed = _TestableJsonFeed()
    feed.mr = testing_helpers.MakeMonorailRequest(
        path='/foo/bar/wee?sna=foo', method='POST',
        params={'a': '123', 'z': 'zebra'})

    feed.post()

    self.assertEqual(True, feed.handle_request_called)
    self.assertEqual(1, len(feed.json_data))

  def testSecurityTokenChecked_BadToken(self):
    feed = _TestableJsonFeed()
    feed.mr = testing_helpers.MakeMonorailRequest(
        user_info={'user_id': 555})
    # Note that feed.mr has no token set.
    self.assertRaises(xsrf.TokenIncorrect, feed.get)
    self.assertRaises(xsrf.TokenIncorrect, feed.post)

    feed.mr.token = 'bad token'
    self.assertRaises(xsrf.TokenIncorrect, feed.get)
    self.assertRaises(xsrf.TokenIncorrect, feed.post)

  def testSecurityTokenChecked_HandlerDoesNotNeedToken(self):
    feed = _TestableJsonFeed()
    feed.mr = testing_helpers.MakeMonorailRequest(
        user_info={'user_id': 555})
    # Note that feed.mr has no token set.
    feed.CHECK_SECURITY_TOKEN = False
    feed.get()
    feed.post()

  def testSecurityTokenChecked_AnonUserDoesNotNeedToken(self):
    feed = _TestableJsonFeed()
    feed.mr = testing_helpers.MakeMonorailRequest()
    # Note that feed.mr has no token set, but also no auth.user_id.
    feed.get()
    feed.post()

  def testSameAppOnly_ExternallyAccessible(self):
    feed = _TestableJsonFeed()
    feed.mr = testing_helpers.MakeMonorailRequest()
    # Note that request has no X-Appengine-Inbound-Appid set.
    feed.get()
    feed.post()

  def testSameAppOnly_InternalOnlyCalledFromSameApp(self):
    feed = _TestableJsonFeed()
    feed.CHECK_SAME_APP = True
    feed.mr = testing_helpers.MakeMonorailRequest()
    app_id = app_identity.get_application_id()
    feed.mr.request.headers['X-Appengine-Inbound-Appid'] = app_id
    feed.get()
    feed.post()

  def testSameAppOnly_InternalOnlyCalledExternally(self):
    feed = _TestableJsonFeed()
    feed.CHECK_SAME_APP = True
    feed.mr = testing_helpers.MakeMonorailRequest()
    # Note that request has no X-Appengine-Inbound-Appid set.
    feed.response = flask.Response()
    self.assertIsNone(feed.get())
    self.assertFalse(feed.handle_request_called)
    self.assertEqual(http_client.FORBIDDEN, feed.response.status_code)
    self.assertIsNone(feed.post())
    self.assertFalse(feed.handle_request_called)
    self.assertEqual(http_client.FORBIDDEN, feed.response.status_code)

  def testSameAppOnly_InternalOnlyCalledFromWrongApp(self):
    feed = _TestableJsonFeed()
    feed.CHECK_SAME_APP = True
    feed.mr = testing_helpers.MakeMonorailRequest()
    feed.mr.request.headers['X-Appengine-Inbound-Appid'] = 'wrong'
    feed.response = flask.Response()
    self.assertIsNone(feed.get())
    self.assertFalse(feed.handle_request_called)
    self.assertEqual(http_client.FORBIDDEN, feed.response.status_code)
    self.assertIsNone(feed.post())
    self.assertFalse(feed.handle_request_called)
    self.assertEqual(http_client.FORBIDDEN, feed.response.status_code)


class _TestableJsonFeed(jsonfeed.JsonFeed):

  def __init__(self):
    super(_TestableJsonFeed, self).__init__(services=service_manager.Services())

    self.response_data = None
    self.handle_request_called = False
    self.json_data = None

  def HandleRequest(self, mr):
    self.handle_request_called = True
    return {'a': mr.GetParam('a')}

  # The output chain is hard to double so we pass on that phase,
  # but save the response data for inspection
  def _RenderJsonResponse(self, json_data):
    self.json_data = json_data
