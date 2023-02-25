# Copyright 2017 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Tests for the warmup servlet."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import unittest

import flask

from framework import warmup


class WarmupTest(unittest.TestCase):

  def testHandleWarmup(self):
    app = flask.Flask(__name__)
    app.add_url_rule('/', view_func=warmup.Warmup)

    with app.test_client() as client:
      response = client.get('/')

    self.assertEqual(response.status_code, 200)
    self.assertEqual(response.data, b'')

  def testHandleStart(self):
    app = flask.Flask(__name__)
    app.add_url_rule('/', view_func=warmup.Start)

    with app.test_client() as client:
      response = client.get('/')

    self.assertEqual(response.status_code, 200)
    self.assertEqual(response.data, b'')

  def testHandleStop(self):
    app = flask.Flask(__name__)
    app.add_url_rule('/', view_func=warmup.Stop)

    with app.test_client() as client:
      response = client.get('/')

    self.assertEqual(response.status_code, 200)
    self.assertEqual(response.data, b'')
