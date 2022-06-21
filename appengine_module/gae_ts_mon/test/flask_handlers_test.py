# Copyright 2022 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

from __future__ import absolute_import
import datetime
import json
import unittest
import os

import gae_ts_mon
import mock
import flask

from .test_support import test_case

from infra_libs.ts_mon import config
from infra_libs.ts_mon import flask_handlers
from infra_libs.ts_mon.common import interface
from infra_libs.ts_mon.common import metrics
from infra_libs.ts_mon.common import targets


class FlaskTSMonJSHandlerTest(test_case.TestCase):

  def setUp(self):
    super(FlaskTSMonJSHandlerTest, self).setUp()

    config.reset_for_unittest()
    target = targets.TaskTarget('test_service', 'test_job', 'test_region',
                                'test_host')
    self.mock_state = interface.State(target=target)
    mock.patch(
        'infra_libs.ts_mon.common.interface.state',
        new=self.mock_state).start()

    self.ts_mon_handler = flask_handlers.TSMonJSFlaskHandler(flask=flask)
    self.ts_mon_handler.register_metrics([
        metrics.BooleanMetric(
            'frontend/boolean_test',
            'Boolean metric test',
            field_spec=[metrics.StringField('client_id')]),
    ])
    self.ts_mon_handler.xsrf_is_valid = mock.Mock(return_value=True)
    self.mock_timestamp = 1537821859

    def time_fn():
      return self.mock_timestamp

    self.ts_mon_handler.time_fn = time_fn

    self.app = flask.Flask('test_app')
    self.app.config['TESTING'] = True
    self.app.add_url_rule(
        '/_/ts_mon_js.do', view_func=self.ts_mon_handler.post, methods=['POST'])

  def tearDown(self):
    mock.patch.stopall()
    config.reset_for_unittest()
    super(FlaskTSMonJSHandlerTest, self).tearDown()

  def test_time_fn(self):
    time_fn_res = flask_handlers.TSMonJSFlaskHandler(flask=flask).time_fn()
    self.assertTrue(isinstance(time_fn_res, float))

  def test_post_metrics_no_metrics(self):
    self.ts_mon_handler._metrics = None
    res = self.app.test_client().post('/_/ts_mon_js.do')
    self.assertEqual(res.status_code, 400)

  def test_post_metrics_invalid_json(self):
    """Test case when JSON request body is invalid."""
    res = self.app.test_client().post('/_/ts_mon_js.do', data='rutabaga')
    self.assertEqual(res.status_code, 400)

  def test_post_metrics_invalid_xsrf(self):
    """Test case when XSRF token is invalid."""
    self.ts_mon_handler.xsrf_is_valid = mock.Mock(return_value=False)

    res = self.app.test_client().post(
        '/_/ts_mon_js.do', data=json.dumps({'metrics': []}))
    self.assertEqual(res.status_code, 403)

    self.ts_mon_handler.xsrf_is_valid.assert_called_once()

  def test_post_metrics_must_be_dict(self):
    """Test case when body is not a dict."""
    res = self.app.test_client().post('/_/ts_mon_js.do', data=json.dumps([]))
    self.assertEqual(res.status_code, 400)

  def test_post_no_metrics_key(self):
    """Test case when body is not a dict."""
    res = self.app.test_client().post(
        '/_/ts_mon_js.do', data=json.dumps({'test': []}))
    self.assertEqual(res.status_code, 400)

  def test_post_metrics_unregistered_metric_name(self):
    """Test case when a metric name isn't registered."""
    body = json.dumps({
        'metrics': [{
            'MetricInfo': {
                'Name': 'frontend/not_defined',
                'ValueType': 2,
            },
            'Cells': [{
                'value': 'rutabaga',
                'fields': {
                    'client_id': '789',
                    'rutabaga_id': '789',
                },
            }],
        },],
    })

    res = self.app.test_client().post('/_/ts_mon_js.do', data=body)
    self.assertEqual(res.status_code, 400)

  def test_post_metrics_malcious_metric_name(self):
    """Metric name is echoed back in a safely escaped form."""
    body = json.dumps({
        'metrics': [{
            'MetricInfo': {
                'Name': 'frontend/not_defined<script>alert(1)</script>',
                'ValueType': 2,
            },
            'Cells': [{
                'value': 'rutabaga',
                'fields': {
                    'client_id': '789',
                    'rutabaga_id': '789',
                },
            }],
        },],
    })
    res = self.app.test_client().post('/_/ts_mon_js.do', data=body)
    self.assertEqual(res.status_code, 400)

  def test_post_metrics_invalid_fields(self):
    """Test case when metric name is fine but fields are not."""
    body = json.dumps({
        'metrics': [{
            'MetricInfo': {
                'Name': 'frontend/boolean_test',
                'ValueType': 2,
            },
            'Cells': [{
                'value': True,
                'fields': {
                    'client_id': '789',
                    'rutabaga_id': '789',
                },
            }],
        },],
    })
    res = self.app.test_client().post('/_/ts_mon_js.do', data=body)
    self.assertEqual(res.status_code, 400)

  def test_post_metrics_maliciou_fields(self):
    """Invalid fields are echoed in a safely escaped format."""
    body = json.dumps({
        'metrics': [{
            'MetricInfo': {
                'Name': 'frontend/boolean_test',
                'ValueType': 2,
            },
            'Cells': [{
                'value': True,
                'fields': {
                    'client_id<script>alert(1)</script>': '789',
                    'rutabaga_id': '789',
                },
            }],
        },],
    })
    res = self.app.test_client().post('/_/ts_mon_js.do', data=body)
    self.assertEqual(res.status_code, 400)

  def test_post_rejects_cumulative_without_start_time(self):
    """Test case where start_time is not supplied for CumulativeDistribution."""
    body = json.dumps({
        'metrics': [{
            'MetricInfo': {
                'Name': 'frontend/cumulative_test',
                'ValueType': 2,
            },
            'Cells': [{
                'value': 'rutabaga',
                'fields': {
                    'client_id': '789',
                },
            }],
        },],
    })
    self.ts_mon_handler.register_metrics([
        metrics.CumulativeDistributionMetric(
            'frontend/cumulative_test',
            'Cumulative metric test',
            field_spec=[metrics.StringField('client_id')]),
    ])

    res = self.app.test_client().post('/_/ts_mon_js.do', data=body)
    self.assertEqual(res.status_code, 400)

  def test_post_rejects_start_time_in_future(self):
    """Test rejects when start_time is in the future."""
    body = json.dumps({
        'metrics': [{
            'MetricInfo': {
                'Name': 'frontend/cumulative_test',
                'ValueType': 2,
            },
            'Cells': [{
                'value': 'rutabaga',
                'fields': {
                    'client_id': '789',
                },
                'start_time': self.mock_timestamp + 1,
            }],
        },],
    })
    self.ts_mon_handler.register_metrics([
        metrics.CumulativeDistributionMetric(
            'frontend/cumulative_test',
            'Cumulative metric test',
            field_spec=[metrics.StringField('client_id')]),
    ])
    res = self.app.test_client().post('/_/ts_mon_js.do', data=body)
    self.assertEqual(res.status_code, 400)

  def test_post_rejects_start_time_in_past(self):
    """Test rejects when start_time is >1 month in the past."""
    one_month_seconds = 60 * 60 * 24 * 30
    body = json.dumps({
        'metrics': [{
            'MetricInfo': {
                'Name': 'frontend/cumulative_test',
                'ValueType': 2,
            },
            'Cells': [{
                'value': 'rutabaga',
                'fields': {
                    'client_id': '789',
                },
                'start_time': self.mock_timestamp - one_month_seconds * 2,
            }],
        },],
    })
    self.ts_mon_handler.register_metrics([
        metrics.CumulativeDistributionMetric(
            'frontend/cumulative_test',
            'Cumulative metric test',
            field_spec=[metrics.StringField('client_id')]),
    ])
    res = self.app.test_client().post('/_/ts_mon_js.do', data=body)
    self.assertEqual(res.status_code, 400)

  def test_post_distribution_metrics_not_a_dict(self):
    """Test case when a distribution metric value is not a dict."""
    body = json.dumps({
        'metrics': [{
            'MetricInfo': {
                'Name': 'frontend/cumulative_test',
                'ValueType': 2,
            },
            'Cells': [{
                'value': 'rutabaga',
                'fields': {
                    'client_id': '789',
                },
                'start_time': self.mock_timestamp - 60,
            }],
        },],
    })
    self.ts_mon_handler.register_metrics([
        metrics.CumulativeDistributionMetric(
            'frontend/cumulative_test',
            'Cumulative metric test',
            field_spec=[metrics.StringField('client_id')]),
    ])
    res = self.app.test_client().post('/_/ts_mon_js.do', data=body)
    self.assertEqual(res.status_code, 400)

  def test_post_metrics_normal(self):
    """Test successful POST case."""
    body = json.dumps({
        'metrics': [
            {
                'MetricInfo': {
                    'Name': 'frontend/boolean_test',
                    'ValueType': 2,
                },
                'Cells': [{
                    'value': True,
                    'fields': {
                        'client_id': '789',
                    },
                }],
            },
            {
                'MetricInfo': {
                    'Name': 'frontend/cumulative_test',
                    'ValueType': 2,
                },
                'Cells': [{
                    'value': {
                        'sum': 1234,
                        'count': 4321,
                        'buckets': {
                            0: 123,
                            1: 321,
                            2: 213,
                        },
                    },
                    'fields': {
                        'client_id': '789',
                    },
                    'start_time': self.mock_timestamp - 60,
                }],
            },
        ],
    })

    self.ts_mon_handler.register_metrics([
        metrics.CumulativeDistributionMetric(
            'frontend/cumulative_test',
            'Cumulative metric test',
            field_spec=[metrics.StringField('client_id')]),
    ])

    res = self.app.test_client().post('/_/ts_mon_js.do', data=body)
    self.assertEqual(res.status_code, 201)
