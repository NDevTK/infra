# Copyright 2017 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file or at
# https://developers.google.com/open-source/licenses/bsd

"""Tests for the warmup servlet."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import unittest

import webapp2

from framework import warmup


class WarmupTest(unittest.TestCase):

  def testHandleWarmup(self):
    app = webapp2.WSGIApplication([('/', warmup.Warmup)])

    request = webapp2.Request.blank('/')
    response = request.get_response(app)

    self.assertEqual(response.status_int, 200)
    self.assertEqual(response.body, '')

  def testHandleStart(self):
    app = webapp2.WSGIApplication([('/', warmup.Start)])

    request = webapp2.Request.blank('/')
    response = request.get_response(app)

    self.assertEqual(response.status_int, 200)
    self.assertEqual(response.body, '')

  def testHandleStop(self):
    app = webapp2.WSGIApplication([('/', warmup.Stop)])

    request = webapp2.Request.blank('/')
    response = request.get_response(app)

    self.assertEqual(response.status_int, 200)
    self.assertEqual(response.body, '')
