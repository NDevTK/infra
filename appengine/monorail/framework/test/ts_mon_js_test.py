# Copyright 2018 The Chromium Authors
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Tests for MonorailTSMonJSHandler."""
from __future__ import print_function
from __future__ import division
from __future__ import absolute_import

import json
import unittest
from mock import patch

import flask
from google.appengine.ext import testbed

from framework.ts_mon_js import MonorailTSMonJSHandler
from services import service_manager


class MonorailTSMonJSHandlerTest(unittest.TestCase):

  def setUp(self):
    self.testbed = testbed.Testbed()
    self.testbed.activate()
    self.testbed.init_user_stub()
    self.services = service_manager.Services()
    self.app = flask.Flask('test_app')
    self.app.config['TESTING'] = True
    self.app.add_url_rule(
        '/_/ts_mon_js.do',
        view_func=MonorailTSMonJSHandler(
            services=self.services).PostMonorailTSMonJSHandler,
        methods=['POST'])

  def tearDown(self):
    self.testbed.deactivate()

  @patch('framework.xsrf.ValidateToken')
  @patch('time.time')
  def testSubmitMetrics(self, _mockTime, _mockValidateToken):
    """Test normal case POSTing metrics."""
    _mockTime.return_value = 1537821859
    res = self.app.test_client().post(
        '/_/ts_mon_js.do',
        data = json.dumps(
            {
                'metrics':
                    [
                        {
                            'MetricInfo':
                                {
                                    'Name':
                                      'monorail/frontend/issue_update_latency',
                                    'ValueType':
                                      2,
                                },
                            'Cells':
                                [
                                    {
                                        'value':
                                            {
                                                'sum': 1234,
                                                'count': 4321,
                                                'buckets':
                                                    {
                                                        0: 123,
                                                        1: 321,
                                                        2: 213,
                                                    },
                                            },
                                        'fields':
                                            {
                                                'client_id': '789',
                                                'host_name': 'rutabaga',
                                                'document_visible': True,
                                            },
                                        'start_time': 1537821799,
                                    }
                                ],
                        }
                    ],
            }))
    self.assertEqual(res.status_code, 201)
